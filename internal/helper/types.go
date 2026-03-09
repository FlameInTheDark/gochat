package helper

import (
	"encoding/json"
	"fmt"
	"strconv"
)

type StringInt64Array []int64

func (a *StringInt64Array) UnmarshalJSON(b []byte) error {
	var raw []interface{}
	if err := json.Unmarshal(b, &raw); err != nil {
		return err
	}
	res := make([]int64, len(raw))
	for i, v := range raw {
		switch val := v.(type) {
		case string:
			parsed, err := strconv.ParseInt(val, 10, 64)
			if err != nil {
				return fmt.Errorf("invalid int64 at string index %d", i)
			}
			res[i] = parsed
		case float64:
			res[i] = int64(val)
		default:
			return fmt.Errorf("invalid int64 at index %d", i)
		}
	}
	*a = res
	return nil
}

type StringInt64 int64

func (s *StringInt64) UnmarshalJSON(b []byte) error {
	var raw interface{}
	if err := json.Unmarshal(b, &raw); err != nil {
		return err
	}
	switch val := raw.(type) {
	case string:
		parsed, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid int64: %w", err)
		}
		*s = StringInt64(parsed)
	case float64:
		*s = StringInt64(int64(val))
	default:
		return fmt.Errorf("invalid int64 type")
	}
	return nil
}
