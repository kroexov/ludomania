# rpcgen: JSON-RPC 2.0 Client Generator Implementation for zenrpc

[![Go Report Card](https://goreportcard.com/badge/github.com/vmkteam/rpcgen)](https://goreportcard.com/report/github.com/vmkteam/rpcgen) [![Go Reference](https://pkg.go.dev/badge/github.com/vmkteam/rpcgen.svg)](https://pkg.go.dev/github.com/vmkteam/rpcgen)

`rpcgen` is a JSON-RPC 2.0 client library generator for [zenrpc](https://github.com/vmkteam/zenrpc). It supports client generation for following languages:
- Golang
- PHP
- TypeScript
- Swift
- OpenRPC schema

## Examples

### Basic usage

```go
package main

import (
	"fmt"
	"log"

	"github.com/vmkteam/rpcgen/v2"
	"github.com/vmkteam/zenrpc/v2"
)

func main() {
	rpc := zenrpc.NewServer(zenrpc.Options{})

	generated, err := rpcgen.FromSMD(rpc.SMD()).GoClient().Generate()
	if err != nil {
		log.Fatal(err)
	}
	
	fmt.Printf("%s", generated)
}
```

### Generate in HTTP handler

```go
package main

import (
	"net/http"

	"github.com/vmkteam/rpcgen/v2"
	"github.com/vmkteam/rpcgen/v2/golang"
	"github.com/vmkteam/rpcgen/v2/swift"
	"github.com/vmkteam/zenrpc/v2"
)

func main() {
	rpc := zenrpc.NewServer(zenrpc.Options{})

	gen := rpcgen.FromSMD(rpc.SMD())

	http.HandleFunc("/client.go", rpcgen.Handler(gen.GoClient(golang.Settings{})))
	http.HandleFunc("/client.ts", rpcgen.Handler(gen.TSClient(nil)))
	http.HandleFunc("/RpcClient.php", rpcgen.Handler(gen.PHPClient("")))
	http.HandleFunc("/client.swift", rpcgen.Handler(gen.SwiftClient(swift.Settings{})))
}
```

### Add custom TypeScript type mapper

```go
package main

import (
	"fmt"
	"net/http"

	"github.com/vmkteam/rpcgen/v2"
	"github.com/vmkteam/rpcgen/v2/typescript"
	"github.com/vmkteam/zenrpc/v2"
	"github.com/vmkteam/zenrpc/v2/smd"
)

func main() {
	rpc := zenrpc.NewServer(zenrpc.Options{})

	gen := rpcgen.FromSMD(rpc.SMD())

	typeMapper := func(in smd.JSONSchema, tsType typescript.Type) typescript.Type {
		if in.Type == "object" {
			if in.Description == "Group" && in.Name == "groups" {
				tsType.Type = fmt.Sprintf("Record<number, I%s>", in.Description)
			}
		}
		
		return tsType
	}

	http.HandleFunc("/client.ts", rpcgen.Handler(gen.TSClient(typeMapper)))
}
```

### Add custom Swift type mapper

```go
package main

import (
	"fmt"
	"net/http"

	"github.com/vmkteam/rpcgen/v2"
	"github.com/vmkteam/rpcgen/v2/swift"
	"github.com/vmkteam/zenrpc/v2"
	"github.com/vmkteam/zenrpc/v2/smd"
)

func main() {
	rpc := zenrpc.NewServer(zenrpc.Options{})

	gen := rpcgen.FromSMD(rpc.SMD())

	typeMapper := func(typeName string, in smd.Property, param swift.Parameter) swift.Parameter {
		switch typeName {
		case "Group":
			switch in.Name {
			case "groups":
				param.Type = fmt.Sprintf("[Int: %s]", param.Type)
				param.DecodableDefault = swift.DefaultMap
			}
		}
		return param
	}

	http.HandleFunc("/client.swift", rpcgen.Handler(gen.SwiftClient(swift.Settings{"", typeMapper})))
}
```
