package jsonrpc

import (
	"fmt"
	"net/http"
	"sync"
)

// DefaultDispatcher is the default Dispatcher used for Register and Dispatch
var DefaultDispatcher = NewMapDispatcher()

// A Dispatcher processes JSONRPC calls from a request.
//
// HandleCall should return a Result based on the method and parameters of the
// Call.
//
// When a call fails Dispatch should return a Result containing an Error
//
// When a method cannot be found that matches the requested the Method in the
// Call then Dispatch should return a Result containing an Error with the
// code -32601
type Dispatcher interface {
	Dispatch(*Response, *Call, *http.Request)
}

// Method is the target for the MapDispatcher
type Method func(*Response, *Call, *http.Request)

// MapDispatcher holds a map of methods and will dispatch based on method name
type MapDispatcher struct {
	methods map[string]Method
	mtx     *sync.Mutex
}

// NewMapDispatcher returns a pointer to a MapDispatcher intialised with no
// methods
func NewMapDispatcher() *MapDispatcher {
	dispatcher := &MapDispatcher{
		methods: make(map[string]Method),
		mtx:     new(sync.Mutex),
	}
	return dispatcher
}

// Register a method with the dispatcher
func (dispatcher *MapDispatcher) Register(name string, method Method) error {
	dispatcher.mtx.Lock()
	defer dispatcher.mtx.Unlock()

	_, ok := dispatcher.methods[name]

	if ok {
		return fmt.Errorf("jsonrpc: unable to register method with name %s as it already exists", name)
	}

	dispatcher.methods[name] = method
	return nil
}

// Dispatch looks for the methods with the given name in the methods map and
// if found calls it with the original parameters.
//
// When the method is not found, it returns an error.
func (dispatcher *MapDispatcher) Dispatch(resp *Response, call *Call, req *http.Request) {
	method, ok := dispatcher.methods[call.Method]
	if !ok {
		resp.Error = &Error{
			Code:    CodeMethodNotFound,
			Message: fmt.Sprintf("jsonrpc: method with name %s not register", call.Method),
		}
		return
	}

	method(resp, call, req)
}

// Register adds the method to the DefaultDispatcher
func Register(name string, method Method) error {
	err := DefaultDispatcher.Register(name, method)
	return err
}

// Dispatch a call with the DefaultDispatcher
func Dispatch(resp *Response, call *Call, req *http.Request) {
	DefaultDispatcher.Dispatch(resp, call, req)
}
