package httpclient

import (
	"errors"
	"github.com/hashnot/function/amqptypes"
	"github.com/rafalkrupinski/rev-api-gw/httplog"
	"net/http"
	"net/url"
)

type HttpClient struct {
	Function *amqptypes.Configuration `yaml:"function"`
	Tasks    map[string]*HttpTask     `yaml:"tasks"`

	verbose bool
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
	Template  *amqptypes.PublishingTemplate
}

type HttpTask struct {
	Source HttpInputSpec
	Output OutputConfig
}

func (task *HttpTask) setup(name string, verbose bool) error {
	err := task.Source.setupTransport(verbose)
	if err != nil {
		return errors.New(name + ": " + err.Error())
	}
	return nil
}

type HttpInputSpec struct {
	Method    string
	Address   TemplateWrapper
	Proxy     string
	RateLimit *RateLimitSpec `yaml:"rateLimit"`

	client *http.Client
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
