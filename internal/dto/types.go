package dto

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
