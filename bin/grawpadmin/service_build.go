package main

import (
	"context"
	"io"
	"strings"

	"github.com/WilkinsonK/grawp/bin/grawpadmin/service_manifest"
	"github.com/WilkinsonK/grawp/bin/grawpadmin/service_models"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
)

type ServiceBuildOpts struct {
	OutDestination io.Writer
	Manifest       *service_manifest.ServiceManifest
	ServiceName    string
	TagName        string
}

// Attempt to build an image from an `ImageManifest`.
func BuildImageFromManifest(cli *client.Client, opts ServiceBuildOpts) ([]service_models.ServiceImage, error) {
	var models []service_models.ServiceImage
	opt, err := opts.Manifest.GetImageBuildOptions()
	if err != nil {
		return models, err
	}
	ctx, err := opts.Manifest.GetImageBuildContext()
	if err != nil {
		return models, err
	}
	defer ctx.Close()

	resp, err := cli.ImageBuild(context.Background(), ctx, opt)
	if err != nil {
		return models, err
	}
	defer resp.Body.Close()
	io.Copy(opts.OutDestination, resp.Body)

	_, err = cli.ImagesPrune(context.Background(), filters.Args{})
	if err != nil {
		return models, err
	}

	for _, tag := range opt.Tags {
		tParts := strings.SplitN(tag, ":", 2)
		tName := tParts[0]
		tTag := "latest"
		if len(tParts) > 1 {
			tTag = tParts[1]
		}

		resp, err := cli.ImageList(context.Background(), image.ListOptions{
			Filters: filters.NewArgs(filters.Arg("reference", tag)),
		})
		if err != nil {
			return models, err
		}

		model_opts := service_models.ServiceImageNewOptions{
			Name:     tName,
			Tag:      tTag,
			DockerId: resp[0].ID,
		}
		model, err := service_models.ServiceImageNew(model_opts)
		if err != nil {
			return models, err
		}
		models = append(models, model)
	}

	return models, nil
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

	ctx := context.Background()
	res, err := cli.ContainerCreate(
		ctx,
		&config,
		&hostc,
		nil, nil, opts.ServiceName)
	return opts.ServiceName, res.ID, err
}
