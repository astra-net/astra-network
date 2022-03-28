package main

import (
	"fmt"
	"os"

	"github.com/Astra-Net/AstraNetwork/internal/cli"
	"github.com/spf13/cobra"
)

const (
	versionFormat = "Astra (C) 2020. %v, version %v-%v (%v %v)"
)

// Version string variables
var (
	version string
	builtBy string
	builtAt string
	commit  string
)

var versionFlag = cli.BoolFlag{
	Name:      "version",
	Shorthand: "V",
	Usage:     "display version info",
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "print version of the astra binary",
	Long:  "print version of the astra binary",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		printVersion()
		os.Exit(0)
	},
}

func getAstraVersion() string {
	return fmt.Sprintf(versionFormat, "astra", version, commit, builtBy, builtAt)
}

func printVersion() {
	fmt.Println(getAstraVersion())
}
