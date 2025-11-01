package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/WilkinsonK/grawp/grawpadmin/service/models"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

type WatchArgs struct {
	Client     *client.Client
	Error      error
	Model      models.ServiceContainer
	Response   *container.InspectResponse
	RetryCount uint
	RetryDelay time.Duration
	RetryMax   uint
	WatchDelay time.Duration
}

type WatchCallback func(*WatchArgs) error

var DoneError = fmt.Errorf("done")

func SetDoneError(args *WatchArgs) {
	args.Error = DoneError
}

func IsDoneError(args *WatchArgs) bool {
	return args.Error != nil && args.Error == DoneError
}

func IsStoppedOk(args *WatchArgs) bool {
	state := args.Response.State
	return state.Status == "exited" && state.Error == ""
}

func IsStoppedErr(args *WatchArgs) bool {
	state := args.Response.State
	return state.Status == "exited" && state.Error != ""
}

func ShouldRestart(args *WatchArgs) bool {
	return IsStoppedErr(args)
}

func ShouldStop(args *WatchArgs) bool {
	return IsDoneError(args)
}

func WatchRestart(args *WatchArgs) error {
	log.Println("Restarting service...")
	return args.Client.ContainerRestart(context.Background(), args.Model.DockerId, container.StopOptions{})
}

func WatchRetryFailure(args *WatchArgs) {
	if args.Error != nil && args.RetryCount != 0 && !IsDoneError(args) {
		log.Printf("Error: %s\n", args.Error)
		WatchRetryWait(args)
		args.Error = nil
		args.RetryCount--
		return
	}
	if args.Error != nil && !IsDoneError(args) {
		log.Printf("Error: maximum retries met\n")
		return
	}
	args.RetryCount = args.RetryMax
}

func WatchRetryWait(args *WatchArgs) {
	time.Sleep(args.RetryDelay)
}

func WatchStart(args *WatchArgs) error {
	log.Println("Service starting...")
	return args.Client.ContainerStart(context.Background(), args.Model.DockerId, container.StartOptions{})
}

func WatchStop(args *WatchArgs) error {
	log.Println("Service stopping...")
	return args.Client.ContainerStop(context.Background(), args.Model.DockerId, container.StopOptions{})
}

func WatchWait(args *WatchArgs, err *error) {
	time.Sleep(args.WatchDelay)
}

func Watch(args *WatchArgs, callback WatchCallback) error {
	ctx := context.Background()
	for {
		resp, err := args.Client.ContainerInspect(ctx, args.Model.DockerId)
		if WatchRetryFailure(args); err != nil {
			return err
		}
		args.Response = &resp

		err = callback(args)
		if ShouldStop(args) || err != nil {
			return err
		}
		WatchWait(args, &err)
	}
}

func WatchImageService(cli *client.Client, db *sql.DB, name string) error {
	models, err := models.ServiceContainerFind(db, models.ServiceContainerFindOpts{
		Name: name,
	})
	if err != nil {
		return err
	}

	args := WatchArgs{
		Client:     cli,
		Error:      nil,
		Model:      models[0],
		RetryCount: 3,
		RetryMax:   3,
		RetryDelay: 10 * time.Second,
		WatchDelay: 500 * time.Millisecond,
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)
	done := make(chan bool, 1)

	go func() {
		_ = <-sigChan
		SetDoneError(&args)
		done <- true
	}()

	log.Println("Waiting for service...")
	Watch(&args, func(wa *WatchArgs) error {
		if wa.Response.State.Running {
			return DoneError
		} else {
			return WatchStart(wa)
		}
	})
	log.Println("Service running")

	Watch(&args, func(wa *WatchArgs) error {
		if IsStoppedErr(wa) || IsStoppedOk(wa) {
			if err := WatchRestart(wa); err != nil {
				return err
			}
		}
		if ShouldStop(wa) {
			WatchStop(wa)
		}
		return nil
	})

	return nil
}
