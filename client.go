package jsonrpc

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type Client struct {
	addr       string
	httpclient *http.Client
	Debug      bool
}

func NewClient(addr string) (client *Client) {
	client = new(Client)
	client.addr = addr
	client.httpclient = new(http.Client)
	return
}

func (client Client) Call(method string, params interface{}, results interface{}) (err error) {
	request := requestData{
		Id:      "1",
		Version: "2.0",
		Params:  params,
		Method:  method,
	}
	reqBytes, err := json.Marshal(request)
	if err != nil {
		return err
	}

	if client.Debug {
		fmt.Printf("REQUEST: %s\n", string(reqBytes))
	}
	req, err := http.NewRequest("POST", client.addr, bytes.NewBuffer(reqBytes))

	resp, err := client.httpclient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if client.Debug {
		fmt.Printf("RESPONSE: %s\n", string(body))
	}

	var response responseData
	err = json.Unmarshal(body, &response)
	if err != nil {
		return err
	}

	if response.Error != nil {
		return fmt.Errorf("jsonrpc: got error in response with code %d: %s",
			response.Error.Code,
			response.Error.Message,
		)
	}

	jdata, err := json.Marshal(&response.Result)

	if err != nil {
		return err
	}

	err = json.Unmarshal(jdata, results)
	if err != nil {
		return err
	}

	return
}
