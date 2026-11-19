package clo

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"testing"
)

func TestResourceImportContracts(t *testing.T) {
	resources := Provider().ResourcesMap
	for name, r := range resources {
		if name == "clo_compute_instance_action" {
			if r.Importer != nil {
				t.Error("action receipts cannot be recovered from CLO")
			}
			continue
		}
		if r.Importer == nil || r.Importer.StateContext == nil {
			t.Errorf("%s has no context importer", name)
		}
	}
	for _, name := range []string{"clo_disks_volume", "clo_network_ip", "clo_compute_keypair", "clo_compute_snapshot_restore", "clo_compute_instance_license"} {
		t.Run(name, func(t *testing.T) {
			r := resources[name]
			for _, id := range []string{"", "object", "/object", "parent/", "a/b/c", "parent/object"} {
				d := schema.TestResourceDataRaw(t, r.Schema, nil)
				d.SetId(id)
				state, err := r.Importer.StateContext(t.Context(), d, nil)
				if id != "parent/object" {
					if err == nil {
						t.Errorf("accepted invalid identifier %q", id)
					}
					continue
				}
				if err != nil || len(state) != 1 || state[0].Id() != "object" {
					t.Fatalf("import %s: %v", id, err)
				}
			}
		})
	}
}
