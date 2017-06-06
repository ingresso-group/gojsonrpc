package jsonrpc

import (
	"fmt"
	"net/http"
	"sync"
)

// DefaultDispatcher is the default Dispatcher used for Register and Dispatch
var DefaultDispatcher = NewMapDispatcher()

// A Dispatcher handles individual JSONRPC method calls.
//
// The Call argument represents the part of the request concerning this
// particular method call.
//
// The target method should write it's response in the Result field of the
// Response argument. There is no need to return anything.
//
// When a call fails Dispatch should write an Error to the Error field of the
// Response arguments. There is no need to return anything.
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
			Message: fmt.Sprintf("jsonrpc: method with name %s not registered", call.Method),
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
