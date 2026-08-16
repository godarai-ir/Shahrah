import { Injectable, ServiceUnavailableException } from '@nestjs/common';
import { PrismaService } from '../database/prisma.service';
import { IndexingService } from '../indexing/indexing.service';
import { ModelsQueryDto } from './dto/models-query.dto';
import { createHash } from 'node:crypto';
import { modelRuntimeProfileSchema } from './runtime-profile.schema';
import {
  activeModelPricingWhere,
  activePricingRateWhere,
  isPricingWindowActive,
  nextPricingTransition,
  sortModelsByActivePrice,
} from './pricing-window';

type OpenAIRuntimePricingModel = {
  canonicalId: string;
  runtimeProfile: {
    profile: {
      requestControls: { serviceTierValues: string[] };
    };
  };
  pricingRates: Array<{
    dimension: string;
    currency: string;
    region: string;
  }>;
};

function missingOpenAITierPricing(model: OpenAIRuntimePricingModel): string[] {
  const dimensions = new Set(
    model.pricingRates
      .filter((rate) => rate.currency === 'USD' && rate.region === 'global')
      .map((rate) => rate.dimension),
  );
  const tiers = new Set(
    model.runtimeProfile.profile.requestControls.serviceTierValues,
  );
  const required = new Set<string>();
  const require = (...values: string[]) =>
    values.forEach((value) => required.add(value));

  if (tiers.has('standard')) require('input_tokens', 'output_tokens');
  if (tiers.has('batch')) {
    require('batch_input_tokens', 'batch_output_tokens');
  }
  if (tiers.has('flex')) require('flex_input_tokens', 'flex_output_tokens');
  if (tiers.has('priority') || tiers.has('fast')) {
    require('speed_input_tokens', 'speed_output_tokens');
  }

  return [...required].filter((dimension) => !dimensions.has(dimension));
}

@Injectable()
export class CatalogService {
  constructor(
    private readonly prisma: PrismaService,
    private readonly indexing: IndexingService,
  ) {}

  private orderBy(sort: string): any[] {
    switch (sort) {
      case 'name':
        return [{ displayName: 'asc' }];
      case 'largest_context':
        return [{ contextWindow: 'desc' }];
      case 'most_popular':
        return [{ popularityScore: 'desc' }];
      case 'newest':
      default:
        return [{ releaseDate: 'desc' }, { createdAt: 'desc' }];
    }
  }

  async runtimeSnapshot(providerSlug: string) {
    const generatedAt = new Date();
    const rows = await this.prisma.model.findMany({
      where: {
        provider: { slug: providerSlug, enabled: true },
        enabled: true,
        runtimeProfiles: { some: {} },
      },
      include: {
        pricingRates: {
          orderBy: [
            { dimension: 'asc' },
            { region: 'asc' },
            { effectiveFrom: 'asc' },
          ],
        },
        runtimeProfiles: {
          where: {
            effectiveFrom: { lte: generatedAt },
            OR: [{ effectiveTo: null }, { effectiveTo: { gt: generatedAt } }],
          },
          orderBy: { version: 'desc' },
          take: 1,
        },
        sources: { orderBy: { url: 'asc' } },
      },
      orderBy: { canonicalId: 'asc' },
    });
    if (rows.length === 0) {
      throw new ServiceUnavailableException({
        code: 'runtime_snapshot_unavailable',
        provider: providerSlug,
        reason: 'no_eligible_models',
      });
    }

    const nextPricingTransitionAt = nextPricingTransition(
      rows.flatMap((row) => row.pricingRates),
      generatedAt,
    );
    const models = rows.map((row) => {
      const storedRuntimeProfile = row.runtimeProfiles[0];
      const runtimeProfile = modelRuntimeProfileSchema.parse(
        storedRuntimeProfile.profileJson,
      );
      if (storedRuntimeProfile.version !== runtimeProfile.version) {
        throw new Error(
          `Runtime profile version mismatch for ${row.canonicalId}`,
        );
      }
      const effectiveStatus =
        row.retirementDate && row.retirementDate <= generatedAt
          ? 'unavailable'
          : row.deprecationDate && row.deprecationDate <= generatedAt
            ? 'deprecated'
            : row.status;
      return {
        canonicalId: row.canonicalId,
        slug: row.slug,
        aliases: runtimeProfile.aliases,
        status: effectiveStatus,
        deprecationDate: row.deprecationDate,
        retirementDate: row.retirementDate,
        replacementCanonicalId: row.replacementCanonicalId,
        runtimeProfile: {
          version: storedRuntimeProfile.version,
          profile: runtimeProfile,
          sourceUrl: storedRuntimeProfile.sourceUrl,
          effectiveFrom: storedRuntimeProfile.effectiveFrom,
        },
        pricingRates: row.pricingRates
          .filter((rate) => isPricingWindowActive(rate, generatedAt))
          .map((rate) => ({
            dimension: rate.dimension,
            unit: rate.unit,
            unitPriceUsd: rate.unitPriceUsd.toString(),
            currency: rate.currency,
            condition: rate.conditionJson,
            region: rate.region,
            effectiveFrom: rate.effectiveFrom,
            effectiveTo: rate.effectiveTo,
            sourceUrl: rate.sourceUrl,
          })),
        sources: row.sources.map((source) => source.url),
      };
    });
    if (providerSlug === 'openai') {
      for (const model of models) {
        const missingDimensions = missingOpenAITierPricing(model);
        if (missingDimensions.length > 0) {
          throw new ServiceUnavailableException({
            code: 'runtime_snapshot_unavailable',
            provider: providerSlug,
            model: model.canonicalId,
            reason: 'missing_tier_pricing',
            missingDimensions,
          });
        }
      }
    }
    if (providerSlug === 'anthropic' || providerSlug === 'xai') {
      for (const model of models) {
        const coreDimensions = new Set(
          model.pricingRates
            .filter(
              (rate) => rate.currency === 'USD' && rate.region === 'global',
            )
            .map((rate) => rate.dimension),
        );
        const requiredDimensions =
          providerSlug === 'xai' &&
          model.runtimeProfile.profile.modalities.output.includes('image')
            ? ['input_images', 'output_images']
            : ['input_tokens', 'output_tokens'];
        const missingDimensions = requiredDimensions.filter(
          (dimension) => !coreDimensions.has(dimension),
        );
        if (missingDimensions.length > 0) {
          throw new ServiceUnavailableException({
            code: 'runtime_snapshot_unavailable',
            provider: providerSlug,
            model: model.canonicalId,
            reason: 'missing_core_pricing',
            missingDimensions,
          });
        }
      }
    }
    const revision = createHash('sha256')
      .update(JSON.stringify({ nextPricingTransitionAt, models }))
      .digest('hex');

    return {
      provider: providerSlug,
      revision,
      generatedAt,
      nextPricingTransitionAt,
      models,
    };
  }

  async listModels(query: ModelsQueryDto) {
    // Pricing is temporal, while the Typesense document contains a scalar
    // projection. Keep the public priced list on the authoritative database
    // path so a scheduled boundary never depends on a reindex completing.
    const asOf = new Date();
    const activePricing = activeModelPricingWhere(asOf);

    const where: any = {
      enabled: true,
      provider: {
        ...(query.provider ? { slug: query.provider } : {}),
        enabled: true,
      },
      category: query.category ? { slug: query.category } : undefined,
      status: query.status as any,
      free: query.free_only || undefined,
      openSource: query.open_source_only || undefined,
      contextWindow: {
        gte: query.min_context_window,
        lte: query.max_context_window,
      },
      maxOutputTokens: {
        gte: query.min_output_tokens,
        lte: query.max_output_tokens,
      },
      mode: query.mode?.length ? { hasSome: query.mode } : undefined,
      OR: query.q
        ? [
            { displayName: { contains: query.q, mode: 'insensitive' } },
            { canonicalId: { contains: query.q, mode: 'insensitive' } },
            { shortDescription: { contains: query.q, mode: 'insensitive' } },
          ]
        : undefined,
      inputModalities: query.input_modalities?.length
        ? { some: { modality: { slug: { in: query.input_modalities } } } }
        : undefined,
      outputModalities: query.output_modalities?.length
        ? { some: { modality: { slug: { in: query.output_modalities } } } }
        : undefined,
      supportedParams: query.supported_parameters?.length
        ? {
            some: {
              supportedParameter: { slug: { in: query.supported_parameters } },
            },
          }
        : undefined,
      capabilities: query.capabilities?.length
        ? { some: { capability: { slug: { in: query.capabilities } } } }
        : undefined,
      regions: query.region
        ? { some: { region: { slug: query.region } } }
        : undefined,
      pricing:
        query.min_input_price !== undefined ||
        query.max_input_price !== undefined ||
        query.max_output_price !== undefined
          ? {
              some: {
                ...activePricing,
                inputPricePerMillionTokens:
                  query.min_input_price !== undefined ||
                  query.max_input_price !== undefined
                    ? {
                        gte: query.min_input_price,
                        lte: query.max_input_price,
                      }
                    : undefined,
                outputPricePerMillionTokens:
                  query.max_output_price !== undefined
                    ? { lte: query.max_output_price }
                    : undefined,
              },
            }
          : undefined,
    };

    const priceSort =
      query.sort === 'cheapest_input' || query.sort === 'cheapest_output';
    const [total, queriedRows] = await Promise.all([
      this.prisma.model.count({ where }),
      this.prisma.model.findMany({
        where,
        include: {
          provider: true,
          modelSeries: true,
          category: true,
          pricing: {
            where: activePricing,
            take: 1,
            orderBy: { effectiveFrom: 'desc' },
          },
          pricingRates: {
            where: activePricingRateWhere(asOf),
            orderBy: [
              { dimension: 'asc' },
              { region: 'asc' },
              { effectiveFrom: 'desc' },
            ],
          },
          runtimeProfiles: {
            where: {
              effectiveFrom: { lte: asOf },
              OR: [{ effectiveTo: null }, { effectiveTo: { gt: asOf } }],
            },
            orderBy: { version: 'desc' },
            take: 1,
          },
          inputModalities: { include: { modality: true } },
          outputModalities: { include: { modality: true } },
          supportedParams: { include: { supportedParameter: true } },
          capabilities: { include: { capability: true } },
          regions: { include: { region: true } },
          benchmarks: true,
        },
        orderBy: this.orderBy(query.sort),
        skip: priceSort ? undefined : (query.page - 1) * query.limit,
        take: priceSort ? undefined : query.limit,
      }),
    ]);
    const rows = priceSort
      ? sortModelsByActivePrice(
          queriedRows,
          query.sort === 'cheapest_input'
            ? 'inputPricePerMillionTokens'
            : 'outputPricePerMillionTokens',
        ).slice((query.page - 1) * query.limit, query.page * query.limit)
      : queriedRows;

    const data = rows.map((row) => {
      const runtimeProfile = row.runtimeProfiles?.[0]
        ? modelRuntimeProfileSchema.parse(row.runtimeProfiles[0].profileJson)
        : null;
      return {
        id: row.id,
        canonicalId: row.canonicalId,
        slug: row.slug,
        aliases: runtimeProfile?.aliases ?? [],
        name: row.displayName,
        provider: {
          slug: row.provider.slug,
          name: row.provider.name,
          logoUrl: row.provider.logoUrl,
        },
        modelSeries: row.modelSeries
          ? {
              slug: row.modelSeries.slug,
              name: row.modelSeries.name,
              logoUrl: row.modelSeries.logoUrl,
            }
          : null,
        description: row.shortDescription,
        status: row.status,
        mode: row.mode,
        contextWindow: row.contextWindow,
        maxOutputTokens: row.maxOutputTokens,
        inputModalities: row.inputModalities.map((m) => m.modality.slug),
        outputModalities: row.outputModalities.map((m) => m.modality.slug),
        supportedParameters: row.supportedParams.map(
          (m) => m.supportedParameter.slug,
        ),
        capabilities: row.capabilities.map((m) => m.capability.slug),
        pricing: row.pricing[0]
          ? {
              currency: row.pricing[0].currency,
              inputPerMillionTokens: row.pricing[0].inputPricePerMillionTokens,
              outputPerMillionTokens:
                row.pricing[0].outputPricePerMillionTokens,
              cachedInputPerMillionTokens:
                row.pricing[0].cachedInputPricePerMillionTokens,
            }
          : null,
        pricingRates: (row.pricingRates ?? []).map((rate) => ({
          dimension: rate.dimension,
          unit: rate.unit,
          unitPriceUsd: rate.unitPriceUsd.toString(),
          currency: rate.currency,
          condition: rate.conditionJson,
          region: rate.region,
          effectiveFrom: rate.effectiveFrom,
          effectiveTo: rate.effectiveTo,
          sourceUrl: rate.sourceUrl,
        })),
        regions: row.regions.map((r) => r.region.slug),
        openSource: row.openSource,
        free: row.free,
        createdAt: row.createdAt,
        updatedAt: row.updatedAt,
        runtimeProfile: runtimeProfile
          ? {
              version: row.runtimeProfiles[0].version,
              profile: runtimeProfile,
              sourceUrl: row.runtimeProfiles[0].sourceUrl,
              effectiveFrom: row.runtimeProfiles[0].effectiveFrom,
            }
          : null,
      };
    });

    return {
      data,
      pagination: {
        page: query.page,
        limit: query.limit,
        total,
        totalPages: Math.max(1, Math.ceil(total / query.limit)),
      },
      meta: {
        source: 'database',
        fallbackUsed: this.indexing.isReadPathEnabled(),
      },
    };
  }

  async getModelByCanonicalId(canonicalId: string) {
    const asOf = new Date();

    return this.prisma.model.findFirst({
      where: { canonicalId, enabled: true, provider: { enabled: true } },
      include: {
        provider: true,
        modelSeries: true,
        category: true,
        pricing: {
          where: activeModelPricingWhere(asOf),
          include: { region: true },
          orderBy: { effectiveFrom: 'desc' },
        },
        inputModalities: { include: { modality: true } },
        outputModalities: { include: { modality: true } },
        supportedParams: { include: { supportedParameter: true } },
        capabilities: { include: { capability: true } },
        regions: { include: { region: true } },
        sources: true,
        changelog: { orderBy: { changedAt: 'desc' }, take: 30 },
        benchmarks: { include: { benchmark: true } },
      },
    });
  }

  compare(ids: string[]) {
    const asOf = new Date();

    return this.prisma.model.findMany({
      where: {
        canonicalId: { in: ids },
        enabled: true,
        provider: { enabled: true },
      },
      include: {
        provider: true,
        pricing: {
          where: activeModelPricingWhere(asOf),
          take: 1,
          orderBy: { effectiveFrom: 'desc' },
        },
        capabilities: { include: { capability: true } },
        inputModalities: { include: { modality: true } },
        outputModalities: { include: { modality: true } },
        supportedParams: { include: { supportedParameter: true } },
        regions: { include: { region: true } },
        benchmarks: { include: { benchmark: true } },
      },
    });
  }

  recent() {
    return this.prisma.model.findMany({
      where: { enabled: true, provider: { enabled: true } },
      orderBy: { releaseDate: 'desc' },
      take: 20,
      include: { provider: true },
    });
  }

  freeModels() {
    const asOf = new Date();

    return this.prisma.model.findMany({
      where: { free: true, enabled: true, provider: { enabled: true } },
      include: {
        provider: true,
        pricing: { where: activeModelPricingWhere(asOf) },
      },
    });
  }

  deprecatedModels() {
    return this.prisma.model.findMany({
      where: {
        status: 'deprecated',
        enabled: true,
        provider: { enabled: true },
      },
      include: { provider: true },
    });
  }
}
