package sse

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/universal-tool-calling-protocol/go-utcp/src/providers/base"
	. "github.com/universal-tool-calling-protocol/go-utcp/src/providers/sse"
)

func TestSSEClientTransport_RegisterAndCall(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/tools":
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"version":"1.0","tools":[{"name":"echo","description":"Echo"}]}`))
		case r.Method == http.MethodPost && r.URL.Path == "/echo":
			var in map[string]interface{}
			json.NewDecoder(r.Body).Decode(&in)
			out, _ := json.Marshal(map[string]interface{}{"result": in["msg"]})
			w.Header().Set("Content-Type", "application/json")
			w.Write(out)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	prov := &SSEProvider{
		BaseProvider: BaseProvider{Name: "sse", ProviderType: ProviderSSE},
		URL:          server.URL + "/tools",
	}
	tr := NewSSETransport(nil)
	ctx := context.Background()
	tools, err := tr.RegisterToolProvider(ctx, prov)
	if err != nil {
		t.Fatalf("register error: %v", err)
	}
	if len(tools) != 1 || tools[0].Name != "echo" {
		t.Fatalf("unexpected tools: %+v", tools)
	}

	prov.URL = server.URL
	res, err := tr.CallTool(ctx, "echo", map[string]interface{}{"msg": "hi"}, prov, nil)
	if err != nil {
		t.Fatalf("call error: %v", err)
	}
	m, ok := res.(map[string]interface{})
	if !ok || m["result"] != "hi" {
		t.Fatalf("unexpected result: %#v", res)
	}
}

func TestSSEDeregister(t *testing.T) {
	sse := NewSSETransport(nil)
	sh := &SSEProvider{BaseProvider: BaseProvider{Name: "s", ProviderType: ProviderSSE}}
	if err := sse.DeregisterToolProvider(context.Background(), sh); err != nil {
		t.Fatalf("sse deregister error: %v", err)
	}
}
