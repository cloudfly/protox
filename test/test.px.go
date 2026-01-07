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


func (*TestCommon) Table() string { return "Test" }

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

func (x TestOrmx) OrmxColumnOption(fieldName string) string {
	switch fieldName {
		case "Noop": 
			return "noop"
		case "Id": 
			return "id,insert:true"
		case "Name": 
			return "name,insert:true,select:true,update:false,op:eq"
	}
	return ""
}

func (x TestOrmx) OrmxColumn(fieldName string) string {
	switch fieldName {
		case "Noop": 
			return "noop"
		case "Id": 
			return "id"
		case "Name": 
			return "name"
	}
	return ""
}

func (x Error) Error() string {
	switch x {
		case Error_NotFound: 
			return "not_found"
		case Error_InternalServerError: 
			return "internal_server_error"
		case Error_PermissionDenied: 
			return "permission_denied"
	}
	return fmt.Sprintf("unknown Error %d", x)
}
