package jsonrpc

import (
	"bytes"
	"encoding/json"
)

var (
	// IndentOutput controls wether or not JSON output is indented.
	IndentOutput = true
	// EscapeHTML controls wether json output is HTML escaped.
	EscapeHTML = false
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
