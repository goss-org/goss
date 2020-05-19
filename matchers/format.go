// Some of the contents of this file is copied/tweaked from gomega/format.go
package matchers

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/onsi/gomega/format"
)

//The default indentation string emitted by the format package
var Indent = "    "

func Object(object interface{}, indentation uint) string {
	indent := strings.Repeat(Indent, int(indentation))
	switch v := object.(type) {
	case fmt.Stringer:
		return indent + fmt.Sprint(v)
		//return indent + formatString(v, indentation)
	default:
		return format.Object(object, indentation)
	}
}

func Message(actual interface{}, message string, expected ...interface{}) string {
	if len(expected) == 0 {
		return fmt.Sprintf("Expected\n%s\n%s", Object(actual, 1), message)
	}
	return fmt.Sprintf("Expected\n%s\n%s\n%s", Object(actual, 1), message, Object(expected[0], 1))
}

func formatString(object interface{}, indentation uint) string {
	if indentation == 1 {
		s := fmt.Sprintf("%s", object)
		components := strings.Split(s, "\n")
		result := ""
		for i, component := range components {
			if i == 0 {
				result += component
			} else {
				result += Indent + component
			}
			if i < len(components)-1 {
				result += "\n"
			}
		}

		return result
	} else {
		return fmt.Sprintf("%q", object)
	}
}

func isArrayOrSlice(a interface{}) bool {
	if a == nil {
		return false
	}
	switch reflect.TypeOf(a).Kind() {
	case reflect.Array, reflect.Slice:
		return true
	default:
		return false
	}
}

func isMap(a interface{}) bool {
	if a == nil {
		return false
	}
	return reflect.TypeOf(a).Kind() == reflect.Map
}
