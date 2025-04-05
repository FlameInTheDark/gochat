package sendpulse_sdk_go

import (
	"fmt"
	"strings"
	"time"
)

type DateTimeType time.Time

const dtFormat = "2006-01-02 15:04:05"

func (d *DateTimeType) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), "\"")
	if s == "null" {
		*d = DateTimeType(time.Time{})
		return nil
	}
	t, err := time.Parse(dtFormat, s)
	if err != nil {
		return err
	}

	*d = DateTimeType(t)
	return nil
}

func (d DateTimeType) MarshalJSON() ([]byte, error) {
	return []byte(d.String()), nil
}

func (d *DateTimeType) String() string {
	t := time.Time(*d)
	return fmt.Sprintf("%q", t.Format(dtFormat))
}
