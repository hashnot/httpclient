package httpclient

import (
	"bitbucket.org/hashnot/httpclient/httptask"
	"bytes"
	"errors"
	"github.com/hashnot/function"
	"github.com/hashnot/function/amqptypes"
	"github.com/rafalkrupinski/rev-api-gw/httplog"
	"github.com/streadway/amqp"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"text/template"
	"log"
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
func (c *HttpClient) Handle(i function.InputMessage, out function.OutChan) error {
	input := &httptask.HttpInputMessage{}
	err := i.DecodeBody(input)
	if err != nil {
		return err
	}

	wrapped := &httpInputMessage{
		HttpInputMessage: input,
		delivery:         i.Body(),
	}

	taskName := input.Task
	task, ok := c.Tasks[taskName]

	if !ok {
		return errors.New("Task '" + taskName + "' not found in configuration")
	}

	output, err := c.do(task, wrapped)
	if err != nil {
		return err
	}

	out <- &function.OutputMessage{output}

	return nil
}

func (c *HttpClient) do(task *HttpTask, in *httpInputMessage) (*httptask.HttpOutputMessage, error) {
	source := task.Source

	address, err := apply(source.addressTempl, in.Data)
	if err != nil {
		return nil, err
	}

	log.Print(source.Method + " " + address)

	req, err := http.NewRequest(source.Method, address, in.Body())
	if err != nil {
		return nil, err
	}
	// TODO headers
	hc := source.client
	resp, err := hc.Do(req)

	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	result := &httptask.HttpOutputMessage{
		Headers: resp.Header,
		Payload: body,
	}

	return result, err
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
	Source *HttpInputSpec `yaml:"source"`
	*amqptypes.Output `yaml:"output"`
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

func apply(t *template.Template, data map[string]interface{}) (string, error) {
	buffer := bytes.Buffer{}
	err := t.Execute(&buffer, data)
	return buffer.String(), err
}

type httpInputMessage struct {
	*httptask.HttpInputMessage
	delivery *amqp.Delivery
}

func (i *httpInputMessage) ContentType() string {
	return i.delivery.ContentType
}

func (i *httpInputMessage) Body() io.Reader {
	return bytes.NewBufferString(i.Payload)
}
