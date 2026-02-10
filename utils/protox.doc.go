package utils


import "github.com/cloudfly/protox/utils/doc"


var ProtoxMessages = []*doc.Message{
	{
		Name: "Pagination",
		Comment: []string{
		},
		Package: "utils",
		Fields: []doc.Field{
			{
				Name: "page",
				Comment: []string{
				},
				Type: "uint64",
				Required: true,
			},
			{
				Name: "pageSize",
				Comment: []string{
				},
				Type: "uint64",
				Required: true,
			},
			{
				Name: "order",
				Comment: []string{
				},
				Type: "[]string",
				Required: true,
			},
		},
	},
	{
		Name: "Empty",
		Comment: []string{
		},
		Package: "utils",
		Fields: []doc.Field{
		},
	},
}

var ProtoxServices = []*doc.Service{
}
