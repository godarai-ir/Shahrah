"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { useLocale, useTranslations } from "next-intl";
import type { ComponentType } from "react";
import {
  BadgeCheck,
  BriefcaseBusiness,
  Building2,
  KeyRound,
  UserCircle2,
  Users,
} from "lucide-react";

import { Tabs, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { canAnyPermission } from "@/lib/authz";
import { usePermissions } from "@/app/components/permissions-context";

type AccessTabKey =
  | "users"
  | "teams"
  | "orgUnits"
  | "personalAccessTokens"
  | "virtualAccounts"
  | "roles";

const tabs: Array<{
  key: AccessTabKey;
  href: string;
  labelKey: "roles" | "virtualAccounts" | "personalAccessTokens" | "orgUnits" | "teams" | "users";
  icon: ComponentType<{ className?: string }>;
  requiredPermissions: string[];
}> = [
  { key: "roles", href: "/roles", labelKey: "roles", icon: BadgeCheck, requiredPermissions: ["role:list", "role:manage"] },
  {
    key: "virtualAccounts",
    href: "/virtual-users",
    labelKey: "virtualAccounts",
    icon: BriefcaseBusiness,
    requiredPermissions: ["virtual_account:read", "virtual_account:manage", "virtual_account:create"],
  },
  // Always visible by requirement.
  { key: "personalAccessTokens", href: "/personal-access-tokens", labelKey: "personalAccessTokens", icon: KeyRound, requiredPermissions: [] },
  {
    key: "orgUnits",
    href: "/org-units",
    labelKey: "orgUnits",
    icon: Building2,
    requiredPermissions: ["org_unit:list", "org_unit:read", "org_unit:manage"],
  },
  { key: "teams", href: "/teams", labelKey: "teams", icon: Users, requiredPermissions: ["team:read", "team:manage"] },
  { key: "users", href: "/users", labelKey: "users", icon: UserCircle2, requiredPermissions: ["user:list", "user:manage"] },
];

export default function AccessManagementHeader({ activeTab }: { activeTab: AccessTabKey }) {
  const t = useTranslations("accessManagement");
  const locale = useLocale();
  const permissions = usePermissions();
  const pathname = usePathname();
  const direction = locale.startsWith("fa") ? "rtl" : "ltr";

  return (
    <nav aria-label={t("title")} className="min-w-0">
      <Tabs value={activeTab} orientation="horizontal" className="min-w-0 w-full" dir={direction}>
        <div className="overflow-x-auto pb-1">
          <TabsList variant="line" className="h-auto w-max min-w-full justify-start gap-2 rounded-none border-b p-0">
          {tabs.map((tab) => {
            const Icon = tab.icon;
            const label = t(`tabs.${tab.labelKey}`);
            const isActive = pathname === tab.href || activeTab === tab.key;
            const canAccess =
              tab.requiredPermissions.length === 0 ||
              (permissions ? canAnyPermission(permissions, tab.requiredPermissions) : true);

            if (!canAccess) {
              return (
                <TabsTrigger
                  key={tab.key}
                  value={tab.key}
                  disabled
                  className="rounded-none px-2 py-2 text-muted-foreground opacity-60 data-[state=active]:text-muted-foreground"
                >
                  <Icon aria-hidden="true" />
                  <span>{label}</span>
                </TabsTrigger>
              );
            }

            return (
              <TabsTrigger key={tab.key} value={tab.key} asChild className="rounded-none px-2 py-2 data-[state=active]:text-foreground">
                <Link href={tab.href} aria-current={isActive ? "page" : undefined}>
                  <Icon aria-hidden="true" />
                  <span>{label}</span>
                </Link>
              </TabsTrigger>
            );
          })}
          </TabsList>
        </div>
      </Tabs>
    </nav>
  );
}
