package main

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	yaml "gopkg.in/yaml.v2"

	log "github.com/Sirupsen/logrus"
	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use:   "cfn-tmpl",
	Short: "Helper utility for managing Cloudformation templates",
}

var flags struct {
	sourceDir  string
	outputFile string
	overwrite  bool
}

var mergeCmd = &cobra.Command{
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

func main() {
	log.SetLevel(log.DebugLevel)

	RootCmd.PersistentFlags().StringVar(&flags.sourceDir, "source-dir", "", "Directory containing the Cloudformation template files")
	RootCmd.PersistentFlags().StringVar(&flags.outputFile, "output-file", "", "File to which the merged template will be written")
	RootCmd.PersistentFlags().BoolVar(&flags.overwrite, "overwrite", false, "Overwrite existing output file")

	RootCmd.AddCommand(mergeCmd)

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
	Next() ([]byte, error)
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

func (r *DirectoryReader) Next() ([]byte, error) {
	source, err := ioutil.ReadFile(r.fileNames[r.idx])
	if err != nil {
		return nil, err
	}
	return source, nil
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
		source, err := reader.Next()
		if err != nil {
			return nil, err
		}

		m, err := unmarshalCfnYaml(source)
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

	d, err := marshalCfnYaml(mergedMap)
	if err != nil {
		return nil, err
	}

	return d, nil
}

func unmarshalCfnYaml(source []byte) (CfnMap, error) {
	// Escape quotes so that it can be unescaped later to make it appear later
	// in the final generated yaml
	str := strings.Replace(string(source), "\"", "\\\"", -1)

	var m = make(CfnMap)
	if err := yaml.Unmarshal([]byte(str), &m); err != nil {
		return m, err
	}
	return m, nil
}

func marshalCfnYaml(m CfnMap) ([]byte, error) {
	d, err := yaml.Marshal(m)
	if err != nil {
		return nil, err
	}

	// Unescape quotes which were escaped during unmarshalling
	str := strings.Replace(string(d), "\\\"", "\"", -1)
	return []byte(str), nil
}
