package cloapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestServerCreateFromVolumeAndRetainAttachments(t *testing.T) {
	var calls int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		var got map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
			t.Error(err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		switch r.Method + " " + r.URL.Path {
		case "POST /v2/projects/project/servers":
			want := map[string]interface{}{"name": "demo", "volume": "boot", "user_data": "aGk=", "flavor": map[string]interface{}{"ram": float64(2), "vcpus": float64(1), "cpu_type": "SHARED"}}
			if !reflect.DeepEqual(want, got) {
				t.Errorf("create body %#v", got)
			}
			w.WriteHeader(http.StatusAccepted)
			_, _ = w.Write([]byte(`{"result":{"id":"server"}}`))
		case "DELETE /v2/servers/server":
			want := map[string]interface{}{"delete_addresses": []interface{}{}, "delete_volumes": []interface{}{"vol"}, "clear_fstab": true}
			if !reflect.DeepEqual(want, got) {
				t.Errorf("delete body %#v", got)
			}
			w.WriteHeader(http.StatusAccepted)
		default:
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
		}
	}))
	defer srv.Close()
	cli, err := New("test", srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	id, err := cli.CreateServer(t.Context(), ServerCreateParams{ProjectID: "project", Name: "demo", VolumeID: "boot", UserData: "aGk=", CPUType: "SHARED", FlavorRam: 2, FlavorVcpus: 1})
	if err != nil || id != "server" {
		t.Fatalf("create returned ID %q, err %v", id, err)
	}
	if err := cli.DeleteServer(t.Context(), id, []string{}, []string{"vol"}, true); err != nil {
		t.Fatal(err)
	}
	if calls != 2 {
		t.Errorf("received %d requests", calls)
	}
}
