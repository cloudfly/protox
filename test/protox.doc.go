package test


import "github.com/cloudfly/protox/utils/doc"


var ProtoxMessages = []*doc.Message{
	{
		Name: "TestCommon",
		Comment: []string{
		},
		Package: "test",
		Fields: []doc.Field{
			{
				Name: "id",
				Comment: []string{
				},
				Type: "int64",
				Required: true,
			},
			{
				Name: "name",
				Comment: []string{
				},
				Type: "string",
				Required: false,
				Tags: map[string]string{
					"db": "name",
					"op": "eq",
				},
			},
			{
				Name: "Public",
				Comment: []string{
				},
				Type: "string",
				Required: false,
				Tags: map[string]string{
					"db": "public",
				},
			},
			{
				Name: "publicId",
				Comment: []string{
					"@gotags: db:\"-\" json:\"id\"\n",
				},
				Type: "string",
				Required: true,
			},
			{
				Name: "type",
				Comment: []string{
					"@gotags: db:\"type\"\n",
					"'enum' | 'message'\n",
				},
				Type: "string",
				Required: true,
			},
			{
				Name: "description",
				Comment: []string{
					"@gotags: db:\"description\"\n",
				},
				Type: "string",
				Required: true,
			},
			{
				Name: "projectId",
				Comment: []string{
					"@gotags: db:\"fields\"\n",
				},
				Type: "int64",
				Required: true,
			},
			{
				Name: "createTime",
				Comment: []string{
					"@gotags: db:\"createTime\"\n",
				},
				Type: "int64",
				Required: true,
			},
			{
				Name: "updateTime",
				Comment: []string{
					"@gotags: db:\"updateTime\"\n",
				},
				Type: "int64",
				Required: true,
			},
		},
	},
	{
		Name: "TestInherit",
		Comment: []string{
		},
		Package: "test",
		Fields: []doc.Field{
			{
				Name: "type",
				Comment: []string{
				},
				Type: "int64",
				Required: true,
			},
		},
	},
	{
		Name: "TestRequest",
		Comment: []string{
		},
		Package: "test",
		Fields: []doc.Field{
			{
				Name: "query",
				Comment: []string{
				},
				Type: "string",
				Required: true,
			},
			{
				Name: "limit",
				Comment: []string{
				},
				Type: "int64",
				Required: true,
			},
			{
				Name: "nested",
				Comment: []string{
				},
				Type: "test.Nested",
				Required: true,
			},
			{
				Name: "tags",
				Comment: []string{
				},
				Type: "[]string",
				Required: true,
			},
			{
				Name: "metadata",
				Comment: []string{
				},
				Type: "map[string]string",
				Required: true,
			},
		},
	},
	{
		Name: "Nested",
		Comment: []string{
		},
		Package: "test",
		Fields: []doc.Field{
			{
				Name: "active",
				Comment: []string{
				},
				Type: "bool",
				Required: true,
			},
		},
	},
	{
		Name: "TestResponse",
		Comment: []string{
		},
		Package: "test",
		Fields: []doc.Field{
			{
				Name: "result",
				Comment: []string{
				},
				Type: "string",
				Required: true,
			},
		},
	},
}

var ProtoxServices = []*doc.Service{
	{
		Name: "TestService",
		Comment: []string{
		},
		Package: "test",
		Methods: []doc.Method{
			{
				Name: "TestMethod",
				Comment: []string{
				},
				Input: &doc.Message{
					Name: "TestRequest",
					Comment: []string{
					},
					Package: "test",
					Fields: []doc.Field{
						{
							Name: "query",
							Comment: []string{
							},
							Type: "string",
							Required: true,
						},
						{
							Name: "limit",
							Comment: []string{
							},
							Type: "int64",
							Required: true,
						},
						{
							Name: "nested",
							Comment: []string{
							},
							Type: "test.Nested",
							Required: true,
						},
						{
							Name: "tags",
							Comment: []string{
							},
							Type: "[]string",
							Required: true,
						},
						{
							Name: "metadata",
							Comment: []string{
							},
							Type: "map[string]string",
							Required: true,
						},
					},
				},
				Output: &doc.Message{
					Name: "TestResponse",
					Comment: []string{
					},
					Package: "test",
					Fields: []doc.Field{
						{
							Name: "result",
							Comment: []string{
							},
							Type: "string",
							Required: true,
						},
					},
				},
			},
			{
				Name: "TestMethod22",
				Comment: []string{
				},
				Input: &doc.Message{
					Name: "TestRequest",
					Comment: []string{
					},
					Package: "test",
					Fields: []doc.Field{
						{
							Name: "query",
							Comment: []string{
							},
							Type: "string",
							Required: true,
						},
						{
							Name: "limit",
							Comment: []string{
							},
							Type: "int64",
							Required: true,
						},
						{
							Name: "nested",
							Comment: []string{
							},
							Type: "test.Nested",
							Required: true,
						},
						{
							Name: "tags",
							Comment: []string{
							},
							Type: "[]string",
							Required: true,
						},
						{
							Name: "metadata",
							Comment: []string{
							},
							Type: "map[string]string",
							Required: true,
						},
					},
				},
				Output: &doc.Message{
					Name: "TestResponse",
					Comment: []string{
					},
					Package: "test",
					Fields: []doc.Field{
						{
							Name: "result",
							Comment: []string{
							},
							Type: "string",
							Required: true,
						},
					},
				},
			},
		},
	},
}
