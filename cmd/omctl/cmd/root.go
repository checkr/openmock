package cmd

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:          "omctl",
	Short:        "CLI for openmock",
	SilenceUsage: true,
	Long: `
		A simple CLI for interacting with openmock.
`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		logrus.Fatal(err)
	}
}

// local openmock directory
var localDirectory string

// URL of remote openmock instance to control
var openMockURL string

// set key used when post / delete directory with a specific key
var setKey string

func init() {
	RootCmd.PersistentFlags().StringVarP(&localDirectory, "directory", "d", "./demo_templates", "Local directory in filesystem to upload to remote openmock")
	RootCmd.PersistentFlags().StringVarP(&openMockURL, "url", "u", "http://localhost:9998", "URL for remote openmock instance to control")
	RootCmd.PersistentFlags().StringVarP(&setKey, "set-key", "k", "", "'set' key to use when manipulating directory with a specific key")
}
