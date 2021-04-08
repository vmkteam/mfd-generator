package dartclient

import (
	"bytes"
	"fmt"
	"reflect"
	"sort"
	"strings"
	"text/template"
	"unicode"

	"github.com/semrush/zenrpc/v2/smd"
)

const (
	definitionsPrefix = "#/definitions/"
	voidResponse      = "void"
	listType          = "List"
	objectType        = "object"
)

type dartClass struct {
	Name       string
	Parameters []dartType
}

type dartType struct {
	Name     string
	Comment  string // TODO
	Type     string
	SubType  string
	Optional bool // TODO
}

type dartNamespace struct {
	Name     string
	Services []dartService
}

type dartService struct {
	Namespace string // TODO
	Name      string // TODO
	Args      []dartType
	Response  dartType
}

type Client struct {
	smd    smd.Schema
	client struct {
		Models     []dartClass
		Namespaces []dartNamespace
	}
	models     map[string]struct{}
	namespaces map[string]struct{}
}

func NewClient(schema smd.Schema) *Client {
	return &Client{
		smd:        schema,
		models:     map[string]struct{}{},
		namespaces: map[string]struct{}{},
	}
}

// Run converts SMD client to Dart model.
func (c *Client) Run() ([]byte, error) {
	c.convert()

	var fns = template.FuncMap{
		"len": func(a interface{}) int {
			return reflect.ValueOf(a).Len() - 1
		},
	}

	tmpl, err := template.New("test").Funcs(fns).Parse(
		`// Code generated from zenrpc smd. DO NOT EDIT.

// To update this file:
// 1. Pull fresh version and start apisvc locally on 8080 port.
// 2. Navigate in terminal to root directory of this flutter project.
// 3. ` + "`curl http://localhost:8080/doc/api_client.dart --output ./lib/services/api/api_client.dart`" + `
// 4. ` + "`dart format --fix -l 150 ./lib/services/api/api_client.dart`" + `
// 5. ` + "`flutter pub run build_runner build --delete-conflicting-outputs`" + `

import 'package:json_annotation/json_annotation.dart';

part 'api_client.g.dart';

// JSONRPCClient is the main interface of executor class.
// Implementations may use different transports: http, websockets, nats, etc.
abstract class JSONRPCClient {
  Future<dynamic> call(String method, dynamic params) async {}
}

// ----- main api client class -----

class ApiClient {
  ApiClient(JSONRPCClient client) {{$lenN := len .Namespaces}}{{range $i,$e := .Namespaces}}{{if eq $i 0}}:{{end}}
    {{.Name}} = {{.ServiceName}}(client){{if ne $i $lenN}},{{else}};{{end}}{{end}}
{{range .Namespaces}}
  final {{.ServiceName}} {{.Name}}; {{end}}
}

// ----- namespace classes -----

{{range .Namespaces}}
class {{.ServiceName}} {
  {{.ServiceName}}(this._client);

  final JSONRPCClient _client;
{{$shortName := .Name}}{{range .Services}}
  Future<{{.FutureType}}> {{.NameLCF}}({{.ArgsType}} args) {
    return Future(() async {
      {{- if ne .FutureType "void"}}final response ={{- end }} await _client.call('{{$shortName}}.{{.NameLCF}}', args) {{.AsType}};
      {{- if eq .ResponseType "List"}}
      final responseList = response?.map((e) {
        if (e == null) {
          return null;
        }
		{{ if .IsResponseSubTypeScalar -}}
		return e as {{.ResponseSubType}};
		{{- else -}}
        return {{.ResponseSubType}}.fromJson(e as Map<String, dynamic>);
		{{- end }}
      });
      return responseList?.toList(); 
      {{- else if eq .ResponseType "object"}}
      return {{.ResponseSubType}}.fromJson(response);
      {{- else if ne .FutureType "void"}}
      return response;
      {{- end}}
    });
  }
  {{end}} 
} 
{{ range .Services }}
@JsonSerializable(includeIfNull: false, explicitToJson: true)
class {{.ArgsType}} {
  {{.ArgsType}}( {{- $len := len .Args}}{{if gt $len -1 -}} { {{range $i,$e := .Args}}this.{{.Name}}{{if ne $len $i }}, {{end}}{{end}} } {{- end}});

  factory {{.ArgsType}}.fromJson(Map<String, dynamic> json) => _${{.ArgsType}}FromJson(json);

  Map<String, dynamic> toJson() => _${{.ArgsType}}ToJson(this);
{{range $i,$e := .Args}}
  final {{if eq .Type "object"}}{{.SubType}}?{{else if eq .Type "List"}}List<{{.SubType}}?>?{{else}}{{.Type}}?{{end}} {{.Name}}; {{end}}
}
{{end}}{{end}}

// ----- models -----

{{range .Models}}
@JsonSerializable(includeIfNull: false, explicitToJson: true)
class {{.Name}} {
  {{.Name}}( {{- $len := len .Parameters}}{{if gt $len -1 -}} { {{range $i,$e := .Parameters}}this.{{.Name}}{{if ne $len $i }}, {{end}}{{end}} } {{- end}});

  factory {{.Name}}.fromJson(Map<String, dynamic> json) => _${{.Name}}FromJson(json);

  Map<String, dynamic> toJson() => _${{.Name}}ToJson(this);
{{range $i,$e := .Parameters}}
  final {{if eq .Type "object"}}{{.SubType}}?{{else if eq .Type "List"}}List<{{.SubType}}?>?{{else}}{{.Type}}?{{end}} {{.Name}}; {{end}}
}
{{end}}
`)
	if err != nil {
		return nil, err
	}

	// compile template
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, c.client); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// convert converts SMD services to Dart.
func (c *Client) convert() {
	// iterate over all services
	for serviceName, service := range c.smd.Services {
		serviceNameParts := strings.Split(serviceName, ".")
		if len(serviceNameParts) != 2 {
			continue
		}
		namespace := serviceNameParts[0]
		method := serviceNameParts[1]

		// add service args as Dart model
		args := make([]dartType, len(service.Parameters))
		for i := range service.Parameters {
			args[i] = c.convertType(service.Parameters[i], "")
		}

		// add service "returns" as Dart model
		resp := c.convertType(service.Returns, "")

		// add service to dart services
		respService := dartService{
			Namespace: namespace,
			Name:      method,
			Args:      args,
			Response:  resp,
		}

		// add service to namespace
		var index int
		for i := range c.client.Namespaces {
			if c.client.Namespaces[i].Name == namespace {
				index = i
			}
		}
		if _, ok := c.namespaces[namespace]; !ok {
			c.namespaces[namespace] = struct{}{}
			c.client.Namespaces = append(c.client.Namespaces, dartNamespace{
				Name:     namespace,
				Services: nil,
			})
			index = len(c.client.Namespaces) - 1
		}
		c.client.Namespaces[index].Services = append(c.client.Namespaces[index].Services, respService)
	}

	// sort models
	sort.Slice(c.client.Models, func(i, j int) bool {
		return c.client.Models[i].Name < c.client.Models[j].Name
	})

	// sort models args
	for idx := range c.client.Models {
		sort.Slice(c.client.Models[idx].Parameters, func(i, j int) bool {
			return c.client.Models[idx].Parameters[i].Name < c.client.Models[idx].Parameters[j].Name
		})
	}

	// sort namespaces
	sort.Slice(c.client.Namespaces, func(i, j int) bool {
		return c.client.Namespaces[i].Name < c.client.Namespaces[j].Name
	})

	// sort methods
	for idx := range c.client.Namespaces {
		sort.Slice(c.client.Namespaces[idx].Services, func(i, j int) bool {
			return c.client.Namespaces[idx].Services[i].Name < c.client.Namespaces[idx].Services[j].Name
		})
		// sort args
		for si := range c.client.Namespaces[idx].Services {
			sort.Slice(c.client.Namespaces[idx].Services[si].Args, func(i, j int) bool {
				return c.client.Namespaces[idx].Services[si].Args[i].Name < c.client.Namespaces[idx].Services[si].Args[j].Name
			})
		}
	}
}

// addModel adds Dart class model to client.
func (c *Client) addModel(dc dartClass) {
	if len(dc.Parameters) == 0 {
		return
	}
	switch dc.Name {
	case "PorebrikTime":
		return
	}

	if _, ok := c.models[dc.Name]; !ok {
		c.client.Models = append(c.client.Models, dc)
		c.models[dc.Name] = struct{}{}
	}
}

// convertScalar converts scalars from go to dart.
func (c *Client) convertScalar(t, description string) string {
	switch t {
	case "integer", "int":
		return "int"
	case "string":
		return "String"
	case "number":
		return "double"
	case "boolean":
		return "bool"
	default:
		switch description {
		case "PorebrikTime":
			return "DateTime"
		}
		return ""
	}
}

// convertType converts smd.JSONSchema to dartType.
func (c *Client) convertType(in smd.JSONSchema, comment string) dartType {
	result := dartType{
		Name:     in.Name,
		Comment:  comment,
		Type:     c.convertScalar(in.Type, in.Description),
		Optional: in.Optional,
	}

	// detect array sub type
	if in.Type == "array" {
		var subType string
		if scalar, ok := in.Items["type"]; ok {
			subType = c.convertScalar(scalar, in.Description)
		}
		if ref, ok := in.Items["$ref"]; ok {
			subType = strings.TrimPrefix(ref, definitionsPrefix)
		}

		result.Type = listType
		result.SubType = subType
	}

	// add object as complex type
	if result.Type == "" && in.Type == objectType && in.Description != "" {
		c.addComplexInterface(in)
		result.Type = objectType
		result.SubType = in.Description
	}

	// add definitions as complex types
	for name, d := range in.Definitions {
		c.addComplexInterface(smd.JSONSchema{
			Name:        name,
			Description: name,
			Type:        d.Type,
			Properties:  d.Properties,
		})
	}

	// dirty hacks for dictionaries
	// TODO we need map support in SMD schema
	//if in.Type == "object" {
	//	if in.Description == "ApiPharmacy" && in.Name == "pharmacies" {
	//		result.Type = fmt.Sprintf("Record<number, %s>", in.Description)
	//	}
	//	if in.Description == "ApiExtendedPickupPrice" && in.Name == "extendedPickups" {
	//		result.Type = fmt.Sprintf("Record<number, %s>", in.Description)
	//	}
	//	if in.Description == "ApiPharmacyPrice" && in.Name == "pharmacies" {
	//		result.Type = fmt.Sprintf("Record<number, %s>", in.Description)
	//	}
	//}

	return result
}

// addComplexInterface converts complex type stored in smd.JSONSchema to dartClass and adds it to client.
func (c *Client) addComplexInterface(in smd.JSONSchema) {
	var dartTypes []dartType

	for name, p := range in.Properties {
		dartTypes = append(dartTypes, c.convertType(smd.JSONSchema{
			Name:        name,
			Description: strings.TrimPrefix(p.Ref, definitionsPrefix),
			Type:        p.Type,
			Items:       p.Items,
		}, p.Description))
	}

	c.addModel(dartClass{
		Name:       in.Description,
		Parameters: dartTypes,
	})
}

func (n *dartNamespace) ServiceName() string {
	return "_Service" + ucFirst(n.Name)
}

func (s *dartService) FutureType() string {
	switch s.Response.Type {
	case listType:
		return fmt.Sprintf("List<%s?>?", s.Response.SubType)
	case objectType:
		return s.Response.SubType + "?"
	case "":
		return voidResponse
	default:
		return s.Response.Type + "?"
	}
}

func (s dartService) AsType() string {
	switch s.Response.Type {
	case listType:
		return "as List?"
	case objectType:
		return "as Map<String, dynamic>?"
	case "":
		return ""
	default:
		return "as " + s.Response.Type + "?"
	}
}

func (s dartService) NameLCF() string {
	return lcFirst(s.Name)
}

func (s dartService) HasArgs() bool {
	return len(s.Args) > 0
}

func (s dartService) ResponseType() string {
	if s.Response.Type == "" {
		return voidResponse
	}
	return s.Response.Type
}

func (s dartService) ResponseSubType() string {
	return s.Response.SubType
}

var subtypeScalars = map[string]struct{}{
	"String": {},
}

func (s dartService) IsResponseSubTypeScalar() bool {
	_, ok := subtypeScalars[s.Response.SubType]
	return ok
}

func (s dartService) ArgsType() string {
	return fmt.Sprintf("%s%sArgs", strings.Title(s.Namespace), strings.Title(s.Name))
}

func lcFirst(str string) string {
	for _, v := range str {
		u := string(unicode.ToLower(v))
		return u + str[len(u):]
	}
	return ""
}

func ucFirst(str string) string {
	for _, v := range str {
		u := string(unicode.ToUpper(v))
		return u + str[len(u):]
	}
	return ""
}
