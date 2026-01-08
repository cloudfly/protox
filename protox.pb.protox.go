package protox

import (
	"database/sql/driver"
	"fmt"
	"time"
)

func (x *Timestamp) Scan(src any) error {
	switch v := src.(type) {
	case time.Time:
		x.Millis = uint64(v.UnixNano() / int64(time.Millisecond))
	case int64:
		x.Millis = uint64(v)
	case string:
		t, err := time.Parse(time.DateTime, v)
		if err != nil {
			return err
		}
		x.Millis = uint64(t.UnixNano() / int64(time.Millisecond))
	case []byte:
		t, err := time.Parse(time.DateTime, string(v))
		if err != nil {
			return err
		}
		x.Millis = uint64(t.UnixNano() / int64(time.Millisecond))
	default:
		return fmt.Errorf("unknown value type %d", src)
	}
	return nil
}

func (x Timestamp) Value() (driver.Value, error) {
	return time.Unix(int64(x.Millis), 0), nil
}

func (x *Timestamp) Time() time.Time {
	if x == nil {
		return time.Unix(0, 0)
	}
	return time.UnixMilli(int64(x.Millis))
}
