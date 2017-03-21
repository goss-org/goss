package goss

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"text/template"
)

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

var funcMap = map[string]interface{}{
	"mkSlice":  mkSlice,
	"readFile": readFile,
	"getEnv":   getEnv,
}

func NewTemplateFilter(varsFile string) func([]byte) []byte {
	vars, err := varsFromFile(varsFile)
	if err != nil {
		fmt.Printf("Error: loading vars file '%s'\n%v\n", varsFile, err)
		os.Exit(1)
	}
	tVars := &TmplVars{Vars: vars}

	f := func(data []byte) []byte {
		funcMap := funcMap
		t := template.New("test").Funcs(template.FuncMap(funcMap))
		tmpl, err := t.Parse(string(data))
		if err != nil {
			log.Fatal(err)
		}
		tmpl.Option("missingkey=error")
		var doc bytes.Buffer
		err = tmpl.Execute(&doc, tVars)
		if err != nil {
			log.Fatal(err)
		}
		return doc.Bytes()
	}
	return f
}
