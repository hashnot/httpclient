package httptask

type HttpInputMessage struct {
	Task    string                 `json:"task"`
	Headers map[string][]string    `json:"headers"`
	Payload string                 `json:"payload,string"`
	Data    map[string]interface{} `json:"data"`
}

type HttpOutputMessage struct {
	Headers map[string][]string `json:"headers,omitempty"`
	Payload []byte              `json:"payload,omitempty,string"`
}
