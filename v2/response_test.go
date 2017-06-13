package jsonrpc

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestError(t *testing.T) {
	err := &Error{
		Code:    CodeInvalidParameters,
		Message: "number1 is not a number",
		Data:    "number1 == \"abc\"",
	}

	errorStr := err.Error()

	assert.Equal(t, errorStr, "jsonrpc: number1 is not a number (-32602)")

}
