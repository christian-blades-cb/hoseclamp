package main

import (
	"bytes"
	"fmt"
)

func logfmtMap(structuredLog map[string]interface{}) string {
	line := bytes.NewBuffer(nil)
	for k, v := range structuredLog {
		line.WriteString(logfmtPair(k, v))
	}

	return line.String()
}

func logfmtPair(key, value interface{}) string {
	switch value.(type) {
	case string:
		if needsQuoting(value.(string)) {
			return fmt.Sprintf("%v=%q ", key, value)
		} else {
			return fmt.Sprintf("%v=%s ", key, value)
		}
	default:
		return fmt.Sprintf("%v=%v ", key, value)
	}
}

// lifted from logrus. thanks guys!
func needsQuoting(text string) bool {
	for _, ch := range text {
		if !((ch >= 'a' && ch <= 'z') ||
			(ch >= 'A' && ch <= 'Z') ||
			(ch >= '0' && ch <= '9') ||
			ch == '-' || ch == '.') {
			return true
		}
	}
	return false
}
