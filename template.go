package cform

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"regexp"
	"strings"

	yaml "gopkg.in/yaml.v2"
)

// TemplateReader defines the operations performed by the reader which reads
// one or more YAML file contents.
type TemplateReader interface {
	Next() ([]byte, error)
	HasNext() bool
}

// TemplateMap is a representation of the YAML struct returned after reading
// contents from the YAML file.
type TemplateMap map[string]map[string]interface{}

// DirectoryReader implements the `TemplateReader` interface to read yaml files
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

func MergeTemplates(reader TemplateReader) ([]byte, error) {
	mergedMap := make(TemplateMap)

	for reader.HasNext() {
		source, err := reader.Next()
		if err != nil {
			return nil, err
		}

		if err := validateYaml(source); err != nil {
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

func validateYaml(yaml []byte) error {
	// Check if the yaml contains any shortcode intrinsic functions
	r, err := regexp.Compile(`(!(?:Base64|FindInMap|GetAtt|GetAZs|ImportValue|Join|Select|Split|Sub|Ref))`)
	if err != nil {
		return err
	}
	matches := r.FindAllString(string(yaml), -1)
	if len(matches) > 0 {
		return fmt.Errorf("Found shortcode intrinsic functions: %s; Use the Fn form instead", strings.Join(matches, ","))
	}
	return nil
}

func unmarshalCfnYaml(source []byte) (TemplateMap, error) {
	// Escape quotes so that it can be unescaped later to make it appear later
	// in the final generated yaml
	str := strings.Replace(string(source), "\"", "\\\"", -1)

	var m = make(TemplateMap)
	if err := yaml.Unmarshal([]byte(str), &m); err != nil {
		return m, err
	}
	return m, nil
}

func marshalCfnYaml(m TemplateMap) ([]byte, error) {
	d, err := yaml.Marshal(m)
	if err != nil {
		return nil, err
	}

	// Unescape quotes which were escaped during unmarshalling
	str := strings.Replace(string(d), "\\\"", "\"", -1)
	return []byte(str), nil
}
