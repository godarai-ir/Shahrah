import { NextIntlClientProvider } from "next-intl";
import { getLocale, getMessages } from "next-intl/server";

import { AdminShellPageBridge, type ContentLayout } from "@/features/admin-shell/page-context";
import type { ShellPageMetadata } from "@/components/app-ui/ShellPageHeading";
import { buildDocsBaseHref } from "@/lib/docs-href";
import { pickMessages } from "@/lib/pick-messages";

const SHELL_MESSAGE_NAMESPACES = ["common", "theme", "language", "shell"] as const;

type AdminShellProps = {
  children: React.ReactNode;
  tenantSlug: string;
  pageHeader?: ShellPageMetadata;
  contentPadding?: "default" | "none";
  contentLayout?: ContentLayout;
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
  contentLayout,
  messageNamespaces = [],
}: AdminShellProps) {
  const locale = await getLocale();
  const messages = pickMessages(
    await getMessages(),
    [...SHELL_MESSAGE_NAMESPACES, ...messageNamespaces],
  );
  return (
    <NextIntlClientProvider locale={locale} messages={messages}>
      <AdminShellPageBridge pageHeader={pageHeader} contentPadding={contentPadding} contentLayout={contentLayout}>
        {children}
      </AdminShellPageBridge>
    </NextIntlClientProvider>
  );
}

/** The documentation base URL with a trailing slash, for appending relative doc paths. */
export function buildDocsHref(headersList: Headers, tenantSlug: string): string {
  return `${buildDocsBaseHref(headersList, tenantSlug)}/`;
}
