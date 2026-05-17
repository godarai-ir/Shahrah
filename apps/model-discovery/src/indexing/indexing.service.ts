import { InjectQueue } from '@nestjs/bullmq';
import { Injectable } from '@nestjs/common';
import { ConfigService } from '@nestjs/config';
import { Queue } from 'bullmq';
import Typesense from 'typesense';
import { PrismaService } from '../database/prisma.service';

export const INDEXING_QUEUE = 'catalog-indexing';

@Injectable()
export class IndexingService {
  private readonly client: any;
  private readonly collectionName: string;

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
    });
    this.collectionName = cfg.get<string>(
      'TYPESENSE_COLLECTION_MODELS',
      'models',
    );
  }

  enqueueFullReindex() {
    return this.queue.add('full-reindex', {});
  }
  enqueueModelIndex(modelId: string) {
    return this.queue.add('index-model', { modelId });
  }

  async ensureCollection() {
    const schema = {
      name: this.collectionName,
      fields: [
        { name: 'id', type: 'string' },
        { name: 'canonicalId', type: 'string' },
        { name: 'name', type: 'string' },
        { name: 'slug', type: 'string' },
        { name: 'providerSlug', type: 'string', facet: true },
        { name: 'providerName', type: 'string', facet: true },
        { name: 'description', type: 'string', optional: true },
        { name: 'status', type: 'string', facet: true },
        { name: 'lifecycleStatus', type: 'string', facet: true },
        { name: 'category', type: 'string', facet: true, optional: true },
        { name: 'contextWindow', type: 'int32', sort: true, optional: true },
        { name: 'maxOutputTokens', type: 'int32', sort: true, optional: true },
        {
          name: 'inputPricePerMillionTokens',
          type: 'float',
          sort: true,
          optional: true,
        },
        {
          name: 'outputPricePerMillionTokens',
          type: 'float',
          sort: true,
          optional: true,
        },
        {
          name: 'cachedInputPricePerMillionTokens',
          type: 'float',
          optional: true,
        },
        { name: 'free', type: 'bool', facet: true },
        { name: 'openSource', type: 'bool', facet: true },
        { name: 'inputModalities', type: 'string[]', facet: true },
        { name: 'outputModalities', type: 'string[]', facet: true },
        { name: 'supportedParameters', type: 'string[]', facet: true },
        { name: 'capabilities', type: 'string[]', facet: true },
        { name: 'regions', type: 'string[]', facet: true },
        { name: 'tags', type: 'string[]', facet: true },
        { name: 'benchmarkScore', type: 'float', sort: true, optional: true },
        {
          name: 'releaseDateTimestamp',
          type: 'int64',
          sort: true,
          optional: true,
        },
        { name: 'updatedAtTimestamp', type: 'int64', sort: true },
      ],
      default_sorting_field: 'updatedAtTimestamp',
    };

    try {
      await this.client.collections(this.collectionName).retrieve();
    } catch {
      await this.client.collections().create(schema as any);
    }
  }

  async indexOneModel(modelId: string) {
    const model = await this.prisma.model.findUnique({
      where: { id: modelId },
      include: {
        provider: true,
        category: true,
        pricing: { take: 1, orderBy: { effectiveFrom: 'desc' } },
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
    const p = model.pricing[0];
    const doc = {
      id: model.id,
      canonicalId: model.canonicalId,
      name: model.displayName,
      slug: model.slug,
      providerSlug: model.provider.slug,
      providerName: model.provider.name,
      description: model.shortDescription || '',
      status: model.status,
      lifecycleStatus: model.lifecycleStatus,
      category: model.category?.slug || '',
      contextWindow: model.contextWindow || 0,
      maxOutputTokens: model.maxOutputTokens || 0,
      inputPricePerMillionTokens: p?.inputPricePerMillionTokens ?? null,
      outputPricePerMillionTokens: p?.outputPricePerMillionTokens ?? null,
      cachedInputPricePerMillionTokens:
        p?.cachedInputPricePerMillionTokens ?? null,
      free: model.free,
      openSource: model.openSource,
      inputModalities: model.inputModalities.map((x) => x.modality.slug),
      outputModalities: model.outputModalities.map((x) => x.modality.slug),
      supportedParameters: model.supportedParams.map(
        (x) => x.supportedParameter.slug,
      ),
      capabilities: model.capabilities.map((x) => x.capability.slug),
      regions: model.regions.map((x) => x.region.slug),
      tags: model.tags.map((x) => x.tag.slug),
      benchmarkScore: model.benchmarks[0]?.score ?? 0,
      releaseDateTimestamp: model.releaseDate
        ? Math.floor(model.releaseDate.getTime() / 1000)
        : 0,
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
}
