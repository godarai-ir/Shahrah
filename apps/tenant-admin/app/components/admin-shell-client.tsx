"use client";

import { useMemo } from "react";
import Image from "next/image";
import Link from "next/link";
import { usePathname, useRouter } from "next/navigation";
import { useLocale, useTranslations } from "next-intl";
import { BadgeCheck, BookOpen, Building2, ChevronDown, Eye, LogOut, Plug, ShieldCheck, User, Users, WalletCards } from "lucide-react";

import { Avatar, AvatarFallback } from "@/components/ui/avatar";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Collapsible, CollapsibleContent, CollapsibleTrigger } from "@/components/ui/collapsible";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuGroup,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
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
  SidebarTrigger,
  useSidebar,
} from "@/components/ui/sidebar";
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from "@/components/ui/tooltip";
import { ThemeToggle } from "@/components/theme-toggle";
import { ACCESS_NAV_ITEM_KEYS, buildAdminNavItems, filterVisibleNavItems, shouldHandleActiveNavClick, type AdminNavItem } from "@/app/components/admin-shell-nav";
import { PageHeaderInShellProvider } from "@/components/app-ui/PageHeader";
import { ShellPageHeading, type ShellPageMetadata } from "@/components/app-ui/ShellPageHeading";
import { PermissionsProvider } from "@/app/components/permissions-context";
import { mobileAuthUIEnabled } from "@/lib/auth/mobile-auth-ui";
import { cn } from "@/lib/utils";

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
  walletSummary: {
    availableBalanceIRR: number | null;
    ledgerBalanceIRR: number | null;
    committedUnsettledIRR: number;
    reservedIRR: number;
    syncStatus: string | null;
  };
  subscriptionSummary: {
    planName: string | null;
    status: string | null;
    periodEnd: string | null;
  };
  docsHref: string;
  appVersion: string;
  pageHeader?: ShellPageMetadata;
  contentPadding?: "default" | "none";
};

const moneyFormatter = new Intl.NumberFormat("fa-IR", { maximumFractionDigits: 0 });

const SIDEBAR_NAV_GROUPS: Array<{
  labelKey: string;
  itemKeys: AdminNavItem["key"][];
  icon: React.ComponentType<{ className?: string }>;
}> = [
  {
    labelKey: "navGroups.workspace",
    itemKeys: ["dashboard", "quickStart"],
    icon: Building2,
  },
  {
    labelKey: "navGroups.runtime",
    itemKeys: ["providerAccounts", "virtualModels", "playground", "prompts"],
    icon: Plug,
  },
  {
    labelKey: "navGroups.observability",
    itemKeys: ["logs", "traces", "metrics"],
    icon: Eye,
  },
  {
    labelKey: "navGroups.policies",
    itemKeys: ["rateLimitPolicies", "budgetLimitPolicies", "guardrails"],
    icon: ShieldCheck,
  },
  {
    labelKey: "navGroups.access",
    itemKeys: ACCESS_NAV_ITEM_KEYS,
    icon: Users,
  },
  {
    labelKey: "navGroups.billing",
    itemKeys: ["billing", "topup", "bankTransfer", "billingReports", "billingProfile", "billingDocuments"],
    icon: WalletCards,
  },
];

export default function AdminShellClient({
  children,
  tenantSlug,
  currentUser,
  permissions,
  walletSummary,
  subscriptionSummary,
  docsHref,
  appVersion,
  pageHeader,
  contentPadding = "default",
}: AdminShellClientProps) {
  const pathname = usePathname();
  const router = useRouter();
  const t = useTranslations("shell");

  const userIdentity =
    currentUser.email ?? (mobileAuthUIEnabled ? currentUser.mobile : undefined) ?? currentUser.id;

  const navItems = useMemo(() => buildAdminNavItems(pathname, t), [pathname, t]);

  const visibleNavItems = useMemo(() => {
    return filterVisibleNavItems(navItems, permissions);
  }, [navItems, permissions]);

  const currentPageLabel = visibleNavItems.find((item) => item.isActive)?.title ?? t("nav.dashboard");

  return (
    <PermissionsProvider permissions={permissions}>
      <SidebarProvider className="bg-background">
        <a href="#main-content" className="skip-link">
          {t("skipToContent")}
        </a>
        <AdminSidebar
          tenantSlug={tenantSlug}
          navItems={visibleNavItems}
          subscriptionSummary={subscriptionSummary}
          appVersion={appVersion}
          onNavigate={(href) => {
            router.replace(href);
            router.refresh();
          }}
        />

        <SidebarInset
          id="main-content"
          tabIndex={-1}
          className="atlas-canvas min-w-0 rounded-none bg-background"
        >
          <header className="sticky top-0 z-10 border-b border-border/80 bg-background/90 px-4 py-2.5 backdrop-blur-md supports-[backdrop-filter]:bg-background/80 md:px-6">
            <div className="flex min-w-0 items-center gap-3">
              <SidebarTrigger
                className="border border-border bg-card shadow-xs"
                aria-label={t("menu.open")}
              />

              <div className="min-w-0 flex-1">
                {pageHeader ? (
                  <ShellPageHeading title={pageHeader.title} description={pageHeader.description} />
                ) : (
                  <p className="truncate text-base font-semibold text-foreground md:text-lg">{currentPageLabel}</p>
                )}
              </div>

              <div className="flex shrink-0 items-center gap-1.5 md:gap-2">
                <WalletBalanceLink summary={walletSummary} />
                <UserProfileMenu userIdentity={userIdentity} currentUser={currentUser} />
                <ThemeToggle />
                <div className="hidden sm:block">
                  <DocsLink href={pageHeader?.helpHref ?? docsHref} />
                </div>
              </div>
            </div>
          </header>

          <div className={cn(
            "relative min-h-0 min-w-0 flex-1",
            contentPadding === "default" && "p-4 pb-[max(1rem,env(safe-area-inset-bottom))] md:p-6 md:pb-[max(1.5rem,env(safe-area-inset-bottom))]",
            contentPadding === "none" && "overflow-hidden",
          )}>
            <PageHeaderInShellProvider enabled={Boolean(pageHeader)}>
              <div className={cn("w-full animate-fade-in", contentPadding === "none" && "h-full")}>{children}</div>
            </PageHeaderInShellProvider>
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
  const userInitials = getUserInitials(userIdentity);

  async function handleLogout() {
    await fetch("/api/auth/logout", { method: "POST" });
    router.replace("/login");
    router.refresh();
  }

  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button
          type="button"
          variant="outline"
          size="icon"
          className="shrink-0 p-1"
          aria-label={t("profile.menu")}
        >
          <Avatar className="size-7 rounded-md">
            <AvatarFallback
              className="rounded-md bg-primary-soft text-xs font-semibold text-primary-soft-foreground"
              dir="ltr"
            >
              {userInitials}
            </AvatarFallback>
          </Avatar>
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="start" sideOffset={8} className="w-72 p-2 text-start">
        <DropdownMenuGroup>
          <DropdownMenuLabel className="p-0 font-normal">
            <div className="flex items-start gap-3 rounded-md bg-muted/50 p-3">
              <Avatar className="size-11 rounded-lg">
                <AvatarFallback className="rounded-lg" dir="ltr">
                  {userInitials}
                </AvatarFallback>
              </Avatar>
              <span className="flex min-w-0 flex-1 flex-col gap-1.5">
                <span dir="ltr" className="truncate text-sm font-semibold text-popover-foreground">
                  {userIdentity}
                </span>
                <Badge variant="secondary" className="w-fit max-w-full truncate text-xs" dir="ltr">
                  {currentUser.role}
                </Badge>
              </span>
            </div>
          </DropdownMenuLabel>
        </DropdownMenuGroup>
        <DropdownMenuSeparator />
        <DropdownMenuGroup>
          <DropdownMenuItem asChild className="min-h-10 justify-start gap-3 rounded-md px-3">
            <Link href="/settings/account">
              <User aria-hidden="true" />
              <span>{t("profile.account")}</span>
            </Link>
          </DropdownMenuItem>
        </DropdownMenuGroup>
        <DropdownMenuSeparator />
        <DropdownMenuGroup>
          <DropdownMenuItem
            className="min-h-10 justify-start gap-3 rounded-md px-3 text-destructive focus:text-destructive"
            onSelect={(event) => {
              event.preventDefault();
              void handleLogout();
            }}
          >
            <LogOut aria-hidden="true" />
            <span>{t("logout")}</span>
          </DropdownMenuItem>
        </DropdownMenuGroup>
      </DropdownMenuContent>
    </DropdownMenu>
  );
}

type SidebarContentProps = {
  tenantSlug: string;
  navItems: AdminNavItem[];
  subscriptionSummary: {
    planName: string | null;
    status: string | null;
    periodEnd: string | null;
  };
  appVersion: string;
  onNavigate: (href: string) => void;
};

function AdminSidebar({
  tenantSlug,
  navItems,
  subscriptionSummary,
  appVersion,
  onNavigate,
}: SidebarContentProps) {
  const t = useTranslations("shell");
  const locale = useLocale();
  const { setOpenMobile } = useSidebar();
  const direction = locale === "fa" ? "rtl" : "ltr";
  const sidebarSide = direction === "rtl" ? "right" : "left";
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
    <Sidebar
      side={sidebarSide}
      dir={direction}
      collapsible="icon"
      mobileTitle={t("menu.title")}
      mobileDescription={t("menu.description")}
      mobileCloseLabel={t("menu.close")}
    >
      <SidebarHeader className="gap-3 border-b border-sidebar-border/80 p-3">
        <SidebarMenu>
          <SidebarMenuItem>
            <SidebarMenuButton
              size="lg"
              tooltip={`گُدارAI · ${tenantSlug}`}
              className="items-center gap-3 after:hidden hover:bg-sidebar-accent/70"
              asChild
            >
              <Link href="/" onClick={() => setOpenMobile(false)}>
                <span className="flex size-10 shrink-0 items-center justify-center overflow-hidden rounded-lg bg-sidebar-primary/15 p-1 ring-1 ring-sidebar-primary/35">
                  <Image
                    src="/brand/mini_logo.png"
                    alt={t("logoAlt")}
                    width={36}
                    height={36}
                    className="h-full w-full object-contain"
                    priority
                  />
                </span>
                <span className="flex min-w-0 flex-col gap-0.5">
                  <span className="truncate text-base font-semibold text-sidebar-foreground">گُدارAI</span>
                  <span className="truncate text-xs text-sidebar-foreground/55">{t("title")}</span>
                </span>
              </Link>
            </SidebarMenuButton>
          </SidebarMenuItem>
          <SidebarMenuItem>
            <SidebarMenuButton
              size="lg"
              variant="outline"
              tooltip={`${t("tenant")} · ${tenantSlug}`}
              className="after:hidden"
              asChild
            >
              <div>
                <span className="flex size-8 shrink-0 items-center justify-center rounded-md bg-sidebar-primary/20 text-xs font-semibold text-sidebar-primary-foreground ring-1 ring-sidebar-primary/30">
                  {getTenantInitials(tenantSlug)}
                </span>
                <span className="flex min-w-0 flex-col gap-0.5">
                  <span className="truncate text-xs text-sidebar-foreground/55">{t("tenant")}</span>
                  <span className="truncate font-mono text-xs font-medium text-sidebar-foreground" dir="ltr">
                    {tenantSlug}
                  </span>
                </span>
              </div>
            </SidebarMenuButton>
          </SidebarMenuItem>
        </SidebarMenu>
      </SidebarHeader>

      <SidebarContent className="custom-scrollbar gap-1 px-1 py-3" role="navigation" aria-label={t("nav.title")}>
        {navGroups.map((group) => (
          <Collapsible key={group.labelKey} defaultOpen={group.items.some((item) => item.isActive)} asChild>
            <SidebarGroup className="py-1">
              <SidebarGroupLabel asChild>
                <CollapsibleTrigger className="group/collapsible flex w-full items-center justify-between gap-3 hover:bg-sidebar-accent hover:text-sidebar-accent-foreground">
                  <span className="flex min-w-0 items-center gap-2">
                    <group.icon className="size-4 shrink-0 text-sidebar-foreground/55" aria-hidden="true" />
                    <span className="truncate">{t(group.labelKey)}</span>
                  </span>
                  <ChevronDown className="size-3.5 shrink-0 transition-transform group-data-[state=open]/collapsible:rotate-180" aria-hidden="true" />
                </CollapsibleTrigger>
              </SidebarGroupLabel>
              <CollapsibleContent>
                <SidebarGroupContent>
                  <SidebarMenu aria-label={t(group.labelKey)} className="gap-0.5">
                    {group.items.map((item) => {
                      const Icon = item.icon;

                      return (
                        <SidebarMenuItem key={item.href}>
                          <SidebarMenuButton
                            isActive={item.isActive}
                            tooltip={item.title}
                            asChild
                          >
                            <Link
                              href={item.href}
                              aria-current={item.isActive ? "page" : undefined}
                              onClick={(event) => {
                                setOpenMobile(false);
                                if (!item.isActive || !shouldHandleActiveNavClick(event.nativeEvent)) return;
                                event.preventDefault();
                                onNavigate(item.href);
                              }}
                            >
                              <Icon aria-hidden="true" />
                              <span>{item.title}</span>
                            </Link>
                          </SidebarMenuButton>
                        </SidebarMenuItem>
                      );
                    })}
                  </SidebarMenu>
                </SidebarGroupContent>
              </CollapsibleContent>
            </SidebarGroup>
          </Collapsible>
        ))}
      </SidebarContent>

      <SidebarFooter className="border-t border-sidebar-border/80 bg-sidebar-accent/25 p-3 group-data-[collapsible=icon]:hidden">
        <SidebarPlanSummary summary={subscriptionSummary} onNavigate={() => setOpenMobile(false)} />
        <SidebarAppVersion version={appVersion} />
      </SidebarFooter>
      <SidebarRail aria-label={t("menu.toggle")} title={t("menu.toggle")} />
    </Sidebar>
  );
}

function WalletBalanceLink({
  summary,
}: {
  summary: {
    availableBalanceIRR: number | null;
    ledgerBalanceIRR: number | null;
    committedUnsettledIRR: number;
    reservedIRR: number;
    syncStatus: string | null;
  };
}) {
  const t = useTranslations("shell");
  const balanceLabel =
    summary.availableBalanceIRR === null
      ? t("wallet.unknown")
      : t("wallet.available", { amount: moneyFormatter.format(summary.availableBalanceIRR) });
  const isLowBalance = summary.availableBalanceIRR !== null && summary.availableBalanceIRR <= 0;

  return (
    <Button
      type="button"
      variant="outline"
      size="icon"
      className={cn(
        "hidden h-10 w-auto min-w-10 gap-2 px-3 md:inline-flex",
        isLowBalance && "border-destructive/40 text-destructive hover:text-destructive",
      )}
      asChild
    >
      <Link href="/billing" aria-label={t("wallet.open")}>
        <WalletCards data-icon="inline-start" aria-hidden="true" />
        <span className="hidden truncate text-xs font-medium sm:inline">{balanceLabel}</span>
      </Link>
    </Button>
  );
}

function DocsLink({ href }: { href: string }) {
  const t = useTranslations("shell");
  const label = t("docs.open");

  return (
    <TooltipProvider delayDuration={120}>
      <Tooltip>
        <TooltipTrigger asChild>
          <Button type="button" variant="outline" size="icon" className="shrink-0" asChild>
            <Link href={href} target="_blank" rel="noopener noreferrer" aria-label={label}>
              <BookOpen data-icon="inline-start" aria-hidden="true" />
            </Link>
          </Button>
        </TooltipTrigger>
        <TooltipContent side="bottom">{label}</TooltipContent>
      </Tooltip>
    </TooltipProvider>
  );
}

function getUserInitials(value: string): string {
  const trimmed = value.trim();
  if (!trimmed) return "؟";

  const [localPart] = trimmed.split("@");
  const source = localPart || trimmed;
  const parts = source
    .split(/[._\s-]+/)
    .map((part) => part.trim())
    .filter(Boolean);

  if (parts.length >= 2) {
    return `${parts[0][0] ?? ""}${parts[1][0] ?? ""}`.toUpperCase();
  }

  return source.slice(0, 2).toUpperCase();
}

function getTenantInitials(value: string): string {
  const normalized = value.trim().replace(/[^a-zA-Z0-9]/g, "");
  return (normalized.slice(0, 2) || "WS").toUpperCase();
}

function SidebarPlanSummary({
  summary,
  onNavigate,
}: {
  summary: {
    planName: string | null;
    status: string | null;
    periodEnd: string | null;
  };
  onNavigate: () => void;
}) {
  const t = useTranslations("shell");
  const planName = t("plan.label", { name: summary.planName ?? t("plan.unknown") });
  const isExpired = ["expired", "canceled"].includes(summary.status?.trim().toLowerCase() ?? "");
  const periodLabel = isExpired
    ? t("plan.expired")
    : summary.periodEnd
      ? formatDaysRemaining(summary.periodEnd, t)
      : t("plan.noEndDate");

  return (
    <SidebarMenu>
      <SidebarMenuItem>
        <SidebarMenuButton size="lg" variant="outline" tooltip={t("nav.plan")} asChild>
          <Link href="/billing/plan" onClick={onNavigate}>
            <BadgeCheck aria-hidden="true" />
            <span className="flex min-w-0 flex-col gap-0.5">
              <span className="truncate text-sm font-medium">{planName}</span>
              <span className={cn("truncate text-xs font-normal", isExpired ? "text-warning" : "text-sidebar-foreground/60")}>{periodLabel}</span>
            </span>
          </Link>
        </SidebarMenuButton>
      </SidebarMenuItem>
    </SidebarMenu>
  );
}

function SidebarAppVersion({ version }: { version: string }) {
  const t = useTranslations("shell");

  return (
    <p className="border-t border-sidebar-border/70 pt-3 text-center text-xs text-sidebar-foreground/70">
      {t("appVersion.label")} <span dir="ltr" className="font-mono">{version}</span>
    </p>
  );
}

function formatDaysRemaining(value: string, t: (key: string, values?: Record<string, string | number>) => string): string {
  const timestamp = Date.parse(value);
  if (Number.isNaN(timestamp)) return value;
  const days = Math.max(0, Math.ceil((timestamp - Date.now()) / (24 * 60 * 60 * 1000)));
  return t("plan.daysRemaining", { days: moneyFormatter.format(days) });
}
