package matchers

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"strconv"
	"strings"

	"github.com/onsi/gomega/format"
	"github.com/tidwall/gjson"
)

type Transformer interface {
	Transform(interface{}) (interface{}, error)
}

type ToNumeric struct{}

func (t ToNumeric) Transform(e interface{}) (interface{}, error) {
	switch v := e.(type) {
	case float64, int:
		return v, nil
	case string:
		return strconv.ParseFloat(strings.TrimSpace(v), 64)
	case []string:
		i, err := ToString{}.Transform(v)
		if err != nil {
			return 0, err
		}
		s := i.(string)
		return strconv.ParseFloat(strings.TrimSpace(s), 64)
	default:
		return 0, fmt.Errorf("Expected numeric, Got:%s", format.Object(e, 1))

	}
}
func (t ToNumeric) MarshalJSON() ([]byte, error) {
	j := map[string]interface{}{
		"to-numeric": map[string]string{},
	}
	return json.Marshal(j)
}

type ToString struct{}

func (t ToString) Transform(e interface{}) (interface{}, error) {
	switch v := e.(type) {
	case []string:
		return strings.Join(v, "\n"), nil
	default:
		return fmt.Sprintf("%v", v), nil
	}
}

func (t ToString) MarshalJSON() ([]byte, error) {
	j := map[string]interface{}{
		"to-string": map[string]string{},
	}
	return json.Marshal(j)
}

type ToArray struct{}

func (t ToArray) Transform(i interface{}) (interface{}, error) {
	switch v := i.(type) {
	case string:
		return strings.Split(v, "\n"), nil
	default:
		return i, nil
	}
}
func (matcher ToArray) MarshalJSON() ([]byte, error) {
	j := map[string]interface{}{
		"to-array": map[string]string{},
	}
	return json.Marshal(j)
}

type ReaderToStrings struct{}

func (t ReaderToStrings) Transform(i interface{}) (interface{}, error) {
	r, ok := i.(io.Reader)
	if !ok {
		return nil, fmt.Errorf("Expected io.reader, Got:%s", format.Object(i, 1))
	}
	var lines []string
	i, err := ReaderToString{}.Transform(r)
	if err != nil {
		return lines, err
	}
	s := i.(string)
	return strings.Split(s, "\n"), nil
}

type ReaderToString struct{}

func (t ReaderToString) Transform(i interface{}) (interface{}, error) {
	r, ok := i.(io.Reader)
	if !ok {
		return nil, fmt.Errorf("Expected io.reader, Got:%s", format.Object(i, 1))
	}

	b, err := ioutil.ReadAll(r)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

type Gjson struct {
	Path string
}

func (g Gjson) Transform(i interface{}) (interface{}, error) {
	s, ok := i.(string)
	if !ok {
		return nil, fmt.Errorf("Expected string, Got:%s", format.Object(i, 1))
	}
	if !gjson.Valid(s) {
		return nil, fmt.Errorf("Invalid json")
	}
	r := gjson.Get(s, g.Path)
	if !r.Exists() {
		return nil, fmt.Errorf("Path not found: %s", g.Path)
	}

	return r.Value(), nil
}
func (g Gjson) MarshalJSON() ([]byte, error) {
	j := map[string]interface{}{
		"gjson": map[string]string{
			"Path": g.Path,
		},
	}
	return json.Marshal(j)
}
