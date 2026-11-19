package clo

import (
	"reflect"
	"testing"

	"github.com/clo-ru/terraform-provider-clo/v2/internal/cloapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestServerDeletionPolicy(t *testing.T) {
	srv := &cloapi.Server{Addresses: []string{"ip"}, Disks: []cloapi.ServerDisk{{ID: "local", StorageType: "local"}, {ID: "vol", StorageType: "volume"}}}
	for _, tc := range []struct {
		name               string
		raw                map[string]interface{}
		addresses, volumes []string
		clear              bool
	}{
		{"legacy", nil, []string{"ip"}, []string{"vol"}, false},
		{"retain", map[string]interface{}{"deletion": []interface{}{map[string]interface{}{}}}, []string{}, []string{}, false},
		{"selected", map[string]interface{}{"deletion": []interface{}{map[string]interface{}{"volume_ids": []interface{}{"vol", "local", "other"}, "clear_fstab": true}}}, []string{}, []string{"vol"}, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			for _, r := range []*schema.Resource{resourceInstance(), resourceSnapshotRestore()} {
				d := schema.TestResourceDataRaw(t, r.Schema, tc.raw)
				addresses, volumes, clear := serverDeletion(d, srv)
				if !reflect.DeepEqual(addresses, tc.addresses) || !reflect.DeepEqual(volumes, tc.volumes) || clear != tc.clear {
					t.Fatalf("got addresses %v volumes %v clear_fstab %v", addresses, volumes, clear)
				}
			}
		})
	}
}
