package httppost

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/cofxlabs/cofx/functiondriver/go/spec"
	"github.com/cofxlabs/cofx/service/resource"
	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"
)

func TestPost(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// request
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.Equal(t, "Bearer 12345", r.Header.Get("Authorization"))
		assert.Equal(t, "application/vnd.github+json", r.Header.Get("Accept"))
		var buff bytes.Buffer
		io.Copy(&buff, r.Body)
		v := gjson.Get(buff.String(), "base").String()
		assert.Equal(t, "master", v)

		// response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(`{"status": "ok"}`))
	}))
	defer server.Close()

	_, ep, custom := New()

	bundle := spec.EntrypointBundle{
		Version: "latest",
		Custom:  custom(),
		Resources: resource.Resources{
			Logwriter:    os.Stdout,
			CronTrigger:  nil,
			HttpTrigger:  nil,
			OutputParser: nil,
			Labels:       nil,
		},
	}
	bundle.Custom.(*Custom).client = server.Client()

	args := spec.EntrypointArgs{
		"url":             server.URL,
		"set_headers":     "Accept: application/vnd.github+json, Authorization: Bearer 12345",
		"json_file_path":  "./testdata/data.json",
		"query_json_path": "status",
	}

	rets, err := ep(context.Background(), bundle, args)
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	assert.Equal(t, "200", rets["status_code"])
	assert.Equal(t, "ok", rets["status"])
}
