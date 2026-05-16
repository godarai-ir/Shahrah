"use client";

import { useState } from "react";
import Link from "next/link";
import { useLocale, useTranslations } from "next-intl";
import { Plus } from "lucide-react";

import AccessManagementHeader from "@/app/components/access-management-header";
import { usePermissions } from "@/app/components/permissions-context";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";

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

type Banner = {
  kind: "success" | "error";
  message: string;
};

type RolesPermissionsManagementProps = {
  initialRoles: RoleDefinition[];
  initialError: string | null;
};

const protectedSystemRoles = new Set(["proxy_admin", "tenant_admin"]);
const hiddenRolesInUI = new Set(["proxy_admin"]);

export default function RolesPermissionsManagement(props: RolesPermissionsManagementProps) {
  const t = useTranslations("accessManagement");
  const locale = useLocale();
  const permissions = usePermissions();
  const canManageRoles = (permissions ?? []).some((permission) => permission.trim().toLowerCase() === "role:manage");

  const [roles] = useState<RoleDefinition[]>(props.initialRoles.filter((role) => !hiddenRolesInUI.has(role.name)));
  const [banner] = useState<Banner | null>(props.initialError ? { kind: "error", message: props.initialError } : null);

  return (
    <div className="space-y-6">
      <AccessManagementHeader activeTab="roles" />
      <p className="text-sm text-muted-foreground">
        {locale.startsWith("fa") ? "فهرست نقش‌ها را مدیریت کنید و برای ساخت یا ویرایش نقش از صفحه اختصاصی استفاده کنید." : "Manage role list here and use dedicated pages for creating and editing roles."}
      </p>

      {banner ? (
        <Alert variant={banner.kind === "error" ? "destructive" : "default"}>
          <AlertTitle>{banner.kind === "error" ? t("alerts.error") : t("alerts.success")}</AlertTitle>
          <AlertDescription>{banner.message}</AlertDescription>
        </Alert>
      ) : null}

      <Card>
        <CardHeader className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
          <div>
            <CardTitle>{t("rolesPage.listTitle")}</CardTitle>
            <CardDescription>
              {locale.startsWith("fa") ? "برای هر نقش از دکمه ویرایش جهت مدیریت نام نمایشی، توضیحات و مجوزها در صفحه جداگانه استفاده کنید." : "Use edit action to manage display name, description, and permissions in a dedicated page."}
            </CardDescription>
          </div>
          {canManageRoles ? (
            <Button asChild className="gap-2">
              <Link href="/roles/new">
                <Plus className="h-4 w-4" />
                {t("rolesPage.form.create")}
              </Link>
            </Button>
          ) : null}
        </CardHeader>
        <CardContent>
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>{t("rolesPage.table.name")}</TableHead>
                <TableHead>{t("rolesPage.table.displayName")}</TableHead>
                <TableHead>{t("rolesPage.table.scope")}</TableHead>
                <TableHead>{t("rolesPage.table.permissions")}</TableHead>
                <TableHead>{t("rolesPage.table.actions")}</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {roles.length === 0 ? (
                <TableRow>
                  <TableCell colSpan={5}>{t("rolesPage.table.empty")}</TableCell>
                </TableRow>
              ) : (
                roles.map((role) => {
                  const isProtected = role.is_system && protectedSystemRoles.has(role.name);
                  return (
                    <TableRow key={role.id}>
                      <TableCell className="font-medium">{role.name}</TableCell>
                      <TableCell>{role.display_name}</TableCell>
                      <TableCell>
                        <Badge variant={role.is_system ? "secondary" : "outline"}>
                          {role.is_system ? t("rolesPage.systemRole") : t("rolesPage.customRole")}
                        </Badge>
                      </TableCell>
                      <TableCell>{(role.permissions ?? []).length}</TableCell>
                      <TableCell>
                        <div className="flex items-center gap-2">
                          {canManageRoles ? (
                            isProtected ? (
                              <Button size="sm" variant="outline" disabled>
                                {t("rolesPage.actions.edit")}
                              </Button>
                            ) : (
                              <Button asChild size="sm" variant="outline">
                                <Link href={`/roles/${encodeURIComponent(role.id)}/edit`}>
                                  {t("rolesPage.actions.edit")}
                                </Link>
                              </Button>
                            )
                          ) : null}
                          {canManageRoles && isProtected ? (
                            <p className="text-xs text-muted-foreground">
                              {t("rolesPage.actions.protected")} - {locale.startsWith("fa") ? "این نقش سیستمی قابل تغییر نیست." : "System role is read-only."}
                            </p>
                          ) : null}
                        </div>
                      </TableCell>
                    </TableRow>
                  );
                })
              )}
            </TableBody>
          </Table>
        </CardContent>
      </Card>
    </div>
  );
}
