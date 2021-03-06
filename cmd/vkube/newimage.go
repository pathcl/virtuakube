package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"go.universe.tf/virtuakube"
)

var newimageCmd = &cobra.Command{
	Use:   "newimage",
	Short: "Create a base VM disk image",
	Args:  cobra.NoArgs,
	Run:   withUniverse(&imageFlags.universe, newimage),
}

var imageFlags = struct {
	universe universeFlags
	name     string
	script   string
	k8s      bool
	prepull  bool
}{}

func init() {
	rootCmd.AddCommand(newimageCmd)

	addUniverseFlags(newimageCmd, &imageFlags.universe, false, true)
	newimageCmd.Flags().StringVar(&imageFlags.name, "name", "", "name of the new disk image")
	newimageCmd.Flags().StringVar(&imageFlags.script, "script", "", "path to a shell script to customize the disk image")
	newimageCmd.Flags().BoolVar(&imageFlags.k8s, "install-k8s", true, "install prerequisites for Kubernetes cluster setup")
	newimageCmd.Flags().BoolVar(&imageFlags.prepull, "prepull-k8s", true, "pre-pull docker images required to run Kubernetes")
}

func newimage(u *virtuakube.Universe, verbose bool) error {
	if imageFlags.prepull && !imageFlags.k8s {
		return errors.New("Cannot prepull k8s images if I'm not installing k8s")
	}

	cfg := &virtuakube.ImageConfig{
		Name: imageFlags.name,
	}
	if imageFlags.k8s {
		cfg.CustomizeFuncs = append(cfg.CustomizeFuncs, virtuakube.CustomizeInstallK8s)
	}
	if imageFlags.prepull {
		cfg.CustomizeFuncs = append(cfg.CustomizeFuncs, virtuakube.CustomizePreloadK8sImages)
	}
	if imageFlags.script != "" {
		cfg.CustomizeFuncs = append(cfg.CustomizeFuncs, virtuakube.CustomizeScript(imageFlags.script))
	}
	if verbose {
		cfg.BuildLog = os.Stdout
	}

	fmt.Printf("Creating VM base image %q...\n", imageFlags.name)

	if _, err := u.NewImage(cfg); err != nil {
		return fmt.Errorf("Creating image: %v", err)
	}

	fmt.Printf("Created VM base image %q\n", imageFlags.name)

	return nil
}
