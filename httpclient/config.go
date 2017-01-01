package httpclient

import (
	"github.com/hashnot/function"
	"github.com/hashnot/function/amqptypes"
	"net/url"
	"github.com/rafalkrupinski/rev-api-gw/httplog"
	"bytes"
	"text/template"
	"net/http"
	"errors"
)

type HttpClient struct {
	Function *amqptypes.Configuration `yaml:"function"`
	Tasks    map[string]*HttpTask     `yaml:"tasks"`

	verbose  bool
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

type OutputConfig struct {
	OmitEmpty bool `yaml:"omitEmpty"`
	function.Message
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
