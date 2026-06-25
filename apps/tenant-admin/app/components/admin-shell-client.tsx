"use client";

import { useMemo } from "react";
import Image from "next/image";
import Link from "next/link";
import { usePathname, useRouter } from "next/navigation";
import { useTranslations } from "next-intl";
import { LogOut, User } from "lucide-react";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import {
  Sidebar,
  SidebarContent,
  SidebarGroup,
  SidebarGroupContent,
  SidebarGroupLabel,
  SidebarHeader,
  SidebarInset,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  SidebarProvider,
  SidebarRail,
  SidebarTrigger,
} from "@/components/ui/sidebar";
import { ThemeToggle } from "@/components/theme-toggle";
import { buildAdminNavItems, filterVisibleNavItems, type AdminNavItem } from "@/app/components/admin-shell-nav";
import { PermissionsProvider } from "@/app/components/permissions-context";

type AdminShellClientProps = {
  children: React.ReactNode;
  tenantSlug: string;
  currentUser: {
    id: string;
    email?: string;
    mobile?: string;
    role: string;
  };
  permissions: string[];
};

const SIDEBAR_NAV_GROUPS: Array<{
  labelKey: string;
  itemKeys: AdminNavItem["key"][];
}> = [
  {
    labelKey: "navGroups.overview",
    itemKeys: ["dashboard", "topup"],
  },
  {
    labelKey: "navGroups.runtime",
    itemKeys: ["providerAccounts", "virtualModels", "virtualUsers"],
  },
  {
    labelKey: "navGroups.observability",
    itemKeys: ["traces", "metrics"],
  },
  {
    labelKey: "navGroups.governance",
    itemKeys: ["rateLimitPolicies", "budgetLimitPolicies", "guardrails"],
  },
  {
    labelKey: "navGroups.access",
    itemKeys: ["accessManagement", "orgUnits"],
  },
];

export default function AdminShellClient({
  children,
  tenantSlug,
  currentUser,
  permissions,
}: AdminShellClientProps) {
  const pathname = usePathname();
  const t = useTranslations("shell");

  const userIdentity =
    currentUser.email ?? currentUser.mobile ?? currentUser.id;

  const navItems = useMemo(() => buildAdminNavItems(pathname, t), [pathname, t]);

  const visibleNavItems = useMemo(() => {
    return filterVisibleNavItems(navItems, permissions);
  }, [navItems, permissions]);

  const sidebarNavItems = useMemo(() => {
    return visibleNavItems.filter((item) => item.key !== "settings");
  }, [visibleNavItems]);

  const currentPageLabel = visibleNavItems.find((item) => item.isActive)?.title ?? t("nav.dashboard");

  return (
    <PermissionsProvider permissions={permissions}>
      <SidebarProvider className="bg-background">
        <AdminSidebar
          tenantSlug={tenantSlug}
          navItems={sidebarNavItems}
        />

        <SidebarInset className="min-w-0 rounded-none bg-background">
          <header className="sticky top-0 z-10 border-b border-border bg-background/95 px-4 py-3 backdrop-blur supports-[backdrop-filter]:bg-background/85 md:px-6">
            <div className="flex min-w-0 items-center gap-3">
              <SidebarTrigger aria-label={t("menu.open")} />

              <div className="min-w-0 flex-1">
                <p className="text-xs text-muted-foreground">{t("tenant")}</p>
                <h1 className="truncate text-base font-semibold md:text-lg">{currentPageLabel}</h1>
              </div>

              <div className="flex shrink-0 items-center gap-2">
                <ThemeToggle />
                <UserProfileMenu userIdentity={userIdentity} currentUser={currentUser} />
              </div>
            </div>
          </header>

          <div className="min-w-0 flex-1 p-2 pb-[max(0.5rem,env(safe-area-inset-bottom))] md:p-3 md:pb-[max(0.75rem,env(safe-area-inset-bottom))]">
            <div className="w-full animate-fade-in">{children}</div>
          </div>
        </SidebarInset>
      </SidebarProvider>
    </PermissionsProvider>
  );
}

type UserProfileMenuProps = {
  userIdentity: string;
  currentUser: {
    role: string;
  };
};

function UserProfileMenu({ userIdentity, currentUser }: UserProfileMenuProps) {
  const router = useRouter();
  const t = useTranslations("shell");

  async function handleLogout() {
    await fetch("/api/auth/logout", { method: "POST" });
    router.replace("/login");
    router.refresh();
  }

  return (
    <DropdownMenu dir="rtl">
      <DropdownMenuTrigger asChild>
        <Button
          type="button"
          variant="outline"
          size="icon"
          className="shrink-0"
          aria-label={t("profile.menu")}
        >
          <User data-icon="inline-start" aria-hidden="true" />
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="start" className="w-56 text-right">
        <DropdownMenuLabel className="flex flex-col items-start gap-1">
          <span dir="ltr" className="max-w-full truncate text-sm font-medium">
            {userIdentity}
          </span>
          <Badge variant="secondary" className="w-fit text-xs" dir="ltr">
            {currentUser.role}
          </Badge>
        </DropdownMenuLabel>
        <DropdownMenuSeparator />
        <DropdownMenuItem asChild className="justify-start">
          <Link href="/settings">
            <User data-icon="inline-start" aria-hidden="true" />
            <span>{t("profile.account")}</span>
          </Link>
        </DropdownMenuItem>
        <DropdownMenuSeparator />
        <DropdownMenuItem
          className="text-destructive focus:text-destructive"
          onSelect={(event) => {
            event.preventDefault();
            void handleLogout();
          }}
        >
          <LogOut data-icon="inline-start" aria-hidden="true" />
          <span>{t("logout")}</span>
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  );
}

type SidebarContentProps = {
  tenantSlug: string;
  navItems: AdminNavItem[];
};

function AdminSidebar({
  tenantSlug,
  navItems,
}: SidebarContentProps) {
  const t = useTranslations("shell");
  const navItemsByKey = useMemo(() => {
    return new Map(navItems.map((item) => [item.key, item]));
  }, [navItems]);

  const navGroups = useMemo(() => {
    return SIDEBAR_NAV_GROUPS.map((group) => ({
      ...group,
      items: group.itemKeys
        .map((key) => navItemsByKey.get(key))
        .filter((item): item is AdminNavItem => Boolean(item)),
    })).filter((group) => group.items.length > 0);
  }, [navItemsByKey]);

  return (
    <Sidebar side="right" collapsible="icon">
      <SidebarHeader>
        <SidebarMenu>
          <SidebarMenuItem>
            <SidebarMenuButton
              size="lg"
              tooltip={t("title")}
              className="h-12 items-center gap-2 rounded-lg border border-sidebar-border bg-sidebar"
              asChild
            >
              <Link href="/">
                <span className="flex size-8 shrink-0 items-center justify-center overflow-hidden rounded-lg border border-sidebar-border bg-sidebar-accent p-1">
                  <Image
                    src="/brand/shahrah-logo.png"
                    alt={t("logoAlt")}
                    width={32}
                    height={32}
                    className="h-full w-full object-contain"
                    priority
                  />
                </span>
                <span className="flex min-w-0 flex-col gap-0.5">
                  <span className="truncate text-sm font-semibold">{t("title")}</span>
                  <span className="truncate text-xs text-sidebar-foreground/70">{tenantSlug}</span>
                </span>
              </Link>
            </SidebarMenuButton>
          </SidebarMenuItem>
        </SidebarMenu>
      </SidebarHeader>

      <SidebarContent>
        {navGroups.map((group) => (
          <SidebarGroup key={group.labelKey} className="py-1">
            <SidebarGroupLabel>{t(group.labelKey)}</SidebarGroupLabel>
            <SidebarGroupContent>
              <SidebarMenu aria-label={t(group.labelKey)}>
                {group.items.map((item) => {
                  const Icon = item.icon;

                  return (
                    <SidebarMenuItem key={item.href}>
                      <SidebarMenuButton
                        isActive={item.isActive}
                        tooltip={item.title}
                        asChild
                      >
                        <Link href={item.href}>
                          <Icon aria-hidden="true" />
                          <span>{item.title}</span>
                        </Link>
                      </SidebarMenuButton>
                    </SidebarMenuItem>
                  );
                })}
              </SidebarMenu>
            </SidebarGroupContent>
          </SidebarGroup>
        ))}
      </SidebarContent>

      <SidebarRail />
    </Sidebar>
  );
}
