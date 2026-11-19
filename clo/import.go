package clo

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"strings"
)

// scopedImporter accepts a parent ID followed by the managed object's ID where
// the detail endpoint cannot recover the parent during refresh.
func scopedImporter(parent string) *schema.ResourceImporter {
	return &schema.ResourceImporter{StateContext: func(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
		scope, id, ok := strings.Cut(d.Id(), "/")
		if !ok || scope == "" || id == "" || strings.Contains(id, "/") {
			return nil, fmt.Errorf("expected %s/object_id", parent)
		}
		if err := d.Set(parent, scope); err != nil {
			return nil, err
		}
		d.SetId(id)
		return []*schema.ResourceData{d}, nil
	}}
}
