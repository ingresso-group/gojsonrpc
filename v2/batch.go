package jsonrpc

import (
	"bytes"
	"errors"
	"net/http"
	"strconv"
	"sync"
)

type batchCall struct {
	call   *clientCall
	result interface{}
	id     string
}

// Batch represents a collection of method calls that will be sent to the
// server in a single HTTP call.
type Batch struct {
	order         []*batchCall
	calls         map[string]*batchCall
	id            int
	mtx           *sync.Mutex
	DiscardErrors bool
}

func (batch *Batch) nextID() string {
	id := strconv.Itoa(batch.id)
	batch.id = batch.id + 1
	return id
}

func (batch *Batch) resultForID(rawid interface{}) (interface{}, bool) {

	id, ok := rawid.(string)
	if !ok {
		return nil, false
	}
	call, ok := batch.calls[id]
	if !ok {
		return nil, false
	}

	return call.result, true
}

// AddCall adds a call to a batch
func (batch *Batch) AddCall(method string, params interface{}, result interface{}) {
	batch.mtx.Lock()
	defer batch.mtx.Unlock()

	id := batch.nextID()

	call := &clientCall{
		Version: "2.0",
		ID:      id,
		Method:  method,
		Params:  params,
	}

	batchCall := &batchCall{
		call:   call,
		result: result,
		id:     id,
	}

	batch.order = append(batch.order, batchCall)
	batch.calls[id] = batchCall
}

// NewRequest returns a pointer to a new http.Request containing the calls in
// the JSONRPC format
//
// the request can then be executed with Client.DoBatch
func (batch *Batch) NewRequest(url string) (*http.Request, error) {
	batch.mtx.Lock()
	defer batch.mtx.Unlock()

	if len(batch.order) == 0 {
		return nil, errors.New("jsonrpc: no calls in batch, cowardly refusing to send empty request")
	}

	var calls []*clientCall

	for _, v := range batch.order {
		calls = append(calls, v.call)
	}

	data, err := Marshal(calls)

	if err != nil {
		return nil, err
	}

	buf := bytes.NewBuffer(data)

	req, err := http.NewRequest(http.MethodPost, url, buf)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	return req, nil
}

// NewBatch creates a new empty batch for making multiple method calls in one
// HTTP request.
func NewBatch() *Batch {
	batch := &Batch{
		calls:         make(map[string]*batchCall),
		id:            1,
		mtx:           new(sync.Mutex),
		DiscardErrors: true,
	}
	return batch
}
