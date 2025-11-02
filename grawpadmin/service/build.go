package service

import (
	"context"
	"database/sql"
	"io"
	"strings"

	"github.com/WilkinsonK/grawp/grawpadmin/manifest"
	"github.com/WilkinsonK/grawp/grawpadmin/service/models"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
)

// Attempt to build an image from an `ImageManifest`.
func BuildImageFromManifest(cli *client.Client, sm manifest.ServiceManifest) ([]models.ServiceImage, error) {
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
	resp, err := cli.ImageBuild(context.Background(), ctx, opt)
	if err != nil {
		return sModels, err
	}
	defer resp.Body.Close()
	io.Copy(settings.OutDestination, resp.Body)

	_, err = cli.ImagesPrune(context.Background(), filters.Args{})
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

		resp, err := cli.ImageList(context.Background(), image.ListOptions{
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

	err = models.WithDatabase(settings.DataPath, func(db *sql.DB) error {
		_, err = models.ServiceImagePut(db, sModels...)
		return err
	})
	return sModels, err
}

// Attempt to create a service container from an
// `ImageManifest`.
//
// Returns the `name` of the container and the container
// `ID`.
func BuildServiceFromManifest(cli *client.Client, sm manifest.ServiceManifest) (models.ServiceContainer, error) {
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
	res, err := cli.ContainerCreate(
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
	err = models.WithDatabase(settings.DataPath, func(db *sql.DB) error {
		_, err = models.ServiceContainerPut(db, model)
		return err
	})
	return model, err
}
