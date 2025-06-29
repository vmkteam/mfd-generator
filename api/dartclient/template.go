package dartclient

const (
	dartCliTmpl = `// Code generated from zenrpc smd. DO NOT EDIT.

// To update this file:
// 1. Start mfd-generator locally on 8080 port: ` + "`mfd-generator server -a=:8080`" + `.
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
  {{if not (eq .Comment "") }}// {{ .Comment }}{{ end }}
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
`
)
