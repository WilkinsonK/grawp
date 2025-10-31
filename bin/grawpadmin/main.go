package main

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	"github.com/WilkinsonK/grawp/bin/grawpadmin/service_manifest"
	"github.com/WilkinsonK/grawp/bin/grawpadmin/service_models"
	"github.com/docker/docker/client"
	"github.com/spf13/cobra"
)

var (
	dataName            string
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
	Use:     "build",
	Short:   "Build a container image",
	Args:    cobra.ExactArgs(0),
	PreRunE: initDatabase,
	RunE:    BuildImage,
}

var initImageService = &cobra.Command{
	Use:     "init",
	Short:   "Build and start a container from an image or image manifest",
	Args:    cobra.ExactArgs(0),
	PreRunE: initDatabase,
	RunE:    InitImageService,
}

var listImages = &cobra.Command{
	Use:     "listi",
	Short:   "List service images available",
	Args:    cobra.ExactArgs(0),
	PreRunE: initDatabase,
	RunE:    ListImages,
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
	commonImageFlags(listImages)

	rootCommand.PersistentFlags().StringVarP(&dataName, "data-name", "d", "data.db", "Path to service data")
	rootCommand.PersistentFlags().StringVarP(&servicesPath, "services-path", "S", ".", "Service definitions path")
	rootCommand.AddCommand(buildImage, initImageService, listImages)
}

func initDatabase(cmd *cobra.Command, _ []string) error {
	return withDatabase(func(db *sql.DB) error {
		return service_models.InitDatabaseTables(db)
	})
}

func withDatabase(callback func(db *sql.DB) error) error {
	db, err := sql.Open("sqlite3", filepath.Join(servicesPath, dataName))
	if err != nil {
		return err
	}
	defer db.Close()
	return callback(db)
}

func BuildImage(cmd *cobra.Command, _ []string) error {
	sm, err := service_manifest.LoadManifest(filepath.Join(servicesPath, imagePath, imageName))
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
	models, err := BuildImageFromManifest(cli, opts)
	if err != nil {
		return err
	}

	return withDatabase(func(db *sql.DB) error {
		_, err = service_models.ServiceImagePut(db, models...)
		return err
	})
}

func InitImageService(cmd *cobra.Command, _ []string) error {
	// TODO: Still need to be able to rebuild the container
	// (if necessary or user wants to).
	// TODO: Still need to be able to build the container
	// image if it does not exist.
	sm, err := service_manifest.LoadManifest(filepath.Join(servicesPath, imagePath, imageName))
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

func ListImages(cmd *cobra.Command, _ []string) error {
	return withDatabase(func(db *sql.DB) error {
		models, err := service_models.ServiceImagesList(db)
		if err != nil {
			return err
		}
		for idx, model := range models {
			fmt.Printf("MODEL(%d): %s\n", idx, model)
		}
		return nil
	})
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
