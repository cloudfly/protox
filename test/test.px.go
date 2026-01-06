package test

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"context"
	"errors"
)

var (
	_ = fmt.Printf
	_ = driver.Bool
	_ = json.Marshal
	_ = context.Background
	_ = errors.New
)


func (*TestTest) Table() string { return "Test" }

func (data *TestTest) Scan(src any) error {
	*data = TestTest{}
	if src == nil {
		return nil
	}
	var content []byte
	switch value := src.(type) {
	case string:
		content = []byte(value)
	case []byte:
		content = value
	default:
		return fmt.Errorf("can not convert %#v into TestTest", src)
	}
	if len(content) == 0 {
		return nil
	}
	return json.Unmarshal(content, data)
}

func (data TestTest) Value() (driver.Value, error) {
	content, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	return string(content), nil
}


func (x TestTest) MarshalJSON() ([]byte, error) {
	data := map[string]any{
		"NAME": x.Name,
		"PublicId": x.PublicId,
		"Type": x.Type,
		"Description": x.Description,
		"ProjectId": x.ProjectId,
		"CreateTime": x.CreateTime,
		"UpdateTime": x.UpdateTime,
	}
	return json.Marshal(data)
}

func (x *TestTest) UnmarshalJSON(content []byte) (error) {
	data := map[string]any{
		"_public": &x.Public,
		"PublicId": &x.PublicId,
		"Type": &x.Type,
		"Description": &x.Description,
		"ProjectId": &x.ProjectId,
		"CreateTime": &x.CreateTime,
		"UpdateTime": &x.UpdateTime,
	}
	return json.Unmarshal(content, &data)
}
