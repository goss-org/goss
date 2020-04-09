package matchers

import (
	"fmt"
	"strconv"

	"github.com/onsi/gomega/format"
)

func ToFloat64(e interface{}) (float64, error) {
	switch v := e.(type) {
	case float64:
		return e.(float64), nil
	case int:
		return float64(v), nil
	case string:
		return strconv.ParseFloat(e.(string), 64)
	default:
		return 0, fmt.Errorf("Expected numeric, Got:%s", format.Object(e, 1))
		//return 0, fmt.Errorf("expected numeric, got: %v", e)

	}
}

func ToString(e interface{}) string { return fmt.Sprintf("%v", e) }
