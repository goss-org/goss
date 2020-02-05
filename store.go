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
	"github.com/urfave/cli"
)

const (
	UNSET = iota
	JSON
	YAML
)

var OutStoreFormat = UNSET
var TemplateFilter func(data []byte) []byte
var debug = false

func getStoreFormatFromFileName(f string) int {
	ext := filepath.Ext(f)
	switch ext {
	case ".json":
		return JSON
	case ".yaml", ".yml":
		return YAML
	default:
		log.Fatalf("Unknown file extension: %v", ext)
	}
	return 0
}

func getStoreFormatFromData(data []byte) int {
	var v interface{}
	if err := unmarshalJSON(data, &v); err == nil {
		return JSON
	}
	if err := unmarshalYAML(data, &v); err == nil {
		return YAML
	}
	log.Fatalf("Unable to determine format from content")
	return 0
}

// ReadJSON Reads json file returning GossConfig
func ReadJSON(filePath string) GossConfig {
	file, err := ioutil.ReadFile(filePath)
	if err != nil {
		fmt.Printf("File error: %v\n", err)
		os.Exit(1)
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
	format := getStoreFormatFromData(data)
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
	format := getStoreFormatFromData(data)
	if err := unmarshal(data, &vars, format); err != nil {
		return vars, err
	}
	return vars, nil
}

// ReadJSONData Reads json byte array returning GossConfig
func ReadJSONData(data []byte, detectFormat bool) GossConfig {
	if TemplateFilter != nil {
		data = TemplateFilter(data)
		if debug {
			fmt.Println("DEBUG: file after text/template render")
			fmt.Println(string(data))
		}
	}
	format := OutStoreFormat
	if detectFormat == true {
		format = getStoreFormatFromData(data)
	}
	gossConfig := NewGossConfig()
	// Horrible, but will do for now
	if err := unmarshal(data, gossConfig, format); err != nil {
		// FIXME: really dude.. this is so ugly
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	return *gossConfig
}

// RenderJSON Reads json file recursively returning string
func RenderJSON(c *cli.Context) string {
	filePath := c.GlobalString("gossfile")
	varsFile := c.GlobalString("vars")
	varsInline := c.GlobalString("vars-inline")
	debug = c.Bool("debug")
	TemplateFilter = NewTemplateFilter(varsFile, varsInline)
	path := filepath.Dir(filePath)
	OutStoreFormat = getStoreFormatFromFileName(filePath)
	gossConfig := mergeJSONData(ReadJSON(filePath), 0, path)

	b, err := marshal(gossConfig)
	if err != nil {
		log.Fatalf("Error rendering: %v\n", err)
	}
	return string(b)
}

func mergeJSONData(gossConfig GossConfig, depth int, path string) GossConfig {
	depth++
	if depth >= 50 {
		fmt.Println("Error: Max depth of 50 reached, possibly due to dependency loop in goss file")
		os.Exit(1)
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
		if strings.HasPrefix(g.ID(), "/") {
			fpath = g.ID()
		} else {
			fpath = filepath.Join(path, g.ID())
		}
		matches, err := filepath.Glob(fpath)
		if err != nil {
			fmt.Printf("Error in expanding glob pattern: \"%s\"\n", err.Error())
			os.Exit(1)
		}
		if matches == nil {
			fmt.Printf("No matched files were found: \"%s\"\n", fpath)
			os.Exit(1)
		}
		for _, match := range matches {
			fdir := filepath.Dir(match)
			j := mergeJSONData(ReadJSON(match), depth, fdir)
			ret = mergeGoss(ret, j)
		}
	}
	return ret
}

func WriteJSON(filePath string, gossConfig GossConfig) error {
	jsonData, err := marshal(gossConfig)
	if err != nil {
		log.Fatalf("Error writing: %v\n", err)
	}

	// check if the auto added json data is empty before writing to file.
	emptyConfig := *NewGossConfig()
	emptyData, err := marshal(emptyConfig)
	if err != nil {
		log.Fatalf("Error writing: %v\n", err)
	}

	if string(emptyData) == string(jsonData) {
		log.Printf("Can't write empty configuration file. Please check resource name(s).")
		return nil
	}

	if err := ioutil.WriteFile(filePath, jsonData, 0644); err != nil {
		log.Fatalf("Error writing: %v\n", err)
	}

	return nil
}

func resourcePrint(fileName string, res resource.ResourceRead) {
	resMap := map[string]resource.ResourceRead{res.ID(): res}

	oj, _ := marshal(resMap)
	typ := reflect.TypeOf(res)
	typs := strings.Split(typ.String(), ".")[1]

	fmt.Printf("Adding %s to '%s':\n\n%s\n\n", typs, fileName, string(oj))
}

func marshal(gossConfig interface{}) ([]byte, error) {
	switch OutStoreFormat {
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
	return yaml.Unmarshal(data, v)
}
