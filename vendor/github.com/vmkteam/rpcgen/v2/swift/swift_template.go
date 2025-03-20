package swift

const client = `/// Code generated from jsonrpc schema by rpcgen v{{ .Version }}; DO NOT EDIT.

import Foundation

extension {{ .Class }}: RPCMethod {
    public var rpcMethod: String {
        switch self {
       	{{- range .Methods }}
        case .{{ .SafeName }}: return "{{ .Name }}"
       	{{- end }}
        }
    }
}

extension {{ .Class }}: RPCParameters {
    public var rpcParameters: [String: Any?]? {
        switch self {
{{- $methodsLen := len .Methods }}
{{- range $idx, $m := .Methods }}{{- $paramsLen := len .Parameters }}
        case {{ if gt $paramsLen 0 }}let {{ end }}.{{ .SafeName }}{{ if gt $paramsLen 0 }}({{ range $index, $item := .Parameters }}{{ .Name }}{{ if (notLast $index $paramsLen) }}, {{ end }}{{ end }}){{ end }}:
            return {{ if eq $paramsLen 0 }}nil{{ else }}[{{ range $index, $item := .Parameters }}"{{ .Name }}": {{ .Name }}{{ if or .IsArray .IsObject }}.any{{ end }}{{ if (notLast $index $paramsLen) }}, {{ end }}{{ end }}]{{ end }}
{{- if (notLast $idx $methodsLen) }}{{ print "\n" }}{{- end }}
{{- end }}
        }
    }
}

public enum {{ .Class }} {
{{- range .Methods }}{{- $paramsLen := len .Parameters }}
    {{- range .Description }}
    {{- if ne . "" }}
    /// {{ . }}
    {{- end }}
    {{- end }}
    {{- if ne .Returns.Type ""}}
    /// - Returns: {{ .Returns.Type }}{{ if .Returns.Optional }}?{{ end }}
    {{- end }}
    case {{ .SafeName }}{{ if gt $paramsLen 0 }}({{ range $index, $item := .Parameters }}{{ .Name }}: {{ .Type }}{{ if .Optional }}?{{ end }}{{ if (notLast $index $paramsLen) }}, {{ end }}{{ end }}){{ end }}
{{ end }}
}

{{ range .Models }}
{{- if .Description }}
/// {{ .Description }}
{{- end }}
public struct {{ .Name }}: Codable, Hashable {
    {{- $fieldsLen := len .Fields }}{{- range .Fields }}
    {{- if .Description }}
    /// {{ .Description }}
    {{- end }}
    {{- if and (not .Optional) (ne .DecodableDefault "") }}
    @DecodableDefault.{{ .DecodableDefault }}
    {{- end }}
    var {{ if .NeedEscaping }}{{ .SafeName }}{{ else }}{{ .Name }}{{ end }}: {{ .Type }}{{ if .Optional }}?{{ end }}
    {{- end }}
    init({{ range $index, $f := .Fields }}{{ .Name }}: {{ .Type }}{{ if .Optional }}? = nil{{ end }}{{ if (notLast $index $fieldsLen) }}, {{ end }}{{ end }}) {
        {{- range .Fields }}
        self.{{ .Name }} = {{ if .NeedEscaping }}{{ .SafeName }}{{ else }}{{ .Name }}{{ end }}
        {{- end }}
    }
}
{{ end }}
`
