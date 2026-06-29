package resource

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

const defaultRetryDelay = time.Second

// RetryDelay interprets bare numeric retry_delay values as milliseconds while
// also allowing duration strings like 500ms, 2s, or 1m.
type RetryDelay time.Duration

func (d RetryDelay) Duration() time.Duration {
	return time.Duration(d)
}

func (d RetryDelay) IsZero() bool {
	return d.Duration() == 0
}

func (d RetryDelay) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.marshalValue())
}

func (d RetryDelay) MarshalYAML() (any, error) {
	return d.marshalValue(), nil
}

func (d *RetryDelay) UnmarshalJSON(data []byte) error {
	var value any
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}

	delay, err := parseRetryDelay(value)
	if err != nil {
		return err
	}

	*d = RetryDelay(delay)
	return nil
}

func (d *RetryDelay) UnmarshalYAML(value *yaml.Node) error {
	var raw any
	if err := value.Decode(&raw); err != nil {
		return err
	}

	delay, err := parseRetryDelay(raw)
	if err != nil {
		return err
	}

	*d = RetryDelay(delay)
	return nil
}

func (d RetryDelay) marshalValue() any {
	delay := d.Duration()
	if delay%time.Millisecond == 0 {
		return int64(delay / time.Millisecond)
	}
	return delay.String()
}

func parseRetryDelay(value any) (time.Duration, error) {
	switch v := value.(type) {
	case nil:
		return 0, nil
	case int:
		return millisecondsToDuration(float64(v)), nil
	case int64:
		return millisecondsToDuration(float64(v)), nil
	case float64:
		return millisecondsToDuration(v), nil
	case string:
		return parseRetryDelayString(v)
	default:
		return 0, fmt.Errorf("invalid retry_delay value %v", value)
	}
}

func parseRetryDelayString(value string) (time.Duration, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0, nil
	}

	if delay, err := time.ParseDuration(value); err == nil {
		return delay, nil
	}

	seconds, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid retry_delay value %q", value)
	}

	return millisecondsToDuration(seconds), nil
}

func millisecondsToDuration(milliseconds float64) time.Duration {
	return time.Duration(milliseconds * float64(time.Millisecond))
}

func normalizedRetryDelay(delay RetryDelay) time.Duration {
	if delay.Duration() <= 0 {
		return defaultRetryDelay
	}
	return delay.Duration()
}

func retryAttempts(retryCount int) int {
	if retryCount < 0 {
		return 1
	}
	return retryCount + 1
}

func runWithRetry(retryCount int, retryDelay RetryDelay, validate func() bool) {
	attempts := retryAttempts(retryCount)
	delay := normalizedRetryDelay(retryDelay)

	for attempt := 0; attempt < attempts; attempt++ {
		if validate() {
			return
		}

		if attempt < attempts-1 {
			time.Sleep(delay)
		}
	}
}
