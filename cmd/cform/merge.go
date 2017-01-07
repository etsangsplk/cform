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
		if err := mergeFromDir(rootCmdFlags.tmplSrc, rootCmdFlags.tmplOut); err != nil {
			os.Exit(-1)
		}
	},
}

func mergeFromDir(tmplSrc, tmplOut string) error {
	dirReader, err := cform.NewDirectoryReader(tmplSrc)
	if err != nil {
		log.WithError(err).Error("Could not create directory reader")
		return err
	}

	yaml := []byte(fmt.Sprintf("# THIS FILE HAS BEEN GENERATED AUTOMATICALLY BY cform AT %s.\n# DO NOT MODIFY THIS FILE MANUALLY.\n\n", time.Now()))

	merged, err := cform.MergeTemplates(dirReader)
	if err != nil {
		log.WithError(err).Error("Could not merge yaml files")
		return err
	}

	yaml = append(yaml, merged...)

	if err = ioutil.WriteFile(tmplOut, yaml, 0644); err != nil {
		log.WithField("template-out", tmplOut).Error("Could not write to output file")
		return err
	}
	return nil
}

func init() {
	rootCmd.AddCommand(mergeCmd)
}
