package goss

import (
	"bytes"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"strings"
	"text/template"
)

//TemplateFilter is the type of the Goss Template Filter which include custom variables and functions.
type TemplateFilter func([]byte) ([]byte, error)

//NewTemplateFilter creates a new Template Filter based in the file and inline variables.
func NewTemplateFilter(varsFile string, varsInline string) TemplateFilter {
	vars, err := loadVars(varsFile, varsInline)
	if err != nil {
		log.Fatal(err)
	}

	tVars := &TmplVars{Vars: vars}

	return func(data []byte) ([]byte, error) {
		funcMap := funcMap
		t := template.New("goss").Funcs(funcMap)

		tmpl, err := t.Parse(string(data))
		if err != nil {
			return nil, err
		}

		tmpl.Option("missingkey=error")
		var doc bytes.Buffer

		err = tmpl.Execute(&doc, tVars)
		return doc.Bytes(), err
	}
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
