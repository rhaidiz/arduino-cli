// This file is part of arduino-cli.
//
// Copyright 2020 ARDUINO SA (http://www.arduino.cc/)
//
// This software is released under the GNU General Public License version 3,
// which covers the main part of arduino-cli.
// The terms of this license can be found at:
// https://www.gnu.org/licenses/gpl-3.0.en.html
//
// You can be released from the requirements of the above licenses by purchasing
// a commercial license. Buying such a license is mandatory if you want to
// modify or otherwise use the software for commercial activities involving the
// Arduino software without disclosing the source code of your own applications.
// To purchase a commercial license, send an email to license@arduino.cc.

package core

import (
	"context"
	"os"

	"github.com/arduino/arduino-cli/cli/errorcodes"
	"github.com/arduino/arduino-cli/cli/feedback"
	"github.com/arduino/arduino-cli/cli/globals"
	"github.com/arduino/arduino-cli/cli/instance"
	"github.com/arduino/arduino-cli/cli/output"
	"github.com/arduino/arduino-cli/commands/core"
	"github.com/arduino/arduino-cli/configuration"
	rpc "github.com/arduino/arduino-cli/rpc/commands"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func initInstallCommand() *cobra.Command {
	installCommand := &cobra.Command{
		Use:   "install PACKAGER:ARCH[@VERSION] ...",
		Short: "Installs one or more cores and corresponding tool dependencies.",
		Long:  "Installs one or more cores and corresponding tool dependencies.",
		Example: "  # download the latest version of Arduino SAMD core.\n" +
			"  " + os.Args[0] + " core install arduino:samd\n\n" +
			"  # download a specific version (in this case 1.6.9).\n" +
			"  " + os.Args[0] + " core install arduino:samd@1.6.9",
		Args: cobra.MinimumNArgs(1),
		Run:  runInstallCommand,
	}
	addPostInstallFlagsToCommand(installCommand)
	return installCommand
}

var postInstallFlags struct {
	runPostInstall  bool
	skipPostInstall bool
}

func addPostInstallFlagsToCommand(cmd *cobra.Command) {
	cmd.Flags().BoolVar(&postInstallFlags.runPostInstall, "run-post-install", false, "Force run of post-install scripts (if the CLI is not running interactively).")
	cmd.Flags().BoolVar(&postInstallFlags.skipPostInstall, "skip-post-install", false, "Force skip of post-install scripts (if the CLI is running interactively).")
}

func detectSkipPostInstallValue() bool {
	if postInstallFlags.runPostInstall && postInstallFlags.skipPostInstall {
		feedback.Errorf("The flags --run-post-install and --skip-post-install can't be both set at the same time.")
		os.Exit(errorcodes.ErrBadArgument)
	}
	if postInstallFlags.runPostInstall {
		logrus.Info("Will run post-install by user request")
		return false
	}
	if postInstallFlags.skipPostInstall {
		logrus.Info("Will skip post-install by user request")
		return true
	}

	if !configuration.IsInteractive {
		logrus.Info("Not running from console, will skip post-install by default")
		return true
	}
	logrus.Info("Running from console, will run post-install by default")
	return false
}

func runInstallCommand(cmd *cobra.Command, args []string) {
	inst, err := instance.CreateInstance()
	if err != nil {
		feedback.Errorf("Error installing: %v", err)
		os.Exit(errorcodes.ErrGeneric)
	}

	logrus.Info("Executing `arduino core install`")

	platformsRefs, err := globals.ParseReferenceArgs(args, true)
	if err != nil {
		feedback.Errorf("Invalid argument passed: %v", err)
		os.Exit(errorcodes.ErrBadArgument)
	}

	for _, platformRef := range platformsRefs {
		platformInstallReq := &rpc.PlatformInstallReq{
			Instance:        inst,
			PlatformPackage: platformRef.PackageName,
			Architecture:    platformRef.Architecture,
			Version:         platformRef.Version,
			SkipPostInstall: detectSkipPostInstallValue(),
		}
		_, err := core.PlatformInstall(context.Background(), platformInstallReq, output.ProgressBar(), output.TaskProgress())
		if err != nil {
			feedback.Errorf("Error during install: %v", err)
			os.Exit(errorcodes.ErrGeneric)
		}
	}
}
