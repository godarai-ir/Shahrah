import { Injectable } from '@nestjs/common';
import { PrismaService } from '../database/prisma.service';
import { IndexingService } from '../indexing/indexing.service';
import { ModelsQueryDto } from './dto/models-query.dto';

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
    const indexed = await this.indexing.searchModels(query);
    if (indexed) {
      return indexed;
    }

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
        query.min_input_price || query.max_input_price || query.max_output_price
          ? {
              some: {
                inputPricePerMillionTokens:
                  query.min_input_price !== undefined ||
                  query.max_input_price !== undefined
                    ? {
                        gte: query.min_input_price,
                        lte: query.max_input_price,
                      }
                    : undefined,
                outputPricePerMillionTokens: query.max_output_price
                  ? { lte: query.max_output_price }
                  : undefined,
              },
            }
          : undefined,
    };

    const [total, rows] = await Promise.all([
      this.prisma.model.count({ where }),
      this.prisma.model.findMany({
        where,
        include: {
          provider: true,
          modelSeries: true,
          category: true,
          pricing: { take: 1, orderBy: { effectiveFrom: 'desc' } },
          inputModalities: { include: { modality: true } },
          outputModalities: { include: { modality: true } },
          supportedParams: { include: { supportedParameter: true } },
          capabilities: { include: { capability: true } },
          regions: { include: { region: true } },
          benchmarks: true,
        },
        orderBy: this.orderBy(query.sort),
        skip: (query.page - 1) * query.limit,
        take: query.limit,
      }),
    ]);

    const data = rows.map((row) => ({
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
      pricing: row.pricing[0]
        ? {
            currency: row.pricing[0].currency,
            inputPerMillionTokens: row.pricing[0].inputPricePerMillionTokens,
            outputPerMillionTokens: row.pricing[0].outputPricePerMillionTokens,
            cachedInputPerMillionTokens:
              row.pricing[0].cachedInputPricePerMillionTokens,
          }
        : null,
      regions: row.regions.map((r) => r.region.slug),
      openSource: row.openSource,
      free: row.free,
      createdAt: row.createdAt,
      updatedAt: row.updatedAt,
    }));

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
    return this.prisma.model.findFirst({
      where: { canonicalId, enabled: true, provider: { enabled: true } },
      include: {
        provider: true,
        modelSeries: true,
        category: true,
        pricing: {
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
    return this.prisma.model.findMany({
      where: {
        canonicalId: { in: ids },
        enabled: true,
        provider: { enabled: true },
      },
      include: {
        provider: true,
        pricing: { take: 1, orderBy: { effectiveFrom: 'desc' } },
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
    return this.prisma.model.findMany({
      where: { free: true, enabled: true, provider: { enabled: true } },
      include: { provider: true, pricing: true },
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
