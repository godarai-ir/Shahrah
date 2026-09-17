import type { ProviderCatalogProvider, TenantProviderAccount } from "@/features/provider-accounts/types";
import { formatProviderDisplayName } from "@/features/provider-accounts/utils";

export type ProviderAccountGroup = {
  provider: string;
  displayName: string;
  metadata?: ProviderCatalogProvider;
  accounts: TenantProviderAccount[];
};

export function groupAccountsByProvider(
  accounts: TenantProviderAccount[],
  providerBySlug: Map<string, ProviderCatalogProvider>,
): ProviderAccountGroup[] {
  const groups = new Map<string, ProviderAccountGroup>();
  for (const account of accounts) {
    let group = groups.get(account.provider);
    if (!group) {
      const metadata = providerBySlug.get(account.provider);
      group = {
        provider: account.provider,
        displayName: formatProviderDisplayName(account.provider, metadata?.display_name),
        metadata,
        accounts: [],
      };
      groups.set(account.provider, group);
    }
    group.accounts.push(account);
  }
  return [...groups.values()]
    .sort((a, b) => a.displayName.localeCompare(b.displayName))
    .map((group) => ({ ...group, accounts: group.accounts.sort((a, b) => a.name.localeCompare(b.name)) }));
}
