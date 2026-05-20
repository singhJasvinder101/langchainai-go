package types

import (
	"fmt"
	"strconv"
	"strings"
)

type Config map[string]any

func (c Config) getPath(path string) (any, bool) {
	if c == nil {
		return nil, false
	}
	parts := strings.Split(path, ".")
	var cur any = map[string]any(c)
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		switch m := cur.(type) {
		case Config:
			next, ok := m[part]
			if !ok {
				return nil, false
			}
			cur = next
		case map[string]any:
			next, ok := m[part]
			if !ok {
				return nil, false
			}
			cur = next
		case map[any]any:
			next, ok := m[part]
			if !ok {
				return nil, false
			}
			cur = next
		default:
			return nil, false
		}
	}
	return cur, true
}

func (c Config) Get(path string) any {
	v, _ := c.getPath(path)
	return v
}

func (c Config) GetString(path string) string {
	v, ok := c.getPath(path)
	if !ok || v == nil {
		return ""
	}
	switch s := v.(type) {
	case string:
		return s
	default:
		return fmt.Sprint(v)
	}
}

func (c Config) GetInt(path string) int {
	v, ok := c.getPath(path)
	if !ok || v == nil {
		return 0
	}
	switch n := v.(type) {
	case int:
		return n
	case int64:
		return int(n)
	case float64:
		return int(n)
	case string:
		i, _ := strconv.Atoi(n)
		return i
	default:
		return 0
	}
}

func (c Config) GetBool(path string) bool {
	v, ok := c.getPath(path)
	if !ok || v == nil {
		return false
	}
	switch b := v.(type) {
	case bool:
		return b
	case string:
		x, err := strconv.ParseBool(b)
		return err == nil && x
	default:
		return false
	}
}
