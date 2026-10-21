import { NextIntlClientProvider } from "next-intl";
import { getLocale, getMessages } from "next-intl/server";
import { headers } from "next/headers";

import { loadAdminShellData, getAppVersion } from "@/features/admin-shell/data";
import AdminShellClient from "@/features/admin-shell/AdminShellClient";
import { buildDocsHref } from "@/features/admin-shell/AdminShell";
import { requireAuth } from "@/lib/auth/session";
import { pickMessages } from "@/lib/pick-messages";

const SHELL_MESSAGE_NAMESPACES = ["common", "theme", "language", "shell"] as const;

export default async function AdminLayout({ children }: { children: React.ReactNode }) {
  const auth = await requireAuth("/login");
  const [locale, allMessages, requestHeaders, shellData] = await Promise.all([
    getLocale(),
    getMessages(),
    headers(),
    loadAdminShellData(auth.tenantSlug, auth.user.id),
  ]);
  const messages = pickMessages(allMessages, SHELL_MESSAGE_NAMESPACES);
  const currentUser = {
    id: auth.user.id,
    email: auth.user.email,
    mobile: auth.user.mobile,
    role: auth.user.role,
  };

  return (
    <NextIntlClientProvider locale={locale} messages={messages}>
      <AdminShellClient
        tenantSlug={auth.tenantSlug}
        currentUser={currentUser}
        permissions={shellData.permissions}
        subscriptionSummary={shellData.subscriptionSummary}
        docsHref={buildDocsHref(requestHeaders, auth.tenantSlug)}
        appVersion={getAppVersion()}
      >
        {children}
      </AdminShellClient>
    </NextIntlClientProvider>
  );
}
