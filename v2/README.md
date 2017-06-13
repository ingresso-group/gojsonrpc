V2
==

this is a rewrite of V1 as it was bloated and unidiomatic.

Install
-------

```bash
$ go get github.com/ingresso-group/gojsonrpc
```

Client Examples
---------------
```golang
package main

import (
	"log"

	"github.com/ingresso-group/gojsonrpc/v2"
)

func main() {
	var a int
	err := jsonrpc.MethodCall("https://foobar.com", "add", []int{1, 2, 3}, &a)

	if err != nil {
		log.Fatal(err)
	}

	log.Printf("result of the add call is %d", a)
}
```

batch calls are also supported:

```golang
package main

import (
	"log"

	"github.com/ingresso-group/gojsonrpc/v2"
)

func main() {
	var a, b, c int

	batch := jsonrpc.NewBatch()
	batch.AddCall("add", []int{1, 2, 3}, &a)
	batch.AddCall("multiply", []int{4, 5, 6}, &b)
	batch.AddCall("add", []int{7, 8, 9}, &c)

	err := jsonrpc.MethodBatch("https://foobar.com", batch)

	if err != nil {
		log.Fatal(err)
	}

	log.Printf("results of the batch call is a=%d, b=%d, c=%d", a, b, c)
}
```

Both regular and batch requests can expose the underlying `http.Request`
before making the actual call allowing for adding headers/logging/etc:


```golang 
package main

import (
	"log"

	"github.com/ingresso-group/gojsonrpc/v2"
)

func main() {
	var a int

	req, err := jsonrpc.NewRequest("https://foobar.com", "add", []int{1, 2, 3})
	if err != nil {
		log.Fatal(err)
	}

	req.Header.Set("X-Request-ID", "abc123")

	err = jsonrpc.Do(req, &a)

	if err != nil {
		log.Fatal(err)
	}

	log.Printf("result of the add call is %d", a)
}
```

```golang
package main

import (
	"log"

	"github.com/ingresso-group/gojsonrpc/v2"
)

func main() {
	var a, b, c int

	batch := jsonrpc.NewBatch()
	batch.AddCall("add", []int{1, 2, 3}, &a)
	batch.AddCall("multiply", []int{4, 5, 6}, &b)
	batch.AddCall("add", []int{7, 8, 9}, &c)

	req, err := batch.NewRequest("https://foobar.com")

	if err != nil {
		log.Fatal(err)
	}

	req.SetBasicAuth("foo", "bar")

	err = jsonrpc.DoBatch(req, batch)

	if err != nil {
		log.Fatal(err)
	}

	log.Printf("results of the batch call is a=%d, b=%d, c=%d", a, b, c)
}
```


Server Examples
---------------

the package has a default handler similar to the http.DefaultHandler

```golang
package main

import (
	"encoding/json"
	"net/http"

	"github.com/ingresso-group/gojsonrpc/v2"
)

func Add(resp *jsonrpc.Response, call *jsonrpc.Call, req *http.Request) {
	var result = 0
	var params []int
	json.Unmarshal(call.Params, &params)
	for _, n := range params {
		result = result + n
	}
	resp.Result = result
}

func Multiply(resp *jsonrpc.Response, call *jsonrpc.Call, req *http.Request) {
	var result = 1
	var params []int
	json.Unmarshal(call.Params, &params)
	for _, n := range params {
		result = result * n
	}
	resp.Result = result
}

func main() {
	jsonrpc.Register("add", Add)
	jsonrpc.Register("multiply", Multiply)
	jsonrpc.ListenAndServe("localhost:8000")
}
```

You can use a custom dispatcher if you want to do something differently

```golang
package main

import (
	"fmt"
	"net/http"

	"github.com/ingresso-group/gojsonrpc/v2"
)

type myDispatcher struct {
	prefix string
}

func (d *myDispatcher) Dispatch(resp *jsonrpc.Response, call *jsonrpc.Call, req *http.Request) {
	resp.Result = fmt.Sprintf("%s: method %s params: %s", d.prefix, call.Method, call.Params)
}

func main() {
	dispatcher := &myDispatcher{"my dispatcher"}
	handler := &jsonrpc.Handler{dispatcher}

	http.ListenAndServe("localhost:8000", handler)
}
```

[prometheus](github.com/prometheus/client_golang) example:

```golang
package main

import (
	"encoding/json"
	"net/http"

	"github.com/ingresso-group/gojsonrpc/v2"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	methodCalls = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "rpc_method_calls_total",
			Help: "The number of method calls made to the service",
		},
		[]string{"method"},
	)
)

func init() {
	prometheus.MustRegister(methodCalls)
}

type myDispatcher struct {
	dispatcher *jsonrpc.MapDispatcher
}

func (d *myDispatcher) Dispatch(resp *jsonrpc.Response, call *jsonrpc.Call, req *http.Request) {
	methodCalls.WithLabelValues(call.Method).Add(1)
	d.dispatcher.Dispatch(resp, call, req)
}

func (d *myDispatcher) Register(name string, method jsonrpc.Method) error {
	return d.dispatcher.Register(name, method)
}

func Add(resp *jsonrpc.Response, call *jsonrpc.Call, req *http.Request) {
	var result = 0
	var params []int
	json.Unmarshal(call.Params, &params)
	for _, n := range params {
		result = result + n
	}
	resp.Result = result
}

func Multiply(resp *jsonrpc.Response, call *jsonrpc.Call, req *http.Request) {
	var result = 1
	var params []int
	json.Unmarshal(call.Params, &params)
	for _, n := range params {
		result = result * n
	}
	resp.Result = result
}

func main() {
	dispatcher := &myDispatcher{
		dispatcher: jsonrpc.NewMapDispatcher(),
	}
	dispatcher.Register("add", Add)
	dispatcher.Register("multiply", Multiply)

	rpc := &jsonrpc.Handler{dispatcher}
	http.Handle("/rpc", rpc)
	http.Handle("/metrics", promhttp.Handler())

	http.ListenAndServe("localhost:8000", nil)
}
```

You could also wrap the handler as it just implements the http.Handler
interface, which gives you access both to the original request and the
individual method calls.

```golang
package main

import (
	"encoding/json"
	"net/http"

	"github.com/ingresso-group/gojsonrpc/v2"
)

func Add(resp *jsonrpc.Response, call *jsonrpc.Call, req *http.Request) {
	var result = 0
	var params []int
	json.Unmarshal(call.Params, &params)
	for _, n := range params {
		result = result + n
	}
	resp.Result = result
}

func Multiply(resp *jsonrpc.Response, call *jsonrpc.Call, req *http.Request) {
	var result = 1
	var params []int
	json.Unmarshal(call.Params, &params)
	for _, n := range params {
		result = result * n
	}
	resp.Result = result
}

func main() {

	jsonrpc.Register("add", Add)
	jsonrpc.Register("multiply", Multiply)

	http.HandleFunc("/rpc", func(w http.ResponseWriter, r *http.Request) {
		username, password, ok := r.BasicAuth()

		if !ok || (username != "foo" && password != "bar") {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("you are not foo!"))
			return
		}

		jsonrpc.DefaultHandler.ServeHTTP(w, r)
	})

	http.ListenAndServe("localhost:8000", nil)
}
```

JSON Output
-----------

By default the server will output indented json and will not convert certain
characters to uft-8 escape codes.

The server's JSON output can be controlled `jsonrpc.IndentOutput` and
`jsonrpc.EscapeHTML`.


```golang

func init() {
    jsonrpc.IndentOutput = false
    jsonrpc.EscapeHTML = true
}

func main() {
    jsonrpc.ListenAndServe(":8080")
}
```

Error handling in batch requests
--------------------------------

By default the client will discard errors in the results of a batch request.

If you would prefer to have errors returned you can enable this functionality
on batch by batch basis:

```golang
batch := jsonrpc.NewBatch()
batch.DiscardErrors = false
```
