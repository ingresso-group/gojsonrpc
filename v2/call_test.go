package jsonrpc

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCall_UnmarshalParams(t *testing.T) {
	call := Call{
		Params: json.RawMessage(`{"foo": "bar"}`),
	}

	var params struct {
		Foo string `json:"foo"`
	}

	err := call.UnmarshalParams(&params)

	if assert.Nil(t, err) {
		assert.Equal(t, params.Foo, "bar")
	}
}

func TestCall_UnmarshalParams_with_bad_data(t *testing.T) {
	call := Call{
		Params: json.RawMessage(`hahahahahahah`),
	}

	var params struct {
		Foo string `json:"foo"`
	}

	err := call.UnmarshalParams(&params)

	assert.NotNil(t, err)
}
