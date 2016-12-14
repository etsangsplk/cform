package main

import (
	"io/ioutil"
	"os"
	"path/filepath"

	yaml "gopkg.in/yaml.v2"

	log "github.com/Sirupsen/logrus"
	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use:   "cfn-tmpl",
	Short: "Cloudformation helper utility",
}

var flags struct {
	sourceDir  string
	outputFile string
	overwrite  bool
}

var createCmd = &cobra.Command{
	Use:   "merge",
	Short: "Merge templates in a source directory",
	Run: func(cmd *cobra.Command, args []string) {
		// Check if output file exists, can it be overwritten
		if _, err := os.Stat(flags.outputFile); err == nil && !flags.overwrite {
			log.WithField("outputFile", flags.outputFile).Error("File already exists")
			os.Exit(-1)
		}

		dirReader, err := NewDirectoryReader(flags.sourceDir)
		if err != nil {
			log.WithError(err).Error("Could not create directory reader")
			os.Exit(-1)
		}

		yaml, err := MergeYaml(dirReader)
		if err != nil {
			log.WithError(err).Error("Could not merge yaml files")
			os.Exit(-1)
		}

		if err = ioutil.WriteFile(flags.outputFile, yaml, 0644); err != nil {
			log.WithField("output-file", flags.outputFile).Error("Could not write to output file")
			os.Exit(-1)
		}
	},
}

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update a Cloudformation stack",
}

func main() {
	log.SetLevel(log.DebugLevel)

	RootCmd.PersistentFlags().StringVar(&flags.sourceDir, "source-dir", "", "Directory containing the Cloudformation template files")
	RootCmd.PersistentFlags().StringVar(&flags.outputFile, "output-file", "", "File to which the merged template will be written")
	RootCmd.PersistentFlags().BoolVar(&flags.overwrite, "overwrite", false, "Overwrite existing output file")

	RootCmd.AddCommand(createCmd)
	RootCmd.AddCommand(updateCmd)

	if err := RootCmd.Execute(); err != nil {
		log.WithError(err).Error("Failed to initialize cmd")
		os.Exit(-1)
	}
}

// CfnMap is a representation of the YAML struct returned after reading the
// contents from the YAML file.
type CfnMap map[string]map[string]interface{}

// YamlReader represents the interface to read and return yaml file contents.
type YamlReader interface {
	Next() (CfnMap, error)
	HasNext() bool
}

// DirectoryReader implements the `YamlReader` interface to read yaml files
// from an input source directory.
type DirectoryReader struct {
	SourceDir string
	fileNames []string
	idx       int
}

func NewDirectoryReader(sourceDir string) (*DirectoryReader, error) {
	r := &DirectoryReader{SourceDir: sourceDir}
	files, err := ioutil.ReadDir(sourceDir)
	if err != nil {
		return r, err
	}

	for _, f := range files {
		r.fileNames = append(r.fileNames, filepath.Join(sourceDir, f.Name()))
	}
	r.idx = -1
	return r, nil
}

func (r *DirectoryReader) Next() (CfnMap, error) {
	var m = make(CfnMap)

	source, err := ioutil.ReadFile(r.fileNames[r.idx])
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal(source, &m)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func (r *DirectoryReader) HasNext() bool {
	r.idx++
	return r.idx < len(r.fileNames)
}

// MergeYaml reads yaml files from the input reader and returns a merged
// yaml string.
func MergeYaml(reader YamlReader) ([]byte, error) {
	mergedMap := make(CfnMap)

	for reader.HasNext() {
		m, err := reader.Next()
		if err != nil {
			return nil, err
		}

		for k, v := range m {
			if val, ok := mergedMap[k]; ok {
				for kk, vv := range v {
					val[kk] = vv
				}
			} else {
				mergedMap[k] = v
			}
		}
	}

	d, err := yaml.Marshal(mergedMap)
	if err != nil {
		return nil, err
	}

	return d, nil
}
