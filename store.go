package goss

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"sync"

	"dario.cat/mergo"
	"gopkg.in/yaml.v3"

	"github.com/goss-org/goss/resource"
	"github.com/goss-org/goss/util"
)

const (
	UNSET = iota
	JSON
	YAML
)

// These package-level variables hold state derived from the gossfile currently
// being processed. `goss serve` and library consumers can drive the load path
// from more than one goroutine, so every access goes through storeStateMu via
// the helpers below rather than touching the variables directly.
var (
	storeStateMu          sync.RWMutex
	outStoreFormat        = UNSET
	currentTemplateFilter TemplateFilter
	debug                 bool
)

func setStoreFormat(f int) {
	storeStateMu.Lock()
	defer storeStateMu.Unlock()
	outStoreFormat = f
}

// storeFormat atomically reads the package-level store format.
func storeFormat() int {
	storeStateMu.RLock()
	defer storeStateMu.RUnlock()
	return outStoreFormat
}

// setTemplateFilter atomically updates the package-level template filter.
func setTemplateFilter(tf TemplateFilter) {
	storeStateMu.Lock()
	defer storeStateMu.Unlock()
	currentTemplateFilter = tf
}

// templateFilter atomically reads the package-level template filter.
func templateFilter() TemplateFilter {
	storeStateMu.RLock()
	defer storeStateMu.RUnlock()
	return currentTemplateFilter
}

// setDebug atomically updates the package-level debug flag.
func setDebug(d bool) {
	storeStateMu.Lock()
	defer storeStateMu.Unlock()
	debug = d
}

// debugEnabled atomically reads the package-level debug flag.
func debugEnabled() bool {
	storeStateMu.RLock()
	defer storeStateMu.RUnlock()
	return debug
}

var (
	errCannotDetermineFormat = errors.New("unable to determine format from content")
	errMaxDepth              = errors.New("max depth of 50 reached, possibly due to dependency loop in goss file")
	errStoreFormatUnset      = errors.New("StoreFormat unset")
)

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
	var v any
	if err := unmarshalJSON(data, &v); err == nil {
		return JSON, nil
	}
	if err := unmarshalYAML(data, &v); err == nil {
		return YAML, nil
	}

	return 0, errCannotDetermineFormat
}

// ReadJSON Reads json file returning GossConfig
func ReadJSON(filePath string) (GossConfig, error) {
	file, err := os.ReadFile(filePath)
	if err != nil {
		return GossConfig{}, fmt.Errorf("file error: %w", err)
	}

	return ReadJSONData(file, false)
}

type TmplVars struct {
	Vars map[string]any
}

func (t *TmplVars) Env() map[string]string {
	env := make(map[string]string)
	for _, i := range os.Environ() {
		sep := strings.Index(i, "=")
		env[i[0:sep]] = i[sep+1:]
	}
	return env
}

func loadVars(varsFiles []string, varsInline string) (map[string]any, error) {
	mergedVars := map[string]any{}

	// Later defined vars file overwrites values of the previous files
	// in places where non-empty keys overlap.
	for _, varsFile := range varsFiles {
		vars, err := varsFromFile(varsFile)
		if err != nil {
			return nil, fmt.Errorf("loading vars file '%s'\n%w", varsFile, err)
		}
		mergo.Merge(&mergedVars, vars, mergo.WithOverride)
	}

	varsExtra, err := varsFromString(varsInline)
	if err != nil {
		return nil, fmt.Errorf("loading inline vars\n%w", err)
	}

	// Note: This algorithm replaces value under key even if it's nested map
	for k, v := range varsExtra {
		mergedVars[k] = v
	}

	return mergedVars, nil
}

func varsFromFile(varsFile string) (map[string]any, error) {
	vars := make(map[string]any)
	if varsFile == "" {
		return vars, nil
	}
	data, err := os.ReadFile(varsFile)
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

func varsFromString(varsString string) (map[string]any, error) {
	vars := make(map[string]any)
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
	if tf := templateFilter(); tf != nil {
		data, err = tf(data)
		if err != nil {
			return GossConfig{}, err
		}
		if debugEnabled() {
			fmt.Println("DEBUG: file after text/template render")
			fmt.Println(string(data))
		}
	}

	format := storeFormat()
	if detectFormat {
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

// RenderJSON reads json file recursively returning string. Any merge
// warnings accumulated while processing the spec (for example, duplicate
// resource keys across gossfiles) are emitted via c.Log(); RenderJSON is
// the edge layer between the pure config-merging core and the rest of the
// application, so it is the place where warnings become log lines.
func RenderJSON(c *util.Config) (string, error) {
	setDebug(c.Debug)
	tf, err := NewTemplateFilter(c.VarsFiles, c.VarsInline)
	if err != nil {
		return "", err
	}
	setTemplateFilter(tf)

	format, err := getStoreFormatFromFileName(c.Spec)
	if err != nil {
		return "", err
	}
	setStoreFormat(format)

	j, err := ReadJSON(c.Spec)
	if err != nil {
		return "", err
	}

	gossConfig, warnings, err := mergeJSONData(j, 0, filepath.Dir(c.Spec))
	if err != nil {
		return "", err
	}
	logWarnings(c.Log(), warnings)

	b, err := marshal(gossConfig)
	if err != nil {
		return "", fmt.Errorf("rendering failed: %w", err)
	}

	return string(b), nil
}

// mergeJSONData walks the gossfile graph (up to a fixed recursion depth),
// merging the configs it reads along the way. It is intentionally pure with
// respect to logging: any warnings produced during merging are accumulated
// and returned for the caller to emit at the appropriate architectural
// boundary (typically the edge layer that holds a *util.Config).
func mergeJSONData(gossConfig GossConfig, depth int, path string) (GossConfig, []string, error) {
	depth++
	if depth >= 50 {
		return GossConfig{}, nil, errMaxDepth
	}
	// Our return gossConfig
	ret := *NewGossConfig()
	var warnings []string
	ret, w := mergeGoss(ret, gossConfig)
	warnings = append(warnings, w...)

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
			return ret, warnings, fmt.Errorf("error in expanding glob pattern: %w", err)
		}
		if matches == nil {
			return ret, warnings, fmt.Errorf("no matched files were found: %q", fpath)
		}
		for _, match := range matches {
			fdir := filepath.Dir(match)
			j, err := ReadJSON(match)
			if err != nil {
				return GossConfig{}, warnings, fmt.Errorf("could not read json data in %s: %w", match, err)
			}
			var childWarnings []string
			j, childWarnings, err = mergeJSONData(j, depth, fdir)
			warnings = append(warnings, childWarnings...)
			if err != nil {
				return ret, warnings, fmt.Errorf("could not write json data: %w", err)
			}
			var mergeWarnings []string
			ret, mergeWarnings = mergeGoss(ret, j)
			warnings = append(warnings, mergeWarnings...)
		}
	}
	return ret, warnings, nil
}

// WriteJSON marshals gossConfig and writes it to filePath. It is a pure
// function with respect to logging: if the configuration is empty (and
// therefore nothing is written), a human-readable warning string is
// returned alongside a nil error so the caller -- which has a *util.Config
// in scope -- can emit it via c.Log(). An empty warning string means the
// file was written normally.
func WriteJSON(filePath string, gossConfig GossConfig) (string, error) {
	jsonData, err := marshal(gossConfig)
	if err != nil {
		return "", fmt.Errorf("failed to write %s: %w", filePath, err)
	}

	// check if the auto added json data is empty before writing to file.
	emptyConfig := *NewGossConfig()
	emptyData, err := marshal(emptyConfig)
	if err != nil {
		return "", fmt.Errorf("failed to write %s: %w", filePath, err)
	}

	if string(emptyData) == string(jsonData) {
		return "Can't write empty configuration file. Please check resource name(s).", nil
	}

	if err := os.WriteFile(filePath, jsonData, 0644); err != nil {
		return "", fmt.Errorf("failed to write %s: %w", filePath, err)
	}

	return "", nil
}

// logWarnings emits each entry of warnings via logger.Printf, preserving
// the pre-refactor format (warnings already include a "[WARN] " prefix
// where appropriate). It is a no-op for empty or nil slices.
func logWarnings(logger util.Logger, warnings []string) {
	for _, w := range warnings {
		logger.Printf("%s", w)
	}
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

func marshal(gossConfig any) ([]byte, error) {
	switch storeFormat() {
	case JSON:
		return marshalJSON(gossConfig)
	case YAML:
		return marshalYAML(gossConfig)
	default:
		return nil, errStoreFormatUnset
	}
}

func unmarshal(data []byte, v any, storeFormat int) error {
	switch storeFormat {
	case JSON:
		return unmarshalJSON(data, v)
	case YAML:
		return unmarshalYAML(data, v)
	default:
		return errStoreFormatUnset
	}
}

func marshalJSON(gossConfig any) ([]byte, error) {
	return json.MarshalIndent(gossConfig, "", "    ")
}

func unmarshalJSON(data []byte, v any) error {
	return json.Unmarshal(data, v)
}

func marshalYAML(gossConfig any) ([]byte, error) {
	return yaml.Marshal(gossConfig)
}

func unmarshalYAML(data []byte, v any) error {
	return yaml.Unmarshal(data, v)
}
