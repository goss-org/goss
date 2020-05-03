package matchers

import (
	"fmt"
	"io"
	"io/ioutil"
	"strconv"
	"strings"

	"github.com/onsi/gomega/format"
)

func ToFloat64(e interface{}) (float64, error) {
	switch v := e.(type) {
	case float64:
		return v, nil
	case int:
		return float64(v), nil
	case string:
		return strconv.ParseFloat(strings.TrimSpace(v), 64)
	case []string:
		return strconv.ParseFloat(strings.TrimSpace(ToString(v)), 64)
	default:
		return 0, fmt.Errorf("Expected numeric, Got:%s", format.Object(e, 1))
		//return 0, fmt.Errorf("expected numeric, got: %v", e)

	}
}

func ToInt(e interface{}) (int, error) {
	v, err := ToFloat64(e)
	if err != nil {
		return 0, err
	}
	return int(v), nil
}

func ToString(e interface{}) string {
	switch v := e.(type) {
	case []string:
		return strings.Join(v, "\n")
	default:
		return fmt.Sprintf("%v", v)
	}
}

func ReaderToString(r io.Reader) (string, error) {
	b, err := ioutil.ReadAll(r)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

//func ReaderToStrings(r io.Reader) ([]string, error) {
//	var lines []string
//	scanner := bufio.NewScanner(r)
//	for scanner.Scan() {
//		lines = append(lines, scanner.Text())
//	}
//	return lines, scanner.Err()
//}
func ReaderToStrings(r io.Reader) ([]string, error) {
	var lines []string
	s, err := ReaderToString(r)
	if err != nil {
		return lines, err
	}
	return strings.Split(s, "\n"), nil
}
