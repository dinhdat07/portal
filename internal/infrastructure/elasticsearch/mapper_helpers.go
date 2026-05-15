package elasticsearch

import (
	"fmt"
	"time"
)

func mapFromMap(m map[string]any, key string) map[string]any {
	v, ok := m[key]
	if !ok || v == nil {
		return nil
	}

	result, ok := v.(map[string]any)
	if !ok {
		return nil
	}

	return result
}

func stringFromMap(m map[string]any, key string) string {
	v, ok := m[key]
	if !ok || v == nil {
		return ""
	}

	s, ok := v.(string)
	if !ok {
		return ""
	}

	return s
}

func stringPtrFromMap(m map[string]any, key string) *string {
	v, ok := m[key]
	if !ok || v == nil {
		return nil
	}

	s, ok := v.(string)
	if !ok || s == "" {
		return nil
	}

	return &s
}

func dateFromDebeziumDays(v any) (*string, error) {
	if v == nil {
		return nil, nil
	}

	var days int64

	switch n := v.(type) {
	case float64:
		days = int64(n)
	case int:
		days = int64(n)
	case int64:
		days = n
	default:
		return nil, fmt.Errorf("unsupported debezium date type: %T", v)
	}

	t := time.Unix(0, 0).UTC().AddDate(0, 0, int(days))
	s := t.Format("2006-01-02")

	return &s, nil
}
