package helper

import (
	"encoding/json"
	"fmt"
	"strconv"
)

type StringInt64Array []int64

func (a *StringInt64Array) UnmarshalJSON(b []byte) error {
	var raw []json.RawMessage
	if err := json.Unmarshal(b, &raw); err != nil {
		return err
	}
	res := make([]int64, len(raw))
	for i, v := range raw {
		var str string
		if json.Unmarshal(v, &str) == nil {
			parsed, err := strconv.ParseInt(str, 10, 64)
			if err != nil {
				return fmt.Errorf("invalid int64 at string index %d", i)
			}
			res[i] = parsed
			continue
		}
		var num json.Number
		if err := json.Unmarshal(v, &num); err != nil {
			return fmt.Errorf("invalid int64 at index %d", i)
		}
		parsed, err := num.Int64()
		if err != nil {
			return fmt.Errorf("invalid int64 at index %d: %w", i, err)
		}
		res[i] = parsed
	}
	*a = res
	return nil
}

type StringInt64 int64

func (s *StringInt64) UnmarshalJSON(b []byte) error {
	var str string
	if json.Unmarshal(b, &str) == nil {
		parsed, err := strconv.ParseInt(str, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid int64: %w", err)
		}
		*s = StringInt64(parsed)
		return nil
	}
	var num json.Number
	if err := json.Unmarshal(b, &num); err != nil {
		return fmt.Errorf("invalid int64 type")
	}
	parsed, err := num.Int64()
	if err != nil {
		return fmt.Errorf("invalid int64: %w", err)
	}
	*s = StringInt64(parsed)
	return nil
}

func (s StringInt64) MarshalJSON() ([]byte, error) {
	return []byte(`"` + strconv.FormatInt(int64(s), 10) + `"`), nil
}
