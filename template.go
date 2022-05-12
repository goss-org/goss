package goss

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig/v3"
)

// TemplateFilter is the type of the Goss Template Filter which include custom variables and functions.
type TemplateFilter func([]byte) ([]byte, error)

// NewTemplateFilter creates a new Template Filter based in the file and inline variables.
func NewTemplateFilter(varsFile string, varsInline string) (func([]byte) ([]byte, error), error) {
	vars, err := loadVars(varsFile, varsInline)
	if err != nil {
		return nil, fmt.Errorf("failed while loading vars file %q: %v", varsFile, err)
	}

	tVars := &TmplVars{Vars: vars}

	f := func(data []byte) ([]byte, error) {
		t := template.New("test").Funcs(sprig.TxtFuncMap()).Funcs(funcMap)

		tmpl, err := t.Parse(string(data))
		if err != nil {
			return []byte{}, err
		}

		tmpl.Option("missingkey=error")
		var doc bytes.Buffer

		err = tmpl.Execute(&doc, tVars)
		if err != nil {
			return []byte{}, err
		}

		return doc.Bytes(), nil
	}

	return f, nil
}

func mkSlice(args ...interface{}) []interface{} {
	return args
}

func readFile(f string) (string, error) {
	b, err := ioutil.ReadFile(f)
	if err != nil {
		return "", err

	}
	return strings.TrimSpace(string(b)), nil
}

func getEnv(key string, def ...string) string {
	val := os.Getenv(key)
	if val == "" && len(def) > 0 {
		return def[0]
	}

	return os.Getenv(key)
}

func regexMatch(re, s string) (bool, error) {
	compiled, err := regexp.Compile(re)
	if err != nil {
		return false, err
	}

	return compiled.MatchString(s), nil
}

var funcMap = template.FuncMap{
	"mkSlice":    mkSlice,
	"readFile":   readFile,
	"getEnv":     getEnv,
	"regexMatch": regexMatch,
	"toUpper":    strings.ToUpper,
	"toLower":    strings.ToLower,
}
