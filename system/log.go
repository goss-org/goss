package system

import (
	"bytes"
	"log"
)

func logBytes(b []byte, prefix string) {
	if len(b) == 0 {
		return
	}
	lines := bytes.Split(b, []byte("\n"))
	for _, l := range lines {
		log.Printf("[DEBUG] %s %s", prefix, l)
	}
}
