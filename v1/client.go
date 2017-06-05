package jsonrpc

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type clientRequestData struct {
	Id      string      `json:"id"`
	Version string      `json:"jsonrpc"`
	Params  interface{} `json:"params"`
	Method  string      `json:"method"`
}

type Client struct {
	addr       string
	httpclient *http.Client
	Debug      bool
}

func NewClient(addr string) (client *Client) {
	client = &Client{
		addr:       addr,
		httpclient: new(http.Client),
	}
	return
}

func (client Client) Call(method string, username string, password string,
	params interface{}, results interface{}) (err error) {
	request := clientRequestData{
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
	req.Header.Set("Content-Type", "application/json")
	if username != "" || password != "" {
		req.SetBasicAuth(username, password)
	}
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
