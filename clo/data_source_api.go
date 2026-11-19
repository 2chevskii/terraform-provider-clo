package clo

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/clo-ru/terraform-provider-clo/v2/internal/cloapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func computedField(kind schema.ValueType) *schema.Schema {
	return &schema.Schema{Type: kind, Computed: true}
}
func recordFields(fields map[string]*schema.Schema) *schema.Schema {
	return &schema.Schema{Type: schema.TypeList, Computed: true, Elem: &schema.Resource{Schema: fields}}
}

func listQuerySchema() *schema.Schema {
	return &schema.Schema{Type: schema.TypeMap, Optional: true, Elem: &schema.Schema{Type: schema.TypeString}, Description: "API query parameters, including filters and order where supported. Set limit or offset to read a single page."}
}
func listOptions(d *schema.ResourceData) *cloapi.ListOptions {
	query := map[string]string{}
	for key, value := range d.Get("query").(map[string]interface{}) {
		query[key] = value.(string)
	}
	return &cloapi.ListOptions{Query: query}
}
func apiDataSource(operation, scope, description string, collection, enveloped bool, result *schema.Schema) *schema.Resource {
	fields := map[string]*schema.Schema{"result": result}
	if scope != "" {
		fields[scope] = &schema.Schema{Type: schema.TypeString, Required: true, Description: "CLO object ID used by this lookup."}
	}
	if collection {
		fields["query"] = listQuerySchema()
		fields["total_count"] = computedField(schema.TypeInt)
	}
	return &schema.Resource{Description: description, Schema: fields, ReadContext: func(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
		id := ""
		if scope != "" {
			id = d.Get(scope).(string)
		}
		query := map[string]string{}
		if collection {
			for k, v := range d.Get("query").(map[string]interface{}) {
				query[k] = v.(string)
			}
		}
		rows, count, err := readQuery(ctx, m.(*providerMeta).v3, operation, id, query, collection, enveloped)
		if err != nil {
			return diag.FromErr(err)
		}
		for _, row := range rows {
			if err := flattenQueryRecord(row, result.Elem.(*schema.Resource).Schema); err != nil {
				return diag.FromErr(err)
			}
		}
		if err := d.Set("result", rows); err != nil {
			return diag.FromErr(err)
		}
		if collection {
			if err := d.Set("total_count", count); err != nil {
				return diag.FromErr(err)
			}
		}
		d.SetId(operation + "/" + id)
		return nil
	}}
}

func readQuery(ctx context.Context, cli *cloapi.Client, operation, id string, query map[string]string, collection, enveloped bool) ([]map[string]interface{}, int, error) {
	if collection {
		options := &cloapi.ListOptions{Query: query}
		rows, err := cloapi.CollectPages(ctx, options, func(params map[string]string) ([]map[string]interface{}, int, error) {
			body, err := cli.Query(ctx, operation, id, params)
			if err != nil {
				return nil, 0, err
			}
			var page struct {
				Count  int
				Result []map[string]interface{}
			}
			if err := decodeQueryJSON(body, &page); err != nil {
				return nil, 0, err
			}
			return page.Result, page.Count, nil
		})
		return rows, options.Total, err
	}
	body, err := cli.Query(ctx, operation, id, query)
	if err != nil {
		return nil, 0, err
	}
	if enveloped {
		var envelope struct{ Result json.RawMessage }
		if err := json.Unmarshal(body, &envelope); err != nil {
			return nil, 0, err
		}
		body = envelope.Result
	}
	var row map[string]interface{}
	if err := decodeQueryJSON(body, &row); err != nil {
		return nil, 0, err
	}
	if row == nil {
		return nil, 0, fmt.Errorf("%s returned an empty object", operation)
	}
	return []map[string]interface{}{row}, 1, nil
}
func decodeQueryJSON(raw []byte, target interface{}) error {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	return decoder.Decode(target)
}

// SDKv2 represents nested objects as single-element lists. Keep arbitrary license
// fields as JSON and convert integer values without a float64 round trip.
func flattenQueryRecord(row map[string]interface{}, fields map[string]*schema.Schema) error {
	for name, f := range fields {
		key := strings.TrimSuffix(name, "_json")
		value, ok := row[key]
		if !ok || value == nil {
			continue
		}
		if key != name {
			encoded, err := json.Marshal(value)
			if err != nil {
				return err
			}
			delete(row, key)
			row[name] = string(encoded)
			continue
		}
		if n, ok := value.(json.Number); ok {
			if f.Type == schema.TypeInt {
				i, err := strconv.Atoi(n.String())
				if err != nil {
					return err
				}
				row[name] = i
			} else {
				v, err := n.Float64()
				if err != nil {
					return err
				}
				row[name] = v
			}
			continue
		}
		if f.Type != schema.TypeList {
			continue
		}
		children, ok := value.([]interface{})
		if !ok {
			children = []interface{}{value}
		}
		for _, child := range children {
			obj, ok := child.(map[string]interface{})
			if !ok {
				return fmt.Errorf("expected an object in %s", name)
			}
			if err := flattenQueryRecord(obj, f.Elem.(*schema.Resource).Schema); err != nil {
				return err
			}
		}
		row[name] = children
	}
	for key := range row {
		if _, ok := fields[key]; !ok {
			delete(row, key)
		}
	}
	return nil
}
