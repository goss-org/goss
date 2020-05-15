package matchers

import (
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

type ToFloat64 struct{}

func (t ToFloat64) Transform(e interface{}) (interface{}, error) {
	switch v := e.(type) {
	case float64:
		return v, nil
	case int:
		return float64(v), nil
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

type ToString struct{}

func (t ToString) Transform(e interface{}) (interface{}, error) {
	switch v := e.(type) {
	case []string:
		return strings.Join(v, "\n"), nil
	default:
		return fmt.Sprintf("%v", v), nil
	}
}

type ToArray struct{}

func (t ToArray) Transform(i interface{}) (interface{}, error) {
	switch v := i.(type) {
	case string:
		return strings.Split(v, "\n"), nil
	default:
		return i, nil
	}
	//	if !ok {
	//		return nil, fmt.Errorf("Expected io.reader, Got:%s", format.Object(i, 1))
	//	}
	//	var lines []string
	//	i, err := ReaderToString{}.Transform(r)
	//	if err != nil {
	//		return lines, err
	//	}
	//	s := i.(string)
	//return strings.Split(s, "\n"), nil
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

type GJson struct {
	Path string
}

func (g GJson) Transform(i interface{}) (interface{}, error) {
	s, ok := i.(string)
	if !ok {
		return nil, fmt.Errorf("Expected string, Got:%s", format.Object(i, 1))
	}
	r := gjson.Get(s, g.Path)
	//if !r.Exists() {
	//	return nil, fmt.Errorf("gjson failed to find value at %s", g.Path)
	//}

	return r.Value(), nil
}
