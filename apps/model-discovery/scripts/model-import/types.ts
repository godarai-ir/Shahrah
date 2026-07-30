import type { Prisma } from '../../src/generated/prisma/client';

export type ProviderPricingRate = {
  dimension: string;
  unit: string;
  unitPriceUsd: string;
  conditionJson?: Prisma.InputJsonObject;
  region: string;
  sourceUrl: string;
  effectiveFrom?: Date;
  effectiveTo?: Date;
};

export type ProviderRateMapping = {
  key: string;
  dimension: string;
  unit: string;
  conditionJson?: Prisma.InputJsonObject;
};
