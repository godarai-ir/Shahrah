"use client";

import Link from "next/link";
import type { ComponentType } from "react";
import { BadgeCheck, BriefcaseBusiness, Building2, KeyRound, UserCircle2, Users } from "lucide-react";
import { canAnyPermission } from "@/lib/authz";
import { usePermissions } from "@/app/components/permissions-context";

type AccessTabKey = "users" | "teams" | "orgUnits" | "personalAccessTokens" | "virtualAccounts" | "roles";

const tabs: Array<{
  key: AccessTabKey;
  href: string;
  label: string;
  icon: ComponentType<{ className?: string }>;
  requiredPermissions: string[];
}> = [
  { key: "users", href: "/users", label: "کاربران", icon: UserCircle2, requiredPermissions: ["user:list", "user:manage"] },
  { key: "teams", href: "/teams", label: "تیم‌ها", icon: Users, requiredPermissions: ["team:read", "team:manage"] },
  { key: "orgUnits", href: "/org-units", label: "واحدهای سازمانی", icon: Building2, requiredPermissions: ["org_unit:list", "org_unit:read", "org_unit:manage"] },
  { key: "personalAccessTokens", href: "/personal-access-tokens", label: "کلیدهای دسترسی شخصی", icon: KeyRound, requiredPermissions: ["settings:list", "settings:manage"] },
  { key: "virtualAccounts", href: "/virtual-models", label: "اکانت‌های مجازی", icon: BriefcaseBusiness, requiredPermissions: ["virtual_account:read", "virtual_account:manage", "virtual_account:create"] },
  { key: "roles", href: "/roles", label: "نقش‌ها", icon: BadgeCheck, requiredPermissions: ["role:list", "role:manage"] },
];

export default function AccessManagementHeader({ activeTab }: { activeTab: AccessTabKey }) {
  const permissions = usePermissions();
  const visibleTabs = permissions
    ? tabs.filter((tab) => canAnyPermission(permissions, tab.requiredPermissions))
    : tabs;

  return (
    <header className="space-y-4">
      <h1 className="text-4xl font-bold tracking-tight">مدیریت دسترسی</h1>
      <nav className="border-b">
        <ul className="flex flex-wrap items-center gap-5 pb-2">
          {visibleTabs.map((tab) => {
            const Icon = tab.icon;
            const active = tab.key === activeTab;
            return (
              <li key={tab.key}>
                <Link
                  href={tab.href}
                  className={`inline-flex items-center gap-2 border-b-2 px-1 py-2 ${
                    active
                      ? "border-indigo-500 text-foreground"
                      : "border-transparent text-muted-foreground hover:text-foreground"
                  }`}
                  aria-current={active ? "page" : undefined}
                >
                  <Icon className="h-4 w-4" />
                  <span className="text-sm font-semibold">{tab.label}</span>
                </Link>
              </li>
            );
          })}
        </ul>
      </nav>
    </header>
  );
}
