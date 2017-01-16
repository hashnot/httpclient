package httpclient

import (
	"github.com/hashnot/function"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"text/template"
)

type handler struct {
}

func (*handler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("hello"))

}

func TestGet(t *testing.T) {
	server := httptest.NewServer(&handler{})
	defer server.Close()

	task := &HttpTask{
		Source: HttpInputSpec{
			Method:  http.MethodGet,
			Address: "http://" + server.Listener.Addr().String() + "/",
		},
	}
	task.setup("test", true)
	out, err := task.do(&function.Message{})

	t.Log("Error", err)

	t.Log(out)

}

type postHandler struct {
	body []byte
}

func (h *postHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(http.StatusOK)
	body, err := ioutil.ReadAll(req.Body)
	h.body = body
	if err != nil {
		panic(err)
	}
}

func TestPost(t *testing.T) {
	h := &postHandler{}
	server := httptest.NewServer(h)
	defer server.Close()

	task := &HttpTask{
		Source: HttpInputSpec{
			Method:  http.MethodPost,
			Address: "http://" + server.Listener.Addr().String() + "/",
		},
		Output: OutputConfig{OmitEmpty: true},
	}
	task.setup("test", true)
	data := "hello"
	out, err := task.do(&function.Message{
		Body: []byte(data),
	})

	if err != nil {
		t.Error(err)
	}

	if out != nil {
		t.Error("Got non-nil output")
	}

	if string(h.body) != data {
		t.Error("Expected", data, ", got", string(h.body))
	}
}

func TestOutput(t *testing.T) {
	h := &postHandler{}
	server := httptest.NewServer(h)
	defer server.Close()

	task := &HttpTask{
		Source: HttpInputSpec{
			Method:  http.MethodPost,
			Address: "http://" + server.Listener.Addr().String() + "/",
		},
		Output: OutputConfig{
			Message: function.Message{
				Headers: map[string]interface{}{"task": "test"},
			},
		},
	}

	client := &HttpClient{
		Tasks: map[string]*HttpTask{"test": task},
	}

	client.Setup(true)

	data := "hello"
	out, err := client.Handle(&function.Message{
		Headers: map[string]interface{}{"task": "test"},
		Body:    []byte(data),
	})

	if err != nil {
		t.Error(err)
	}

	if out == nil || len(out) == 0 {
		t.Error("Got nil output", out)
	}

	outTaskName := out[0].Headers["task"].(string)
	if outTaskName != "test" {
		t.Error("Expected out task 'test' got", outTaskName)
	}
}

func TestTemplate(t *testing.T) {
	addressTempl := template.New("test")
	_, err := addressTempl.Parse("Int64: {{.Headers.shopId}}{{if ne (or .Headers.to 0) 0}} optional: {{.Headers.to}}{{end}}")
	if err != nil {
		t.Error(err)
	}

	data := make(map[string]interface{})
	data["shopId"] = int64(1)
	//data["to"] = time.Now().Unix()

	result, err := apply(addressTempl, &function.Message{
		Headers: data,
	})
	if err != nil {
		t.Error(err)
	}

	expected := "Int64: 1"
	if result != expected {
		t.Error("Expected", expected, "got", result)
	}
}
