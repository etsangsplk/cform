package main

import (
	"os"

	log "github.com/Sirupsen/logrus"

	"github.com/spf13/cobra"
)

var rootCmdFlags struct {
	tmplSrc       string
	tmplOut       string
	tmplOverwrite bool
}

var rootCmd = &cobra.Command{
	Use:   "cform",
	Short: "CloudFormation utility which provides Terraform-like CLI functionalities",
}

func main() {
	log.SetLevel(log.DebugLevel)

	rootCmd.PersistentFlags().StringVar(&rootCmdFlags.tmplSrc, "template-src", "templates", "Directory containing CloudFormation templates")
	rootCmd.PersistentFlags().StringVar(&rootCmdFlags.tmplOut, "template-out", "", "Location to which the merged template will be written")
	rootCmd.PersistentFlags().BoolVar(&rootCmdFlags.tmplOverwrite, "template-overwrite", false, "Overwrite existing template output file")

	if err := rootCmd.Execute(); err != nil {
		log.WithError(err).Error("Failed to initialize cform ctl")
		os.Exit(-1)
	}
}
