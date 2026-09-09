import { NextIntlClientProvider } from "next-intl";
import { getLocale, getMessages } from "next-intl/server";

import { AdminShellPageBridge } from "@/features/admin-shell/page-context";
import type { ShellPageMetadata } from "@/components/app-ui/ShellPageHeading";
import { pickMessages } from "@/lib/pick-messages";

const SHELL_MESSAGE_NAMESPACES = ["common", "theme", "language", "shell"] as const;

type AdminShellProps = {
  children: React.ReactNode;
  tenantSlug: string;
  pageHeader?: ShellPageMetadata;
  contentPadding?: "default" | "none";
  messageNamespaces?: readonly string[];
  currentUser: {
    id: string;
    email?: string;
    mobile?: string;
    role: string;
  };
};

export default async function AdminShell({
  children,
  pageHeader,
  contentPadding,
  messageNamespaces = [],
}: AdminShellProps) {
  const locale = await getLocale();
  const messages = pickMessages(
    await getMessages(),
    [...SHELL_MESSAGE_NAMESPACES, ...messageNamespaces],
  );
  return (
    <NextIntlClientProvider locale={locale} messages={messages}>
      <AdminShellPageBridge pageHeader={pageHeader} contentPadding={contentPadding}>
        {children}
      </AdminShellPageBridge>
    </NextIntlClientProvider>
  );
}

export function buildDocsHref(headersList: Headers, tenantSlug: string): string {
  const forwardedProto = headersList.get("x-forwarded-proto")?.split(",")[0]?.trim();
  const proto = forwardedProto || (headersList.get("host")?.includes("localhost") ? "http" : "https");
  const host = headersList.get("x-forwarded-host")?.split(",")[0]?.trim() || headersList.get("host") || "";
  const tenantPrefix = `${tenantSlug}.`;
  const docsHost = host.startsWith(tenantPrefix) ? host.slice(tenantPrefix.length) : host;

  return docsHost ? `${proto}://${docsHost}/docs/` : "/docs/";
}
