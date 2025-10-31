package main

import (
	"context"
	"fmt"
	"os"
	"slices"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/spf13/cobra"
)

type DoWhileInspectArgs struct {
	apiClient   *client.Client
	callback    func(*DoWhileInspectArgs) error
	ctr         *container.Summary
	response    *container.InspectResponse
	retry_delay time.Duration
	retry_max   int
}

func DoWhileInspect(args DoWhileInspectArgs) {
	var ctx context.Context
	retries := args.retry_max
	for {
		ctx = context.Background()
		resp, err := args.apiClient.ContainerInspect(ctx, args.ctr.ID)

		// Handle bad inspection calls. Could potentially
		// fail. Try up to N times after waiting for D
		// amount of time.
		if err != nil && retries != 0 {
			fmt.Fprintf(os.Stderr, "error: %s\n", err)
			time.Sleep(args.retry_delay)
			retries--
			continue
		} else if err != nil {
			fmt.Fprintf(os.Stderr, "error: maximum retries met\n")
			panic(err)
		}
		retries = args.retry_max
		args.response = &resp

		err = args.callback(&args)
		if err != nil && err.Error() == "done" {
			break
		} else if err != nil {
			panic(err)
		}
	}
}

func DoWhileWaitingForContainer(args *DoWhileInspectArgs) error {
	if args.response.State.Running {
		return fmt.Errorf("done")
	}
	time.Sleep(500 * time.Millisecond)
	return nil
}

func DoWhileWatchingContainer(args *DoWhileInspectArgs) error {
	state := args.response.State
	if state.Status == "exited" && state.Error == "" {
		fmt.Fprintf(os.Stderr, "Server stopped\n")
		return fmt.Errorf("done")
	}

	if state.Status == "exited" && state.Error != "" {
		fmt.Fprintf(os.Stderr, "Server container down; %s; restarting...\n", state.Error)
		ctx := context.Background()
		args.apiClient.ContainerRestart(ctx, args.ctr.ID, container.StopOptions{})
	}
	return nil
}

func FindServerByName(client *client.Client, name string) (*container.Summary, error) {
	containers, err := client.ContainerList(context.Background(), container.ListOptions{All: true})
	if err != nil {
		return nil, err
	}

	for _, ctr := range containers {
		if slices.Contains(ctr.Names, name) {
			return &ctr, nil
		}
	}

	return nil, fmt.Errorf("No such container '%s'", name)
}

func WatchServer(cmd *cobra.Command, args []string) {
	apiClient, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		panic(err)
	}
	defer apiClient.Close()

	ctr, err := FindServerByName(apiClient, args[0])
	if err != nil {
		panic(err)
	}

	inspectArgs := DoWhileInspectArgs{
		apiClient:   apiClient,
		ctr:         ctr,
		retry_delay: 10 * time.Second,
		retry_max:   3,
	}

	fmt.Fprintf(os.Stderr, "Waiting for container... ")
	inspectArgs.callback = DoWhileWaitingForContainer
	DoWhileInspect(inspectArgs)
	fmt.Fprintf(os.Stderr, "Container available\n")
	inspectArgs.callback = DoWhileWatchingContainer
	DoWhileInspect(inspectArgs)
}

func main() {
	// Get the container we are looking for.
	// Stream the status of the container we want to watch.
	// Detect if the container is failing
	// - If container is failing: restart container.
	// - If container is not failing: do nothing.
	// Peform cleanups on SIGKILL.

	root := &cobra.Command{
		Use:     "watch_server <container-name>",
		Short:   "Watch the server container",
		Long:    "Watches the server container, restarting it if failure is dectected.",
		Args:    cobra.ExactArgs(1),
		Version: "1.0.0",
		Run:     WatchServer,
	}

	err := root.Execute()
	if err != nil {
		panic(err)
	}
}
