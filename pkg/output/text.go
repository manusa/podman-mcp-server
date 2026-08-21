package output

import "fmt"

type TextOutput struct{}

func (t *TextOutput) Format(data interface{}) (string, error) {
	if s, ok := data.(string); ok {
		return s, nil
	}
	return fmt.Sprintf("%v", data), nil
}
