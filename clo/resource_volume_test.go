package clo

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/clo-ru/terraform-provider-clo/v2/internal/cloapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

const (
	volumeName = "volume_1"
)

func TestAccCloVolume_basic(t *testing.T) {
	var volume cloapi.Volume
	const address = "clo_disks_volume." + volumeName
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccCloPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckVolumeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCloVolumeBasic(volumeName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVolumeExists(address, &volume),
				),
			},
			{Config: testAccCloVolumeBasic("volume_renamed"), Check: func(state *terraform.State) error {
				var current cloapi.Volume
				if err := testAccCheckVolumeExists(address, &current)(state); err != nil {
					return err
				}
				if current.ID != volume.ID || current.Name != "volume_renamed" {
					return fmt.Errorf("expected renamed volume %s, got %s named %q", volume.ID, current.ID, current.Name)
				}
				return nil
			}},
			{ResourceName: address, ImportState: true, ImportStateIdFunc: func(*terraform.State) (string, error) { return getTestProject() + "/" + volume.ID, nil }, ImportStateCheck: func(states []*terraform.InstanceState) error {
				if len(states) != 1 || states[0].ID != volume.ID || states[0].Attributes["project_id"] != getTestProject() || states[0].Attributes["name"] != "volume_renamed" {
					return fmt.Errorf("import did not recover the renamed volume and its project")
				}
				return nil
			}},
		},
	})
}

func testAccCloVolumeBasic(name string) string {
	return fmt.Sprintf(`resource "clo_disks_volume" "%s"{
			project_id = "%s"
			name = "%s"
			size = 10
	}`, volumeName, os.Getenv("CLO_API_PROJECT_ID"), name)
}

func testAccCheckVolumeExists(n string, volumeItem *cloapi.Volume) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("volume with ID is not set")
		}
		cli := testAccProvider.Meta().(*providerMeta).v3
		vol, e := cli.GetVolume(context.Background(), rs.Primary.ID)
		if e != nil {
			return e
		}
		*volumeItem = *vol
		return nil
	}
}

func testAccCheckVolumeDestroy(st *terraform.State) error {
	cli := testAccProvider.Meta().(*providerMeta).v3
	for _, rs := range st.RootModule().Resources {
		if rs.Type != "clo_disks_volume" {
			continue
		}
		_, e := cli.GetVolume(context.Background(), rs.Primary.ID)
		if cloapi.IsNotFound(e) {
			continue
		}
		if e != nil {
			return e
		}
		return fmt.Errorf("%s %s still exists", rs.Type, rs.Primary.ID)
	}
	return nil
}
