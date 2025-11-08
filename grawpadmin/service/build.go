package service

import (
	"context"
	"io"
	"strings"

	"github.com/WilkinsonK/grawp/grawpadmin/manifest"
	"github.com/WilkinsonK/grawp/grawpadmin/service/models"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/image"
)

// Attempt to build an image from an `ImageManifest`.
func BuildImageFromManifest(broker *ServiceBroker, sm manifest.ServiceManifest) ([]models.ServiceImage, error) {
	var sModels []models.ServiceImage

	opt, err := sm.GetImageBuildOptions()
	if err != nil {
		return sModels, err
	}
	ctx, err := sm.GetImageBuildContext()
	if err != nil {
		return sModels, err
	}
	defer ctx.Close()

	settings := sm.GetImageBuildSettings()
	resp, err := broker.Client.ImageBuild(context.Background(), ctx, opt)
	if err != nil {
		return sModels, err
	}
	defer resp.Body.Close()
	io.Copy(settings.OutDestination, resp.Body)

	_, err = broker.Client.ImagesPrune(context.Background(), filters.Args{})
	if err != nil {
		return sModels, err
	}

	for _, tag := range opt.Tags {
		tParts := strings.SplitN(tag, ":", 2)
		tName := tParts[0]
		tTag := "latest"
		if len(tParts) > 1 {
			tTag = tParts[1]
		}

		resp, err := broker.Client.ImageList(context.Background(), image.ListOptions{
			Filters: filters.NewArgs(filters.Arg("reference", tag)),
		})
		if err != nil {
			return sModels, err
		}

		model_opts := models.ServiceImageNewOptions{
			Name:     tName,
			Tag:      tTag,
			DockerId: resp[0].ID,
		}
		model, err := models.ServiceImageNew(model_opts)
		if err != nil {
			return sModels, err
		}
		sModels = append(sModels, model)
	}

	_, err = models.ServiceImagePut(broker.Database, sModels...)
	return sModels, err
}

// Attempt to create a service container from an
// `ImageManifest`.
//
// Returns the `name` of the container and the container
// `ID`.
func BuildServiceFromManifest(broker *ServiceBroker, sm manifest.ServiceManifest) (models.ServiceContainer, error) {
	var model models.ServiceContainer

	settings := sm.GetImageBuildSettings()
	if settings.ServiceName == "" {
		settings.ServiceName = sm.GetServiceName()
	}

	config, err := sm.GetServiceBuildConfig(settings.TagName)
	if err != nil {
		return model, err
	}
	hostc, err := sm.GetServiceHostConfig()
	if err != nil {
		return model, err
	}

	ctx := context.Background()
	res, err := broker.Client.ContainerCreate(
		ctx,
		&config,
		&hostc,
		nil, nil, settings.ServiceName)

	model_opts := models.ServiceContainerNewOpts{
		Name:     settings.ServiceName,
		DockerId: res.ID,
	}
	model, err = models.ServiceContainerNew(model_opts)
	if err != nil {
		return model, err
	}
	_, err = models.ServiceContainerPut(broker.Database, model)
	return model, err
}
