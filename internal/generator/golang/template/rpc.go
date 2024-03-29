package template

const (
	RpcTemplate = `// Code generated by https://github.com/meerkat-io/disorder; DO NOT EDIT.
package {{PackageName .Schema.Package}}

import (
{{- range .RpcImports}}
	"{{.}}"
{{- end}}
)
{{- range $index, $service := .Schema.Services}}

type {{CamelCase $service.Name}}Handler func(*rpc.Context, *disorder.Decoder) (interface{}, *rpc.Error)

type {{PascalCase $service.Name}} interface {
	{{- range .Rpc}}
	{{PascalCase .Name}}(*rpc.Context, {{Type .Input}}) ({{Type .Output}}, *rpc.Error)
	{{- end}}
}

func New{{PascalCase $service.Name}}Client(client *rpc.Client) {{PascalCase $service.Name}} {
	return &{{CamelCase $service.Name}}Client{
		name:   "{{$service.Name}}",
		client: client,
	}
}

type {{CamelCase $service.Name}}Client struct {
	name   string
	client *rpc.Client
}
{{- range .Rpc}}

func (c *{{CamelCase $service.Name}}Client) {{PascalCase .Name}}(context *rpc.Context, request {{Type .Input}}) ({{Type .Output}}, *rpc.Error) {
	var response {{Type .Output}}{{InitType .Output}}
	err := c.client.Send(context, c.name, "{{.Name}}", request, {{if not (IsPointer .Output)}}&{{end}}response)
	return response, err
}
{{- end}}

type {{CamelCase $service.Name}}Server struct {
	name    string
	service {{PascalCase $service.Name}}
	methods map[string]{{CamelCase $service.Name}}Handler
}

func Register{{PascalCase $service.Name}}Server(s *rpc.Server, service {{PascalCase $service.Name}}) {
	server := &{{CamelCase $service.Name}}Server{
		name:    "{{$service.Name}}",
		service: service,
	}
	server.methods = map[string]{{CamelCase $service.Name}}Handler{
{{- range .Rpc}}
		"{{.Name}}": server.{{CamelCase .Name}},
{{- end}}
	}
	s.RegisterService("{{$service.Name}}", server)
}

func (s *{{CamelCase $service.Name}}Server) Handle(context *rpc.Context, method string, d *disorder.Decoder) (interface{}, *rpc.Error) {
	handler, ok := s.methods[method]
	if ok {
		return handler(context, d)
	}
	return nil, &rpc.Error{
		Code:  code.Unimplemented,
		Error: fmt.Errorf("unimplemented method \"%s\" under service \"%s\"", method, s.name),
	}
}
{{- range .Rpc}}

func (s *{{CamelCase $service.Name}}Server) {{CamelCase .Name}}(context *rpc.Context, d *disorder.Decoder) (interface{}, *rpc.Error) {
	var request {{Type .Input}}{{InitType .Input}}
	err := d.Decode({{if not (IsPointer .Input)}}&{{end}}request)
	if err != nil {
		return nil, &rpc.Error{
			Code:  code.InvalidRequest,
			Error: err,
		}
	}
	response, rpcErr := s.service.{{PascalCase .Name}}(context, request)
	if rpcErr != nil {
		return nil, rpcErr
	}
	return response, nil
}
{{- end}}
{{- end}}`
)
