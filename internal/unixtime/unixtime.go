package unixtime

import (
	"encoding/json"
	"time"
)

// UnixTime is a time.Time that is serialized as a Unix timestamp.
type UnixTime time.Time

// UnmarshalJSON implements the json.Unmarshaler interface.
func (t *UnixTime) UnmarshalJSON(b []byte) error {
	var n int64
	if err := json.Unmarshal(b, &n); err != nil {
		return err
	}
	*t = UnixTime(time.Unix(n, 0))
	return nil
}

// MarshalJSON implements the json.Marshaler interface.
func (t UnixTime) MarshalJSON() ([]byte, error) {
	return json.Marshal(time.Time(t).Unix())
}
