package loghose

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/christian-blades-cb/gojsonexplode"
)

type LoghoseLine struct {
	Version       int                    `json:"v"`
	ContainerId   string                 `json:"id"`
	Image         string                 `json:"image"`
	ContainerName string                 `json:"name"`
	Logline       map[string]interface{} `json:"line"`
}

func Parse(data []byte) (*LoghoseLine, error) {
	var ll LoghoseLine

	err := json.Unmarshal(data, &ll)
	if err != nil {
		return nil, err
	}

	// flatten logline, nesting is stupid
	line, err := gojsonexplode.ExplodeMap(ll.Logline, "line", ".")
	if err != nil {
		return nil, err
	}
	ll.Logline = line

	return &ll, nil
}

func (ll *LoghoseLine) LogfmtLine() string {
	line := bytes.NewBuffer(nil)
	for k, v := range ll.Logline {
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
