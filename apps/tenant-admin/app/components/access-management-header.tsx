"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
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
  label: string;
  icon: ComponentType<{ className?: string }>;
  requiredPermissions: string[];
}> = [
  { key: "roles", href: "/roles", label: "نقش‌ها", icon: BadgeCheck, requiredPermissions: ["role:list", "role:manage"] },
  {
    key: "virtualAccounts",
    href: "/virtual-users",
    label: "اکانت‌های مجازی",
    icon: BriefcaseBusiness,
    requiredPermissions: ["virtual_account:read", "virtual_account:manage", "virtual_account:create"],
  },
  // Always visible by requirement.
  { key: "personalAccessTokens", href: "/personal-access-tokens", label: "کلیدهای دسترسی شخصی", icon: KeyRound, requiredPermissions: [] },
  {
    key: "orgUnits",
    href: "/org-units",
    label: "واحدهای سازمانی",
    icon: Building2,
    requiredPermissions: ["org_unit:list", "org_unit:read", "org_unit:manage"],
  },
  { key: "teams", href: "/teams", label: "تیم‌ها", icon: Users, requiredPermissions: ["team:read", "team:manage"] },
  { key: "users", href: "/users", label: "کاربران", icon: UserCircle2, requiredPermissions: ["user:list", "user:manage"] },
];

export default function AccessManagementHeader({ activeTab }: { activeTab: AccessTabKey }) {
  const permissions = usePermissions();
  const pathname = usePathname();

  return (
    <header className="space-y-4">
      <h1 className="text-4xl font-bold tracking-tight">مدیریت دسترسی</h1>
      <Tabs value={activeTab} orientation="horizontal" className="w-full">
        <TabsList variant="line" className="h-auto w-full justify-start gap-2 rounded-none border-b p-0">
          {tabs.map((tab) => {
            const Icon = tab.icon;
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
                  <Icon className="size-4" />
                  <span className="text-sm font-semibold">{tab.label}</span>
                </TabsTrigger>
              );
            }

            return (
              <TabsTrigger key={tab.key} value={tab.key} asChild className="rounded-none px-2 py-2 data-[state=active]:text-foreground">
                <Link href={tab.href} aria-current={isActive ? "page" : undefined}>
                  <Icon className="size-4" />
                  <span className="text-sm font-semibold">{tab.label}</span>
                </Link>
              </TabsTrigger>
            );
          })}
        </TabsList>
      </Tabs>
    </header>
  );
}
