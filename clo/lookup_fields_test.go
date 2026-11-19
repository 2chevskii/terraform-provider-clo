package clo

import (
	"fmt"
	"github.com/clo-ru/terraform-provider-clo/v2/internal/cloapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLookupResponseFields(t *testing.T) {
	cases := []struct {
		name, key, body string
		checks          map[string]interface{}
	}{
		{"clo_compute_instance", "id", `{"result":{"id":"s","flavor":{"cpu_type":"SHARED","disk":9},"image":{"min_ram":4,"operation_system":{"os_family":"Linux"}},"snapshots":["snap"],"primary_address":null,"switch_status":"ON"}}`, map[string]interface{}{"cpu_type": "SHARED", "flavor_disk": 9, "image_min_ram": 4, "os_family": "Linux", "snapshots.0": "snap", "switch_status": "ON"}},
		{"clo_disks_volume", "volume_id", `{"result":{"id":"v","file_system":"ext4","image":{"min_disk":12},"recipe":{"name":"test"},"licenses":[{"dynamic":[true,3]}]}}`, map[string]interface{}{"file_system": "ext4", "image_min_disk": 12, "recipe_name": "test", "licenses_json": `[{"dynamic":[true,3]}]`}},
		{"clo_network_ip", "address_id", `{"result":{"id":"a","mac_addr":"AA:BB","external":true,"bill_item_id":13,"version":6,"bandwidth_max_mbps":1024}}`, map[string]interface{}{"mac_addr": "AA:BB", "external": true, "bill_item_id": 13, "version": 6, "bandwidth": 1024}},
		{"clo_dbaas_backups", "project_id", `{"count":1,"result":[{"datastore":{"id":"store"},"databases":[{"id":"db","name":"app","admin_username":"owner"}]}]}`, map[string]interface{}{"result.0.datastore_id": "store", "result.0.databases.0.admin_username": "owner"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				fmt.Fprint(w, tc.body)
			}))
			defer srv.Close()
			c, err := cloapi.New("test", srv.URL)
			if err != nil {
				t.Fatal(err)
			}
			ds := Provider().DataSourcesMap[tc.name]
			d := schema.TestResourceDataRaw(t, ds.Schema, map[string]interface{}{tc.key: "id"})
			if diags := ds.ReadContext(t.Context(), d, &providerMeta{v3: c}); diags.HasError() {
				t.Fatal(diags)
			}
			for field, want := range tc.checks {
				if got := d.Get(field); got != want {
					t.Errorf("%s: got %#v want %#v", field, got, want)
				}
			}
		})
	}
}
