"use client";

import { useEffect, useMemo, useState } from "react";
import Image from "next/image";
import Link from "next/link";
import { usePathname } from "next/navigation";
import { useTranslations } from "next-intl";
import { Activity, BarChart3, Building2, CreditCard, LayoutDashboard, LogOut, Menu, Route, Search, Settings, ShieldCheck, Users } from "lucide-react";

import { Avatar, AvatarFallback } from "@/components/ui/avatar";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Separator } from "@/components/ui/separator";
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle,
  SheetTrigger,
} from "@/components/ui/sheet";
import { ThemeToggle } from "@/components/theme-toggle";
import { canAnyPermission, type MePermissionsResponse } from "@/lib/authz";
import { cn } from "@/lib/utils";

type AdminShellProps = {
  children: React.ReactNode;
  tenantSlug: string;
  currentUser: {
    id: string;
    email?: string;
    mobile?: string;
    role: string;
  };
};

type NavItem = {
  href: string;
  title: string;
  icon: React.ComponentType<{ className?: string }>;
  isActive: boolean;
  requiredPermissions: string[];
};

function getInitials(input: string) {
  const value = input.trim();
  if (!value) {
    return "TA";
  }

  if (value.includes("@")) {
    return value.slice(0, 2).toUpperCase();
  }

  const parts = value.split(/\s+/).filter(Boolean);
  if (parts.length >= 2) {
    return `${parts[0][0] ?? ""}${parts[1][0] ?? ""}`.toUpperCase();
  }

  return value.slice(0, 2).toUpperCase();
}

export default function AdminShell({
  children,
  tenantSlug,
  currentUser,
}: AdminShellProps) {
  const pathname = usePathname();
  const t = useTranslations("shell");
  const [permissions, setPermissions] = useState<string[] | null>(null);

  const userIdentity =
    currentUser.email ?? currentUser.mobile ?? currentUser.id;

  useEffect(() => {
    let mounted = true;
    async function loadPermissions() {
      try {
        const response = await fetch("/api/me/permissions", { cache: "no-store" });
        if (!response.ok) {
          if (mounted) setPermissions([]);
          return;
        }
        const payload = (await response.json()) as MePermissionsResponse;
        if (mounted) {
          setPermissions((payload.permissions ?? []).map((item) => item.trim().toLowerCase()));
        }
      } catch {
        if (mounted) setPermissions([]);
      }
    }
    void loadPermissions();
    return () => {
      mounted = false;
    };
  }, []);

  const navItems: NavItem[] = useMemo(() => [
    {
      href: "/",
      title: t("nav.dashboard"),
      icon: LayoutDashboard,
      isActive: pathname === "/",
      requiredPermissions: ["settings:list"],
    },
    {
      href: "/billing/topup",
      title: t("nav.topup"),
      icon: CreditCard,
      isActive: pathname.startsWith("/billing/topup"),
      requiredPermissions: ["settings:list"],
    },
    {
      href: "/provider-accounts",
      title: t("nav.providerAccounts"),
      icon: ShieldCheck,
      isActive: pathname.startsWith("/provider-accounts"),
      requiredPermissions: ["provider_account:read", "provider_account:manage", "provider_account:create"],
    },
    {
      href: "/virtual-models",
      title: t("nav.virtualModels"),
      icon: Route,
      isActive: pathname.startsWith("/virtual-models"),
      requiredPermissions: ["virtual_account:read", "virtual_account:manage", "virtual_account:create"],
    },
    {
      href: "/users",
      title: t("nav.accessManagement"),
      icon: Users,
      isActive: pathname.startsWith("/users") || pathname.startsWith("/roles"),
      requiredPermissions: ["user:list", "role:list"],
    },
    {
      href: "/traces",
      title: t("nav.traces"),
      icon: Activity,
      isActive: pathname.startsWith("/traces"),
      requiredPermissions: ["settings:list"],
    },
    {
      href: "/metrics",
      title: t("nav.metrics"),
      icon: BarChart3,
      isActive: pathname.startsWith("/metrics"),
      requiredPermissions: ["settings:list"],
    },
    {
      href: "/teams",
      title: t("nav.teams"),
      icon: Users,
      isActive: pathname.startsWith("/teams"),
      requiredPermissions: ["team:read", "team:manage"],
    },
    {
      href: "/org-units",
      title: t("nav.orgUnits"),
      icon: Building2,
      isActive: pathname.startsWith("/org-units"),
      requiredPermissions: ["org_unit:list", "org_unit:read"],
    },
    {
      href: "/gateway-policies/rate-limiting-rules",
      title: t("nav.rateLimitPolicies"),
      icon: ShieldCheck,
      isActive: pathname.startsWith("/gateway-policies/rate-limiting-rules"),
      requiredPermissions: ["policy:list"],
    },
    {
      href: "/gateway-policies/budget-limiting-rules",
      title: t("nav.budgetLimitPolicies"),
      icon: ShieldCheck,
      isActive: pathname.startsWith("/gateway-policies/budget-limiting-rules"),
      requiredPermissions: ["policy:list"],
    },
    {
      href: "/settings",
      title: t("nav.settings"),
      icon: Settings,
      isActive: pathname.startsWith("/settings"),
      requiredPermissions: ["settings:list"],
    },
  ], [pathname, t]);

  const visibleNavItems = useMemo(() => {
    if (permissions === null) return [];
    return navItems.filter((item) => canAnyPermission(permissions, item.requiredPermissions));
  }, [navItems, permissions]);

  const currentPageLabel = visibleNavItems.find((item) => item.isActive)?.title ?? t("nav.dashboard");

  return (
    <div className="min-h-screen bg-background app-background">
      <div className="grid min-h-screen lg:grid-cols-[280px_1fr]">
        <aside className="hidden lg:block">
          <SidebarContent
            tenantSlug={tenantSlug}
            currentUser={currentUser}
            userIdentity={userIdentity}
            navItems={visibleNavItems}
          />
        </aside>

        <div className="min-w-0">
          <header className="sticky top-0 z-sticky border-b bg-background/80 px-4 py-3 backdrop-blur md:px-6">
            <div className="flex items-center gap-2 md:gap-3">
              <Sheet>
                <SheetTrigger asChild>
                  <Button variant="outline" size="icon" className="lg:hidden" aria-label={t("menu.open")}>
                    <Menu className="h-5 w-5" />
                    <span className="sr-only">{t("menu.open")}</span>
                  </Button>
                </SheetTrigger>
                <SheetContent side="right" className="w-[86vw] max-w-sm p-0">
                  <SheetHeader className="px-4 pt-4 text-right">
                    <SheetTitle>{t("title")}</SheetTitle>
                    <SheetDescription>{t("menu.description")}</SheetDescription>
                  </SheetHeader>
                  <SidebarContent
                    tenantSlug={tenantSlug}
                    currentUser={currentUser}
                    userIdentity={userIdentity}
                    navItems={visibleNavItems}
                    className="border-none"
                  />
                </SheetContent>
              </Sheet>

              <div className="min-w-0 flex-1">
                <p className="text-xs text-muted-foreground">{t("tenant")}</p>
                <h1 className="truncate text-base font-semibold md:text-lg">{currentPageLabel}</h1>
              </div>

              <div className="hidden w-full max-w-sm items-center gap-2 md:flex">
                <Search className="h-4 w-4 text-muted-foreground" aria-hidden="true" />
                <Input
                  placeholder={t("searchPlaceholder")}
                  aria-label={t("searchPlaceholder")}
                  className="max-w-xs"
                />
              </div>

              <ThemeToggle />
            </div>
          </header>

          <main className="p-4 md:p-6">
            <div className="mx-auto w-full max-w-7xl animate-fade-in">{children}</div>
          </main>
        </div>
      </div>
    </div>
  );
}

type SidebarContentProps = {
  className?: string;
  tenantSlug: string;
  userIdentity: string;
  currentUser: {
    role: string;
  };
  navItems: NavItem[];
};

function SidebarContent({
  className,
  tenantSlug,
  userIdentity,
  currentUser,
  navItems,
}: SidebarContentProps) {
  const t = useTranslations("shell");

  return (
    <div className={cn("flex h-full flex-col bg-sidebar p-4", className)}>
      <div className="flex items-center gap-3 rounded-xl border border-sidebar-border bg-sidebar p-3 shadow-xs">
        <div className="flex h-10 w-10 items-center justify-center overflow-hidden rounded-lg border border-sidebar-border bg-sidebar-accent p-1.5">
          <Image
            src="/brand/shahrah-logo.png"
            alt={t("logoAlt")}
            width={32}
            height={32}
            className="h-full w-full object-contain"
            priority
          />
        </div>
        <div className="min-w-0">
          <p className="truncate text-sm font-semibold text-sidebar-foreground">{t("title")}</p>
          <p className="truncate text-xs text-sidebar-accent-foreground">{tenantSlug}</p>
        </div>
      </div>

      <nav className="mt-6 grid gap-1.5" aria-label={t("nav.title")}>
        {navItems.map((item) => {
          const Icon = item.icon;

          return (
            <Button
              key={item.href}
              variant={item.isActive ? "default" : "ghost"}
              size="default"
              className={cn(
                "h-10 justify-start gap-3 px-3 transition-all duration-150",
                item.isActive && "shadow-xs",
              )}
              asChild
            >
              <Link href={item.href}>
                <Icon className="h-4 w-4 flex-shrink-0" aria-hidden="true" />
                <span className="truncate">{item.title}</span>
              </Link>
            </Button>
          );
        })}
      </nav>

      <div className="flex-1" />

      <Separator className="my-4" />

      <div className="rounded-xl border border-sidebar-border bg-sidebar p-3 shadow-xs">
        <div className="flex items-center gap-3">
          <Avatar className="h-10 w-10">
            <AvatarFallback className="text-xs font-semibold">
              {getInitials(userIdentity)}
            </AvatarFallback>
          </Avatar>
          <div className="min-w-0 flex-1">
            <p className="truncate text-sm font-medium text-sidebar-foreground">{userIdentity}</p>
            <Badge variant="secondary" className="mt-1 text-xs">
              {currentUser.role}
            </Badge>
          </div>
        </div>

        <form action="/logout" method="post" className="mt-3">
          <Button
            variant="outline"
            size="sm"
            className="w-full justify-start gap-2"
          >
            <LogOut className="h-4 w-4" aria-hidden="true" />
            {t("logout")}
          </Button>
        </form>
      </div>
    </div>
  );
}
