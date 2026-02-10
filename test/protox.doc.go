package test

import "github.com/cloudfly/protox/utils/doc"

var ProtoxMessages = []*doc.Message{
	{
		Name:    "TestCommon",
		Comment: []string{},
		Package: "test",
		Fields: []doc.Field{
			{
				Name:     "id",
				Comment:  []string{},
				Type:     "int64",
				Required: true,
			},
			{
				Name:     "name",
				Comment:  []string{},
				Type:     "string",
				Required: false,
				Tags: map[string]string{
					"db": "name",
					"op": "eq",
				},
			},
			{
				Name:     "Public",
				Comment:  []string{},
				Type:     "string",
				Required: false,
				Tags: map[string]string{
					"db": "public",
				},
			},
			{
				Name:     "publicId",
				Comment:  []string{},
				Type:     "string",
				Required: true,
			},
			{
				Name:     "type",
				Comment:  []string{},
				Type:     "string",
				Required: true,
			},
			{
				Name:     "description",
				Comment:  []string{},
				Type:     "string",
				Required: true,
			},
			{
				Name:     "projectId",
				Comment:  []string{},
				Type:     "int64",
				Required: true,
			},
			{
				Name:     "createTime",
				Comment:  []string{},
				Type:     "protox.Timestamp",
				Required: true,
			},
			{
				Name:     "updateTime",
				Comment:  []string{},
				Type:     "protox.Timestamp",
				Required: true,
			},
		},
	},
	{
		Name:    "TestInherit",
		Comment: []string{},
		Package: "test",
		Fields: []doc.Field{
			{
				Name:     "type",
				Comment:  []string{},
				Type:     "int64",
				Required: true,
			},
		},
	},
}

var ProtoxServices = []*doc.Service{}
