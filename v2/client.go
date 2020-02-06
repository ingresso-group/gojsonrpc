package jsonrpc

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"go.opencensus.io/trace"
)

var (
	// DefaultClient is the client that fulfills the the Call method
	DefaultClient = NewClient()
)

// Client is a JSONRPC client that faciliates the calling of methods on a
// JSONRPC server.
type Client struct {
	HTTPClient *http.Client
}

// NewClient creates a new client that makes use of the http.DefaultClient as
// it's underlying HTTPClient
func NewClient() *Client {
	client := &Client{
		HTTPClient: http.DefaultClient,
	}
	return client
}

func (client *Client) do(req *http.Request, result interface{}) error {

	rawresp, err := client.HTTPClient.Do(req)

	if err != nil {
		return err
	}

	body, err := ioutil.ReadAll(rawresp.Body)

	if err != nil {
		return err
	}

	var resp clientResponse

	err = json.Unmarshal(body, &resp)

	if err != nil {
		return err
	}

	if resp.Error != nil {
		return resp.Error
	}

	err = json.Unmarshal(resp.Result, result)

	if err != nil {
		return err
	}
	return nil
}

func (client *Client) doBatch(req *http.Request, batch *Batch) error {
	rawresp, err := client.HTTPClient.Do(req)

	if err != nil {
		return err
	}

	body, err := ioutil.ReadAll(rawresp.Body)

	if err != nil {
		return err
	}

	var responses []*clientResponse

	err = json.Unmarshal(body, &responses)
	if err != nil {
		return err
	}

	for _, resp := range responses {
		result, ok := batch.resultForID(resp.ID)
		if !ok {
			if batch.DiscardErrors {
				continue
			}
			err := fmt.Errorf("jsonrpc: unable to find a call with the response ID %s", resp.ID)
			return err
		}

		if resp.Error != nil {
			if batch.DiscardErrors {
				continue
			}
			return resp.Error
		}

		err = json.Unmarshal(resp.Result, result)
		if err != nil {
			if batch.DiscardErrors {
				continue
			}
			return err
		}
	}
	return nil
}

// Do executes a http.Request and attempts to deserialise the response to the
// the given result argument.
//
// While this function will execute any http.Request object, it's a good idea to
// use the NewRequest function to generate a http.Request in the correct format
// for your method call.
func (client *Client) Do(req *http.Request, result interface{}) error {
	ctx, span := trace.StartSpan(req.Context(), "(*jsonrpc.Client).Do")
	defer span.End()

	methodName, ok := getMethod(ctx)
	if ok {
		span.SetName("(*jsonrpc.Client).%s" methodName)
		span.AddAttributes(
			trace.StringAttribute("rpc.method", methodName),
		)
	}

	return client.do(req.WithContext(ctx), result)
}

// Batch executes a batch request and attempts to deserialise the response to
// the appropreate result argument provided when creating the calls.
func (client *Client) Batch(url string, batch *Batch) error {
	req, err := batch.NewRequest(url)
	if err != nil {
		return err
	}
	return client.doBatch(req, batch)
}

// DoBatch executes a batch request and attempts to deserialise the response to
// the appropreate result argument provided when creating the calls.
func (client *Client) DoBatch(req *http.Request, batch *Batch) error {
	return client.doBatch(req, batch)
}

// Call makes a single JSONRPC request to the server
func (client *Client) Call(url string, method string, params interface{}, result interface{}) error {

	req, err := NewRequest(url, method, params)

	if err != nil {
		return err
	}

	return client.do(req, result)
}

// NewRequest wraps NewRequestWithContext using the background context
func NewRequest(url string, method string, params interface{}) (*http.Request, error) {
	return NewRequestWithContext(context.Background(), url, method, params)
}

// NewRequestWithContext returns a pointer to a new http.Request containing the call in
// the JSONRPC format.
//
// The request can then be executed with Client.Do.
func NewRequestWithContext(ctx context.Context, url string, method string, params interface{}) (*http.Request, error) {

	call := &clientCall{
		Version: "2.0",
		ID:      "1",
		Method:  method,
		Params:  params,
	}

	data, err := Marshal(call)

	if err != nil {
		return nil, err
	}

	buf := bytes.NewBuffer(data)

	ctx = setMethod(ctx, method)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, buf)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	return req, nil
}

// MethodCall makes a single call using the DefaultClient
func MethodCall(url string, method string, params interface{}, result interface{}) error {
	return DefaultClient.Call(url, method, params, result)
}

// Do executes a http.Request using the DefaultClient
func Do(req *http.Request, result interface{}) error {
	return DefaultClient.Do(req, result)
}

// MethodBatch executes a Batch using the DefaultClient
func MethodBatch(url string, batch *Batch) error {
	return DefaultClient.Batch(url, batch)
}

// DoBatch executes a http.Request using the DefaultClient and attempts to
// deserialise it as a Batch of methods calls.
func DoBatch(req *http.Request, batch *Batch) error {
	return DefaultClient.DoBatch(req, batch)
}
