"use client";

import { useEffect, useMemo, useState } from "react";
import { useRouter } from "next/navigation";
import { useLocale, useTranslations } from "next-intl";
import { Plus, ShieldCheck } from "lucide-react";

import { usePermissions } from "@/app/components/permissions-context";
import {
  AppCard,
  AppSelectField,
  CommercialPlanUpgradeDialog,
  EmptyState,
  PageHeader,
  RowActionsMenu,
  SearchField,
} from "@/components/app-ui";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from "@/components/ui/tooltip";

type RoleDefinition = {
  id: string;
  tenant_id: string;
  name: string;
  display_name: string;
  description?: string;
  is_system: boolean;
  permissions: string[];
  created_at: string;
  updated_at: string;
};

type RolesPermissionsManagementProps = {
  initialRoles: RoleDefinition[];
  initialError: string | null;
};

type RoleTypeFilter = "all" | "system" | "custom";
type EffectiveEntitlements = { features?: Record<string, unknown> } | null;

const protectedSystemRoles = new Set(["proxy_admin", "tenant_admin"]);
const hiddenRolesInUI = new Set(["proxy_admin"]);
const FEATURE_RBAC_ENABLED = "rbac.enabled";

export default function RolesPermissionsManagement(props: RolesPermissionsManagementProps) {
  const t = useTranslations("accessManagement");
  const locale = useLocale();
  const router = useRouter();
  const permissions = usePermissions();
  const canManageRoles = (permissions ?? []).some(
    (permission) => permission.trim().toLowerCase() === "role:manage",
  );

  const [roles] = useState<RoleDefinition[]>(
    props.initialRoles.filter((role) => !hiddenRolesInUI.has(role.name)),
  );
  const [searchQuery, setSearchQuery] = useState("");
  const [roleType, setRoleType] = useState<RoleTypeFilter>("all");
  const [entitlements, setEntitlements] = useState<EffectiveEntitlements>(null);
  const [hasLoadedEntitlements, setHasLoadedEntitlements] = useState(false);
  const [isUpgradeDialogOpen, setIsUpgradeDialogOpen] = useState(false);

  const numberFormatter = useMemo(() => new Intl.NumberFormat(locale), [locale]);
  const direction = locale.startsWith("fa") ? "rtl" : "ltr";
  const normalizedSearch = searchQuery.trim().toLocaleLowerCase();
  const filteredRoles = useMemo(() => {
    return roles.filter((role) => {
      const matchesType =
        roleType === "all" ||
        (roleType === "system" && role.is_system) ||
        (roleType === "custom" && !role.is_system);
      const matchesSearch =
        normalizedSearch.length === 0 ||
        [role.name, role.display_name, role.description ?? ""].some((value) =>
          value.toLocaleLowerCase().includes(normalizedSearch),
        );

      return matchesType && matchesSearch;
    });
  }, [normalizedSearch, roleType, roles]);

  const hasActiveFilters = normalizedSearch.length > 0 || roleType !== "all";
  const columnCount = canManageRoles ? 4 : 3;
  const isRbacFeatureEnabled = hasLoadedEntitlements
    ? readBooleanFeature(entitlements?.features?.[FEATURE_RBAC_ENABLED], false)
    : null;

  useEffect(() => {
    let isMounted = true;

    void getEntitlements()
      .then((response) => {
        if (isMounted) setEntitlements(response);
      })
      .catch(() => {
        if (isMounted) setEntitlements(null);
      })
      .finally(() => {
        if (isMounted) setHasLoadedEntitlements(true);
      });

    return () => {
      isMounted = false;
    };
  }, []);

  function clearFilters() {
    setSearchQuery("");
    setRoleType("all");
  }

  function handleCreateRole() {
    if (isRbacFeatureEnabled === false) {
      setIsUpgradeDialogOpen(true);
      return;
    }
    router.push("/roles/new");
  }

  const createRoleAction = canManageRoles ? (
    <Button type="button" onClick={handleCreateRole} disabled={!hasLoadedEntitlements}>
      <Plus data-icon="inline-start" aria-hidden="true" />
      {t("rolesPage.form.create")}
    </Button>
  ) : null;

  return (
    <div className="mx-auto flex w-full max-w-7xl flex-col gap-6">
      <PageHeader
        title={t("rolesPage.listTitle")}
        description={t("rolesPage.description")}
        titleAddon={createRoleAction}
      />

      {props.initialError ? (
        <Alert variant="destructive">
          <AlertTitle>{t("alerts.error")}</AlertTitle>
          <AlertDescription className="flex flex-col items-start gap-3 sm:flex-row sm:items-center sm:justify-between">
            <p>{props.initialError}</p>
            <Button type="button" size="sm" variant="outline" onClick={() => router.refresh()}>
              {t("rolesPage.actions.retry")}
            </Button>
          </AlertDescription>
        </Alert>
      ) : roles.length === 0 ? (
        <EmptyState
          icon={ShieldCheck}
          title={t("rolesPage.empty.title")}
          description={t("rolesPage.empty.description")}
          primaryAction={createRoleAction}
        />
      ) : (
        <AppCard
          className="min-w-0 shadow-none"
          contentClassName="flex min-w-0 flex-col gap-4"
        >
          <div className="flex flex-wrap items-center gap-3">
            <SearchField
              containerClassName="min-w-56 flex-1"
              type="text"
              value={searchQuery}
              onChange={(event) => setSearchQuery(event.target.value)}
              placeholder={t("rolesPage.filters.searchPlaceholder")}
              aria-label={t("rolesPage.filters.searchLabel")}
              dir={direction}
            />
            <AppSelectField
              value={roleType}
              onValueChange={(value) => setRoleType(value as RoleTypeFilter)}
              options={[
                { value: "all", label: t("rolesPage.filters.all") },
                { value: "system", label: t("rolesPage.filters.system") },
                { value: "custom", label: t("rolesPage.filters.custom") },
              ]}
              ariaLabel={t("rolesPage.filters.typeLabel")}
              className="min-w-48 flex-1 sm:flex-none"
              triggerClassName="h-10 w-full"
              dir={direction}
            />
            {hasActiveFilters ? (
              <Button type="button" size="sm" variant="ghost" onClick={clearFilters}>
                {t("rolesPage.filters.clear")}
              </Button>
            ) : null}
          </div>

          <TooltipProvider delayDuration={150}>
            <Table dir={direction}>
              <TableHeader>
                <TableRow className="bg-muted/40 hover:bg-muted/40">
                  <TableHead scope="col" className="text-start">{t("rolesPage.table.displayName")}</TableHead>
                  <TableHead scope="col" className="text-start">
                    {t("rolesPage.table.scope")}
                  </TableHead>
                  <TableHead scope="col" className="w-20 text-center">
                    {t("rolesPage.table.permissions")}
                  </TableHead>
                  {canManageRoles ? (
                    <TableHead scope="col" className="w-14 text-end">
                      {t("rolesPage.table.actions")}
                    </TableHead>
                  ) : null}
                </TableRow>
              </TableHeader>
              <TableBody>
                {filteredRoles.length === 0 ? (
                  <TableRow>
                    <TableCell colSpan={columnCount} className="py-12 text-center">
                      <div className="flex flex-col items-center gap-3">
                        <p className="text-sm text-muted-foreground">
                          {t("rolesPage.table.filteredEmpty")}
                        </p>
                        <Button type="button" size="sm" variant="outline" onClick={clearFilters}>
                          {t("rolesPage.filters.clear")}
                        </Button>
                      </div>
                    </TableCell>
                  </TableRow>
                ) : (
                  filteredRoles.map((role) => {
                    const isProtected = role.is_system && protectedSystemRoles.has(role.name);
                    return (
                      <TableRow key={role.id}>
                        <TableCell dir="auto" className="min-w-36">
                          <div className="flex max-w-lg flex-col gap-1">
                            <span className="font-medium text-foreground">{role.display_name || role.name}</span>
                            <span dir="ltr" className="text-start font-mono text-xs text-muted-foreground">
                              {role.name}
                            </span>
                          </div>
                        </TableCell>
                        <TableCell>
                          <div className="flex flex-wrap items-center gap-2">
                            <Badge variant={role.is_system ? "secondary" : "outline"}>
                              {role.is_system ? t("rolesPage.systemRole") : t("rolesPage.customRole")}
                            </Badge>
                            {isProtected ? (
                              <Tooltip>
                                <TooltipTrigger asChild>
                                  <Badge variant="outline" tabIndex={0}>
                                    {t("rolesPage.actions.protected")}
                                  </Badge>
                                </TooltipTrigger>
                                <TooltipContent>{t("rolesPage.actions.protectedDescription")}</TooltipContent>
                              </Tooltip>
                            ) : null}
                          </div>
                        </TableCell>
                        <TableCell className="text-center tabular-nums">
                          {numberFormatter.format((role.permissions ?? []).length)}
                        </TableCell>
                        {canManageRoles ? (
                          <TableCell className="text-end">
                            <RowActionsMenu
                              label={t("rolesPage.actions.openMenu", { role: role.display_name || role.name })}
                              dir={direction}
                              actions={[
                                {
                                  label: t("rolesPage.actions.edit"),
                                  disabled: isProtected,
                                  onSelect: () => router.push(`/roles/${encodeURIComponent(role.id)}/edit`),
                                },
                              ]}
                            />
                          </TableCell>
                        ) : null}
                      </TableRow>
                    );
                  })
                )}
              </TableBody>
            </Table>
          </TooltipProvider>
        </AppCard>
      )}

      <CommercialPlanUpgradeDialog
        open={isUpgradeDialogOpen}
        onOpenChange={setIsUpgradeDialogOpen}
        title={t("rolesPage.upgradeDialog.title")}
        description={t("rolesPage.upgradeDialog.description")}
        cancelLabel={t("rolesPage.upgradeDialog.cancel")}
        actionLabel={t("rolesPage.upgradeDialog.action")}
      />
    </div>
  );
}

async function getEntitlements(): Promise<EffectiveEntitlements> {
  const response = await fetch("/api/entitlements", { cache: "no-store" });
  if (!response.ok) throw new Error("Failed to load entitlements");
  return response.json() as Promise<EffectiveEntitlements>;
}

function readBooleanFeature(value: unknown, fallback: boolean): boolean {
  if (typeof value === "boolean") return value;
  if (typeof value === "string") {
    if (value.trim().toLowerCase() === "true") return true;
    if (value.trim().toLowerCase() === "false") return false;
  }
  return fallback;
}
