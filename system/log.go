package system

import (
	"bytes"
	"log"
)

func logBytes(b []byte, prefix string) {
	if len(b) == 0 {
		return
	}
	lines := bytes.SplitSeq(b, []byte("\n"))
	for l := range lines {
		log.Printf("[DEBUG]%s %s", prefix, l)
	}
}
