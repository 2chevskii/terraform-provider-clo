package clo

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/clo-ru/terraform-provider-clo/v2/internal/cloapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestAPIDataSources(t *testing.T) {
	provider := Provider()
	tests := []struct {
		name, path, body string
		input            map[string]interface{}
		key              string
		want             interface{}
	}{
		{"account_balance", "/v2/balance", `{"result":{"balance":12.25,"cost":{"month":7.5}}}`, nil, "result.0.cost.0.month", 7.5},
		{"project", "/v2/projects/p/detail", `{"result":{"name":"example","description":null}}`, map[string]interface{}{"project_id": "p"}, "result.0.name", "example"},
		{"available_licenses", "/v2/licenses", `{"count":1,"result":[{"addon":"x","items":[{"name":"y"}],"fields":{"choices":[1,false]}}]}`, nil, "result.0.fields_json", `{"choices":[1,false]}`},
		{"maintenance_mode", "/v2/maintenance-mode", `{"enable":true,"message":"maintenance"}`, nil, "result.0.enable", true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != "GET" || r.URL.Path != tc.path {
					t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
				}
				w.Header().Set("Content-Type", "application/json")
				fmt.Fprint(w, tc.body)
			}))
			defer srv.Close()
			cli, err := cloapi.New("test", srv.URL)
			if err != nil {
				t.Fatal(err)
			}
			ds := provider.DataSourcesMap["clo_"+tc.name]
			d := schema.TestResourceDataRaw(t, ds.Schema, tc.input)
			if diags := ds.ReadContext(t.Context(), d, &providerMeta{v3: cli}); diags.HasError() {
				t.Fatal(diags)
			}
			if got := d.Get(tc.key); got != tc.want {
				t.Fatalf("%s = %#v, want %#v", tc.key, got, tc.want)
			}
		})
	}
}

func TestAccountLimitsPageTraversal(t *testing.T) {
	for _, explicit := range []bool{false, true} {
		t.Run(fmt.Sprint(explicit), func(t *testing.T) {
			calls := 0
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls++
				q := r.URL.Query()
				if q.Get("order") != "-id" {
					t.Errorf("missing order on page %d", calls)
				}
				if calls == 2 && (q.Get("offset") != "1" || q.Get("limit") != "1") {
					t.Errorf("unexpected next page %v", q)
				}
				w.Header().Set("Content-Type", "application/json")
				fmt.Fprintf(w, `{"count":2,"result":[{"id":%d}]}`, 3-calls)
			}))
			defer srv.Close()
			cli, err := cloapi.New("test", srv.URL)
			if err != nil {
				t.Fatal(err)
			}
			ds := Provider().DataSourcesMap["clo_account_limits"]
			query := map[string]interface{}{"order": "-id"}
			want := 2
			if explicit {
				query["limit"] = "1"
				want = 1
			}
			d := schema.TestResourceDataRaw(t, ds.Schema, map[string]interface{}{"query": query})
			if diags := ds.ReadContext(t.Context(), d, &providerMeta{v3: cli}); diags.HasError() {
				t.Fatal(diags)
			}
			if calls != want || d.Get("result.#") != want || d.Get("total_count") != 2 {
				t.Errorf("calls=%d rows=%v total=%v", calls, d.Get("result.#"), d.Get("total_count"))
			}
		})
	}
}
