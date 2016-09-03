package goss

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"gopkg.in/yaml.v2"

	"github.com/aelsabbahy/goss/resource"
)

const (
	JSON = iota
	YAML
	UNSET
)

var StoreFormat = UNSET

func setStoreFormatFromFileName(f string) {
	ext := filepath.Ext(f)
	switch ext {
	case ".json":
		StoreFormat = JSON
	case ".yaml", ".yml":
		StoreFormat = YAML
	default:
		log.Fatalf("Unknown file extension: %v", ext)
	}
}

func setStoreFormatFromData(data []byte) {
	var v interface{}
	if err := unmarshalJSON(data, &v); err == nil {
		StoreFormat = JSON
		return
	}
	if err := unmarshalYAML(data, &v); err == nil {
		StoreFormat = YAML
		return
	}
	log.Fatalf("Unable to determine format from content")
}

// Reads json file returning GossConfig
func ReadJSON(filePath string) GossConfig {
	// FIXME: Any problems with this?
	setStoreFormatFromFileName(filePath)
	file, err := ioutil.ReadFile(filePath)
	if err != nil {
		fmt.Printf("File error: %v\n", err)
		os.Exit(1)
	}

	return ReadJSONData(file)
}

// Reads json byte array returning GossConfig
func ReadJSONData(data []byte) GossConfig {
	setStoreFormatFromData(data)
	gossConfig := NewGossConfig()
	// Horrible, but will do for now
	if err := unmarshal(data, gossConfig); err != nil {
		// FIXME: really dude.. this is so ugly
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	return *gossConfig
}

// Reads json file recursively returning string
func RenderJSON(filePath string) string {
	path := filepath.Dir(filePath)
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

	for _, g := range gossConfig.Gossfiles {
		var fpath string
		if strings.HasPrefix(g.ID(), "/") {
			fpath = g.ID()
		} else {
			fpath = filepath.Join(path, g.ID())
		}
		fdir := filepath.Dir(fpath)
		j := mergeJSONData(ReadJSON(fpath), depth, fdir)
		gossConfig = mergeGoss(gossConfig, j)
	}
	return gossConfig
}

func WriteJSON(filePath string, gossConfig GossConfig) error {
	jsonData, err := marshal(gossConfig)
	if err != nil {
		log.Fatalf("Error writing: %v\n", err)
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
	switch StoreFormat {
	case JSON:
		return marshalJSON(gossConfig)
	case YAML:
		return marshalYAML(gossConfig)
	default:
		return nil, fmt.Errorf("StoreFormat unset")
	}
}

func unmarshal(data []byte, v interface{}) error {
	switch StoreFormat {
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
