package clo

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/clo-ru/terraform-provider-clo/v2/internal/cloapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestLegacyProjectPagination(t *testing.T) {
	for _, tc := range []struct {
		name                    string
		explicit, repeat, empty bool
		wantCalls, wantRows     int
	}{
		{"all", false, false, false, 2, 3}, {"one_page", true, false, false, 1, 2}, {"repeat", false, true, false, 2, 0}, {"premature_empty", false, false, true, 2, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			calls := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls++
				q := r.URL.Query()
				if q.Get("order") != "name" {
					t.Error("lost ordering on page", calls)
				}
				if calls == 2 && (q.Get("offset") != "2" || q.Get("limit") != "2") {
					t.Errorf("wrong page %v", q)
				}
				w.Header().Set("Content-Type", "application/json")
				page := `[{"id":"1","name":"one"},{"id":"2","name":"two"}]`
				if calls == 2 && !tc.repeat {
					page = `[{"id":"3","name":"three"}]`
					if tc.empty {
						page = "[]"
					}
				}
				fmt.Fprintf(w, `{"count":3,"result":%s}`, page)
			}))
			defer server.Close()
			cli, err := cloapi.New("test", server.URL)
			if err != nil {
				t.Fatal(err)
			}
			r := dataSourceProjects()
			query := map[string]interface{}{"order": "name"}
			if tc.explicit {
				query["limit"] = "2"
			}
			d := schema.TestResourceDataRaw(t, r.Schema, map[string]interface{}{"query": query})
			diagnostics := r.ReadContext(t.Context(), d, &providerMeta{v3: cli})
			if diagnostics.HasError() != (tc.repeat || tc.empty) {
				t.Fatal(diagnostics)
			}
			if calls != tc.wantCalls || d.Get("result.#") != tc.wantRows {
				t.Fatalf("calls %d rows %v", calls, d.Get("result.#"))
			}
			if !diagnostics.HasError() && d.Get("total_count") != 3 {
				t.Errorf("total %v", d.Get("total_count"))
			}
		})
	}
}
