package jsonrpc

import (
	"bytes"
	"context"
	"encoding/json"
)

var (
	// IndentOutput controls wether or not JSON output is indented.
	IndentOutput = true
	// EscapeHTML controls wether json output is HTML escaped.
	EscapeHTML = false
)

type key int

const (
	methodNameKey key = iota
)

// Marshal is a custom json marshaller that conditionally turns off html
// escaping and applies indentation.
func Marshal(v interface{}) ([]byte, error) {

	buf := bytes.NewBuffer([]byte{})
	e := json.NewEncoder(buf)
	if IndentOutput {
		e.SetIndent("", "  ")
	}
	if EscapeHTML {
		e.SetEscapeHTML(false)
	}

	err := e.Encode(v)

	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// setMethod saves a method name into a context
func setMethod(ctx context.Context, method string) context.Context {
	return context.WithValue(ctx, methodNameKey, method)
}

// getMethod gets the method name from the context
func getMethod(ctx context.Context) (string, bool) {
	method, ok := ctx.Value(methodNameKey).(string)
	return method, ok
}
