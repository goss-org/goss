package matchers

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"unsafe"
)

//var json = jsoniter.ConfigCompatibleWithStandardLibrary

type GossMatcher interface {
	Match(actual interface{}) (success bool, err error)
	FailureResult(actual interface{}) MatcherResult
	NegatedFailureResult(actual interface{}) MatcherResult
	fmt.Stringer
}

type MatcherResult struct {
	Actual           interface{}
	Message          string
	Expected         interface{}
	MissingElements  interface{}
	ExtraElements    interface{}
	TransformerChain []Transformer
	Successful       bool
}

//func (m MatcherResult) String() string {
//	if m.Successful {
//		return fmt.Sprintf("matches expectation: %s", m.Expected)
//	}
//	return format.Message(m.Actual, m.Message, m.Expected)
//}

func (m MatcherResult) String() string {
	var s string
	if m.Successful {
		return fmt.Sprintf("%s: %s", m.Message, prettyPrint(m.Expected, false))
	}
	s += fmt.Sprintf("Expected\n%s\n%s\n%s", prettyPrint(m.Actual, true), m.Message, prettyPrint(m.Expected, true))
	if m.MissingElements != nil {
		s += fmt.Sprintf("\nthe missing elements were\n%s", prettyPrint(m.MissingElements, true))
	}
	if m.ExtraElements != nil {
		s += fmt.Sprintf("\nthe extra elements were\n%s", prettyPrint(m.MissingElements, true))
	}
	return s
}

func prettyPrint(i interface{}, indent bool) string {
	//b, err := json.MarshalIndent(ConvertMapI2MapS(i), "", "    ")
	//b, _ := json.MarshalIndent(i, "", "  ")
	//b, _ := json.Marshal(i)
	// fixme: error handling
	b, err := json.Marshal(i)
	if err != nil {
		b = []byte(fmt.Sprint(err))
	}
	if indent {
		return indentLines(string(b))
	} else {
		return string(b)
	}
}

// indents a block of text with an indent string
func indentLines(text string) string {
	indent := "    "
	result := ""
	for _, j := range strings.Split(strings.TrimRight(text, "\n"), "\n") {
		result += indent + j + "\n"
	}
	return result[:len(result)-1]
}

func getUnexported(i interface{}, field string) interface{} {
	rs := reflect.ValueOf(i).Elem()
	rf := rs.FieldByName(field)
	rf = reflect.NewAt(rf.Type(), unsafe.Pointer(rf.UnsafeAddr())).Elem()
	return rf.Interface()
}

// ConvertMapI2MapS walks the given dynamic object recursively, and
// converts maps with interface{} key type to maps with string key type.
// This function comes handy if you want to marshal a dynamic object into
// JSON where maps with interface{} key type are not allowed.
//
// Recursion is implemented into values of the following types:
//   -map[interface{}]interface{}
//   -map[string]interface{}
//   -[]interface{}
//
// When converting map[interface{}]interface{} to map[string]interface{},
// fmt.Sprint() with default formatting is used to convert the key to a string key.
func ConvertMapI2MapS(v interface{}) interface{} {
	switch x := v.(type) {
	case map[interface{}]interface{}:
		m := map[string]interface{}{}
		for k, v2 := range x {
			switch k2 := k.(type) {
			case string: // Fast check if it's already a string
				m[k2] = ConvertMapI2MapS(v2)
			default:
				m[fmt.Sprint(k)] = ConvertMapI2MapS(v2)
			}
		}
		v = m

	case []interface{}:
		for i, v2 := range x {
			x[i] = ConvertMapI2MapS(v2)
		}

	case map[string]interface{}:
		for k, v2 := range x {
			x[k] = ConvertMapI2MapS(v2)
		}
	}

	return v
}
