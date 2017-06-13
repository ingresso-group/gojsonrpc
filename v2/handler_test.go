package jsonrpc

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

type fakeDispatcher struct {
	results map[interface{}]interface{}
}

func (dispatcher *fakeDispatcher) Dispatch(resp *Response, call *Call, req *http.Request) {
	result, ok := dispatcher.results[call.ID]
	if !ok {
		resp.Error = &Error{
			Code:    1234,
			Message: "id missing in results map",
		}
	}
	resp.Result = result
	return
}

func TestServeHTTP(t *testing.T) {
	dispatcher := &fakeDispatcher{
		results: map[interface{}]interface{}{
			"abc123": 6.0,
		},
	}
	handler := &Handler{dispatcher}
	server := httptest.NewServer(handler)
	defer server.Close()

	buf := bytes.NewBufferString(`{"jsonrpc": "2.0", "id": "abc123", "method": "Add", "params": [1, 2, 3]}`)

	response, _ := http.Post(server.URL, "application/json", buf)

	assert.Equal(t, "application/json", response.Header.Get("Content-Type"))
	assert.Equal(t, http.StatusOK, response.StatusCode)

	body, _ := ioutil.ReadAll(response.Body)
	var result Response
	err := json.Unmarshal(body, &result)

	if !assert.Nil(t, err) {
		t.FailNow()
	}

	assert.Equal(t, "2.0", result.Version)
	assert.Equal(t, "abc123", result.ID)
	assert.Equal(t, 6.0, result.Result)
	assert.Nil(t, result.Error)
}

func TestServeHTTP_batch(t *testing.T) {
	dispatcher := &fakeDispatcher{
		results: map[interface{}]interface{}{
			"abc123": 6.0,
			"def456": 20.0,
		},
	}
	handler := &Handler{dispatcher}
	server := httptest.NewServer(handler)
	defer server.Close()

	buf := bytes.NewBufferString(`[
		{"jsonrpc": "2.0", "id": "abc123", "method": "Add", "params": [1, 2, 3]},
		{"jsonrpc": "2.0", "id": "def456", "method": "Multiply", "params": [4, 5]}
	]`)

	response, _ := http.Post(server.URL, "application/json", buf)

	assert.Equal(t, "application/json", response.Header.Get("Content-Type"))
	assert.Equal(t, http.StatusOK, response.StatusCode)

	body, _ := ioutil.ReadAll(response.Body)
	var results []Response
	err := json.Unmarshal(body, &results)

	if !assert.Nil(t, err) {
		t.FailNow()
	}

	assert.Equal(t, "2.0", results[0].Version)
	assert.Equal(t, "abc123", results[0].ID)
	assert.Equal(t, 6.0, results[0].Result)
	assert.Nil(t, results[0].Error)

	assert.Equal(t, "2.0", results[1].Version)
	assert.Equal(t, "def456", results[1].ID)
	assert.Equal(t, 20.0, results[1].Result)
	assert.Nil(t, results[1].Error)
}

func TestServeHTTP_with_get_request(t *testing.T) {
	dispatcher := &fakeDispatcher{}
	handler := &Handler{dispatcher}
	server := httptest.NewServer(handler)
	defer server.Close()

	response, _ := http.Get(server.URL)

	assert.Equal(t, "application/json", response.Header.Get("Content-Type"))
	assert.Equal(t, http.StatusOK, response.StatusCode)

	body, _ := ioutil.ReadAll(response.Body)
	var result Response
	json.Unmarshal(body, &result)

	assert.Equal(t, result.Version, "2.0")
	assert.Nil(t, result.ID)
	assert.Nil(t, result.Result)
	assert.Equal(t, CodeInvalidRequest, result.Error.Code)
	assert.Equal(t, "jsonrpc: rpc calls should be done via a POST request", result.Error.Message)
}

func TestServeHTTP_with_bad_data(t *testing.T) {
	dispatcher := &fakeDispatcher{}
	handler := &Handler{dispatcher}
	server := httptest.NewServer(handler)
	defer server.Close()

	buf := bytes.NewBufferString(`hahahahaha nope!`)

	response, _ := http.Post(server.URL, "application/json", buf)

	assert.Equal(t, "application/json", response.Header.Get("Content-Type"))
	assert.Equal(t, http.StatusOK, response.StatusCode)

	body, _ := ioutil.ReadAll(response.Body)
	var result Response
	err := json.Unmarshal(body, &result)

	if !assert.Nil(t, err) {
		t.FailNow()
	}

	assert.Equal(t, result.Version, "2.0")
	assert.Nil(t, result.ID)
	assert.Nil(t, result.Result)
	assert.Equal(t, CodeParseError, result.Error.Code)
	assert.Equal(t, "invalid character 'h' looking for beginning of value", result.Error.Message)
}
