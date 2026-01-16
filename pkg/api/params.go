package api

import "fmt"

// GetString extracts a string parameter, returns default if not present.
func (p *ToolHandlerParams) GetString(key, defaultValue string) string {
	if value, ok := p.Arguments[key]; ok {
		if strValue, ok := value.(string); ok {
			return strValue
		}
	}
	return defaultValue
}

// RequiredString extracts a required string parameter.
func (p *ToolHandlerParams) RequiredString(key string) (string, error) {
	val, ok := p.Arguments[key]
	if !ok {
		return "", fmt.Errorf("%s parameter required", key)
	}
	str, ok := val.(string)
	if !ok {
		return "", fmt.Errorf("%s parameter must be a string", key)
	}
	return str, nil
}

// GetStringArray extracts a string array parameter.
func (p *ToolHandlerParams) GetStringArray(key string) []string {
	val, ok := p.Arguments[key]
	if !ok {
		return nil
	}
	arr, ok := val.([]interface{})
	if !ok {
		return nil
	}
	result := make([]string, 0, len(arr))
	for _, item := range arr {
		if str, ok := item.(string); ok {
			result = append(result, str)
		}
	}
	return result
}

// GetPortMappings extracts port mappings from a string array (format: "hostPort:containerPort").
func (p *ToolHandlerParams) GetPortMappings(key string) map[int]int {
	arr := p.GetStringArray(key)
	if arr == nil {
		return nil
	}
	result := make(map[int]int)
	for _, mapping := range arr {
		var hostPort, containerPort int
		if _, err := fmt.Sscanf(mapping, "%d:%d", &hostPort, &containerPort); err == nil {
			if hostPort > 0 && containerPort > 0 {
				result[hostPort] = containerPort
			}
		}
	}
	return result
}
