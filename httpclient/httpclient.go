package httpclient

import (
	"bytes"
	"errors"
	"github.com/hashnot/function"
	"github.com/hashnot/function/amqptypes"
	"github.com/rafalkrupinski/rev-api-gw/httplog"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"text/template"
	"reflect"
)

type HttpClient struct {
	Function *amqptypes.Configuration `yaml:"function"`
	Tasks    map[string]*HttpTask     `yaml:"tasks"`

	verbose  bool
}

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

func (client *HttpClient) Setup(verbose bool) error {
	for name, task := range client.Tasks {
		err := task.setup(name, verbose)
		if err != nil {
			return err
		}
	}
	return nil
}

type HttpTask struct {
	Source HttpInputSpec `yaml:"source"`
	Output OutputConfig  `yaml:"output"`
}

func (task *HttpTask) setup(name string, verbose bool) error {
	err := task.Source.parseTemplates(name)
	if err != nil {
		return err
	}

	err = task.Source.setupTransport(verbose)
	if err != nil {
		return errors.New(name + ": " + err.Error())
	}
	return nil
}

type HttpInputSpec struct {
	Method       string `yaml:"method"`
	Address      string `yaml:"address"`
	Proxy        string `yaml:"proxy"`

	client       *http.Client
	addressTempl *template.Template
}

func (spec *HttpInputSpec) parseTemplates(name string) error {
	spec.addressTempl = template.New(name + ".address")
	_, err := spec.addressTempl.Parse(spec.Address)
	return err
}

func (spec *HttpInputSpec) setupTransport(verbose bool) error {
	proxyAddr := spec.Proxy

	var transport http.RoundTripper

	if proxyAddr == "" {
		transport = http.DefaultTransport
	} else {
		proxyUrl, err := url.Parse(proxyAddr)
		if err != nil {
			return nil
		}
		transport = &http.Transport{Proxy: http.ProxyURL(proxyUrl)}
	}

	if verbose {
		transport = &httplog.LoggingRoundTripper{transport}
	}

	httpClient := &http.Client{Transport: transport}
	spec.client = httpClient
	return nil
}

func apply(t *template.Template, data *function.Message) (string, error) {
	buffer := new(bytes.Buffer)
	err := t.Execute(buffer, data)
	return buffer.String(), err
}
