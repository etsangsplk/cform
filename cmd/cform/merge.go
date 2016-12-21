package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/isubuz/cform"
	"github.com/spf13/cobra"
)

var mergeCmd = &cobra.Command{
	Use:   "merge",
	Short: "Merge CloudFormation templates",
	Run: func(cmd *cobra.Command, args []string) {
		// Check if output file exists, can it be overwritten
		if _, err := os.Stat(rootCmdFlags.templateOut); err == nil && !rootCmdFlags.overwrite {
			log.WithField("template-out", rootCmdFlags.templateOut).Error("File already exists")
			os.Exit(-1)
		}

		dirReader, err := cform.NewDirectoryReader(rootCmdFlags.templateSrc)
		if err != nil {
			log.WithError(err).Error("Could not create directory reader")
			os.Exit(-1)
		}

		yaml := []byte(fmt.Sprintf("# THIS FILE HAS BEEN GENERATED AUTOMATICALLY BY cform AT %s.\n# DO NOT MODIFY THIS FILE MANUALLY.\n\n", time.Now()))

		merged, err := cform.MergeTemplates(dirReader)
		if err != nil {
			log.WithError(err).Error("Could not merge yaml files")
			os.Exit(-1)
		}

		yaml = append(yaml, merged...)

		if err = ioutil.WriteFile(rootCmdFlags.templateOut, yaml, 0644); err != nil {
			log.WithField("template-out", rootCmdFlags.templateOut).Error("Could not write to output file")
			os.Exit(-1)
		}
	},
}

func init() {
	rootCmd.AddCommand(mergeCmd)
}
