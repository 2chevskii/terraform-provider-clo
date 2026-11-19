package cloapi

import "testing"

func TestCreateTimeChildren(t *testing.T) {
	rule := RuleCreateParams{AddressID: "address", ExternalProtocolPort: 8443, InternalProtocolPort: 443}
	lb := loadBalancerCreateBody(LoadBalancerCreateParams{Rules: []RuleCreateParams{rule}})
	if lb.Rules == nil || len(*lb.Rules) != 1 {
		t.Fatalf("missing listener rule: %+v", lb.Rules)
	}
	got := (*lb.Rules)[0]
	if got.AddressId != rule.AddressID || got.ExternalProtocolPort != 8443 || got.InternalProtocolPort != 443 {
		t.Fatalf("invalid rule payload %+v", got)
	}
	db := InitialDatabase{Name: "app", AdminUsername: "admin", AdminPassword: "test"}
	cluster := clusterCreateBody(ClusterCreateParams{Database: &db})
	if cluster.Database == nil || *cluster.Database != db || cluster.Backup != nil {
		t.Fatal("invalid initial database request")
	}
}
