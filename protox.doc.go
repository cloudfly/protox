package protox


import "github.com/cloudfly/protox/utils/doc"


var ProtoxMessages = []*doc.Message{
	{
		Name: "Timestamp",
		Comment: []string{
		},
		Package: "protox",
		Fields: []doc.Field{
			{
				Name: "millis",
				Comment: []string{
				},
				Type: "uint64",
				Required: true,
			},
		},
	},
	{
		Name: "GoMethodOption",
		Comment: []string{
		},
		Package: "protox",
		Fields: []doc.Field{
			{
				Name: "name",
				Comment: []string{
				},
				Type: "string",
				Required: true,
			},
			{
				Name: "returnString",
				Comment: []string{
				},
				Type: "string",
				Required: false,
			},
			{
				Name: "returnInt64",
				Comment: []string{
				},
				Type: "int64",
				Required: false,
			},
			{
				Name: "returnBool",
				Comment: []string{
				},
				Type: "bool",
				Required: false,
			},
		},
	},
	{
		Name: "SQLOption",
		Comment: []string{
		},
		Package: "protox",
		Fields: []doc.Field{
			{
				Name: "serializer",
				Comment: []string{
				},
				Type: "string",
				Required: false,
			},
		},
	},
	{
		Name: "InheritOption",
		Comment: []string{
		},
		Package: "protox",
		Fields: []doc.Field{
			{
				Name: "message",
				Comment: []string{
				},
				Type: "string",
				Required: true,
			},
			{
				Name: "omit",
				Comment: []string{
				},
				Type: "[]string",
				Required: true,
			},
		},
	},
	{
		Name: "ApiOption",
		Comment: []string{
		},
		Package: "protox",
		Fields: []doc.Field{
			{
				Name: "enable",
				Comment: []string{
				},
				Type: "bool",
				Required: true,
			},
		},
	},
	{
		Name: "OrmOption",
		Comment: []string{
		},
		Package: "protox",
		Fields: []doc.Field{
			{
				Name: "enable",
				Comment: []string{
				},
				Type: "bool",
				Required: true,
			},
			{
				Name: "table",
				Comment: []string{
				},
				Type: "string",
				Required: false,
			},
			{
				Name: "filterMessage",
				Comment: []string{
				},
				Type: "string",
				Required: false,
			},
			{
				Name: "patchMessage",
				Comment: []string{
				},
				Type: "string",
				Required: false,
			},
		},
	},
}

var ProtoxServices = []*doc.Service{
}
