package dartclient

import (
	"bytes"
	"fmt"
	"reflect"
	"sort"
	"strings"
	"text/template"
	"unicode"

	"github.com/vmkteam/zenrpc/v2/smd"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
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
	Comment   string
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

	tmpl, err := template.New("test").Funcs(fns).Parse(dartCliTmpl)
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
			Comment:   service.Description,
			Args:      args,
			Response:  resp,
		}

		// add service to namespace
		var index int
		for i := range c.client.Namespaces {
			if c.client.Namespaces[i].Name == cleanSymbols(namespace) {
				index = i
				break
			}
		}
		if _, ok := c.namespaces[namespace]; !ok {
			c.namespaces[namespace] = struct{}{}
			c.client.Namespaces = append(c.client.Namespaces, dartNamespace{
				Name:     cleanSymbols(namespace),
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
	if dc.Name == "AnyCustomType" {
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
		if description == "AnyCustomType" {
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
		if _, ok := subtypeScalars[subType]; ok {
			result.SubType = subType
		} else {
			result.SubType = ucFirst(cleanSymbols(subType))
		}
	}

	// add object as complex type
	if result.Type == "" && in.Type == objectType && in.Description != "" {
		c.addComplexInterface(in)
		result.Type = objectType
		result.SubType = ucFirst(cleanSymbols(in.Description))
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

	return result
}

// addComplexInterface converts complex type stored in smd.JSONSchema to dartClass and adds it to client.
func (c *Client) addComplexInterface(in smd.JSONSchema) {
	dartTypes := make([]dartType, len(in.Properties))

	for i := range in.Properties {
		dartTypes[i] = c.convertType(smd.JSONSchema{
			Name:        in.Properties[i].Name,
			Description: strings.TrimPrefix(in.Properties[i].Ref, definitionsPrefix),
			Type:        in.Properties[i].Type,
			Items:       in.Properties[i].Items,
		}, in.Properties[i].Description)
	}

	c.addModel(dartClass{
		Name:       ucFirst(cleanSymbols(in.Description)),
		Parameters: dartTypes,
	})
}

func (n *dartNamespace) ServiceName() string {
	return "_Service" + ucFirst(cleanSymbols(n.Name))
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
		return "as Map<String, dynamic>"
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
	"int":    {},
}

func (s dartService) IsResponseSubTypeScalar() bool {
	_, ok := subtypeScalars[s.Response.SubType]
	return ok
}

func (s dartService) ArgsType() string {
	c := cases.Title(language.Und, cases.NoLower)
	return fmt.Sprintf("%s%sArgs", c.String(cleanSymbols(s.Namespace)), c.String(s.Name))
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

var cleanSymbolsRep = strings.NewReplacer("-", "", ".", "")

func cleanSymbols(s string) string {
	return cleanSymbolsRep.Replace(s)
}
