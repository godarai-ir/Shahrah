"use client";

import { useMemo } from "react";
import Image from "next/image";
import Link from "next/link";
import { usePathname } from "next/navigation";
import { useTranslations } from "next-intl";
import { LogOut, Search } from "lucide-react";

import { Avatar, AvatarFallback } from "@/components/ui/avatar";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
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
  SidebarSeparator,
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

  const currentPageLabel = visibleNavItems.find((item) => item.isActive)?.title ?? t("nav.dashboard");

  return (
    <PermissionsProvider permissions={permissions}>
      <SidebarProvider className="bg-background app-background">
        <AdminSidebar
          tenantSlug={tenantSlug}
          currentUser={currentUser}
          userIdentity={userIdentity}
          navItems={visibleNavItems}
        />

        <SidebarInset className="min-w-0 bg-transparent">
          <header className="sticky top-0 z-10 border-b border-border bg-background/95 px-4 py-3 backdrop-blur supports-[backdrop-filter]:bg-background/85 md:px-6">
            <div className="flex min-w-0 items-center gap-3">
              <SidebarTrigger aria-label={t("menu.open")} />

              <div className="min-w-0 flex-1">
                <p className="text-xs text-muted-foreground">{t("tenant")}</p>
                <h1 className="truncate text-base font-semibold md:text-lg">{currentPageLabel}</h1>
              </div>

              <div className="hidden w-full max-w-sm items-center gap-2 md:flex">
                <Search className="size-4 text-muted-foreground" aria-hidden="true" />
                <Input
                  placeholder={t("searchPlaceholder")}
                  aria-label={t("searchPlaceholder")}
                  className="max-w-xs"
                />
              </div>

              <ThemeToggle />
            </div>
          </header>

          <div className="p-4 pb-[max(1rem,env(safe-area-inset-bottom))] md:p-6 md:pb-[max(1.5rem,env(safe-area-inset-bottom))]">
            <div className="mx-auto w-full max-w-7xl animate-fade-in">{children}</div>
          </div>
        </SidebarInset>
      </SidebarProvider>
    </PermissionsProvider>
  );
}

type SidebarContentProps = {
  tenantSlug: string;
  userIdentity: string;
  currentUser: {
    role: string;
  };
  navItems: AdminNavItem[];
};

function AdminSidebar({
  tenantSlug,
  userIdentity,
  currentUser,
  navItems,
}: SidebarContentProps) {
  const t = useTranslations("shell");

  return (
    <Sidebar side="right" collapsible="icon">
      <SidebarHeader>
        <SidebarMenu>
          <SidebarMenuItem>
            <SidebarMenuButton
              size="lg"
              tooltip={t("title")}
              className="h-14 items-center gap-3 rounded-lg border border-sidebar-border bg-sidebar"
              asChild
            >
              <Link href="/">
                <span className="flex size-10 shrink-0 items-center justify-center overflow-hidden rounded-lg border border-sidebar-border bg-sidebar-accent p-1.5">
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
        <SidebarGroup>
          <SidebarGroupLabel>{t("nav.title")}</SidebarGroupLabel>
          <SidebarGroupContent>
            <SidebarMenu aria-label={t("nav.title")}>
              {navItems.map((item) => {
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
      </SidebarContent>

      <SidebarFooter>
        <SidebarSeparator />
        <SidebarMenu>
          <SidebarMenuItem>
            <div className="flex items-center gap-3 rounded-lg border border-sidebar-border bg-sidebar p-3 group-data-[collapsible=icon]:justify-center group-data-[collapsible=icon]:border-transparent group-data-[collapsible=icon]:bg-transparent group-data-[collapsible=icon]:p-0">
              <Avatar className="size-10">
                <AvatarFallback className="text-xs font-semibold">
                  {getInitials(userIdentity)}
                </AvatarFallback>
              </Avatar>
              <div className="min-w-0 flex-1 group-data-[collapsible=icon]:hidden">
                <p className="truncate text-sm font-medium text-sidebar-foreground">{userIdentity}</p>
                <Badge variant="secondary" className="mt-1 text-xs">
                  {currentUser.role}
                </Badge>
              </div>
            </div>
          </SidebarMenuItem>
          <SidebarMenuItem>
            <form action="/logout" method="post">
              <SidebarMenuButton asChild tooltip={t("logout")}>
                <Button
                  type="submit"
                  variant="outline"
                  size="sm"
                  className="w-full justify-start"
                >
                  <LogOut aria-hidden="true" />
                  <span>{t("logout")}</span>
                </Button>
              </SidebarMenuButton>
            </form>
          </SidebarMenuItem>
        </SidebarMenu>
      </SidebarFooter>

      <SidebarRail />
    </Sidebar>
  );
}
