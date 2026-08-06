type VersionedRuntimeProfile = {
  version: number;
};

type StoredRuntimeProfile = {
  version: number;
  profileJson: unknown;
};

export function stableJson(value: unknown): string {
  if (Array.isArray(value)) {
    return `[${value.map((item) => stableJson(item)).join(',')}]`;
  }
  if (value !== null && typeof value === 'object') {
    const record = value as Record<string, unknown>;
    return `{${Object.keys(record)
      .sort()
      .map((key) => `${JSON.stringify(key)}:${stableJson(record[key])}`)
      .join(',')}}`;
  }

  return JSON.stringify(value) ?? 'undefined';
}

export function assertRuntimeProfileVersionIsImmutable<
  RuntimeProfile extends VersionedRuntimeProfile,
>(
  canonicalId: string,
  existingProfile: StoredRuntimeProfile | null | undefined,
  incomingProfile: RuntimeProfile,
): void {
  if (
    !existingProfile ||
    existingProfile.version !== incomingProfile.version ||
    stableJson(existingProfile.profileJson) === stableJson(incomingProfile)
  ) {
    return;
  }

  throw new Error(
    `${canonicalId}: runtime profile version ${incomingProfile.version} is immutable; increment runtimeProfile.version`,
  );
}
