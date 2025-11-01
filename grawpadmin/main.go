package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	"github.com/WilkinsonK/grawp/bin/grawpadmin/service/build"
	"github.com/WilkinsonK/grawp/bin/grawpadmin/service/manifest"
	"github.com/WilkinsonK/grawp/bin/grawpadmin/service/models"
	"github.com/docker/docker/client"
	"github.com/spf13/cobra"
)

var (
	dataName            string
	imageName           string
	imagePath           string
	serviceFindOpts     models.ServiceContainerFindOpts
	serviceExposedPorts []string
	serviceImageTagName string
	serviceLocalVolume  string
	serviceName         string
	servicesPath        string
)

var rootCommand = &cobra.Command{
	Use:     "grawpadmin",
	Short:   "Manage processes of this project",
	Long:    "Grawpadmin is an application meant for maintaining this project and its processes",
	Version: "1.0.0",
}

var buildImageCommand = &cobra.Command{
	Use:   "build",
	Short: "Build a container image",
	Args:  cobra.ExactArgs(0),
	RunE:  BuildImage,
}

var buildImageServiceCommand = &cobra.Command{
	Use:   "build",
	Short: "Build and start a container from an image or image manifest",
	Args:  cobra.ExactArgs(0),
	RunE:  InitImageService,
}

var imagesCommand = &cobra.Command{
	Use:               "images",
	Short:             "manage service images",
	Args:              cobra.ExactArgs(0),
	PersistentPreRunE: initDatabase,
}

var imageServicesCommand = &cobra.Command{
	Use:               "services",
	Short:             "manage service containers",
	Args:              cobra.ExactArgs(0),
	PersistentPreRunE: initDatabase,
}

var listImagesCommand = &cobra.Command{
	Use:   "list",
	Short: "List service images available",
	Args:  cobra.ExactArgs(0),
	RunE:  ListImages,
}

var listImageServicesCommand = &cobra.Command{
	Use:   "list",
	Short: "List services available",
	Args:  cobra.ExactArgs(0),
	RunE:  ListServices,
}

var watchImageServiceCommand = &cobra.Command{
	Aliases: []string{"start"},
	Use:     "watch <name>",
	Short:   "Start and watch a running service container",
	Long:    "Watches a service container, restarting it if failure is detected.",
	Args:    cobra.ExactArgs(1),
	PreRunE: initDatabase,
	RunE:    WatchService,
}

func buildImage(cli *client.Client) error {
	sm, err := manifest.LoadManifest(filepath.Join(servicesPath, imagePath, imageName))
	if err != nil {
		return err
	}
	opts := build.ServiceBuildOpts{
		DataPath:       getDataSource(),
		Manifest:       &sm,
		OutDestination: os.Stdout,
	}

	_, err = build.BuildImageFromManifest(cli, opts)
	return err
}

func commonImagePersistentFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().StringVarP(&dataName, "data-name", "d", "data.db", "Path to service data")
	cmd.PersistentFlags().StringVarP(&servicesPath, "services-path", "S", ".", "Service definitions path")
}

func commonImageFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&imagePath, "manifest-path", "M", "", "Service manifest path")
	cmd.Flags().StringVarP(&imageName, "manifest-name", "m", "service.yaml", "Service manifest file name")
}

func init() {
	buildImageServiceCommand.Flags().StringVarP(&serviceName, "name", "N", "", "Name of the service to be created")
	buildImageServiceCommand.Flags().StringSliceVarP(&serviceExposedPorts, "publish", "p", []string{}, "Additional ports to expose on service intialization")
	buildImageServiceCommand.Flags().StringVarP(&serviceImageTagName, "image-tag", "t", "latest", "Service image tag name to create service from")
	buildImageServiceCommand.Flags().StringVarP(&serviceLocalVolume, "local-volume", "v", "server", "The output directory where server assets are managed")

	commonImageFlags(buildImageServiceCommand)
	commonImageFlags(buildImageCommand)

	listImageServicesCommand.Flags().StringVarP(&serviceFindOpts.Name, "name", "N", "", "Name of the container")
	listImageServicesCommand.Flags().StringVarP(&serviceFindOpts.DockerID, "id", "I", "", "Docker ID of the container")
	listImageServicesCommand.Flags().UintVarP(&serviceFindOpts.Limit, "limit", "l", 0, "Max number of items to return")

	commonImagePersistentFlags(imagesCommand)
	commonImagePersistentFlags(imageServicesCommand)

	watchImageServiceCommand.Flags().StringVarP(&dataName, "data-name", "d", "data.db", "Path to service data")

	imagesCommand.AddCommand(buildImageCommand, listImagesCommand)
	imageServicesCommand.AddCommand(buildImageServiceCommand, listImageServicesCommand)
	rootCommand.AddCommand(imagesCommand, imageServicesCommand, watchImageServiceCommand)
}

func initImageService(cli *client.Client) error {
	// TODO: Still need to be able to rebuild the container
	// (if necessary or user wants to).
	// TODO: Still need to be able to build the container
	// image if it does not exist.
	sm, err := manifest.LoadManifest(filepath.Join(servicesPath, imagePath, imageName))
	if err != nil {
		return err
	}

	sm.Ports = append(sm.Ports, serviceExposedPorts...)
	if sm.LocalVolume == "" {
		sm.LocalVolume = serviceLocalVolume
	}

	opts := build.ServiceBuildOpts{
		DataPath:    getDataSource(),
		Manifest:    &sm,
		ServiceName: serviceName,
		TagName:     serviceImageTagName,
	}
	model, err := build.BuildServiceFromManifest(cli, opts)
	if err != nil {
		return err
	}

	fmt.Print(model.Name)
	if model.DockerId != "" {
		fmt.Printf(" - %s", model.DockerId)
	}
	fmt.Println()
	return err
}

func getDataSource() string {
	return filepath.Join(servicesPath, dataName)
}

func initDatabase(cmd *cobra.Command, _ []string) error {
	return models.WithDatabase(getDataSource(), func(db *sql.DB) error {
		return models.InitDatabaseTables(db)
	})
}

func listImages(db *sql.DB) error {
	models, err := models.ServiceImagesList(db)
	if err != nil {
		return err
	}
	for _, model := range models {
		fmt.Printf("%s\t%s   \t%s   \t%s\t%v\n", model.Uuid, model.Name, model.Tag, model.DockerID, model.IsAvailable)
	}
	return nil
}

func listServices(db *sql.DB) error {
	return withClient(func(cli *client.Client) error {
		return listServicesInner(cli, db)
	})
}

func listServicesInner(cli *client.Client, db *sql.DB) error {
	var status string
	models, err := models.ServiceContainerFind(db, serviceFindOpts)
	if err != nil {
		return err
	}
	for _, model := range models {
		resp, err := cli.ContainerInspect(context.Background(), model.DockerId)
		if err != nil {
			status = "unknown"
		} else {
			status = resp.State.Status
		}
		fmt.Printf("%s\t%s   \t%s   \t%s\n", model.Uuid, model.Name, model.DockerId, status)
	}
	return nil
}

func watchServiceWithClient(db *sql.DB, args []string) func(cli *client.Client) error {
	return func(cli *client.Client) error {
		return WatchImageService(cli, db, args[0])
	}
}

func watchServiceWithDatabase(args []string) func(db *sql.DB) error {
	return func(db *sql.DB) error {
		return withClient(watchServiceWithClient(db, args))
	}
}

func withClient(callback func(cli *client.Client) error) error {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return err
	}
	defer cli.Close()
	return callback(cli)
}

func BuildImage(cmd *cobra.Command, _ []string) error {
	return withClient(buildImage)
}

func InitImageService(cmd *cobra.Command, _ []string) error {
	return withClient(initImageService)
}

func ListImages(cmd *cobra.Command, _ []string) error {
	return models.WithDatabase(getDataSource(), listImages)
}

func ListServices(cmd *cobra.Command, _ []string) error {
	return models.WithDatabase(getDataSource(), listServices)
}

func WatchService(cmd *cobra.Command, args []string) error {
	return models.WithDatabase(getDataSource(), watchServiceWithDatabase(args))
}

// Commands:
// - Archive server assets
// - Start container
// - Watch container
func main() {
	if err := rootCommand.Execute(); err != nil {
		os.Exit(1)
	}
}
