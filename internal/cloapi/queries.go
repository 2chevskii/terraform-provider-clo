package cloapi

import (
	"context"
	"fmt"
	gen "github.com/clo-ru/cloapi-go-client/v3"
)

// Query returns the original JSON so open-ended fields and nullable values survive SDK decoding.
func (c *Client) Query(ctx context.Context, operation, id string, params map[string]string) ([]byte, error) {
	editor := gen.WithQuery(params)
	switch operation {
	case "ServerConsole":
		r, err := c.gen.ServerConsoleWithResponse(ctx, id, editor)
		if err != nil {
			return nil, err
		}
		return r.Body, nil
	case "ServerCheckGuestAgent":
		r, err := c.gen.ServerCheckGuestAgentWithResponse(ctx, id, editor)
		if err != nil {
			return nil, err
		}
		return r.Body, nil
	case "AccountBalance":
		r, err := c.gen.AccountBalanceWithResponse(ctx, editor)
		if err != nil {
			return nil, err
		}
		return r.Body, nil
	case "AccountLimits":
		r, err := c.gen.AccountLimitsWithResponse(ctx, editor)
		if err != nil {
			return nil, err
		}
		return r.Body, nil
	case "AccountProjectsLimits":
		r, err := c.gen.AccountProjectsLimitsWithResponse(ctx, editor)
		if err != nil {
			return nil, err
		}
		return r.Body, nil
	case "AccountStat":
		r, err := c.gen.AccountStatWithResponse(ctx, editor)
		if err != nil {
			return nil, err
		}
		return r.Body, nil
	case "ProjectDetail":
		r, err := c.gen.ProjectDetailWithResponse(ctx, id, editor)
		if err != nil {
			return nil, err
		}
		return r.Body, nil
	case "ProjectLimitsList":
		r, err := c.gen.ProjectLimitsListWithResponse(ctx, id, editor)
		if err != nil {
			return nil, err
		}
		return r.Body, nil
	case "ProjectServerConfigDep":
		r, err := c.gen.ProjectServerConfigDepWithResponse(ctx, id, editor)
		if err != nil {
			return nil, err
		}
		return r.Body, nil
	case "ProjectDbaasConfigDep":
		r, err := c.gen.ProjectDbaasConfigDepWithResponse(ctx, id, editor)
		if err != nil {
			return nil, err
		}
		return r.Body, nil
	case "ProjectInfrastructureModuleConstants":
		r, err := c.gen.ProjectInfrastructureModuleConstantsWithResponse(ctx, id, editor)
		if err != nil {
			return nil, err
		}
		return r.Body, nil
	case "ProjectConsumption":
		r, err := c.gen.ProjectConsumptionWithResponse(ctx, id, editor)
		if err != nil {
			return nil, err
		}
		return r.Body, nil
	case "NetworksList":
		r, err := c.gen.NetworksListWithResponse(ctx, id, editor)
		if err != nil {
			return nil, err
		}
		return r.Body, nil
	case "ProjectLocalDisksList":
		r, err := c.gen.ProjectLocalDisksListWithResponse(ctx, id, editor)
		if err != nil {
			return nil, err
		}
		return r.Body, nil
	case "LocalDiskDetail":
		r, err := c.gen.LocalDiskDetailWithResponse(ctx, id, editor)
		if err != nil {
			return nil, err
		}
		return r.Body, nil
	case "AvailableLicensesList":
		r, err := c.gen.AvailableLicensesListWithResponse(ctx, editor)
		if err != nil {
			return nil, err
		}
		return r.Body, nil
	case "ServerLicenses":
		r, err := c.gen.ServerLicensesWithResponse(ctx, id, editor)
		if err != nil {
			return nil, err
		}
		return r.Body, nil
	case "LicenseDetails":
		r, err := c.gen.LicenseDetailsWithResponse(ctx, id, editor)
		if err != nil {
			return nil, err
		}
		return r.Body, nil
	case "ProjectRuleList":
		r, err := c.gen.ProjectRuleListWithResponse(ctx, id, editor)
		if err != nil {
			return nil, err
		}
		return r.Body, nil
	case "LoadBalancerStat":
		r, err := c.gen.LoadBalancerStatWithResponse(ctx, id, editor)
		if err != nil {
			return nil, err
		}
		return r.Body, nil
	case "MaintenanceModeStatus":
		r, err := c.gen.MaintenanceModeStatusWithResponse(ctx, editor)
		if err != nil {
			return nil, err
		}
		return r.Body, nil
	default:
		return nil, fmt.Errorf("unknown CLO read operation %q", operation)
	}
}
