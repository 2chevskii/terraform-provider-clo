package clo

import (
	"fmt"
	"github.com/clo-ru/terraform-provider-clo/v2/internal/cloapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLimitAndLicenseRefresh(t *testing.T) {
	p := Provider()
	if err := p.InternalValidate(); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name, path, response, key string
		cfg                       map[string]interface{}
		want                      interface{}
	}{
		{"clo_project_limits", "/v2/projects/p/limits", `{"count":2,"result":[{"name":"servers","max":0,"enabled":true},{"name":"addresses","max":20,"enabled":true}]}`, "limits.servers", map[string]interface{}{"project_id": "p", "limits": map[string]interface{}{"servers": 3}}, 0},
		{"clo_account_limits", "/v2/limits/projects", `{"count":2,"result":[{"project_id":"51b30da6-64b5-44c1-a2ed-e7d5102835bb","name":"disk-space","max":24,"enabled":true},{"project_id":"64f6e71a-ce5b-40e6-bbce-749187f170f0","name":"disk-space","max":17,"enabled":true}]}`, "disk_space.51b30da6-64b5-44c1-a2ed-e7d5102835bb", map[string]interface{}{"disk_space": map[string]interface{}{"51b30da6-64b5-44c1-a2ed-e7d5102835bb": 16}}, 24},
		{"clo_compute_instance_license", "/v2/licenses/p", `{"result":{"id":"p","addon":"app","name":"new","value":3,"status":"ACTIVE"}}`, "value", map[string]interface{}{"instance_id": "s", "addon": "app", "name": "old"}, 3},
	} {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != "GET" || r.URL.Path != tc.path {
					t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
				}
				w.Header().Set("Content-Type", "application/json")
				fmt.Fprint(w, tc.response)
			}))
			defer srv.Close()
			cli, err := cloapi.New("test", srv.URL)
			if err != nil {
				t.Fatal(err)
			}
			res := p.ResourcesMap[tc.name]
			d := schema.TestResourceDataRaw(t, res.Schema, tc.cfg)
			d.SetId("p")
			if diagnostics := res.ReadContext(t.Context(), d, &providerMeta{v3: cli}); diagnostics.HasError() {
				t.Fatal(diagnostics)
			}
			if got := d.Get(tc.key); got != tc.want {
				t.Fatalf("%s got %#v want %#v", tc.key, got, tc.want)
			}
		})
	}
}
