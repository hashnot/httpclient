package httpclient

import (
	"bytes"
	"errors"
	"github.com/hashnot/function"
	"io/ioutil"
	"log"
	"net/http"
	"reflect"
	"strings"
	"time"
)

//Input:
//http body from payload
//headers from configuration file
//http response body to output payload
func (c *HttpClient) Handle(i *function.Message, p function.Publisher) error {
	taskValue, ok := i.Headers["task"]
	if !ok {
		return errors.New("No `task` header in message")
	}

	taskName, ok := taskValue.(string)
	if !ok {
		return errors.New("`task` header not a string " + reflect.TypeOf(taskValue).String())
	}

	task, ok := c.Tasks[taskName]
	if !ok {
		return errors.New("Task '" + taskName + "' not found in configuration")
	}

	if limit := task.Source.RateLimit; limit != nil {
		limiter, err := limit.Get(i)
		if err != nil {
			return err
		}
		limitRate(limiter)
	}

	err := task.do((*httpMessage)(i), p)

	return err
}

type httpMessage function.Message

func (in *httpMessage) newRequest(source *HttpInputSpec) (*http.Request, error) {
	address, err := source.Address.Apply(in)
	if err != nil {
		return nil, err
	}

	log.Print(source.Method + " " + address)

	req, err := http.NewRequest(source.Method, address, bytes.NewReader(in.Body))

	if contentType := in.ContentType; contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}

	if ts := in.Timestamp; (ts != time.Time{}) {
		req.Header.Set("Date", ts.Format(http.TimeFormat))
	}

	return req, err
}

func (task *HttpTask) do(in *httpMessage, p function.Publisher) error {
	source := task.Source
	req, err := in.newRequest(&source)
	if err != nil {
		return err
	}
	// TODO headers
	hc := source.client
	resp, err := hc.Do(req)

	if err != nil {
		return err
	}

	return task.responseToMessage(resp, p)
}

func (task *HttpTask) responseToMessage(resp *http.Response, p function.Publisher) error {
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if task.Output.OmitEmpty && len(body) == 0 {
		return nil
	}

	date, _ := http.ParseTime(resp.Header.Get("Date"))

	var result = p.NewFrom(task.Output.Template)

	result.Body = body
	result.Timestamp = date
	result.ContentType = resp.Header.Get("Content-Type")

	for k, v := range resp.Header {
		result.Headers["http."+k] = strings.Join(v, ", ")
	}

	p.Publish(result)
	return nil
}
