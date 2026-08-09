import type { ProviderCredentialValues } from "./credential-schema";
import type { ProviderAuthMode } from "./types";

export function credentialFingerprint(input: {
  authMode: ProviderAuthMode;
  usePlatformManagedCredentials: boolean;
  values: ProviderCredentialValues;
}) {
  return JSON.stringify({
    authMode: input.authMode,
    usePlatformManagedCredentials: input.usePlatformManagedCredentials,
    values: Object.fromEntries(
      Object.entries(input.values).sort(([left], [right]) => left.localeCompare(right)),
    ),
  });
}
