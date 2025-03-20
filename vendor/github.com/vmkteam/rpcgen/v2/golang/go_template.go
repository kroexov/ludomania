package golang

// goTpl contains template for Go client
const goTpl = `// Code generated from jsonrpc schema by rpcgen v{{ .Version }}; DO NOT EDIT.

package {{ .Package }}

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"io"
	"sync/atomic"
	"strconv"
	"time"

	"github.com/vmkteam/zenrpc/v2"
)

var (
	// Always import time package. Generated models can contain time.Time fields.
	_ time.Time
)

type Client struct {
	rpcClient *rpcClient

{{ range .NamespaceNames }}
	{{ title . }} *svc{{ title .}}{{ end }}}

func NewDefaultClient(endpoint string) *Client {
	return NewClient(endpoint, http.Header{}, &http.Client{})
}

func NewClient(endpoint string, header http.Header, httpClient *http.Client) *Client {
	c := &Client{
		rpcClient: newRPCClient(endpoint, header, httpClient),
	}
{{ range .NamespaceNames }}
	c.{{ title . }} = newClient{{ title .}}(c.rpcClient){{ end }}

	return c
}

{{ range .Models }}
type {{ .Name }} struct {
	{{ range .Fields }}{{ if ne .Description "" }} {{ .CommentDescription }}
	{{ end }}{{ title .Name }} {{ if and .Optional (eq .ArrayItemType "")}}*{{ end }}{{ .GoType }} ` + "`json:\"{{ .Name }}{{if .Optional}},omitempty{{end}}\"`" + `
{{ end }}
}
{{ end }} 


{{ range .NamespaceNames }}
 {{ $lTitle := . }}{{ $namespace := . }}
type svc{{ title . }} struct {
	client *rpcClient
}

func newClient{{ title . }} (client *rpcClient) *svc{{ title . }}  {
	return &svc{{ title . }}{
		client: client,
	}
}

{{ range $.NamespaceMethodNames .}} {{ $method := $.MethodByName $namespace . }}

{{ if $method.HasErrors }}
var (
{{ range $method.Errors }}
Err{{ title $namespace }}{{ title $method.Name }}{{ .StringCode }} = zenrpc.NewError({{ .Code }}, fmt.Errorf("{{ .Message }}")){{ end }}
)
{{ end }}

{{ $method.CommentDescription }}
func (c *svc{{ title $lTitle}}) {{ title . }}(ctx context.Context, {{ range $method.Params }}{{ .Name }} {{ if and .Optional (eq .ArrayItemType "")}}*{{ end }}{{ .GoType }}, {{ end }}) ( {{ if $method.HasResult }} res {{ if and $method.Returns.Optional (eq $method.Returns.ArrayItemType "")}}*{{ end }}{{ $method.Returns.GoType }},  {{ else }} {{end}} err error) {
	_req := struct {
		{{ range $method.Params }}{{ title .Name }} {{ if and .Optional (eq .ArrayItemType "")}}*{{ end }}{{ .GoType }}
{{ end }}
	} {
		{{ range $method.Params }}{{ title .Name }}: {{ .Name }}, {{ end }}
	}

	err = c.client.call(ctx, "{{ $namespace }}.{{ . }}", _req, {{ if $method.HasResult }} &res {{ else }} nil {{ end }})
{{ if $method.HasErrors }}
	switch v:= err.(type) {
		case *zenrpc.Error:
				{{- range $method.Errors }}
				if v.Code == {{ .Code }} {
					err = Err{{ title $namespace }}{{ title $method.Name }}{{ .StringCode }}
				}
                {{- end }}
	}
{{- end }}

	return
}
{{ end }}
{{ end }}

type rpcClient struct {
	endpoint string
	cl       *http.Client

	requestID uint64
	header    http.Header
}

func newRPCClient(endpoint string, header http.Header, httpClient *http.Client) *rpcClient {
	return &rpcClient{
		endpoint: endpoint,
		header:   header,
		cl:       httpClient,
	}
}

func (rc *rpcClient) call(ctx context.Context, methodName string, request, result interface{}) error {
	// encode params
	bts, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("encode params: %w", err)
	}

	requestID := atomic.AddUint64(&rc.requestID, 1)
	requestIDBts := json.RawMessage(strconv.Itoa(int(requestID)))

	req := zenrpc.Request{
		Version: zenrpc.Version,
		ID:      &requestIDBts,
		Method:  methodName,
		Params:  bts,
	}

	res, err := rc.Exec(ctx, req)
	if err != nil {
		return err
	}

	if res == nil {
		return nil
	}

	if res.Error != nil {
		return res.Error
	}

	if res.Result == nil {
		return nil
	}

	if result == nil {
		return nil
	}

	return json.Unmarshal(*res.Result, result)
}

// Exec makes http request to jsonrpc endpoint and returns json rpc response.
func (rc *rpcClient) Exec(ctx context.Context, rpcReq zenrpc.Request) (*zenrpc.Response, error) {
	if n, ok := ctx.Value("JSONRPC2-Notification").(bool); ok && n {
		rpcReq.ID = nil
	}

	c, err := json.Marshal(rpcReq)
	if err != nil {
		return nil, fmt.Errorf("json marshal call failed: %w", err)
	}

	buf := bytes.NewReader(c)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, rc.endpoint, buf)
	if err != nil {
		return nil, fmt.Errorf("create request failed: %w", err)
	}

	req.Header = rc.header.Clone()
	req.Header.Add("Content-Type", "application/json")

	if xRequestID, ok := ctx.Value("X-Request-Id").(string); ok && req.Header.Get("X-Request-Id") == "" && xRequestID != "" {
		req.Header.Add("X-Request-Id", xRequestID)
	}

	// Do request
	resp, err := rc.cl.Do(req)
	if resp != nil {
		defer resp.Body.Close()
	}

	if err != nil {
		return nil, fmt.Errorf("make request failed: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad response (%d)", resp.StatusCode)
	}

	var zresp zenrpc.Response
	if rpcReq.ID == nil {
		return &zresp, nil
	}

	bb, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("response body (%s) read failed: %w", bb, err)
	}

	if err = json.Unmarshal(bb, &zresp); err != nil {
		return nil, fmt.Errorf("json decode failed (%s): %w", bb, err)
	}

	return &zresp, nil
}
`
