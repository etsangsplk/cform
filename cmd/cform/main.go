package main

import (
	"io/ioutil"
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

		if rootCmdFlags.tmplOut == "" {
			f, err := ioutil.TempFile("", "cform")
			if err != nil {
				log.Error("Cannot create output file for generated template")
				os.Exit(-1)
			}
			rootCmdFlags.tmplOut = f.Name()
			log.WithField("template-out", f.Name()).Debug("created new output file for template")
		} else {
			// If output file exists, check if it can be overwritten.
			if _, err := os.Stat(rootCmdFlags.tmplOut); err == nil {
				if !rootCmdFlags.tmplOverwrite {
					log.WithField("template-out", rootCmdFlags.tmplOut).Error("File already exists")
					os.Exit(-1)
				}
			}
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
