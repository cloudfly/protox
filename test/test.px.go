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


func (TestCommon) Table() string { return "Test" }

func (TestCommon) Error() string { return "test error" }

func (data *TestCommon) Scan(src any) error {
	*data = TestCommon{}
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
		return fmt.Errorf("can not convert %#v into TestCommon", src)
	}
	if len(content) == 0 {
		return nil
	}
	return json.Unmarshal(content, data)
}

func (data TestCommon) Value() (driver.Value, error) {
	content, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	return string(content), nil
}


func (x TestCommon) JSON() ([]byte, error) {
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

func (x *TestCommon) FromJSON(content []byte) (error) {
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

func (x *TestInherit) FromTestCommon(parent *TestCommon) {
	if x == nil || parent == nil {
		return
	}
	x.Id = parent.Id
	x.Name = parent.Name
	x.PublicId = parent.PublicId
	x.Description = parent.Description
	x.ProjectId = parent.ProjectId
	x.CreateTime = parent.CreateTime
	x.UpdateTime = parent.UpdateTime
	return
}


func (x *TestInherit) ToTestCommon() *TestCommon {
	if x == nil {
		return &TestCommon{}
	}
	target := &TestCommon{}
	target.Id = x.Id
	target.Name = x.Name
	target.PublicId = x.PublicId
	target.Description = x.Description
	target.ProjectId = x.ProjectId
	target.CreateTime = x.CreateTime
	target.UpdateTime = x.UpdateTime
	return target
}


func (x TestInherit) JSON() ([]byte, error) {
	data := map[string]any{
		"Type": x.Type,
		"NAME": x.Name,
		"PublicId": x.PublicId,
		"Description": x.Description,
		"ProjectId": x.ProjectId,
		"CreateTime": x.CreateTime,
		"UpdateTime": x.UpdateTime,
	}
	return json.Marshal(data)
}

func (x *TestInherit) FromJSON(content []byte) (error) {
	data := map[string]any{
		"Type": &x.Type,
		"PublicId": &x.PublicId,
		"Description": &x.Description,
		"ProjectId": &x.ProjectId,
		"CreateTime": &x.CreateTime,
		"UpdateTime": &x.UpdateTime,
	}
	return json.Unmarshal(content, &data)
}

func (x Error) Error() string {
	switch x {
		case Error_InternalServerError: 
			return "internal_server_error"
		case Error_PermissionDenied: 
			return "permission_denied"
		case Error_NotFound: 
			return "not_found"
	}
	return fmt.Sprintf("unknown Error %d", x)
}
