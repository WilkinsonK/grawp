package main

import (
	"context"
	"io"

	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
)

type ServiceBuildOpts struct {
	OutDestination io.Writer
	Manifest       *ServiceManifest
	ServiceName    string
	TagName        string
}

// Attempt to build an image from an `ImageManifest`.
func BuildImageFromManifest(cli *client.Client, opts ServiceBuildOpts) error {
	opt, err := opts.Manifest.GetImageBuildOptions()
	if err != nil {
		return err
	}
	ctx, err := opts.Manifest.GetImageBuildContext()
	if err != nil {
		return err
	}
	defer ctx.Close()

	resp, err := cli.ImageBuild(context.Background(), ctx, opt)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	io.Copy(opts.OutDestination, resp.Body)

	cli.ImagesPrune(context.Background(), filters.Args{})
	return nil
}

// Attempt to create a service container from an
// `ImageManifest`.
//
// Returns the `name` of the container and the container
// `ID`.
func BuildServiceFromManifest(cli *client.Client, opts ServiceBuildOpts) (string, string, error) {
	if opts.ServiceName == "" {
		opts.ServiceName = opts.Manifest.GetServiceName()
	}

	config, err := opts.Manifest.GetServiceBuildConfig(opts.TagName)
	if err != nil {
		return "", "", err
	}
	hostc, err := opts.Manifest.GetServiceHostConfig()
	if err != nil {
		return "", "", err
	}

	res, err := cli.ContainerCreate(
		context.Background(),
		&config,
		&hostc,
		nil, nil, opts.ServiceName)
	return opts.ServiceName, res.ID, err
}
