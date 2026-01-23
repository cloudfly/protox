package protox

import (
	"database/sql/driver"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/v2/bson"
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

func (x Timestamp) MarshalBSON() ([]byte, error) {
	return bson.Marshal(primitive.Timestamp{
		T: uint32(x.Millis / 1000),
		I: uint32(x.Millis % 1000), // I is the increment. You may need to provide a meaningful value.
	})
}

func (x *Timestamp) UnmarshalBSON(data []byte) error {
	var ts primitive.Timestamp
	if err := bson.Unmarshal(data, &ts); err != nil {
		return err
	}
	*x = Timestamp{
		Millis: uint64(ts.T)*1000 + uint64(ts.I),
	}
	return nil
}
