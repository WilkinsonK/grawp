package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/docker/docker/client"
	"github.com/spf13/cobra"
)

var (
	imageName           string
	imagePath           string
	serviceExposedPorts []string
	serviceImageTagName string
	serviceLocalVolume  string
	serviceName         string
	servicesPath        string
)

var rootCommand = &cobra.Command{
	Use:   "grawpadmin",
	Short: "Manage processes of this project",
	Long:  "Grawpadmin is an application meant for maintaining this project and its processes",
}

var buildImage = &cobra.Command{
	Use:   "build <name>",
	Short: "Build a container image",
	Args:  cobra.ExactArgs(0),
	RunE:  BuildImage,
}

var initImageService = &cobra.Command{
	Use:   "init <name>",
	Short: "Build and start a container from an image or image manifest",
	Args:  cobra.ExactArgs(0),
	RunE:  InitImageService,
}

func commonImageFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&imagePath, "manifest-path", "M", "", "Service manifest path")
	cmd.Flags().StringVarP(&imageName, "manifest-name", "m", "service.yaml", "Service manifest file name")
}

func init() {
	initImageService.Flags().StringVarP(&serviceName, "name", "N", "", "Name of the service to be created")
	initImageService.Flags().StringSliceVarP(&serviceExposedPorts, "publish", "p", []string{}, "Additional ports to expose on service intialization")
	initImageService.Flags().StringVarP(&serviceImageTagName, "image-tag", "t", "latest", "Service image tag name to create service from")
	initImageService.Flags().StringVarP(&serviceLocalVolume, "local-volume", "v", "server", "The output directory where server assets are managed")
	commonImageFlags(buildImage)
	commonImageFlags(initImageService)
	// Used for reading all available manifest file(s).
	rootCommand.PersistentFlags().StringVarP(&servicesPath, "services-path", "S", ".", "Service definitions path")
	rootCommand.AddCommand(buildImage, initImageService)
}

func BuildImage(cmd *cobra.Command, _ []string) error {
	sm, err := LoadManifest(filepath.Join(servicesPath, imagePath, imageName))
	if err != nil {
		return err
	}
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return err
	}

	opts := ServiceBuildOpts{
		Manifest:       &sm,
		OutDestination: os.Stdout,
	}
	if err = BuildImageFromManifest(cli, opts); err != nil {
		return err
	}
	return nil
}

func InitImageService(cmd *cobra.Command, _ []string) error {
	// TODO: Still need to be able to rebuild the container
	// (if necessary or user wants to).
	sm, err := LoadManifest(filepath.Join(servicesPath, imagePath, imageName))
	if err != nil {
		return err
	}
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return err
	}

	sm.Ports = append(sm.Ports, serviceExposedPorts...)
	if sm.LocalVolume == "" {
		sm.LocalVolume = serviceLocalVolume
	}

	opts := ServiceBuildOpts{
		Manifest:    &sm,
		ServiceName: serviceName,
		TagName:     serviceImageTagName,
	}
	name, id, err := BuildServiceFromManifest(cli, opts)
	if err != nil {
		return err
	}

	fmt.Print(name)
	if id != "" {
		fmt.Printf(" - %s", id)
	}
	fmt.Println()

	return nil
}

// Commands:
// - Archive server assets
// - Start container
// - Watch container
// - List containers built
func main() {
	if err := rootCommand.Execute(); err != nil {
		os.Exit(1)
	}
}
