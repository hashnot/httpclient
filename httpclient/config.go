package httpclient

import "github.com/hashnot/function"

type OutputConfig struct {
	OmitEmpty bool `yaml:"omitEmpty"`
	function.Message
}
