package jsonrpc

import (
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBatch(t *testing.T) {
	batch := NewBatch()

	var a, b, c int

	batch.AddCall("add", []int{1, 2, 3}, &a)
	batch.AddCall("multiply", []int{4, 5, 6}, &b)
	batch.AddCall("add", []int{7, 8, 9}, &c)

	req, err := batch.NewRequest("https://foobar.com")

	assert.Nil(t, err)
	assert.Equal(t, "application/json", req.Header.Get("Content-Type"))
	body, err := ioutil.ReadAll(req.Body)
	assert.JSONEq(t, `[
		{"jsonrpc": "2.0", "id": "1", "method": "add", "params": [1, 2, 3]},
		{"jsonrpc": "2.0", "id": "2", "method": "multiply", "params": [4, 5, 6]},
		{"jsonrpc": "2.0", "id": "3", "method": "add", "params": [7, 8, 9]}
	]`, string(body))
}

func TestBatch_no_calls(t *testing.T) {
	batch := NewBatch()
	_, err := batch.NewRequest("https://foobar.com")
	assert.NotNil(t, err)
}
