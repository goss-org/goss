package util

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/oleiade/reflections"
)

type Config struct {
	IgnoreList        []string
	RequestHeader     []string
	Timeout           int
	AllowInsecure     bool
	NoFollowRedirects bool
	Server            string
	Username          string
	Password          string
}

type OutputConfig struct {
	FormatOptions []string
}

type format string

const (
	JSON format = "json"
	YAML format = "yaml"
)

func ValidateSections(unmarshal func(interface{}) error, i interface{}, whitelist map[string]bool) error {
	// Get generic input
	var toValidate map[string]map[string]interface{}
	if err := unmarshal(&toValidate); err != nil {
		return err
	}

	// Run input through whitelist
	typ := reflect.TypeOf(i)
	typs := strings.Split(typ.String(), ".")[1]
	for id, v := range toValidate {
		for k, _ := range v {
			if !whitelist[k] {
				return fmt.Errorf("Invalid Attribute for %s:%s: %s", typs, id, k)
			}
		}
	}

	return nil
}

func WhitelistAttrs(i interface{}, format format) (map[string]bool, error) {
	validAttrs := make(map[string]bool)
	tags, err := reflections.Tags(i, string(format))
	if err != nil {
		return nil, err
	}
	for _, v := range tags {
		validAttrs[strings.Split(v, ",")[0]] = true
	}
	return validAttrs, nil
}

func IsValueInList(value string, list []string) bool {
	for _, v := range list {
		if strings.ToLower(v) == strings.ToLower(value) {
			return true
		}
	}
	return false
}
