import { Injectable } from '@nestjs/common';
import { PrismaService } from '../database/prisma.service';

@Injectable()
export class FacetsService {
  constructor(private readonly prisma: PrismaService) {}

  async all() {
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
        select: { slug: true, name: true },
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
      providers,
      categories,
      inputModalities: inputModalities.map((row) => row.modality),
      outputModalities: outputModalities.map((row) => row.modality),
      supportedParameters,
      capabilities,
      regions,
      statuses: statuses.map((s) => s.status),
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
}
