package protox

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

func (x Timestamp) MarshalJSON() ([]byte, error) {
	return json.Marshal(x.Seconds)
}

func (x *Timestamp) UnmarshalJSON(content []byte) error {
	var n uint32
	if err := json.Unmarshal(content, &n); err == nil {
		x.Seconds = n
		return nil
	}
	var v any
	if err := json.Unmarshal(content, &v); err != nil {
		return err
	}
	switch value := v.(type) {
	case string:
		t, err := time.Parse(time.DateTime, value)
		if err != nil {
			return err
		}
		x.Seconds = uint32(t.Unix())
		x.Nanos = uint32(t.Nanosecond())
	default:
		return errors.New("invalid timestamp")
	}
	return nil
}

func (x *Timestamp) Scan(src any) error {
	switch v := src.(type) {
	case time.Time:
		x.Seconds = uint32(v.Unix())
		x.Nanos = uint32(v.Nanosecond())
	case int64:
		x.Seconds = uint32(v)
	case string:
		t, err := time.Parse(time.DateTime, v)
		if err != nil {
			return err
		}
		x.Seconds = uint32(t.Unix())
		x.Nanos = uint32(t.Nanosecond())
	case []byte:
		t, err := time.Parse(time.DateTime, string(v))
		if err != nil {
			return err
		}
		x.Seconds = uint32(t.Unix())
		x.Nanos = uint32(t.Nanosecond())
	default:
		return fmt.Errorf("unknown value type %d", src)
	}
	return nil
}

func (x Timestamp) Value() (driver.Value, error) {
	return time.Unix(int64(x.Seconds), int64(x.Nanos)), nil
}

func (x *Timestamp) Time() time.Time {
	if x == nil {
		return time.Unix(0, 0)
	}
	return time.Unix(int64(x.Seconds), int64(x.Nanos))
}

func TimestampFromTime(t time.Time) *Timestamp {
	return &Timestamp{
		Seconds: uint32(t.Unix()),
		Nanos:   uint32(t.Nanosecond()),
	}
}
