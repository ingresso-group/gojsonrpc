package jsonrpc

import (
	"encoding/json"
)

// Call contains a target method and the parameters that should be deserilised
// for that call.
type Call struct {
	Version string          `json:"jsonrpc"`
	ID      interface{}     `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params"`
}
