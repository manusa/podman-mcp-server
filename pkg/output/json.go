package output

import "encoding/json"

type JSONOutput struct{}

func (j *JSONOutput) Format(data interface{}) (string, error) {
	if s, ok := data.(string); ok {
		return s, nil
	}
	bytes, err := json.Marshal(data)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}
