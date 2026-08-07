package iam

import "sort"

type permissionMeta struct {
	Group       string
	Description string
}

// PermissionCatalogItem describes an IAM permission exposed to management clients.
type PermissionCatalogItem struct {
	Key         string `json:"key"`
	Group       string `json:"group"`
	Description string `json:"description"`
}

// permissionCatalog is the canonical source for valid permissions and UI metadata.
var permissionCatalog = map[string]permissionMeta{
	"policy:list":             {Group: "policies", Description: "Allows listing and viewing policies"},
	"policy:manage":           {Group: "policies", Description: "Allows creating, updating, and deleting policies"},
	"role:list":               {Group: "roles", Description: "Allows listing and viewing roles"},
	"role:manage":             {Group: "roles", Description: "Allows creating, updating, and deleting roles"},
	"user:list":               {Group: "users", Description: "Allows listing and viewing users"},
	"user:manage":             {Group: "users", Description: "Allows creating, updating, and deleting users"},
	"team:create":             {Group: "teams", Description: "Allows creation of new teams"},
	"team:delete":             {Group: "teams", Description: "Allows deleting existing teams"},
	"team:manage":             {Group: "teams", Description: "Allows updating team configuration and members"},
	"team:read":               {Group: "teams", Description: "Allows viewing team details and members"},
	"org_unit:create":         {Group: "organization_units", Description: "Allows creation of new organization units"},
	"org_unit:delete":         {Group: "organization_units", Description: "Allows deleting existing organization units"},
	"org_unit:manage":         {Group: "organization_units", Description: "Allows updating organization unit configuration and members"},
	"org_unit:read":           {Group: "organization_units", Description: "Allows viewing organization unit details and members"},
	"org_unit:list":           {Group: "organization_units", Description: "Allows listing and viewing all organization units"},
	"virtual_account:create":  {Group: "virtual_accounts", Description: "Allows creation of new virtual accounts"},
	"virtual_account:read":    {Group: "virtual_accounts", Description: "Allows viewing virtual account details"},
	"virtual_account:manage":  {Group: "virtual_accounts", Description: "Allows updating virtual account configuration and tokens"},
	"virtual_account:delete":  {Group: "virtual_accounts", Description: "Allows deleting existing virtual accounts"},
	"provider_account:create": {Group: "provider_accounts", Description: "Allows creation of new provider accounts"},
	"provider_account:delete": {Group: "provider_accounts", Description: "Allows deleting existing provider accounts"},
	"provider_account:manage": {Group: "provider_accounts", Description: "Allows updating provider account configuration"},
	"provider_account:read":   {Group: "provider_accounts", Description: "Allows viewing provider account details"},
	"settings:list":           {Group: "settings", Description: "Allows listing and viewing platform settings"},
	"settings:manage":         {Group: "settings", Description: "Allows updating platform settings"},
	"billing:read":            {Group: "billing", Description: "Allows viewing wallet, billing reports, and financial profile"},
	"billing:manage":          {Group: "billing", Description: "Allows managing wallet top-ups and financial operations"},
	"billing:profile_manage":  {Group: "billing", Description: "Allows editing tenant billing profile"},
	"billing:documents_read":  {Group: "billing", Description: "Allows viewing invoices, transactions, and discount documents"},
	"billing:codes_redeem":    {Group: "billing", Description: "Allows redeeming gift and discount codes"},
	"logs:read_metadata":      {Group: "logs", Description: "Allows listing request logs metadata and summary fields"},
	"logs:read_content":       {Group: "logs", Description: "Allows viewing stored request and response payload content"},
	"logs:export":             {Group: "logs", Description: "Allows exporting request logs"},
	"logs:manage_views":       {Group: "logs", Description: "Allows creating and managing saved log views"},
	"logs:replay":             {Group: "logs", Description: "Allows creating safe replay drafts from request logs"},
	"support_ticket:create":   {Group: "support", Description: "Allows creating support tickets"},
	"support_ticket:read_own": {Group: "support", Description: "Allows viewing support tickets created by the current user"},
	"support_ticket:read_all": {Group: "support", Description: "Allows viewing all support tickets in the tenant"},
	"support_ticket:reply":    {Group: "support", Description: "Allows replying to visible support tickets"},
	"support_ticket:rate":     {Group: "support", Description: "Allows rating resolved support tickets created by the current user"},
}

// IsValidPermission reports whether a permission is present and documented.
func IsValidPermission(permission string) bool {
	meta, ok := permissionCatalog[permission]

	return ok && meta.Description != ""
}

// ListPermissionCatalog returns a stable copy of the permission catalog.
func ListPermissionCatalog() []PermissionCatalogItem {
	items := make([]PermissionCatalogItem, 0, len(permissionCatalog))
	for key, meta := range permissionCatalog {
		items = append(items, PermissionCatalogItem{Key: key, Group: meta.Group, Description: meta.Description})
	}

	sort.Slice(items, func(i, j int) bool {
		if items[i].Group == items[j].Group {
			return items[i].Key < items[j].Key
		}

		return items[i].Group < items[j].Group
	})

	return items
}
