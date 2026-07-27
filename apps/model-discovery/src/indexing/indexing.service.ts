import { InjectQueue } from '@nestjs/bullmq';
import { Injectable, Logger } from '@nestjs/common';
import { ConfigService } from '@nestjs/config';
import { Queue } from 'bullmq';
import Typesense from 'typesense';
import { PrismaService } from '../database/prisma.service';
import { ModelsQueryDto } from '../catalog/dto/models-query.dto';
import {
  serializePricingProfiles,
  summarizePricing,
} from '../pricing/pricing-response';

export const INDEXING_QUEUE = 'catalog-indexing';

type TypesenseDocument = {
  id: string;
  canonicalId: string;
  name: string;
  slug: string;
  enabled: boolean;
  providerSlug: string;
  providerName: string;
  providerLogoUrl?: string;
  modelSeriesSlug?: string;
  modelSeriesName?: string;
  modelSeriesLogoUrl?: string;
  description?: string;
  status: string;
  lifecycleStatus: string;
  mode: string[];
  category?: string;
  categoryName?: string;
  contextWindow?: number;
  maxOutputTokens?: number;
  pricingStatus: string;
  pricingUsages: string[];
  pricingUnits: string[];
  pricingResolutions: string[];
  pricingRegions: string[];
  pricingServiceTiers: string[];
  pricingJson: string;
  pricingSummaryJson: string;
  openSource: boolean;
  inputModalities: string[];
  outputModalities: string[];
  supportedParameters: string[];
  capabilities: string[];
  regions: string[];
  tags: string[];
  benchmarkScore?: number | null;
  popularityScore?: number | null;
  releaseDateTimestamp?: number | null;
  createdAtTimestamp: number;
  updatedAtTimestamp: number;
};

export type TypesenseListModelsResponse = {
  data: Array<{
    id: string;
    canonicalId: string;
    slug: string;
    name: string;
    provider: {
      slug: string;
      name: string;
      logoUrl?: string | null;
    };
    modelSeries: {
      slug: string;
      name: string;
      logoUrl?: string | null;
    } | null;
    description?: string;
    status: string;
    mode: string[];
    contextWindow?: number | null;
    maxOutputTokens?: number | null;
    inputModalities: string[];
    outputModalities: string[];
    supportedParameters: string[];
    capabilities: string[];
    pricingStatus: string;
    pricingProfiles: ReturnType<typeof serializePricingProfiles>;
    pricingSummary: ReturnType<typeof summarizePricing>;
    regions: string[];
    openSource: boolean;
    createdAt: Date;
    updatedAt: Date;
  }>;
  pagination: {
    page: number;
    limit: number;
    total: number;
    totalPages: number;
  };
  meta: {
    source: 'typesense';
    fallbackUsed: false;
  };
};

export type TypesenseFacetCounts = {
  source: 'typesense';
  providerCounts: Map<string, number>;
  categoryCounts: Map<string, number>;
  inputModalityCounts: Map<string, number>;
  outputModalityCounts: Map<string, number>;
  supportedParameterCounts: Map<string, number>;
  capabilityCounts: Map<string, number>;
  regionCounts: Map<string, number>;
  statusCounts: Map<string, number>;
};

@Injectable()
export class IndexingService {
  private readonly client: any;
  private readonly collectionName: string;
  private readonly readsEnabled: boolean;
  private readonly readyStrict: boolean;
  private readonly logger = new Logger(IndexingService.name);

  constructor(
    @InjectQueue(INDEXING_QUEUE) private readonly queue: Queue,
    private readonly prisma: PrismaService,
    cfg: ConfigService,
  ) {
    this.client = new Typesense.Client({
      nodes: [
        {
          host: cfg.get<string>('TYPESENSE_HOST', '127.0.0.1'),
          port: Number(cfg.get<number>('TYPESENSE_PORT', 8108)),
          protocol: cfg.get<'http' | 'https'>('TYPESENSE_PROTOCOL', 'http'),
        },
      ],
      apiKey: cfg.get<string>('TYPESENSE_API_KEY', 'xyz'),
      connectionTimeoutSeconds: 2,
    });
    const configuredCollectionName = cfg.get<string>(
      'TYPESENSE_COLLECTION_MODELS',
      'models',
    );
    this.collectionName = configuredCollectionName.endsWith('_v2')
      ? configuredCollectionName
      : `${configuredCollectionName}_v2`;
    this.readsEnabled =
      cfg.get<string>('TYPESENSE_READS_ENABLED', 'true') !== 'false';
    this.readyStrict =
      cfg.get<string>('TYPESENSE_READY_STRICT', 'true') !== 'false';
  }

  isReadPathEnabled() {
    return this.readsEnabled;
  }

  isReadyStrict() {
    return this.readyStrict;
  }

  enqueueFullReindex() {
    return this.queue.add('full-reindex', {});
  }

  enqueueModelIndex(modelId: string) {
    return this.queue.add('index-model', { modelId });
  }

  enqueueProviderModelsReindex(providerSlug: string) {
    return this.queue.add('index-provider-models', { providerSlug });
  }

  enqueueModelSeriesReindex(modelSeriesSlug: string) {
    return this.queue.add('index-model-series', { modelSeriesSlug });
  }

  async ping() {
    try {
      await this.client.collections().retrieve();
      return true;
    } catch (error) {
      this.logger.warn(
        `Typesense ping failed: ${error instanceof Error ? error.message : String(error)}`,
      );
      return false;
    }
  }

  async ensureCollection() {
    const schema = {
      name: this.collectionName,
      fields: [
        { name: 'id', type: 'string' },
        { name: 'canonicalId', type: 'string' },
        { name: 'name', type: 'string', sort: true },
        { name: 'slug', type: 'string' },
        { name: 'enabled', type: 'bool', facet: true },
        { name: 'providerSlug', type: 'string', facet: true },
        { name: 'providerName', type: 'string', facet: true },
        { name: 'providerLogoUrl', type: 'string', optional: true },
        {
          name: 'modelSeriesSlug',
          type: 'string',
          facet: true,
          optional: true,
        },
        {
          name: 'modelSeriesName',
          type: 'string',
          facet: true,
          optional: true,
        },
        { name: 'modelSeriesLogoUrl', type: 'string', optional: true },
        { name: 'description', type: 'string', optional: true },
        { name: 'status', type: 'string', facet: true },
        { name: 'lifecycleStatus', type: 'string', facet: true },
        { name: 'mode', type: 'string[]', facet: true },
        { name: 'category', type: 'string', facet: true, optional: true },
        { name: 'categoryName', type: 'string', optional: true },
        { name: 'contextWindow', type: 'int32', sort: true, optional: true },
        { name: 'maxOutputTokens', type: 'int32', sort: true, optional: true },
        { name: 'pricingStatus', type: 'string', facet: true },
        { name: 'pricingUsages', type: 'string[]', facet: true },
        { name: 'pricingUnits', type: 'string[]', facet: true },
        { name: 'pricingResolutions', type: 'string[]', facet: true },
        { name: 'pricingRegions', type: 'string[]', facet: true },
        { name: 'pricingServiceTiers', type: 'string[]', facet: true },
        { name: 'pricingJson', type: 'string', optional: true },
        { name: 'pricingSummaryJson', type: 'string', optional: true },
        { name: 'openSource', type: 'bool', facet: true },
        { name: 'inputModalities', type: 'string[]', facet: true },
        { name: 'outputModalities', type: 'string[]', facet: true },
        { name: 'supportedParameters', type: 'string[]', facet: true },
        { name: 'capabilities', type: 'string[]', facet: true },
        { name: 'regions', type: 'string[]', facet: true },
        { name: 'tags', type: 'string[]', facet: true },
        { name: 'benchmarkScore', type: 'float', sort: true, optional: true },
        { name: 'popularityScore', type: 'float', sort: true, optional: true },
        {
          name: 'releaseDateTimestamp',
          type: 'int64',
          sort: true,
          optional: true,
        },
        { name: 'createdAtTimestamp', type: 'int64', sort: true },
        { name: 'updatedAtTimestamp', type: 'int64', sort: true },
      ],
      default_sorting_field: 'updatedAtTimestamp',
    };

    try {
      await this.client.collections(this.collectionName).retrieve();
      try {
        await this.client
          .collections(this.collectionName)
          .update({ fields: schema.fields } as any);
      } catch (error) {
        this.logger.warn(
          `Typesense schema sync skipped: ${error instanceof Error ? error.message : String(error)}`,
        );
      }
    } catch {
      await this.client.collections().create(schema as any);
    }
  }

  async searchModels(
    query: ModelsQueryDto,
  ): Promise<TypesenseListModelsResponse | null> {
    if (!this.readsEnabled) {
      return null;
    }

    try {
      await this.ensureCollection();
      const searchParameters = {
        q: query.q?.trim() || '*',
        query_by: [
          'name',
          'canonicalId',
          'slug',
          'description',
          'providerName',
          'modelSeriesName',
          'categoryName',
          'tags',
          'capabilities',
          'supportedParameters',
        ].join(','),
        query_by_weights: '8,7,6,3,2,2,2,1,1,1',
        filter_by: this.buildFilterBy(query),
        sort_by: this.mapSort(query.sort),
        page: query.page,
        per_page: query.limit,
      };

      const result = await this.client
        .collections(this.collectionName)
        .documents()
        .search(searchParameters as any);

      const data = (result.hits || []).map((hit: any) =>
        this.toListItem(hit.document as TypesenseDocument),
      );

      return {
        data,
        pagination: {
          page: query.page,
          limit: query.limit,
          total: result.found || 0,
          totalPages: Math.max(1, Math.ceil((result.found || 0) / query.limit)),
        },
        meta: {
          source: 'typesense',
          fallbackUsed: false,
        },
      };
    } catch (error) {
      this.logger.warn(
        `Typesense search failed, falling back to database: ${error instanceof Error ? error.message : String(error)}`,
      );
      return null;
    }
  }

  async facetCounts(): Promise<TypesenseFacetCounts | null> {
    if (!this.readsEnabled) {
      return null;
    }

    try {
      await this.ensureCollection();
      const result = await this.client
        .collections(this.collectionName)
        .documents()
        .search({
          q: '*',
          query_by: 'name',
          filter_by: 'enabled:=true',
          page: 1,
          per_page: 0,
          max_facet_values: 250,
          facet_by: [
            'providerSlug',
            'category',
            'inputModalities',
            'outputModalities',
            'supportedParameters',
            'capabilities',
            'regions',
            'status',
          ].join(','),
        } as any);

      const counts = new Map<string, Map<string, number>>();
      for (const facet of result.facet_counts || []) {
        const values = new Map<string, number>();
        for (const item of facet.counts || []) {
          values.set(String(item.value), Number(item.count));
        }
        counts.set(String(facet.field_name), values);
      }

      return {
        source: 'typesense',
        providerCounts: counts.get('providerSlug') || new Map(),
        categoryCounts: counts.get('category') || new Map(),
        inputModalityCounts: counts.get('inputModalities') || new Map(),
        outputModalityCounts: counts.get('outputModalities') || new Map(),
        supportedParameterCounts:
          counts.get('supportedParameters') || new Map(),
        capabilityCounts: counts.get('capabilities') || new Map(),
        regionCounts: counts.get('regions') || new Map(),
        statusCounts: counts.get('status') || new Map(),
      };
    } catch (error) {
      this.logger.warn(
        `Typesense facet query failed, falling back to database: ${error instanceof Error ? error.message : String(error)}`,
      );
      return null;
    }
  }

  async indexOneModel(modelId: string) {
    const model = await this.prisma.model.findUnique({
      where: { id: modelId },
      include: {
        provider: true,
        modelSeries: true,
        category: true,
        pricing: {
          include: { region: true, rates: true },
          orderBy: [{ region: { slug: 'asc' } }, { effectiveFrom: 'desc' }],
        },
        inputModalities: { include: { modality: true } },
        outputModalities: { include: { modality: true } },
        supportedParams: { include: { supportedParameter: true } },
        capabilities: { include: { capability: true } },
        regions: { include: { region: true } },
        tags: { include: { tag: true } },
        benchmarks: true,
      },
    });
    if (!model) return null;
    const pricingProfiles = serializePricingProfiles(model.pricing);
    const pricingSummary = summarizePricing(pricingProfiles, model.mode);
    const pricingRates = pricingProfiles.flatMap((profile) => profile.rates);
    const doc: TypesenseDocument = {
      id: model.id,
      canonicalId: model.canonicalId,
      name: model.displayName,
      slug: model.slug,
      enabled: model.enabled && model.provider.enabled,
      providerSlug: model.provider.slug,
      providerName: model.provider.name,
      providerLogoUrl: model.provider.logoUrl || undefined,
      modelSeriesSlug: model.modelSeries?.slug || undefined,
      modelSeriesName: model.modelSeries?.name || undefined,
      modelSeriesLogoUrl: model.modelSeries?.logoUrl || undefined,
      description: model.shortDescription || '',
      status: model.status,
      lifecycleStatus: model.lifecycleStatus,
      mode: model.mode,
      category: model.category?.slug || undefined,
      categoryName: model.category?.name || undefined,
      contextWindow: model.contextWindow || undefined,
      maxOutputTokens: model.maxOutputTokens || undefined,
      pricingStatus: model.pricingStatus,
      pricingUsages: Array.from(
        new Set(pricingRates.map((rate) => rate.usage)),
      ),
      pricingUnits: Array.from(new Set(pricingRates.map((rate) => rate.unit))),
      pricingResolutions: Array.from(
        new Set(
          pricingRates
            .map((rate) => rate.resolution)
            .filter((value): value is string => Boolean(value)),
        ),
      ),
      pricingRegions: Array.from(
        new Set(pricingProfiles.map((profile) => profile.region)),
      ),
      pricingServiceTiers: Array.from(
        new Set(pricingRates.map((rate) => rate.serviceTier)),
      ),
      pricingJson: JSON.stringify(pricingProfiles),
      pricingSummaryJson: JSON.stringify(pricingSummary),
      openSource: model.openSource,
      inputModalities: model.inputModalities.map((x) => x.modality.slug),
      outputModalities: model.outputModalities.map((x) => x.modality.slug),
      supportedParameters: model.supportedParams.map(
        (x) => x.supportedParameter.slug,
      ),
      capabilities: model.capabilities.map((x) => x.capability.slug),
      regions: model.regions.map((x) => x.region.slug),
      tags: model.tags.map((x) => x.tag.slug),
      benchmarkScore: model.benchmarks[0]?.score ?? null,
      popularityScore: model.popularityScore ?? null,
      releaseDateTimestamp: model.releaseDate
        ? Math.floor(model.releaseDate.getTime() / 1000)
        : null,
      createdAtTimestamp: Math.floor(model.createdAt.getTime() / 1000),
      updatedAtTimestamp: Math.floor(model.updatedAt.getTime() / 1000),
    };
    await this.client
      .collections(this.collectionName)
      .documents()
      .upsert(doc as any);
    return doc;
  }

  async reindexAll() {
    await this.ensureCollection();
    const models = await this.prisma.model.findMany({ select: { id: true } });
    for (const model of models) {
      await this.indexOneModel(model.id);
    }
    return { indexed: models.length };
  }

  async reindexProviderModels(providerSlug: string) {
    const models = await this.prisma.model.findMany({
      where: { provider: { slug: providerSlug } },
      select: { id: true },
    });
    for (const model of models) {
      await this.indexOneModel(model.id);
    }
    return { indexed: models.length, providerSlug };
  }

  async reindexModelSeries(modelSeriesSlug: string) {
    const models = await this.prisma.model.findMany({
      where: { modelSeries: { slug: modelSeriesSlug } },
      select: { id: true },
    });
    for (const model of models) {
      await this.indexOneModel(model.id);
    }
    return { indexed: models.length, modelSeriesSlug };
  }

  private buildFilterBy(query: ModelsQueryDto) {
    const filters: string[] = ['enabled:=true'];

    if (query.provider) filters.push(`providerSlug:=${query.provider}`);
    if (query.category) filters.push(`category:=${query.category}`);
    if (query.status) filters.push(`status:=${query.status}`);
    if (query.region) filters.push(`regions:=${query.region}`);
    if (query.free_only) filters.push('pricingStatus:=free');
    if (query.price_usage)
      filters.push(`pricingUsages:=${query.price_usage}`);
    if (query.price_unit) filters.push(`pricingUnits:=${query.price_unit}`);
    if (query.price_region)
      filters.push(`pricingRegions:=${query.price_region}`);
    if (query.price_resolution)
      filters.push(`pricingResolutions:=${query.price_resolution}`);
    if (query.price_service_tier)
      filters.push(`pricingServiceTiers:=${query.price_service_tier}`);
    if (query.open_source_only) filters.push('openSource:=true');
    if (query.mode?.length) filters.push(this.arrayFilter('mode', query.mode));
    if (query.input_modalities?.length) {
      filters.push(this.arrayFilter('inputModalities', query.input_modalities));
    }
    if (query.output_modalities?.length) {
      filters.push(
        this.arrayFilter('outputModalities', query.output_modalities),
      );
    }
    if (query.supported_parameters?.length) {
      filters.push(
        this.arrayFilter('supportedParameters', query.supported_parameters),
      );
    }
    if (query.capabilities?.length) {
      filters.push(this.arrayFilter('capabilities', query.capabilities));
    }
    if (query.min_context_window !== undefined) {
      filters.push(`contextWindow:>=${query.min_context_window}`);
    }
    if (query.max_context_window !== undefined) {
      filters.push(`contextWindow:<=${query.max_context_window}`);
    }
    if (query.min_output_tokens !== undefined) {
      filters.push(`maxOutputTokens:>=${query.min_output_tokens}`);
    }
    if (query.max_output_tokens !== undefined) {
      filters.push(`maxOutputTokens:<=${query.max_output_tokens}`);
    }
    return filters.join(' && ');
  }

  private arrayFilter(field: string, values: string[]) {
    if (values.length === 1) {
      return `${field}:=${values[0]}`;
    }
    return `${field}:=[${values.join(',')}]`;
  }

  private mapSort(sort: string) {
    switch (sort) {
      case 'name':
        return 'name:asc,_text_match:desc';
      case 'largest_context':
        return 'contextWindow:desc,_text_match:desc';
      case 'most_popular':
        return 'popularityScore:desc,_text_match:desc';
      case 'highest_benchmark_score':
        return 'benchmarkScore:desc,_text_match:desc';
      case 'newest':
      default:
        return 'releaseDateTimestamp:desc,updatedAtTimestamp:desc';
    }
  }

  private toListItem(document: TypesenseDocument) {
    return {
      id: document.id,
      canonicalId: document.canonicalId,
      slug: document.slug,
      name: document.name,
      provider: {
        slug: document.providerSlug,
        name: document.providerName,
        logoUrl: document.providerLogoUrl || null,
      },
      modelSeries:
        document.modelSeriesSlug && document.modelSeriesName
          ? {
              slug: document.modelSeriesSlug,
              name: document.modelSeriesName,
              logoUrl: document.modelSeriesLogoUrl || null,
            }
          : null,
      description: document.description || '',
      status: document.status,
      mode: document.mode,
      contextWindow: document.contextWindow ?? null,
      maxOutputTokens: document.maxOutputTokens ?? null,
      inputModalities: document.inputModalities,
      outputModalities: document.outputModalities,
      supportedParameters: document.supportedParameters,
      capabilities: document.capabilities,
      pricingStatus: document.pricingStatus,
      pricingProfiles: this.parseJson(document.pricingJson, []),
      pricingSummary: this.parseJson(document.pricingSummaryJson, null),
      regions: document.regions,
      openSource: document.openSource,
      createdAt: new Date(document.createdAtTimestamp * 1000),
      updatedAt: new Date(document.updatedAtTimestamp * 1000),
    };
  }

  private parseJson<T>(value: string, fallback: T): T {
    try {
      return JSON.parse(value) as T;
    } catch {
      return fallback;
    }
  }
}
