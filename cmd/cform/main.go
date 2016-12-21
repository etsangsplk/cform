package main

import (
	"os"

	log "github.com/Sirupsen/logrus"

	"github.com/spf13/cobra"
)

var rootCmdFlags struct {
	templateSrc string
	templateOut string
	overwrite   bool
}

var rootCmd = &cobra.Command{
	Use:   "cform",
	Short: "CloudFormation utility which provides Terraform-like CLI functionalities",
}

func main() {

	rootCmd.PersistentFlags().StringVar(&rootCmdFlags.templateSrc, "template-src", "templates", "Directory containing CloudFormation templates")
	rootCmd.PersistentFlags().StringVar(&rootCmdFlags.templateOut, "template-out", "", "Location to which the merged template will be written")
	rootCmd.PersistentFlags().BoolVar(&rootCmdFlags.overwrite, "overwrite", false, "Overwrite existing template output file")

	if err := rootCmd.Execute(); err != nil {
		log.WithError(err).Error("Failed to initialize cform ctl")
		os.Exit(-1)
	}
}
