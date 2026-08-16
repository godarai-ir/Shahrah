export type PricingWindow = {
  effectiveFrom?: Date | null;
  effectiveTo?: Date | null;
};

export function isPricingWindowActive(
  window: PricingWindow,
  asOf: Date,
): boolean {
  return (
    (!window.effectiveFrom || window.effectiveFrom <= asOf) &&
    (!window.effectiveTo || window.effectiveTo > asOf)
  );
}

export function activeModelPricingWhere(asOf: Date) {
  return {
    AND: [
      {
        OR: [{ effectiveFrom: null }, { effectiveFrom: { lte: asOf } }],
      },
      {
        OR: [{ effectiveTo: null }, { effectiveTo: { gt: asOf } }],
      },
    ],
  };
}

export function activePricingRateWhere(asOf: Date) {
  return {
    effectiveFrom: { lte: asOf },
    OR: [{ effectiveTo: null }, { effectiveTo: { gt: asOf } }],
  };
}

export function nextPricingTransition(
  windows: PricingWindow[],
  asOf: Date,
): Date | null {
  let next: Date | null = null;

  for (const window of windows) {
    for (const boundary of [window.effectiveFrom, window.effectiveTo]) {
      if (boundary && boundary > asOf && (!next || boundary < next)) {
        next = boundary;
      }
    }
  }

  return next;
}

type PricingProjection = {
  displayName: string;
  pricing: Array<{
    inputPricePerMillionTokens?: number | null;
    outputPricePerMillionTokens?: number | null;
  }>;
};

export function sortModelsByActivePrice<T extends PricingProjection>(
  rows: T[],
  field: 'inputPricePerMillionTokens' | 'outputPricePerMillionTokens',
): T[] {
  return [...rows].sort((left, right) => {
    const leftPrice = left.pricing[0]?.[field] ?? Number.POSITIVE_INFINITY;
    const rightPrice = right.pricing[0]?.[field] ?? Number.POSITIVE_INFINITY;

    return (
      leftPrice - rightPrice ||
      left.displayName.localeCompare(right.displayName)
    );
  });
}
