package service

import (
	"context"
	"database/sql"
	"fmt"
	"io"

	"github.com/WilkinsonK/grawp/grawpadmin/manifest"
	"github.com/WilkinsonK/grawp/grawpadmin/service/models"
	"github.com/docker/docker/client"
)

type ServiceManifestCallback func(manifest.ServiceManifest) error

type ServiceBroker struct {
	Client   *client.Client
	Database *sql.DB
}

func (Sb *ServiceBroker) BuildImage(sm manifest.ServiceManifest) error {
	return attempt(sm, Sb.RenderManifestFiles, Sb.BuildImageFromManifest)
}

func (Sb *ServiceBroker) BuildImageFromManifest(sm manifest.ServiceManifest) error {
	_, err := BuildImageFromManifest(Sb.Client, sm)
	return err
}

func (Sb *ServiceBroker) BuildImageServiceFromManifest(sm manifest.ServiceManifest, out io.Writer) error {
	model, err := BuildServiceFromManifest(Sb.Client, sm)
	if err != nil {
		return err
	}

	fmt.Fprint(out, model.Name)
	if model.DockerId != "" {
		fmt.Fprintf(out, " - %s", model.DockerId)
	}
	fmt.Fprintln(out)
	return err
}

func (Sb *ServiceBroker) Close() error {
	Sb.Client.Close()
	Sb.Database.Close()
	return nil
}

func (Sb *ServiceBroker) GetServiceContainerStatus(model models.ServiceContainer) string {
	resp, err := Sb.Client.ContainerInspect(context.Background(), model.DockerId)
	if err != nil {
		return "unknown"
	} else {
		return resp.State.Status
	}
}

func (Sb *ServiceBroker) InitDatabase() error {
	return models.InitDatabaseTables(Sb.Database)
}

func (Sb *ServiceBroker) ListImages(out io.Writer) error {
	models, err := models.ServiceImagesList(Sb.Database)
	if err != nil {
		return err
	}
	for _, model := range models {
		fmt.Printf("%s\t%s   \t%s   \t%s\t%v\n", model.Uuid, model.Name, model.Tag, model.DockerID, model.IsAvailable)
	}
	return nil
}

func (Sb *ServiceBroker) ListServices(out io.Writer, opts models.ServiceContainerFindOpts) error {
	models, err := models.ServiceContainerFind(Sb.Database, opts)
	if err != nil {
		return err
	}
	for _, model := range models {
		status := Sb.GetServiceContainerStatus(model)
		fmt.Fprintf(out, "%s\t%s   \t%s   \t%s\n", model.Uuid, model.Name, model.DockerId, status)
	}
	return nil
}

func (Sb *ServiceBroker) RenderManifestFiles(sm manifest.ServiceManifest) error {
	return manifest.RenderAllFromManifest(&sm)
}

func attempt(sm manifest.ServiceManifest, callbacks ...ServiceManifestCallback) error {
	var err error
	for _, callback := range callbacks {
		if err = callback(sm); err != nil {
			return err
		}
	}
	return err
}

func ServiceBrokerNew(dataSource string) (*ServiceBroker, error) {
	// client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	// sql.Open("sqlite3", dataSource)
	var sb ServiceBroker
	var err error

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}
	dbc, err := sql.Open("sqlite3", dataSource)
	if err != nil {
		return nil, err
	}

	sb.Client = cli
	sb.Database = dbc
	return &sb, nil
}
