package main

import (
	"os"

	"github.com/WilkinsonK/grawp/grawpadmin/manifest"
	"github.com/WilkinsonK/grawp/grawpadmin/service"
	"github.com/WilkinsonK/grawp/grawpadmin/service/models"
	"github.com/spf13/cobra"
)

var (
	grawpManifest   manifest.GrawpManifest
	serviceFindOpts models.ServiceContainerFindOpts
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
	RunE:  BuildImageService,
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

var initImageServiceCommand = &cobra.Command{
	Use:   "init",
	Short: "Create a new image service definition",
	Args:  cobra.ExactArgs(0),
	RunE:  NewService,
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

func commonImagePersistentFlags(commands ...*cobra.Command) {
	for _, cmd := range commands {
		commonImagePersistentFlagsC(cmd)
	}
}

func commonImagePersistentFlagsC(cmd *cobra.Command) {
	path, err := grawpManifest.GetServicesPath()
	if err != nil {
		panic(err)
	}
	cmd.PersistentFlags().StringVarP(&grawpManifest.ServicesPath, "services-path", "S", path, "Service definitions path")
}

func commonImageFlags(commands ...*cobra.Command) {
	for _, cmd := range commands {
		commonImageFlagsC(cmd)
	}
}

func commonImageFlagsC(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&grawpManifest.GetMetadata().Image.Path, "manifest-path", "M", "", "Service manifest path")
	cmd.Flags().StringVarP(&grawpManifest.GetMetadata().Image.Name, "manifest-name", "m", "service.yaml", "Service manifest file name")
}

func init() {
	if gm, err := manifest.LoadGrawpManifest(); err != nil {
		panic(err)
	} else {
		grawpManifest = gm
	}

	buildImageCommand.Flags().StringArrayVarP(&grawpManifest.GetMetadata().Image.BuildArgs, "build-arg", "b", []string{}, "Build arguments, as <key>=<value> pairs, to pass at construction")
	buildImageCommand.Flags().StringArrayVarP(&grawpManifest.GetMetadata().Image.BuildProperties, "property", "P", []string{}, "Build properties, as <key>=<value> pairs, to pass at construction")

	buildImageServiceCommand.Flags().StringVarP(&grawpManifest.GetMetadata().Service.Name, "name", "N", "", "Name of the service to be created")
	buildImageServiceCommand.Flags().StringSliceVarP(&grawpManifest.GetMetadata().Service.ExposedPorts, "publish", "p", []string{}, "Additional ports to expose on service intialization")
	buildImageServiceCommand.Flags().StringVarP(&grawpManifest.GetMetadata().Service.TagName, "image-tag", "t", "latest", "Service image tag name to create service from")
	buildImageServiceCommand.Flags().StringVarP(&grawpManifest.GetMetadata().Service.LocalVolume, "local-volume", "v", "server", "The output directory where server assets are managed")

	commonImageFlags(buildImageCommand, buildImageServiceCommand)

	listImageServicesCommand.Flags().StringVarP(&serviceFindOpts.Name, "name", "N", "", "Name of the container")
	listImageServicesCommand.Flags().StringVarP(&serviceFindOpts.DockerID, "id", "I", "", "Docker ID of the container")
	listImageServicesCommand.Flags().UintVarP(&serviceFindOpts.Limit, "limit", "l", 0, "Max number of items to return")

	initImageServiceCommand.Flags().StringVarP(&grawpManifest.GetMetadata().MinecraftVersion, "mc-version", "X", "1.21.10", "Minecraft version this service is meant for")
	initImageServiceCommand.Flags().StringVarP(&grawpManifest.GetMetadata().Service.Name, "name", "N", "", "Name of the service to be created")

	commonImagePersistentFlags(imagesCommand, imageServicesCommand, initImageServiceCommand)

	watchImageServiceCommand.Flags().StringVarP(&grawpManifest.DataName, "data-name", "d", grawpManifest.DataName, "Path to service data")

	imagesCommand.AddCommand(buildImageCommand, listImagesCommand)
	imageServicesCommand.AddCommand(buildImageServiceCommand, listImageServicesCommand, initImageServiceCommand)
	rootCommand.AddCommand(imagesCommand, imageServicesCommand, watchImageServiceCommand)
}

func initDatabase(cmd *cobra.Command, _ []string) error {
	broker, err := service.ServiceBrokerNew(grawpManifest.GetDataSource())
	if err != nil {
		return err
	}
	defer broker.Close()
	return broker.InitDatabase()
}

func BuildImage(cmd *cobra.Command, _ []string) error {
	broker, err := service.ServiceBrokerNew(grawpManifest.GetDataSource())
	if err != nil {
		return err
	}
	defer broker.Close()

	sm, err := grawpManifest.LoadServiceManifest()
	if err != nil {
		return err
	}
	return broker.BuildImage(sm)
}

func BuildImageService(cmd *cobra.Command, _ []string) error {
	// TODO: Still need to be able to rebuild the container
	// (if necessary or user wants to).
	// TODO: Still need to be able to build the container
	// image if it does not exist.
	broker, err := service.ServiceBrokerNew(grawpManifest.GetDataSource())
	if err != nil {
		return err
	}

	sm, err := grawpManifest.LoadServiceManifest()
	if err != nil {
		return err
	}
	return broker.BuildImageServiceFromManifest(sm, os.Stdout)
}

func ListImages(cmd *cobra.Command, _ []string) error {
	broker, err := service.ServiceBrokerNew(grawpManifest.GetDataSource())
	if err != nil {
		return err
	}
	defer broker.Close()
	return broker.ListImages(os.Stdout)
}

func ListServices(cmd *cobra.Command, _ []string) error {
	broker, err := service.ServiceBrokerNew(grawpManifest.GetDataSource())
	if err != nil {
		return err
	}
	defer broker.Close()
	return broker.ListServices(os.Stdout, serviceFindOpts)
}

func NewService(cmd *cobra.Command, _ []string) error {
	metadata := grawpManifest.GetMetadata()
	return service.ServiceNew(metadata.MinecraftVersion, grawpManifest.ServicesPath, metadata.Service.Name)
}

func WatchService(cmd *cobra.Command, args []string) error {
	broker, err := service.ServiceBrokerNew(grawpManifest.GetDataSource())
	if err != nil {
		return err
	}
	defer broker.Close()
	return WatcherNew(broker).Watch(args[0])
}

// Tasks:
// - Generate core assets to be packed into a new image.
// - Archive server assets
// - Remove assest that are too old
func main() {
	if err := rootCommand.Execute(); err != nil {
		os.Exit(1)
	}
}
