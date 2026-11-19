package cloapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestProjectLifecycleRequests(t *testing.T) {
	type request struct {
		method, path string
		body         map[string]interface{}
	}
	var got []request
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]interface{}
		if r.ContentLength > 0 {
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Error(err)
			}
		}
		got = append(got, request{r.Method, r.URL.Path, body})
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == "DELETE":
			w.WriteHeader(204)
		case r.URL.Path == "/v2/projects":
			w.WriteHeader(201)
			fmt.Fprint(w, `{"result":{"id":"p"}}`)
		case r.Method == "GET":
			fmt.Fprint(w, `{"result":{"id":"p","display_name":"new","description":"text","switch_status":"ON","status":"ACTIVE"}}`)
		default:
			w.WriteHeader(202)
			fmt.Fprint(w, "{}")
		}
	}))
	defer server.Close()
	cli, err := New("test", server.URL)
	if err != nil {
		t.Fatal(err)
	}
	ctx := t.Context()
	name, limit := "old", 0
	id, err := cli.CreateProject(ctx, &name, nil, &limit)
	if err != nil || id != "p" {
		t.Fatalf("create %q %v", id, err)
	}
	name = "new"
	if err := cli.UpdateProject(ctx, id, &name, nil); err != nil {
		t.Fatal(err)
	}
	for _, enabled := range []bool{false, true} {
		if err := cli.SetProjectEnabled(ctx, id, enabled); err != nil {
			t.Fatal(err)
		}
	}
	p, err := cli.GetProject(ctx, id)
	if err != nil || p.DisplayName != "new" || p.Description != "text" || p.SwitchStatus != "ON" {
		t.Fatalf("detail %#v %v", p, err)
	}
	if err := cli.DeleteProject(ctx, id); err != nil {
		t.Fatal(err)
	}
	want := []request{{"POST", "/v2/projects", map[string]interface{}{"display_name": "old", "limit_name": float64(0)}}, {"PATCH", "/v2/projects/p", map[string]interface{}{"display_name": "new"}}, {"POST", "/v2/projects/p/stop", nil}, {"POST", "/v2/projects/p/start", nil}, {"GET", "/v2/projects/p/detail", nil}, {"DELETE", "/v2/projects/p", nil}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected requests %#v", got)
	}
}
