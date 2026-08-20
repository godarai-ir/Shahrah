"use client";

import { useEffect, useMemo, useRef } from "react";
import Link from "next/link";
import { usePathname, useRouter } from "next/navigation";
import { useLocale, useTranslations } from "next-intl";
import { BookOpen, Bot, ChevronDown, Eye, LogOut, Plug, ShieldCheck, User, Users, WalletCards } from "lucide-react";

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
import { ThemeToggleMenuItem } from "@/components/theme-toggle";
import { ACCESS_NAV_ITEM_KEYS, buildAdminNavItems, filterVisibleNavItems, shouldHandleActiveNavClick, type AdminNavItem } from "@/features/admin-shell/navigation";
import { logout } from "@/features/admin-shell/api";
import { PageHeaderInShellProvider } from "@/components/app-ui/PageHeader";
import { ShellPageHeading } from "@/components/app-ui/ShellPageHeading";
import { GodarAILogo } from "@/components/app-ui/GodarAILogo";
import { PermissionsProvider } from "@/components/permissions-context";
import { AdminShellPageProvider, useAdminShellPage } from "@/features/admin-shell/page-context";
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
};

const moneyFormatter = new Intl.NumberFormat("fa-IR", { maximumFractionDigits: 0 });
const ROOT_NAV_ITEM_KEYS: AdminNavItem["key"][] = ["dashboard", "quickStart"];

const SIDEBAR_NAV_GROUPS: Array<{
  labelKey: string;
  itemKeys: AdminNavItem["key"][];
  icon: React.ComponentType<{ className?: string }>;
}> = [
  {
    labelKey: "navGroups.models",
    itemKeys: ["providerAccounts", "virtualModels"],
    icon: Plug,
  },
  {
    labelKey: "navGroups.prompts",
    itemKeys: ["prompts", "playground"],
    icon: BookOpen,
  },
  {
    labelKey: "navGroups.knowledge",
    itemKeys: ["knowledgeBases", "chatbots"],
    icon: Bot,
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
  ...props
}: AdminShellClientProps) {
  return (
    <AdminShellPageProvider>
      <AdminShellClientContent {...props} />
    </AdminShellPageProvider>
  );
}

function AdminShellClientContent({
  children,
  tenantSlug,
  currentUser,
  permissions,
  walletSummary,
  subscriptionSummary,
  docsHref,
  appVersion,
}: AdminShellClientProps) {
  const pathname = usePathname();
  const router = useRouter();
  const t = useTranslations("shell");
  const { pageHeader, contentPadding } = useAdminShellPage();

  const userIdentity =
    currentUser.email ?? (mobileAuthUIEnabled ? currentUser.mobile : undefined) ?? currentUser.id;

  const navItems = useMemo(() => buildAdminNavItems(pathname, t), [pathname, t]);

  const visibleNavItems = useMemo(() => {
    return filterVisibleNavItems(navItems, permissions);
  }, [navItems, permissions]);

  const currentPageLabel = visibleNavItems.find((item) => item.isActive)?.title ?? t("nav.dashboard");
  const supportNavItem = visibleNavItems.find((item) => item.key === "supportTickets");

  return (
    <PermissionsProvider permissions={permissions}>
      <SidebarProvider defaultOpen={!pathname.startsWith("/playground")} className="bg-background">
        <SidebarRouteStateSync pathname={pathname} />
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
          className="theme-canvas min-w-0 rounded-none bg-background"
        >
          <header className="sticky top-0 z-10 h-16 border-b border-border bg-background px-4 md:px-6">
            <div className="flex h-full min-w-0 items-center gap-3">
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
                {supportNavItem ? <SupportLink item={supportNavItem} /> : null}
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
              <div className={cn("w-full", contentPadding === "none" && "h-full")}>{children}</div>
            </PageHeaderInShellProvider>
          </div>
        </SidebarInset>
      </SidebarProvider>
    </PermissionsProvider>
  );
}

function SidebarRouteStateSync({ pathname }: { pathname: string }) {
  const { setOpen, setOpenMobile } = useSidebar();
  const previousPathname = useRef<string | null>(null);

  useEffect(() => {
    if (previousPathname.current === pathname) {
      return;
    }

    previousPathname.current = pathname;

    if (pathname.startsWith("/playground")) {
      setOpen(false);
      setOpenMobile(false);
    }
  }, [pathname, setOpen, setOpenMobile]);

  return null;
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
    await logout();
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
          <ThemeToggleMenuItem />
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
  const { isMobile, setOpenMobile, state } = useSidebar();
  const direction = locale === "fa" ? "rtl" : "ltr";
  const sidebarSide = direction === "rtl" ? "right" : "left";
  const tooltipSide = direction === "rtl" ? "left" : "right";
  const showCompactNavigation = !isMobile && state === "collapsed";
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
  const rootNavItems = useMemo(() => {
    return ROOT_NAV_ITEM_KEYS
      .map((key) => navItemsByKey.get(key))
      .filter((item): item is AdminNavItem => Boolean(item));
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
      <SidebarHeader className="h-16 flex-none justify-center gap-0 border-b border-sidebar-border! px-3 py-0">
        <SidebarMenu>
          <SidebarMenuItem>
            <SidebarMenuButton
              size="lg"
              tooltip={{ children: `گدارAI · ${tenantSlug}`, side: tooltipSide }}
              className="items-center gap-3 after:hidden hover:bg-sidebar-accent/70"
              asChild
            >
              <Link href="/" onClick={() => setOpenMobile(false)}>
                <span className="flex size-10 shrink-0 items-center justify-center overflow-hidden rounded-lg bg-sidebar-primary/15 p-1 ring-1 ring-sidebar-primary/35">
                  <GodarAILogo
                    alt={t("logoAlt")}
                    className="h-full w-auto max-w-full object-contain"
                    priority
                    sizes="2.5rem"
                  />
                </span>
                <span className="flex min-w-0 flex-col gap-0.5">
                  <span className="truncate text-base font-semibold text-sidebar-foreground">گدارAI</span>
                  <span className="truncate text-xs text-sidebar-foreground/55">{t("title")}</span>
                </span>
              </Link>
            </SidebarMenuButton>
          </SidebarMenuItem>
        </SidebarMenu>
      </SidebarHeader>

      <SidebarContent
        className="custom-scrollbar gap-1 overflow-x-hidden px-1 py-3 group-data-[collapsible=icon]:overflow-y-auto!"
        role="navigation"
        aria-label={t("nav.title")}
      >
        {rootNavItems.length > 0 ? (
          <SidebarGroup className="py-1">
            <SidebarGroupContent>
              <SidebarMenu aria-label={t("nav.title")} className="gap-0.5">
                {rootNavItems.map((item) => (
                  <AdminSidebarNavItem
                    key={item.href}
                    item={item}
                    tooltipSide={tooltipSide}
                    onCloseMobile={() => setOpenMobile(false)}
                    onNavigate={onNavigate}
                  />
                ))}
              </SidebarMenu>
            </SidebarGroupContent>
          </SidebarGroup>
        ) : null}

        {showCompactNavigation
          ? navGroups.map((group) => (
              <SidebarGroup key={group.labelKey} className="py-1">
                <SidebarGroupContent>
                  <SidebarMenu aria-label={t(group.labelKey)}>
                    {group.items.map((item) => (
                      <AdminSidebarNavItem
                        key={item.href}
                        item={item}
                        tooltipSide={tooltipSide}
                        onCloseMobile={() => setOpenMobile(false)}
                        onNavigate={onNavigate}
                      />
                    ))}
                  </SidebarMenu>
                </SidebarGroupContent>
              </SidebarGroup>
            ))
          : navGroups.map((group) => (
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
                        {group.items.map((item) => (
                          <AdminSidebarNavItem
                            key={item.href}
                            item={item}
                            tooltipSide={tooltipSide}
                            onCloseMobile={() => setOpenMobile(false)}
                            onNavigate={onNavigate}
                          />
                        ))}
                      </SidebarMenu>
                    </SidebarGroupContent>
                  </CollapsibleContent>
                </SidebarGroup>
              </Collapsible>
            ))}
      </SidebarContent>

      <SidebarFooter className="border-t border-sidebar-border! bg-sidebar-accent/20 p-3 group-data-[collapsible=icon]:hidden">
        <SidebarPlanSummary summary={subscriptionSummary} appVersion={appVersion} onNavigate={() => setOpenMobile(false)} />
      </SidebarFooter>
      <SidebarRail aria-label={t("menu.toggle")} title={t("menu.toggle")} />
    </Sidebar>
  );
}

function AdminSidebarNavItem({
  item,
  tooltipSide,
  onCloseMobile,
  onNavigate,
}: {
  item: AdminNavItem;
  tooltipSide: React.ComponentProps<typeof TooltipContent>["side"];
  onCloseMobile: () => void;
  onNavigate: (href: string) => void;
}) {
  const Icon = item.icon;

  return (
    <SidebarMenuItem>
      <SidebarMenuButton
        isActive={item.isActive}
        tooltip={{ children: item.title, side: tooltipSide }}
        asChild
      >
        <Link
          href={item.href}
          aria-current={item.isActive ? "page" : undefined}
          onClick={(event) => {
            onCloseMobile();
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

function SupportLink({ item }: { item: AdminNavItem }) {
  const Icon = item.icon;

  return (
    <TooltipProvider delayDuration={120}>
      <Tooltip>
        <TooltipTrigger asChild>
          <Button type="button" variant="outline" size="icon" className="shrink-0" asChild>
            <Link href={item.href} aria-label={item.title}>
              <Icon data-icon="inline-start" aria-hidden="true" />
            </Link>
          </Button>
        </TooltipTrigger>
        <TooltipContent side="bottom">{item.title}</TooltipContent>
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

function SidebarPlanSummary({
  summary,
  appVersion,
  onNavigate,
}: {
  summary: {
    planName: string | null;
    status: string | null;
    periodEnd: string | null;
  };
  appVersion: string;
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
        <SidebarMenuButton size="lg" variant="outline" tooltip={t("nav.plan")} className="flex-col gap-1.5" asChild>
          <Link href="/billing/plan" onClick={onNavigate}>
            <span className="grid w-full min-w-0 grid-cols-[minmax(0,1fr)_auto_minmax(0,1fr)] items-center gap-2 text-center">
              <span className="truncate text-xs font-medium">{planName}</span>
              <span className="text-sidebar-foreground/35" aria-hidden="true">|</span>
              <span className={cn("truncate text-xs font-normal", isExpired ? "text-warning" : "text-sidebar-foreground/60")}>{periodLabel}</span>
            </span>
            <span className="w-full text-center text-xs leading-none text-sidebar-foreground/45">
              {t("appVersion.label")} <span dir="ltr" className="font-mono">{appVersion}</span>
            </span>
          </Link>
        </SidebarMenuButton>
      </SidebarMenuItem>
    </SidebarMenu>
  );
}

function formatDaysRemaining(value: string, t: (key: string, values?: Record<string, string | number>) => string): string {
  const timestamp = Date.parse(value);
  if (Number.isNaN(timestamp)) return value;
  const days = Math.max(0, Math.ceil((timestamp - Date.now()) / (24 * 60 * 60 * 1000)));
  return t("plan.daysRemaining", { days: moneyFormatter.format(days) });
}
