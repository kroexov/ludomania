package typescript

const client = `/* Code generated from jsonrpc schema by rpcgen v{{ .Version }}; DO NOT EDIT. */
/* eslint-disable */
{{- range .Interfaces }}
export interface {{ .Name }} {
{{- $len := len .Parameters }}
{{- range $i, $e := .Parameters }}
  {{ .Name }}{{ if .Optional }}?{{ end }}: {{ .Type }}{{ if ne $i $len }},{{ end }}{{ if ne .Comment "" }} // {{ .Comment }}
{{- end }}
{{- end }}
}
{{ end }}

{{- if .WithClasses }}
{{- range .Interfaces }}
export class {{ .ModelName }} implements {{ .Name }} {
  static entityName = "{{ .EntityNameTmpl }}";
{{ $len := len .Parameters }}
{{- range $i,$e := .Parameters }}
  {{ .Name }}{{ if .Optional }}?{{ end }}: {{ .Type }} = {{ .DefaultTmpl }};
{{- end }}
}
{{ end }}

{{- end }}
export const factory = (send: any) => ({
{{- $lenN := len .Namespaces }}
{{- range $i,$e := .Namespaces }}
  {{ .Name }}: {
{{- $lenS := len .Services }}
{{- range $i, $e := .Services }}
    {{ .NameLCF }}({{ if .HasParams }}params: {{ .Params }}{{ end }}): Promise<{{ .Response }}> {
      return send('{{ .Namespace }}.{{ .Name }}'{{ if .HasParams }}, params{{ end }})
    }{{ if ne $i $lenS }},{{ end }}
{{- end }}
  }{{ if ne $i $lenN }},{{ end }}
{{- end }}
})
`
