package jsonrpc

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMapDispatcher(t *testing.T) {
	dispatcher := NewMapDispatcher()
	dispatcher.Register("add", func(resp *Response, call *Call, req *http.Request) {
		var result int
		var params []int
		json.Unmarshal(call.Params, &params)
		for _, n := range params {
			result = result + n
		}
		resp.Result = result
	})
	dispatcher.Register("multiply", func(resp *Response, call *Call, req *http.Request) {
		result := 1
		var params []int
		json.Unmarshal(call.Params, &params)
		for _, n := range params {
			result = result * n
		}
		resp.Result = result
	})

	call := &Call{
		Method: "multiply",
		Params: json.RawMessage("[2, 3, 4]"),
	}
	resp := NewResponse(call)

	dispatcher.Dispatch(resp, call, nil)
	assert.Equal(t, 24, resp.Result)

	call = &Call{
		Method: "add",
		Params: json.RawMessage("[2, 3, 4]"),
	}
	resp = NewResponse(call)

	dispatcher.Dispatch(resp, call, nil)
	assert.Equal(t, 9, resp.Result)

	call = &Call{
		Method: "divide",
		Params: json.RawMessage("[2, 3]"),
	}
	resp = NewResponse(call)

	dispatcher.Dispatch(resp, call, nil)
	assert.Equal(t, CodeMethodNotFound, resp.Error.Code)
}

func TestMapDispatcher_conflict(t *testing.T) {
	dispatcher := NewMapDispatcher()
	err := dispatcher.Register("foo", func(resp *Response, call *Call, req *http.Request) {
		resp.Result = "bar"
	})

	assert.Nil(t, err)

	err = dispatcher.Register("foo", func(resp *Response, call *Call, req *http.Request) {
		resp.Result = "bar"
	})

	assert.NotNil(t, err)
}

func TestRegister(t *testing.T) {
	method := func(resp *Response, call *Call, req *http.Request) {
		var result string
		json.Unmarshal(call.Params, &result)
		resp.Result = result
	}
	Register("echo", method)

	call := &Call{
		Method: "echo",
		Params: json.RawMessage(`"hello world"`),
	}
	resp := NewResponse(call)

	DefaultDispatcher.Dispatch(resp, call, nil)
	assert.Equal(t, "hello world", resp.Result)
}

func TestDispatch(t *testing.T) {
	method := func(resp *Response, call *Call, req *http.Request) {
		var result string
		json.Unmarshal(call.Params, &result)
		resp.Result = result
	}
	DefaultDispatcher.Register("echo", method)

	call := &Call{
		Method: "echo",
		Params: json.RawMessage(`"hello world"`),
	}
	resp := NewResponse(call)

	Dispatch(resp, call, nil)
	assert.Equal(t, "hello world", resp.Result)
}
