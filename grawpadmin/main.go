package main

import (
	"fmt"
	"os"
	"slices"

	"github.com/WilkinsonK/grawp/grawpadmin/manifest"
	"github.com/WilkinsonK/grawp/grawpadmin/service"
	"github.com/WilkinsonK/grawp/grawpadmin/service/models"
	"github.com/WilkinsonK/grawp/grawpadmin/util"
	"github.com/spf13/cobra"
)

var (
	Manifest        manifest.GrawpManifest
	ImageFindOpts   models.ServiceImageFindOpts
	ServiceFindOpts models.ServiceContainerFindOpts
)

var rootCommand = &cobra.Command{
	Use:     "grawpadmin",
	Short:   "Manage processes of this project",
	Long:    "Grawpadmin is an application meant for maintaining this project and its processes",
	Version: "1.1.0",
}

var archiveServiceCommand = &cobra.Command{
	Aliases: []string{"arc"},
	Use:     "archive",
	Short:   "Create tar ball(s) of server assets.",
	Args:    cobra.ExactArgs(0),
	RunE:    ArchiveService,
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

var printManifestCommand = &cobra.Command{
	Use:   "manifest",
	Short: "Print the service manifest to stdout",
	Args:  cobra.ExactArgs(0),
	RunE:  PrintManifest,
}

var rebuildSelf = &cobra.Command{
	Aliases: []string{"rs"},
	Use:     "rebuild-self",
	Short:   "Rebuild this application",
	Args:    cobra.ExactArgs(0),
	RunE:    RebuildSelf,
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
	util.ForEach(slices.Values(commands), commonImagePersistentFlagsC)
}

func commonImagePersistentFlagsC(cmd *cobra.Command) {
	path, err := Manifest.GetServicesPath()
	if err != nil {
		panic(err)
	}
	cmd.PersistentFlags().StringVarP(&Manifest.ServicesPath, "services-path", "S", path, "Service definitions path")
}

func commonImageFlags(commands ...*cobra.Command) {
	util.ForEach(slices.Values(commands), commonImageFlagsC)
}

func commonImageFlagsC(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&Manifest.GetMetadata().Image.Path, "manifest-path", "M", "", "Service manifest path")
	cmd.Flags().StringVarP(&Manifest.GetMetadata().Image.Name, "manifest-name", "m", "service.yaml", "Service manifest file name")
}

func init() {
	if gm, err := manifest.LoadGrawpManifest(); err != nil {
		panic(err)
	} else {
		Manifest = gm
	}

	initCommandArchiveService()
	initCommandBuildImage()
	initCommandBuildImageService()
	initCommandImages()
	initCommandImageServices()
	initCommandInitImageService()
	initCommandListImages()
	initCommandListImageServices()
	initCommandPrintManifest()
	initCommandWatchService()

	subcmds := []*cobra.Command{
		archiveServiceCommand,
		imagesCommand,
		imageServicesCommand,
		printManifestCommand,
		watchImageServiceCommand,
	}
	rootCommand.AddCommand(subcmds...)

	// For project level embedded values.
	initLinkOptions()
	if _DeveloperMode {
		rootCommand.AddCommand(rebuildSelf)
	}
}

func initCommandArchiveService() {
	cmd := archiveServiceCommand
	commonImageFlags(cmd)
}

func initCommandBuildImage() {
	cmd := buildImageCommand
	cmd.Flags().StringArrayVarP(&Manifest.GetMetadata().Image.BuildArgs, "build-arg", "b", []string{}, "Build arguments, as <key>=<value> pairs, to pass at construction")
	cmd.Flags().StringArrayVarP(&Manifest.GetMetadata().Image.BuildProperties, "property", "P", []string{}, "Build properties, as <key>=<value> pairs, to pass at construction")
	commonImageFlags(cmd)
}

func initCommandBuildImageService() {
	cmd := buildImageServiceCommand
	cmd.Flags().StringVarP(&Manifest.GetMetadata().Service.Name, "name", "N", "", "Name of the service to be created")
	cmd.Flags().StringSliceVarP(&Manifest.GetMetadata().Service.ExposedPorts, "publish", "p", []string{}, "Additional ports to expose on service intialization")
	cmd.Flags().StringVarP(&Manifest.GetMetadata().Service.TagName, "image-tag", "t", "latest", "Service image tag name to create service from")
	cmd.Flags().StringVarP(&Manifest.GetMetadata().Service.LocalVolume, "local-volume", "v", "server", "The output directory where server assets are managed")
	commonImageFlags(cmd)
}

func initCommandImages() {
	cmd := imagesCommand
	commonImagePersistentFlags(cmd)
	cmd.AddCommand(buildImageCommand, listImagesCommand)
}

func initCommandImageServices() {
	cmd := imageServicesCommand
	commonImagePersistentFlags(cmd)
	cmd.AddCommand(buildImageServiceCommand, listImageServicesCommand, initImageServiceCommand)
}

func initCommandInitImageService() {
	cmd := initImageServiceCommand
	cmd.Flags().StringVarP(&Manifest.GetMetadata().MinecraftVersion, "mc-version", "X", "1.21.10", "Minecraft version this service is meant for")
	cmd.Flags().StringVarP(&Manifest.GetMetadata().Service.Name, "name", "N", "", "Name of the service to be created")
	commonImagePersistentFlags(initImageServiceCommand)
}

func initCommandListImages() {
	cmd := listImagesCommand
	cmd.Flags().StringVarP(&ImageFindOpts.Name, "name", "N", "", "Name of the image")
	cmd.Flags().StringVarP(&ImageFindOpts.DockerID, "id", "I", "", "Docker ID of the image")
	cmd.Flags().StringVarP(&ImageFindOpts.Tag, "tag", "t", "", "Image tag name")
	cmd.Flags().UintVarP(&ImageFindOpts.Limit, "limit", "l", 0, "Max number of items to return")
}

func initCommandListImageServices() {
	cmd := listImageServicesCommand
	cmd.Flags().StringVarP(&ServiceFindOpts.Name, "name", "N", "", "Name of the container")
	cmd.Flags().StringVarP(&ServiceFindOpts.DockerID, "id", "I", "", "Docker ID of the container")
	cmd.Flags().UintVarP(&ServiceFindOpts.Limit, "limit", "l", 0, "Max number of items to return")
}

func initCommandPrintManifest() {
	cmd := printManifestCommand
	commonImageFlags(cmd)
}

func initCommandWatchService() {
	cmd := watchImageServiceCommand
	cmd.Flags().StringVarP(&Manifest.DataName, "data-name", "d", Manifest.DataName, "Path to service data")
}

func initDatabase(cmd *cobra.Command, _ []string) error {
	broker, err := service.ServiceBrokerNew(&Manifest)
	if err != nil {
		return err
	}
	defer broker.Close()
	return broker.InitDatabase()
}

func ArchiveService(cmd *cobra.Command, _ []string) error {
	broker, err := service.ServiceBrokerNew(&Manifest)
	if err != nil {
		return err
	}
	defer broker.Close()

	sm, err := Manifest.LoadServiceManifest()
	if err != nil {
		return err
	}

	return broker.ArchiveService(sm)
}

func BuildImage(cmd *cobra.Command, _ []string) error {
	broker, err := service.ServiceBrokerNew(&Manifest)
	if err != nil {
		return err
	}
	defer broker.Close()

	sm, err := Manifest.LoadServiceManifest()
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
	broker, err := service.ServiceBrokerNew(&Manifest)
	if err != nil {
		return err
	}
	defer broker.Close()

	sm, err := Manifest.LoadServiceManifest()
	if err != nil {
		return err
	}

	return broker.BuildImageServiceFromManifest(sm, os.Stdout)
}

func ListImages(cmd *cobra.Command, _ []string) error {
	broker, err := service.ServiceBrokerNew(&Manifest)
	if err != nil {
		return err
	}
	defer broker.Close()
	return broker.ListImages(os.Stdout, ImageFindOpts)
}

func ListServices(cmd *cobra.Command, _ []string) error {
	broker, err := service.ServiceBrokerNew(&Manifest)
	if err != nil {
		return err
	}
	defer broker.Close()
	return broker.ListServices(os.Stdout, ServiceFindOpts)
}

func NewService(cmd *cobra.Command, _ []string) error {
	broker, err := service.ServiceBrokerNew(&Manifest)
	if err != nil {
		return err
	}
	defer broker.Close()
	return broker.NewService()
}

func PrintManifest(cmd *cobra.Command, _ []string) error {
	broker, err := service.ServiceBrokerNew(&Manifest)
	if err != nil {
		return err
	}
	defer broker.Close()

	sm, err := Manifest.LoadServiceManifest()
	if err != nil {
		return err
	}

	display, err := sm.Display()
	if err != nil {
		return err
	}
	fmt.Println(display)
	return nil
}

func RebuildSelf(cmd *cobra.Command, _ []string) error {
	return DoRebuildSelf()
}

func WatchService(cmd *cobra.Command, args []string) error {
	broker, err := service.ServiceBrokerNew(&Manifest)
	if err != nil {
		return err
	}
	defer broker.Close()
	return service.WatcherNew(broker).Watch(args[0])
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
