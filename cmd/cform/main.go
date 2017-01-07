package main

import (
	"os"

	log "github.com/Sirupsen/logrus"

	"github.com/spf13/cobra"
)

var rootCmdFlags struct {
	debug         bool
	tmplSrc       string
	tmplOut       string
	tmplOverwrite bool
}

var rootCmd = &cobra.Command{
	Use:   "cform",
	Short: "CloudFormation utility which provides Terraform-like CLI functionalities",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if rootCmdFlags.debug {
			log.SetLevel(log.DebugLevel)
		}
	},
}

func main() {
	rootCmd.PersistentFlags().BoolVar(&rootCmdFlags.debug, "debug", false, "Print debug information")
	rootCmd.PersistentFlags().StringVar(&rootCmdFlags.tmplOut, "template-out", "", "Location to which the merged template will be written")
	rootCmd.PersistentFlags().StringVar(&rootCmdFlags.tmplSrc, "template-src", "templates", "Directory containing CloudFormation templates")
	rootCmd.PersistentFlags().BoolVar(&rootCmdFlags.tmplOverwrite, "template-overwrite", false, "Overwrite existing template output file")

	if err := rootCmd.Execute(); err != nil {
		log.WithError(err).Error("Failed to initialize cform ctl")
		os.Exit(-1)
	}
}
