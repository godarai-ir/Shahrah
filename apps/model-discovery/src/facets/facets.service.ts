import { Injectable } from '@nestjs/common';
import { PrismaService } from '../database/prisma.service';
import {
  IndexingService,
  TypesenseFacetCounts,
} from '../indexing/indexing.service';

@Injectable()
export class FacetsService {
  constructor(
    private readonly prisma: PrismaService,
    private readonly indexing: IndexingService,
  ) {}

  async all() {
    const indexedCounts = await this.indexing.facetCounts();
    if (indexedCounts) {
      return this.allFromTypesense(indexedCounts);
    }

    const [
      providers,
      categories,
      inputModalities,
      outputModalities,
      supportedParameters,
      capabilities,
      regions,
      statuses,
    ] = await Promise.all([
      this.prisma.provider.findMany({
        where: { enabled: true },
        select: {
          slug: true,
          name: true,
          _count: { select: { models: { where: { enabled: true } } } },
        },
      }),
      this.prisma.category.findMany({ select: { slug: true, name: true } }),
      this.prisma.modelInputModality.findMany({
        where: { model: { enabled: true, provider: { enabled: true } } },
        distinct: ['modalityId'],
        select: { modality: { select: { slug: true, name: true } } },
      }),
      this.prisma.modelOutputModality.findMany({
        where: { model: { enabled: true, provider: { enabled: true } } },
        distinct: ['modalityId'],
        select: { modality: { select: { slug: true, name: true } } },
      }),
      this.prisma.supportedParameter.findMany({
        select: { slug: true, name: true },
      }),
      this.prisma.capability.findMany({ select: { slug: true, name: true } }),
      this.prisma.region.findMany({ select: { slug: true, name: true } }),
      this.prisma.model.groupBy({
        by: ['status'],
        where: { enabled: true, provider: { enabled: true } },
      }),
    ]);

    return {
      providers: providers.map(({ _count, ...provider }) => ({
        ...provider,
        count: _count.models,
      })),
      categories,
      inputModalities: inputModalities.map((row) => row.modality),
      outputModalities: outputModalities.map((row) => row.modality),
      supportedParameters,
      capabilities,
      regions,
      statuses: statuses.map((s) => s.status),
      meta: {
        source: 'database',
        fallbackUsed: this.indexing.isReadPathEnabled(),
      },
    };
  }

  categories() {
    return this.prisma.category.findMany();
  }

  capabilities() {
    return this.prisma.capability.findMany();
  }

  supportedParameters() {
    return this.prisma.supportedParameter.findMany();
  }

  modalities() {
    return this.prisma.modality.findMany();
  }

  regions() {
    return this.prisma.region.findMany();
  }

  private async allFromTypesense(counts: TypesenseFacetCounts) {
    const [
      providers,
      categories,
      modalities,
      supportedParameters,
      capabilities,
      regions,
    ] = await Promise.all([
      this.prisma.provider.findMany({
        where: { enabled: true },
        select: { slug: true, name: true },
      }),
      this.prisma.category.findMany({
        where: { slug: { in: Array.from(counts.categoryCounts.keys()) } },
        select: { slug: true, name: true },
      }),
      this.prisma.modality.findMany({
        where: {
          slug: {
            in: Array.from(
              new Set([
                ...counts.inputModalityCounts.keys(),
                ...counts.outputModalityCounts.keys(),
              ]),
            ),
          },
        },
        select: { slug: true, name: true },
      }),
      this.prisma.supportedParameter.findMany({
        where: {
          slug: { in: Array.from(counts.supportedParameterCounts.keys()) },
        },
        select: { slug: true, name: true },
      }),
      this.prisma.capability.findMany({
        where: { slug: { in: Array.from(counts.capabilityCounts.keys()) } },
        select: { slug: true, name: true },
      }),
      this.prisma.region.findMany({
        where: { slug: { in: Array.from(counts.regionCounts.keys()) } },
        select: { slug: true, name: true },
      }),
    ]);

    return {
      providers: providers.map((provider) => ({
        ...provider,
        count: counts.providerCounts.get(provider.slug) || 0,
      })),
      categories: categories.map((category) => ({
        ...category,
        count: counts.categoryCounts.get(category.slug) || 0,
      })),
      inputModalities: modalities
        .filter((modality) => counts.inputModalityCounts.has(modality.slug))
        .map((modality) => ({
          ...modality,
          count: counts.inputModalityCounts.get(modality.slug) || 0,
        })),
      outputModalities: modalities
        .filter((modality) => counts.outputModalityCounts.has(modality.slug))
        .map((modality) => ({
          ...modality,
          count: counts.outputModalityCounts.get(modality.slug) || 0,
        })),
      supportedParameters: supportedParameters.map((parameter) => ({
        ...parameter,
        count: counts.supportedParameterCounts.get(parameter.slug) || 0,
      })),
      capabilities: capabilities.map((capability) => ({
        ...capability,
        count: counts.capabilityCounts.get(capability.slug) || 0,
      })),
      regions: regions.map((region) => ({
        ...region,
        count: counts.regionCounts.get(region.slug) || 0,
      })),
      statuses: Array.from(counts.statusCounts.keys()),
      statusCounts: Array.from(counts.statusCounts.entries()).map(
        ([status, count]) => ({
          status,
          count,
        }),
      ),
      meta: {
        source: counts.source,
        fallbackUsed: false,
      },
    };
  }
}
