package goss

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"

	"gopkg.in/yaml.v2"

	"github.com/aelsabbahy/goss/resource"
	"github.com/aelsabbahy/goss/util"
)

const (
	UNSET = iota
	JSON
	YAML
)

var outStoreFormat = UNSET
var currentTemplateFilter TemplateFilter
var debug = false

func getStoreFormatFromFileName(f string) (int, error) {
	ext := filepath.Ext(f)
	switch ext {
	case ".json":
		return JSON, nil
	case ".yaml", ".yml":
		return YAML, nil
	default:
		return 0, fmt.Errorf("unknown file extension: %v", ext)
	}
}

func getStoreFormatFromData(data []byte) (int, error) {
	var v interface{}
	if err := unmarshalJSON(data, &v); err == nil {
		return JSON, nil
	}
	if err := unmarshalYAML(data, &v); err == nil {
		return YAML, nil
	}

	return 0, fmt.Errorf("unable to determine format from content")
}

// ReadJSON Reads json file returning GossConfig
func ReadJSON(filePath string) (GossConfig, error) {
	file, err := ioutil.ReadFile(filePath)
	if err != nil {
		return GossConfig{}, fmt.Errorf("file error: %v", err)
	}

	return ReadJSONData(file, false)
}

type TmplVars struct {
	Vars map[string]interface{}
}

func (t *TmplVars) Env() map[string]string {
	env := make(map[string]string)
	for _, i := range os.Environ() {
		sep := strings.Index(i, "=")
		env[i[0:sep]] = i[sep+1:]
	}
	return env
}

func loadVars(varsFile string, varsInline string) (map[string]interface{}, error) {
	vars, err := varsFromFile(varsFile)
	if err != nil {
		return nil, fmt.Errorf("Error: loading vars file '%s'\n%w", varsFile, err)
	}

	varsExtra, err := varsFromString(varsInline)
	if err != nil {
		return nil, fmt.Errorf("Error: loading inline vars\n%w", err)
	}

	for k, v := range varsExtra {
		vars[k] = v
	}

	return vars, nil
}

func varsFromFile(varsFile string) (map[string]interface{}, error) {
	vars := make(map[string]interface{})
	if varsFile == "" {
		return vars, nil
	}
	data, err := ioutil.ReadFile(varsFile)
	if err != nil {
		return vars, err
	}
	format, err := getStoreFormatFromData(data)
	if err != nil {
		return nil, err
	}
	if err := unmarshal(data, &vars, format); err != nil {
		return vars, err
	}
	return vars, nil
}

func varsFromString(varsString string) (map[string]interface{}, error) {
	vars := make(map[string]interface{})
	if varsString == "" {
		return vars, nil
	}
	data := []byte(varsString)
	format, err := getStoreFormatFromData(data)
	if err != nil {
		return nil, err
	}

	if err := unmarshal(data, &vars, format); err != nil {
		return vars, err
	}
	return vars, nil
}

// ReadJSONData Reads json byte array returning GossConfig
func ReadJSONData(data []byte, detectFormat bool) (GossConfig, error) {
	var err error
	if currentTemplateFilter != nil {
		data, err = currentTemplateFilter(data)
		if err != nil {
			return GossConfig{}, err
		}
		if debug {
			fmt.Println("DEBUG: file after text/template render")
			fmt.Println(string(data))
		}
	}

	format := outStoreFormat
	if detectFormat == true {
		format, err = getStoreFormatFromData(data)
		if err != nil {
			return GossConfig{}, err
		}
	}

	gossConfig := NewGossConfig()
	// Horrible, but will do for now
	if err := unmarshal(data, gossConfig, format); err != nil {
		return *gossConfig, err
	}

	return *gossConfig, nil
}

// RenderJSON reads json file recursively returning string
func RenderJSON(c *util.Config) (string, error) {
	var err error
	debug = c.Debug
	currentTemplateFilter, err = NewTemplateFilter(c.Vars, c.VarsInline)
	if err != nil {
		return "", err
	}

	outStoreFormat, err = getStoreFormatFromFileName(c.Spec)
	if err != nil {
		return "", err
	}

	j, err := ReadJSON(c.Spec)
	if err != nil {
		return "", err
	}

	gossConfig, err := mergeJSONData(j, 0, filepath.Dir(c.Spec))
	if err != nil {
		return "", err
	}

	b, err := marshal(gossConfig)
	if err != nil {
		return "", fmt.Errorf("rendering failed: %v", err)
	}

	return string(b), nil
}

func mergeJSONData(gossConfig GossConfig, depth int, path string) (GossConfig, error) {
	depth++
	if depth >= 50 {
		return GossConfig{}, fmt.Errorf("max depth of 50 reached, possibly due to dependency loop in goss file")
	}
	// Our return gossConfig
	ret := *NewGossConfig()
	ret = mergeGoss(ret, gossConfig)

	// Sort the gossfiles to ensure consistent ordering
	var keys []string
	for k := range gossConfig.Gossfiles {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Merge gossfiles in sorted order
	for _, k := range keys {
		g := gossConfig.Gossfiles[k]
		var fpath string
		if strings.HasPrefix(g.GetGossfile(), "/") {
			fpath = g.GetGossfile()
		} else {
			fpath = filepath.Join(path, g.GetGossfile())
		}
		if g.GetSkip() {
			// Do not process gossfiles with the skip attribute
			continue
		}
		matches, err := filepath.Glob(fpath)
		if err != nil {
			return ret, fmt.Errorf("error in expanding glob pattern: %q", err)
		}
		if matches == nil {
			return ret, fmt.Errorf("no matched files were found: %q", fpath)
		}
		for _, match := range matches {
			fdir := filepath.Dir(match)
			j, err := ReadJSON(match)
			if err != nil {
				return GossConfig{}, fmt.Errorf("could not read json data in %s: %s", match, err)
			}
			j, err = mergeJSONData(j, depth, fdir)
			if err != nil {
				return ret, fmt.Errorf("could not write json data: %s", err)
			}
			ret = mergeGoss(ret, j)
		}
	}
	return ret, nil
}

func WriteJSON(filePath string, gossConfig GossConfig) error {
	jsonData, err := marshal(gossConfig)
	if err != nil {
		return fmt.Errorf("failed to write %s: %s", filePath, err)
	}

	// check if the auto added json data is empty before writing to file.
	emptyConfig := *NewGossConfig()
	emptyData, err := marshal(emptyConfig)
	if err != nil {
		return fmt.Errorf("failed to write %s: %s", filePath, err)
	}

	if string(emptyData) == string(jsonData) {
		log.Printf("Can't write empty configuration file. Please check resource name(s).")
		return nil
	}

	if err := ioutil.WriteFile(filePath, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write %s: %s", filePath, err)
	}

	return nil
}

func resourcePrint(fileName string, res resource.ResourceRead, announce bool) {
	resMap := map[string]resource.ResourceRead{res.ID(): res}

	oj, _ := marshal(resMap)
	typ := reflect.TypeOf(res)
	typs := strings.Split(typ.String(), ".")[1]

	if announce {
		fmt.Printf("Adding %s to '%s':\n\n%s\n\n", typs, fileName, string(oj))
	}
}

func marshal(gossConfig interface{}) ([]byte, error) {
	switch outStoreFormat {
	case JSON:
		return marshalJSON(gossConfig)
	case YAML:
		return marshalYAML(gossConfig)
	default:
		return nil, fmt.Errorf("StoreFormat unset")
	}
}

func unmarshal(data []byte, v interface{}, storeFormat int) error {
	switch storeFormat {
	case JSON:
		return unmarshalJSON(data, v)
	case YAML:
		return unmarshalYAML(data, v)
	default:
		return fmt.Errorf("StoreFormat unset")
	}
}

func marshalJSON(gossConfig interface{}) ([]byte, error) {
	return json.MarshalIndent(gossConfig, "", "    ")
}

func unmarshalJSON(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

func marshalYAML(gossConfig interface{}) ([]byte, error) {
	return yaml.Marshal(gossConfig)
}

func unmarshalYAML(data []byte, v interface{}) error {
	err := yaml.Unmarshal(data, v)
	if err != nil {
		return fmt.Errorf("could not unmarshal %q as YAML data: %s", string(data), err)
	}

	return nil
}
