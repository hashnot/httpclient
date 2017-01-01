package httpclient

import (
	"bytes"
	"errors"
	"github.com/hashnot/function"
	"io/ioutil"
	"log"
	"net/http"
	"reflect"
)

//Input:
//http body from payload
//headers from configuration file
//http response body to output payload
func (c *HttpClient) Handle(i *function.Message) ([]*function.Message, error) {
	taskValue, ok := i.Headers["task"]
	if !ok {
		return nil, errors.New("No `task` header in message")
	}

	taskName, ok := taskValue.(string)
	if !ok {
		return nil, errors.New("`task` header not a string " + reflect.TypeOf(taskValue).String())
	}

	task, ok := c.Tasks[taskName]
	if !ok {
		return nil, errors.New("Task '" + taskName + "' not found in configuration")
	}

	output, err := task.do(i)

	var result []*function.Message

	if output != nil {
		result = append(result, output)
	}

	return result, err
}

func (task *HttpTask) do(in *function.Message) (*function.Message, error) {
	source := task.Source

	address, err := apply(source.addressTempl, in)
	if err != nil {
		return nil, err
	}

	log.Print(source.Method + " " + address)

	req, err := http.NewRequest(source.Method, address, bytes.NewReader(in.Body))
	if err != nil {
		return nil, err
	}
	// TODO headers
	hc := source.client
	resp, err := hc.Do(req)

	if err != nil {
		return nil, err
	}

	return task.responseToMessage(resp)
}

func (task *HttpTask)responseToMessage(resp *http.Response) (*function.Message, error) {
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if task.Output.OmitEmpty && len(body) == 0 {
		return nil, nil
	}

	date, _ := http.ParseTime(resp.Header.Get("Date"))

	var result = task.Output.Message

	result.Body = body
	result.Timestamp = date
	result.ContentType = resp.Header.Get("Content-Type")

	return &result, nil
}
