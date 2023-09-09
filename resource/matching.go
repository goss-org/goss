package resource

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"strings"

	"github.com/goss-org/goss/system"
	"github.com/goss-org/goss/util"
)

type Matching struct {
	Title    string  `json:"title,omitempty" yaml:"title,omitempty"`
	Meta     meta    `json:"meta,omitempty" yaml:"meta,omitempty"`
	Content  any     `json:"content,omitempty" yaml:"content,omitempty"`
	AsReader bool    `json:"as-reader,omitempty" yaml:"as-reader,omitempty"`
	id       string  `json:"-" yaml:"-"`
	Matches  matcher `json:"matches" yaml:"matches"`
	Skip     bool    `json:"skip,omitempty" yaml:"skip,omitempty"`
}

const (
	MatchingResourceKey  = "mount"
	MatchingResourceName = "Mount"
)

type MatchingMap map[string]*Matching

func (a *Matching) ID() string       { return a.id }
func (a *Matching) SetID(id string)  { a.id = id }
func (a *Matching) SetSkip()         {}
func (a *Matching) TypeKey() string  { return MatchingResourceKey }
func (a *Matching) TypeName() string { return MatchingResourceName }

// FIXME: Can this be refactored?
func (r *Matching) GetTitle() string { return r.Title }
func (r *Matching) GetMeta() meta    { return r.Meta }

func (a *Matching) Validate(sys *system.System) []TestResult {
	ctx := context.Background()
	skip := false
	if a.Skip {
		skip = true
	}

	var stub interface{}
	if a.AsReader {
		s := fmt.Sprintf("%v", a.Content)
		// ValidateValue expects a function
		stub = func(_ context.Context) (io.Reader, error) {
			return strings.NewReader(s), nil
		}
	} else {
		// ValidateValue expects a function
		stub = func(_ context.Context) (any, error) {
			return a.Content, nil
		}
	}

	var results []TestResult
	results = append(results, ValidateValue(ctx, a, "matches", a.Matches, stub, skip))
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
