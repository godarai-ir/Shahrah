"use client";

import Image from "next/image";
import Link from "next/link";
import { usePathname } from "next/navigation";
import { useTranslations } from "next-intl";
import { LayoutDashboard, LogOut, Menu, Search, Settings } from "lucide-react";

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

  const userIdentity =
    currentUser.email ?? currentUser.mobile ?? currentUser.id;

  const navItems = [
    {
      href: "/",
      title: t("nav.dashboard"),
      icon: LayoutDashboard,
      isActive: pathname === "/",
    },
    {
      href: "/settings",
      title: t("nav.settings"),
      icon: Settings,
      isActive: pathname.startsWith("/settings"),
    },
  ];

  const currentPageLabel = navItems.find((item) => item.isActive)?.title ?? t("nav.dashboard");

  return (
    <div className="min-h-screen bg-background">
      <div className="grid min-h-screen lg:grid-cols-[280px_1fr]">
        <aside className="hidden border-l bg-card/70 lg:block">
          <SidebarContent
            tenantSlug={tenantSlug}
            currentUser={currentUser}
            userIdentity={userIdentity}
            navItems={navItems}
          />
        </aside>

        <div className="min-w-0">
          <header className="sticky top-0 z-10 border-b bg-background/85 px-4 py-3 backdrop-blur md:px-6">
            <div className="flex items-center gap-2 md:gap-3">
              <Sheet>
                <SheetTrigger asChild>
                  <Button variant="outline" size="icon" className="lg:hidden">
                    <Menu className="h-4 w-4" />
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
                    navItems={navItems}
                    className="border-none"
                  />
                </SheetContent>
              </Sheet>

              <div className="min-w-0 flex-1">
                <p className="text-xs text-muted-foreground">{t("tenant")}</p>
                <h1 className="truncate text-lg font-semibold">{currentPageLabel}</h1>
              </div>

              <div className="hidden w-full max-w-sm items-center gap-2 md:flex">
                <Search className="h-4 w-4 text-muted-foreground" />
                <Input placeholder={t("searchPlaceholder")} aria-label={t("searchPlaceholder")} />
              </div>

              <ThemeToggle />
            </div>
          </header>

          <main className="p-4 md:p-6">
            <div className="mx-auto w-full max-w-7xl">{children}</div>
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
  navItems: Array<{
    href: string;
    title: string;
    icon: React.ComponentType<{ className?: string }>;
    isActive: boolean;
  }>;
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
    <div className={cn("flex h-full flex-col border-l p-4", className)}>
      <div className="flex items-center gap-3 rounded-lg border bg-background p-3">
        <div className="flex h-10 w-10 items-center justify-center overflow-hidden rounded-lg border bg-muted/40 p-1">
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
          <p className="truncate text-sm font-semibold">{t("title")}</p>
          <p className="truncate text-xs text-muted-foreground">{tenantSlug}</p>
        </div>
      </div>

      <nav className="mt-6 grid gap-1">
        {navItems.map((item) => {
          const Icon = item.icon;

          return (
            <Button
              key={item.href}
              variant={item.isActive ? "secondary" : "ghost"}
              className="h-10 justify-start gap-2"
              asChild
            >
              <Link href={item.href}>
                <Icon className="h-4 w-4" />
                <span>{item.title}</span>
              </Link>
            </Button>
          );
        })}
      </nav>

      <Separator className="my-4" />

      <div className="mt-auto rounded-xl border bg-background p-3">
        <div className="flex items-center gap-2">
          <Avatar>
            <AvatarFallback>{getInitials(userIdentity)}</AvatarFallback>
          </Avatar>
          <div className="min-w-0">
            <p className="truncate text-sm font-medium">{userIdentity}</p>
            <Badge variant="outline" className="mt-1">
              {currentUser.role}
            </Badge>
          </div>
        </div>

        <form action="/logout" method="post" className="mt-3">
          <Button variant="outline" size="sm" className="w-full justify-start gap-2">
            <LogOut className="h-4 w-4" />
            {t("logout")}
          </Button>
        </form>
      </div>
    </div>
  );
}
