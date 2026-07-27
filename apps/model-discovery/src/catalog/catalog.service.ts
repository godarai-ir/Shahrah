import { BadRequestException, Injectable } from '@nestjs/common';
import { PrismaService } from '../database/prisma.service';
import { IndexingService } from '../indexing/indexing.service';
import { ModelsQueryDto } from './dto/models-query.dto';
import {
  serializePricingProfiles,
  summarizePricing,
} from '../pricing/pricing-response';
import { Decimal } from '@prisma/client-runtime-utils';

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

  async listModels(query: ModelsQueryDto) {
    const hasPriceAmountQuery =
      query.min_price !== undefined ||
      query.max_price !== undefined ||
      query.sort === 'price_asc' ||
      query.sort === 'price_desc';
    const hasPricingQuery =
      hasPriceAmountQuery ||
      Boolean(
        query.price_usage ||
          query.price_unit ||
          query.price_region ||
          query.price_resolution ||
          query.price_service_tier,
      );
    if (hasPriceAmountQuery && (!query.price_usage || !query.price_unit)) {
      throw new BadRequestException(
        'price_usage and price_unit are required for price filters and sorting',
      );
    }

    const indexed = hasPricingQuery
      ? null
      : await this.indexing.searchModels(query);
    if (indexed) {
      if (indexed.pagination.total > 0) {
        return indexed;
      }

      const databaseHasEnabledModels = await this.prisma.model.count({
        where: { enabled: true, provider: { enabled: true } },
      });
      if (databaseHasEnabledModels === 0) {
        return indexed;
      }
    }

    const where: any = {
      enabled: true,
      provider: {
        ...(query.provider ? { slug: query.provider } : {}),
        enabled: true,
      },
      category: query.category ? { slug: query.category } : undefined,
      status: query.status as any,
      pricingStatus: query.free_only ? 'free' : undefined,
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
        query.price_usage ||
        query.price_unit ||
        query.price_region ||
        query.price_resolution ||
        query.price_service_tier ||
        query.min_price !== undefined ||
        query.max_price !== undefined
          ? {
              some: {
                region: query.price_region
                  ? { slug: query.price_region }
                  : undefined,
                rates: {
                  some: {
                    usage: query.price_usage as any,
                    unit: query.price_unit,
                    resolution: query.price_resolution,
                    serviceTier: query.price_service_tier as any,
                    amount:
                      query.min_price !== undefined ||
                      query.max_price !== undefined
                        ? {
                            gte: query.min_price,
                            lte: query.max_price,
                          }
                        : undefined,
                  },
                },
              },
            }
          : undefined,
    };

    const priceSorted =
      query.sort === 'price_asc' || query.sort === 'price_desc';
    const [total, allRows] = await Promise.all([
      this.prisma.model.count({ where }),
      this.prisma.model.findMany({
        where,
        include: {
          provider: true,
          modelSeries: true,
          category: true,
          pricing: {
            include: {
              region: true,
              rates: { orderBy: { rateKey: 'asc' } },
            },
            orderBy: [{ region: { slug: 'asc' } }, { effectiveFrom: 'desc' }],
          },
          inputModalities: { include: { modality: true } },
          outputModalities: { include: { modality: true } },
          supportedParams: { include: { supportedParameter: true } },
          capabilities: { include: { capability: true } },
          regions: { include: { region: true } },
          benchmarks: true,
        },
        orderBy: priceSorted ? [{ displayName: 'asc' }] : this.orderBy(query.sort),
        skip: priceSorted ? undefined : (query.page - 1) * query.limit,
        take: priceSorted ? undefined : query.limit,
      }),
    ]);

    const mappedRows = allRows.map((row) => {
      const pricingProfiles = serializePricingProfiles(row.pricing);
      return {
      id: row.id,
      canonicalId: row.canonicalId,
      slug: row.slug,
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
      pricingStatus: row.pricingStatus,
      pricingProfiles,
      pricingSummary: summarizePricing(pricingProfiles, row.mode),
      regions: row.regions.map((r) => r.region.slug),
      openSource: row.openSource,
      createdAt: row.createdAt,
      updatedAt: row.updatedAt,
      };
    });

    const sortedRows = priceSorted
      ? mappedRows.sort((left, right) => {
          const leftValue = this.selectedSummaryAmount(
            left.pricingProfiles,
            query,
          );
          const rightValue = this.selectedSummaryAmount(
            right.pricingProfiles,
            query,
          );
          const direction = query.sort === 'price_desc' ? -1 : 1;
          if (!leftValue && !rightValue) return 0;
          if (!leftValue) return 1;
          if (!rightValue) return -1;
          return leftValue.comparedTo(rightValue) * direction;
        })
      : mappedRows;
    const data = priceSorted
      ? sortedRows.slice(
          (query.page - 1) * query.limit,
          query.page * query.limit,
        )
      : sortedRows;

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
    const row = await this.prisma.model.findFirst({
      where: { canonicalId, enabled: true, provider: { enabled: true } },
      include: {
        provider: true,
        modelSeries: true,
        category: true,
        pricing: {
          include: {
            region: true,
            rates: { orderBy: { rateKey: 'asc' } },
          },
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
    if (!row) return null;
    const { pricing, ...model } = row;
    const pricingProfiles = serializePricingProfiles(pricing);
    return {
      ...model,
      pricingProfiles,
      pricingSummary: summarizePricing(pricingProfiles, row.mode),
    };
  }

  async compare(ids: string[]) {
    const rows = await this.prisma.model.findMany({
      where: {
        canonicalId: { in: ids },
        enabled: true,
        provider: { enabled: true },
      },
      include: {
        provider: true,
        pricing: {
          include: {
            region: true,
            rates: { orderBy: { rateKey: 'asc' } },
          },
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
    return rows.map((row) => {
      const { pricing, ...model } = row;
      const pricingProfiles = serializePricingProfiles(pricing);
      return {
        ...model,
        pricingProfiles,
        pricingSummary: summarizePricing(pricingProfiles, row.mode),
      };
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

  async freeModels() {
    const rows = await this.prisma.model.findMany({
      where: {
        pricingStatus: 'free',
        enabled: true,
        provider: { enabled: true },
      },
      include: {
        provider: true,
        pricing: { include: { region: true, rates: true } },
      },
    });
    return rows.map((row) => {
      const { pricing, ...model } = row;
      const pricingProfiles = serializePricingProfiles(pricing);
      return {
        ...model,
        pricingProfiles,
        pricingSummary: summarizePricing(pricingProfiles, row.mode),
      };
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

  private selectedSummaryAmount(
    profiles: ReturnType<typeof serializePricingProfiles>,
    query: ModelsQueryDto,
  ): Decimal | null {
    const amounts = profiles.flatMap((profile) =>
      profile.rates
        .filter(
          (rate) =>
            rate.usage === query.price_usage &&
            rate.unit === query.price_unit &&
            (!query.price_region || profile.region === query.price_region) &&
            (!query.price_resolution ||
              rate.resolution === query.price_resolution) &&
            (!query.price_service_tier ||
              rate.serviceTier === query.price_service_tier),
        )
        .map((rate) => new Decimal(rate.amount)),
    );
    return amounts.length > 0
      ? amounts.reduce((left, right) => Decimal.min(left, right))
      : null;
  }
}
