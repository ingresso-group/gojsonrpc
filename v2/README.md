V2
==

this is a rewrite of V1 as it was bloated and unidiomatic.

Examples
--------

the package has a default handler similar to the http.DefaultHandler

```golang
package main

import (
    "encoding/json"
    "net/http"

    "github.com/ingresso-group/gojsonprc/v2"
)

func Add (resp *jsonrpc.Response, call *jsonrpc.Call, req *http.Request) {
    var result = 0
    var params []int
    json.Unmarshal(call.Params, &params)
    for _, n := range params {
        result = result + n
    }
    resp.Result = result
}

func Multiply (resp *jsonrpc.Response, call *jsonrpc.Call, req *http.Request) {
    var result = 1
    var params []int
    json.Unmarshal(call.Params, &params)
    for _, n := range params {
        result = result * n
    }
    resp.Result = result
}

func main () {
    jsonrpc.Register("add", Add)
    jsonrpc.Register("multiply", Multiply)
    jsonrpc.ListenAndServe("localhost:8000")
}
```

You can use a custom dispatcher if you want to do something differently

```golang
package main

import (
    "net/http"

    "github.com/ingresso-group/gojsonprc/v2"
)

type myDispatcher struct {
    prefix string 
}

func (d *myDispatcher) Dispatch (resp *jsonrpc.Response, call *jsonrpc.Call, req *http.Request) {
    resp.Result = fmt.Sprintf("%s: method %s params: %s", d.prefix, call.Method, call.Params)
}

func main () {
    dispatcher := myDispatcher{"my dispatcher"}
    handler := jsonrpc.Handler{dispatcher}
        
    http.ListenAndServe("localhost:8000", handler)
}
```
