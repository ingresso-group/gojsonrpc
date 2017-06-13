package jsonrpc

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClient_call(t *testing.T) {
	client := NewClient()
	dispatcher := NewMapDispatcher()
	dispatcher.Register("add", func(resp *Response, call *Call, req *http.Request) {
		var params []int
		err := json.Unmarshal(call.Params, &params)
		assert.Nil(t, err)
		assert.Equal(t, []int{1, 2, 3}, params)

		resp.Result = 6
	})

	server := httptest.NewServer(&Handler{dispatcher})
	defer server.Close()

	var result int
	err := client.Call(server.URL, "add", []int{1, 2, 3}, &result)
	assert.Nil(t, err)
	assert.Equal(t, result, 6)
}

func TestClient_with_error(t *testing.T) {
	client := NewClient()
	dispatcher := NewMapDispatcher()
	dispatcher.Register("add", func(resp *Response, call *Call, req *http.Request) {
		resp.Error = &Error{
			Code:    CodeInvalidParameters,
			Message: "nope!",
		}
	})

	server := httptest.NewServer(&Handler{dispatcher})
	defer server.Close()

	var result int
	err := client.Call(server.URL, "add", "one and two", &result)
	assert.Equal(t, "jsonrpc: nope! (-32602)", err.Error())
}

func TestClient_with_invalid_json_response(t *testing.T) {
	client := NewClient()

	server := httptest.NewServer(http.DefaultServeMux)
	defer server.Close()

	var result int
	err := client.Call(server.URL, "add", []int{1, 2, 3}, &result)
	assert.NotNil(t, err)
}

func TestClient_with_non_pointer(t *testing.T) {
	client := NewClient()
	dispatcher := NewMapDispatcher()
	dispatcher.Register("add", func(resp *Response, call *Call, req *http.Request) {
		var params []int
		err := json.Unmarshal(call.Params, &params)
		assert.Nil(t, err)
		assert.Equal(t, []int{1, 2, 3}, params)

		resp.Result = 6
	})

	server := httptest.NewServer(&Handler{dispatcher})
	defer server.Close()

	var result int
	err := client.Call(server.URL, "add", []int{1, 2, 3}, result)
	assert.NotNil(t, err)
}

func TestClient_with_incorrect_result_type(t *testing.T) {
	client := NewClient()
	dispatcher := NewMapDispatcher()
	dispatcher.Register("add", func(resp *Response, call *Call, req *http.Request) {
		var params []int
		err := json.Unmarshal(call.Params, &params)
		assert.Nil(t, err)
		assert.Equal(t, []int{1, 2, 3}, params)

		resp.Result = 6
	})

	server := httptest.NewServer(&Handler{dispatcher})
	defer server.Close()

	var result string
	err := client.Call(server.URL, "add", []int{1, 2, 3}, result)
	assert.NotNil(t, err)
}

func TestClient_do(t *testing.T) {
	client := NewClient()
	dispatcher := NewMapDispatcher()
	dispatcher.Register("add", func(resp *Response, call *Call, req *http.Request) {
		var params []int
		err := json.Unmarshal(call.Params, &params)
		assert.Nil(t, err)
		assert.Equal(t, []int{1, 2, 3}, params)

		resp.Result = 6
	})

	server := httptest.NewServer(&Handler{dispatcher})
	defer server.Close()

	req, err := NewRequest(server.URL, "add", []int{1, 2, 3})
	assert.Nil(t, err)

	var result int
	err = client.Do(req, &result)
	assert.Nil(t, err)
	assert.Equal(t, result, 6)
}

func TestClient_batch(t *testing.T) {
	client := NewClient()
	dispatcher := NewMapDispatcher()
	dispatcher.Register("add", func(resp *Response, call *Call, req *http.Request) {
		var params []int
		var result = 0
		err := json.Unmarshal(call.Params, &params)
		assert.Nil(t, err)
		for _, n := range params {
			result = result + n
		}
		resp.Result = result
	})

	dispatcher.Register("multiply", func(resp *Response, call *Call, req *http.Request) {
		var params []int
		var result = 1
		err := json.Unmarshal(call.Params, &params)
		assert.Nil(t, err)
		for _, n := range params {
			result = result * n
		}
		resp.Result = result
	})

	server := httptest.NewServer(&Handler{dispatcher})
	defer server.Close()

	batch := NewBatch()

	var a, b, c int
	batch.AddCall("add", []int{1, 2, 3}, &a)
	batch.AddCall("multiply", []int{4, 5, 6}, &b)
	batch.AddCall("add", []int{7, 8, 9}, &c)

	err := client.Batch(server.URL, batch)
	assert.Nil(t, err)
	assert.Equal(t, 6, a)
	assert.Equal(t, 120, b)
	assert.Equal(t, 24, c)
}

func TestClient_batch_ignoring_errors(t *testing.T) {
	client := NewClient()
	dispatcher := NewMapDispatcher()
	dispatcher.Register("add", func(resp *Response, call *Call, req *http.Request) {
		var params []int
		var result = 0
		err := json.Unmarshal(call.Params, &params)
		assert.Nil(t, err)
		for _, n := range params {
			result = result + n
		}
		resp.Result = result
	})

	server := httptest.NewServer(&Handler{dispatcher})
	defer server.Close()

	batch := NewBatch()
	batch.DiscardErrors = true

	var a, b, c int
	batch.AddCall("add", []int{1, 2, 3}, &a)
	batch.AddCall("multiply", []int{4, 5, 6}, &b)
	batch.AddCall("add", []int{7, 8, 9}, &c)

	err := client.Batch(server.URL, batch)
	assert.Nil(t, err)
	assert.Equal(t, 6, a)
	assert.Equal(t, 0, b)
	assert.Equal(t, 24, c)
}

func TestClient_batch_with_error(t *testing.T) {
	client := NewClient()
	dispatcher := NewMapDispatcher()
	dispatcher.Register("add", func(resp *Response, call *Call, req *http.Request) {
		var params []int
		var result = 0
		err := json.Unmarshal(call.Params, &params)
		assert.Nil(t, err)
		for _, n := range params {
			result = result + n
		}
		resp.Result = result
	})

	server := httptest.NewServer(&Handler{dispatcher})
	defer server.Close()

	batch := NewBatch()
	batch.DiscardErrors = false

	var a, b, c int
	batch.AddCall("add", []int{1, 2, 3}, &a)
	batch.AddCall("multiply", []int{4, 5, 6}, &b)
	batch.AddCall("add", []int{7, 8, 9}, &c)

	err := client.Batch(server.URL, batch)
	assert.NotNil(t, err)
}

func TestClient_do_batch(t *testing.T) {
	client := NewClient()
	dispatcher := NewMapDispatcher()
	dispatcher.Register("add", func(resp *Response, call *Call, req *http.Request) {
		var params []int
		var result = 0
		err := json.Unmarshal(call.Params, &params)
		assert.Nil(t, err)
		for _, n := range params {
			result = result + n
		}
		resp.Result = result
	})

	dispatcher.Register("multiply", func(resp *Response, call *Call, req *http.Request) {
		var params []int
		var result = 1
		err := json.Unmarshal(call.Params, &params)
		assert.Nil(t, err)
		for _, n := range params {
			result = result * n
		}
		resp.Result = result
	})

	server := httptest.NewServer(&Handler{dispatcher})
	defer server.Close()

	batch := NewBatch()

	var a, b, c int
	batch.AddCall("add", []int{1, 2, 3}, &a)
	batch.AddCall("multiply", []int{4, 5, 6}, &b)
	batch.AddCall("add", []int{7, 8, 9}, &c)

	req, err := batch.NewRequest(server.URL)
	assert.Nil(t, err)

	err = client.DoBatch(req, batch)
	assert.Nil(t, err)
	assert.Equal(t, 6, a)
	assert.Equal(t, 120, b)
	assert.Equal(t, 24, c)
}

func TestNewRequest(t *testing.T) {
	req, err := NewRequest("https://foobar.com", "add", []int{1, 2, 3})
	assert.Nil(t, err)
	assert.Equal(t, "application/json", req.Header.Get("Content-Type"))
	body, err := ioutil.ReadAll(req.Body)
	assert.JSONEq(t, `{"jsonrpc": "2.0", "id": "1", "method": "add", "params": [1, 2, 3]}`, string(body))
}

func TestMethodCall(t *testing.T) {
	dispatcher := NewMapDispatcher()
	dispatcher.Register("add", func(resp *Response, call *Call, req *http.Request) {
		var params []int
		err := json.Unmarshal(call.Params, &params)
		assert.Nil(t, err)
		assert.Equal(t, []int{1, 2, 3}, params)

		resp.Result = 6
	})

	server := httptest.NewServer(&Handler{dispatcher})
	defer server.Close()

	var result int
	err := MethodCall(server.URL, "add", []int{1, 2, 3}, &result)
	assert.Nil(t, err)
	assert.Equal(t, result, 6)
}

func TestDo(t *testing.T) {
	dispatcher := NewMapDispatcher()
	dispatcher.Register("add", func(resp *Response, call *Call, req *http.Request) {
		var params []int
		err := json.Unmarshal(call.Params, &params)
		assert.Nil(t, err)
		assert.Equal(t, []int{1, 2, 3}, params)

		resp.Result = 6
	})

	server := httptest.NewServer(&Handler{dispatcher})
	defer server.Close()

	req, err := NewRequest(server.URL, "add", []int{1, 2, 3})
	assert.Nil(t, err)

	var result int
	err = Do(req, &result)
	assert.Nil(t, err)
	assert.Equal(t, result, 6)
}

func TestMethodBatch(t *testing.T) {
	dispatcher := NewMapDispatcher()
	dispatcher.Register("add", func(resp *Response, call *Call, req *http.Request) {
		var params []int
		var result = 0
		err := json.Unmarshal(call.Params, &params)
		assert.Nil(t, err)
		for _, n := range params {
			result = result + n
		}
		resp.Result = result
	})

	server := httptest.NewServer(&Handler{dispatcher})
	defer server.Close()

	batch := NewBatch()

	var a, b, c int
	batch.AddCall("add", []int{1, 2, 3}, &a)
	batch.AddCall("add", []int{4, 5, 6}, &b)
	batch.AddCall("add", []int{7, 8, 9}, &c)

	err := MethodBatch(server.URL, batch)
	assert.Nil(t, err)
	assert.Equal(t, 6, a)
	assert.Equal(t, 15, b)
	assert.Equal(t, 24, c)
}

func TestDoBatch(t *testing.T) {
	dispatcher := NewMapDispatcher()
	dispatcher.Register("add", func(resp *Response, call *Call, req *http.Request) {
		var params []int
		var result = 0
		err := json.Unmarshal(call.Params, &params)
		assert.Nil(t, err)
		for _, n := range params {
			result = result + n
		}
		resp.Result = result
	})

	server := httptest.NewServer(&Handler{dispatcher})
	defer server.Close()

	batch := NewBatch()

	var a, b, c int
	batch.AddCall("add", []int{1, 2, 3}, &a)
	batch.AddCall("add", []int{4, 5, 6}, &b)
	batch.AddCall("add", []int{7, 8, 9}, &c)

	req, err := batch.NewRequest(server.URL)

	err = DoBatch(req, batch)
	assert.Nil(t, err)
	assert.Equal(t, 6, a)
	assert.Equal(t, 15, b)
	assert.Equal(t, 24, c)
}
