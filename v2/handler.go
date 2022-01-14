package jsonrpc

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
)

// DefaultHandler is the default Handler for serving requests quickly
var DefaultHandler = &Handler{
	Dispatcher: DefaultDispatcher,
}

// Handler provides the interface between http requests and the Dispatcher
type Handler struct {
	Dispatcher Dispatcher
}

func serverError(w http.ResponseWriter, message string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	result := Response{
		Version: "2.0",
		ID:      nil,
		Error: &Error{
			Code:    code,
			Message: message,
		},
	}

	resp, err := Marshal(&result)

	if err != nil {
		w.Write([]byte(fmt.Sprintf(`{"jsonrpc": "2.0", "id": null, "error": {"code": %d, "message": "something went wrong!"}}`, CodeInternalError)))
		return
	}

	w.Write(resp)
}

func (handler *Handler) dispatch(resp *Response, call *Call, req *http.Request, wg *sync.WaitGroup) {
	defer wg.Done()
	handler.Dispatcher.Dispatch(resp, call, req)
}

// ServeHTTP handles converting a http.Request into a Calls. Implements the
// http.Handler interface
func (handler *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		serverError(w, "jsonrpc: rpc calls should be done via a POST request", CodeInvalidRequest)
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		serverError(w, err.Error(), CodeInvalidRequest)
		return
	}

	var calls []*Call
	var single bool

	err = json.Unmarshal(body, &calls)

	if err != nil {
		var call *Call

		serr := json.Unmarshal(body, &call)

		if serr != nil {
			serverError(w, err.Error(), CodeParseError)
			return
		}
		single = true
		calls = append(calls, call)
	}

	var responses []*Response
	var wg sync.WaitGroup

	known_ids := make([]interface{}, 0)
CALLS_LOOP:
	for _, call := range calls {
		resp := NewResponse(call)
		responses = append(responses, resp)
		if call.ID != nil {
			for _, b := range known_ids {
				if call.ID == b {
					resp.Error = &Error{}
					resp.Error.Code = CodeInvalidRequest
					resp.Error.Message = "The 'id' element is not unique"
					continue CALLS_LOOP
				}
			}
		}
		wg.Add(1)
		go handler.dispatch(resp, call, r, &wg)
		known_ids = append(known_ids, call.ID)
	}

	wg.Wait()

	var data []byte
	if single {
		data, err = Marshal(responses[0])
	} else {
		data, err = Marshal(responses)
	}

	if err != nil {
		serverError(w, err.Error(), CodeInternalError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(data)

	return
}

// ListenAndServe sets up the DefaultHandler to listen on the address given.
func ListenAndServe(addr string) error {
	return http.ListenAndServe(addr, DefaultHandler)
}
