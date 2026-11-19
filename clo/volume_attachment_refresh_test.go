package clo

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/clo-ru/terraform-provider-clo/v2/internal/cloapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestVolumeAttachmentRefresh(t *testing.T) {
	for _, tc := range []struct{ owner, wantID, wantDevice string }{
		{"server-a", "vol", "/dev/vdb"},
		{"server-b", "", ""},
	} {
		t.Run(tc.owner, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != "GET" || r.URL.Path != "/v2/volumes/vol/detail" {
					t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
				}
				w.Header().Set("Content-Type", "application/json")
				fmt.Fprintf(w, `{"result":{"id":"vol","attached_to_server":{"id":%q,"device":"/dev/vdb"}}}`, tc.owner)
			}))
			defer server.Close()
			cli, err := cloapi.New("test", server.URL)
			if err != nil {
				t.Fatal(err)
			}
			res := resourceVolumeAttach()
			d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{"volume_id": "vol", "instance_id": "server-a"})
			d.SetId("vol")
			if diags := res.ReadContext(t.Context(), d, &providerMeta{v3: cli}); diags.HasError() {
				t.Fatal(diags)
			}
			if d.Id() != tc.wantID || d.Get("device") != tc.wantDevice {
				t.Errorf("got ID %q device %q", d.Id(), d.Get("device"))
			}
		})
	}
}
