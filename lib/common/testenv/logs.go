package testenv

import (
	"bytes"
	"encoding/json"
	"strings"
	"sync"

	"github.com/testcontainers/testcontainers-go"
)

// LogBuffer collects JSON logs of a container.
type LogBuffer struct {
	mu     sync.Mutex
	buffer bytes.Buffer
}

func (b *LogBuffer) Accept(log testcontainers.Log) {
	b.mu.Lock()
	defer b.mu.Unlock()
	_, _ = b.buffer.Write(log.Content)
}

func (b *LogBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buffer.String()
}

// Records returns JSON records whose msg equals message and whose attributes contain attrs.
func (b *LogBuffer) Records(message string, attrs map[string]string) []map[string]any {
	var records []map[string]any
	for line := range strings.SplitSeq(b.String(), "\n") {
		var record map[string]any
		if json.Unmarshal([]byte(line), &record) != nil || record["msg"] != message {
			continue
		}
		if matches(record, attrs) {
			records = append(records, record)
		}
	}
	return records
}

// Count returns the number of records matching message and attrs.
func (b *LogBuffer) Count(message string, attrs map[string]string) int {
	return len(b.Records(message, attrs))
}

func matches(record map[string]any, attrs map[string]string) bool {
	for key, value := range attrs {
		if record[key] != value {
			return false
		}
	}
	return true
}
