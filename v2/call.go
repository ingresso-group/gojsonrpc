package jsonrpc

import (
	"encoding/json"
)

// Call contains a target method and the parameters that should be deserilised
// for that call.
//
// Params may not be present
//
// Calls without ID are not expecting a Response and may be safely ignored.
type Call struct {
	Version string          `json:"jsonrpc"`
	ID      interface{}     `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params"`
}

// UnmarshalParams unmarshals the calls parameters into the given interface.
func (call Call) UnmarshalParams(v interface{}) error {
	return json.Unmarshal(call.Params, v)
}

type clientCall struct {
	Version string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params"`
}
