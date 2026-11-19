package cloapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestLimitLicenseAndActionRequests(t *testing.T) {
	ctx := t.Context()
	zero := 0
	tests := []struct {
		method, path string
		body         map[string]interface{}
		run          func(*Client) error
	}{
		{"PATCH", "/v2/projects/p/limits", map[string]interface{}{"disk-space": float64(0)}, func(c *Client) error { return c.UpdateProjectLimits(ctx, "p", map[string]int{"disk-space": 0}) }},
		{"PATCH", "/v2/limits", map[string]interface{}{"servers": map[string]interface{}{"p": float64(0), "q": float64(3)}}, func(c *Client) error {
			return c.UpdateAccountLimits(ctx, map[string]map[string]int{"servers": {"p": 0, "q": 3}})
		}},
		{"POST", "/v2/servers/s/licenses", map[string]interface{}{"addon": "app", "name": "basic", "value": float64(0)}, func(c *Client) error {
			id, err := c.AddServerLicense(ctx, "s", "app", "basic", &zero)
			if err == nil && id != "l" {
				return fmt.Errorf("unexpected ID %q", id)
			}
			return err
		}},
		{"PATCH", "/v2/licenses/l", map[string]interface{}{"name": "pro"}, func(c *Client) error { return c.UpdateLicense(ctx, "l", "pro", nil) }},
		{"DELETE", "/v2/licenses/l", nil, func(c *Client) error { return c.DeleteLicense(ctx, "l") }},
		{"POST", "/v2/servers/s/rescue", nil, func(c *Client) error { return c.RescueServer(ctx, "s") }},
		{"POST", "/v2/servers/s/reboot", nil, func(c *Client) error { return c.RebootServer(ctx, "s") }},
	}
	for _, tc := range tests {
		t.Run(tc.method+tc.path, func(t *testing.T) {
			calls := 0
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls++
				if r.Method != tc.method || r.URL.Path != tc.path {
					t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
				}
				var payload map[string]interface{}
				if r.ContentLength > 0 {
					if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
						t.Error(err)
					}
				}
				if !reflect.DeepEqual(payload, tc.body) {
					t.Errorf("payload %#v, want %#v", payload, tc.body)
				}
				w.Header().Set("Content-Type", "application/json")
				switch {
				case tc.path == "/v2/servers/s/licenses":
					w.WriteHeader(201)
					fmt.Fprint(w, `{"result":{"id":"l"}}`)
				case tc.method == "DELETE":
					w.WriteHeader(204)
				default:
					w.WriteHeader(202)
					fmt.Fprint(w, "{}")
				}
			}))
			defer srv.Close()
			cli, err := New("test", srv.URL)
			if err != nil {
				t.Fatal(err)
			}
			if err := tc.run(cli); err != nil {
				t.Fatal(err)
			}
			if calls != 1 {
				t.Errorf("expected one request, got %d", calls)
			}
		})
	}
}
