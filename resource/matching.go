package resource

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/goss-org/goss/system"
	"github.com/goss-org/goss/util"
)

type Matching struct {
	Title   string  `json:"title,omitempty" yaml:"title,omitempty"`
	Meta    meta    `json:"meta,omitempty" yaml:"meta,omitempty"`
	Content any     `json:"content,omitempty" yaml:"content,omitempty"`
	Id      string  `json:"-" yaml:"-"`
	Matches matcher `json:"matches" yaml:"matches"`
}

const (
	MatchingResourceKey  = "mount"
	MatchingResourceName = "Mount"
)

type MatchingMap map[string]*Matching

func (a *Matching) ID() string       { return a.Id }
func (a *Matching) SetID(id string)  { a.Id = id }
func (a *Matching) SetSkip()         {}
func (a *Matching) TypeKey() string  { return MatchingResourceKey }
func (a *Matching) TypeName() string { return MatchingResourceName }

// FIXME: Can this be refactored?
func (r *Matching) GetTitle() string { return r.Title }
func (r *Matching) GetMeta() meta    { return r.Meta }

func (a *Matching) Validate(sys *system.System) []TestResult {
	skip := false

	// ValidateValue expects a function
	stub := func() (any, error) {
		return a.Content, nil
	}

	var results []TestResult
	results = append(results, ValidateValue(a, "matches", a.Matches, stub, skip))
	return results
}

func (ret *MatchingMap) UnmarshalJSON(data []byte) error {
	// Curried json.Unmarshal
	unmarshal := func(i any) error {
		if err := json.Unmarshal(data, i); err != nil {
			return err
		}
		return nil
	}

	// Validate configuration
	zero := Matching{}
	whitelist, err := util.WhitelistAttrs(zero, util.JSON)
	if err != nil {
		return err
	}
	if err := util.ValidateSections(unmarshal, zero, whitelist); err != nil {
		return err
	}

	var tmp map[string]*Matching
	if err := unmarshal(&tmp); err != nil {
		return err
	}

	typ := reflect.TypeOf(zero)
	typs := strings.Split(typ.String(), ".")[1]
	for id, res := range tmp {
		if res == nil {
			return fmt.Errorf("Could not parse resource %s:%s", typs, id)
		}
		res.SetID(id)
	}

	*ret = tmp
	return nil
}

func (ret *MatchingMap) UnmarshalYAML(unmarshal func(v any) error) error {
	// Validate configuration
	zero := Matching{}
	whitelist, err := util.WhitelistAttrs(zero, util.YAML)
	if err != nil {
		return err
	}
	if err := util.ValidateSections(unmarshal, zero, whitelist); err != nil {
		return err
	}

	var tmp map[string]*Matching
	if err := unmarshal(&tmp); err != nil {
		return err
	}

	typ := reflect.TypeOf(zero)
	typs := strings.Split(typ.String(), ".")[1]
	for id, res := range tmp {
		if res == nil {
			return fmt.Errorf("Could not parse resource %s:%s", typs, id)
		}
		res.SetID(id)
	}

	*ret = tmp
	return nil
}
