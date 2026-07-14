import { Injectable } from '@nestjs/common';
import { randomUUID } from 'crypto';
import { mkdirSync, writeFileSync } from 'fs';
import { extname, join } from 'path';
import { PrismaService } from '../database/prisma.service';
import {
  auditPrincipal,
  type ServicePrincipal,
} from '../common/authenticated-request';
import { IndexingService } from '../indexing/indexing.service';
import { UpsertProviderDto } from './dto/upsert-provider.dto';
import { UpsertProviderConnectionProfileDto } from './dto/upsert-provider-connection-profile.dto';

@Injectable()
export class ProvidersService {
  constructor(
    private readonly prisma: PrismaService,
    private readonly indexing: IndexingService,
  ) {}

  list() {
    return this.prisma.provider.findMany({
      where: { enabled: true },
      orderBy: { name: 'asc' },
      include: {
        models: { where: { enabled: true } },
        connectionProfile: true,
      },
    });
  }

  listAllForAdmin() {
    return this.prisma.provider.findMany({
      orderBy: { name: 'asc' },
      include: {
        models: { orderBy: { displayName: 'asc' } },
        connectionProfile: true,
      },
    });
  }

  detail(slug: string) {
    return this.prisma.provider.findFirst({
      where: { slug, enabled: true },
      include: {
        models: {
          where: { enabled: true },
          include: {
            pricing: { take: 1, orderBy: { effectiveFrom: 'desc' } },
            capabilities: { include: { capability: true } },
          },
          orderBy: { releaseDate: 'desc' },
        },
      },
    });
  }

  async listModels(slug: string) {
    const rows = await this.prisma.model.findMany({
      where: { enabled: true, provider: { slug, enabled: true } },
      include: {
        pricing: { take: 1, orderBy: { effectiveFrom: 'desc' } },
        inputModalities: { include: { modality: true } },
        outputModalities: { include: { modality: true } },
        supportedParams: { include: { supportedParameter: true } },
        capabilities: { include: { capability: true } },
        regions: { include: { region: true } },
      },
      orderBy: [{ releaseDate: 'desc' }, { createdAt: 'desc' }],
    });

    return {
      data: rows.map((row) => ({
        canonicalId: row.canonicalId,
        slug: row.slug,
        name: row.displayName,
        mode: row.mode,
        contextWindow: row.contextWindow,
        maxOutputTokens: row.maxOutputTokens,
        inputModalities: row.inputModalities.map((item) => item.modality.slug),
        outputModalities: row.outputModalities.map(
          (item) => item.modality.slug,
        ),
        supportedParameters: row.supportedParams.map(
          (item) => item.supportedParameter.slug,
        ),
        capabilities: row.capabilities.map((item) => item.capability.slug),
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
        regions: row.regions.map((item) => item.region.slug),
      })),
    };
  }

  async upsert(dto: UpsertProviderDto, principal: ServicePrincipal) {
    const provider = await this.prisma.$transaction(async (tx) => {
      const before = await tx.provider.findUnique({
        where: { slug: dto.slug },
      });
      const provider = await tx.provider.upsert({
        where: { slug: dto.slug },
        create: dto as any,
        update: dto as any,
      });
      await tx.auditLog.create({
        data: {
          entityType: 'provider',
          entityId: provider.id,
          action: before ? 'update' : 'create',
          beforeJson: before as any,
          afterJson: provider as any,
          ...auditPrincipal(principal),
        },
      });
      return provider;
    });
    await this.indexing.enqueueProviderModelsReindex(provider.slug);
    return provider;
  }

  async updateEnabled(
    slug: string,
    enabled: boolean,
    principal: ServicePrincipal,
  ) {
    const provider = await this.prisma.$transaction(async (tx) => {
      const before = await tx.provider.findUniqueOrThrow({ where: { slug } });
      const provider = await tx.provider.update({
        where: { slug },
        data: { enabled },
      });
      await tx.auditLog.create({
        data: {
          entityType: 'provider',
          entityId: provider.id,
          action: 'set_enabled',
          beforeJson: before as any,
          afterJson: provider as any,
          ...auditPrincipal(principal),
        },
      });
      return provider;
    });
    await this.indexing.enqueueProviderModelsReindex(provider.slug);
    return provider;
  }

  async upsertConnectionProfile(
    dto: UpsertProviderConnectionProfileDto,
    principal: ServicePrincipal,
  ) {
    return this.prisma.$transaction(async (tx) => {
      const provider = await tx.provider.update({
        where: { slug: dto.providerSlug },
        data: {
          connectionProfile: {
            upsert: {
              create: {
                providerType: dto.providerType,
                authMode: dto.authMode,
                configJson: dto.configJson as any,
              },
              update: {
                providerType: dto.providerType,
                authMode: dto.authMode,
                configJson: dto.configJson as any,
              },
            },
          },
        },
        include: { connectionProfile: true },
      });
      await tx.auditLog.create({
        data: {
          entityType: 'provider_connection_profile',
          entityId: provider.id,
          action: 'upsert',
          afterJson: {
            providerType: dto.providerType,
            authMode: dto.authMode ?? null,
          },
          ...auditPrincipal(principal),
        },
      });
      return provider;
    });
  }

  getConnectionProfile(providerSlug: string) {
    return this.prisma.provider.findUnique({
      where: { slug: providerSlug },
      include: { connectionProfile: true },
    });
  }

  async saveProviderLogo(
    providerSlug: string,
    file: any,
    principal: ServicePrincipal,
  ) {
    const uploadsDir = join(process.cwd(), 'uploads', 'provider-logos');
    mkdirSync(uploadsDir, { recursive: true });
    const ext = extname(file.originalname || '').toLowerCase() || '.png';
    const fileName = `${providerSlug}-${randomUUID()}${ext}`;
    const absPath = join(uploadsDir, fileName);
    writeFileSync(absPath, file.buffer);
    const logoUrl = `/api/v1/providers/assets/provider-logos/${fileName}`;
    const provider = await this.prisma.$transaction(async (tx) => {
      const provider = await tx.provider.update({
        where: { slug: providerSlug },
        data: { logoUrl },
      });
      await tx.auditLog.create({
        data: {
          entityType: 'provider',
          entityId: provider.id,
          action: 'logo_update',
          afterJson: provider as any,
          ...auditPrincipal(principal),
        },
      });
      return provider;
    });
    await this.indexing.enqueueProviderModelsReindex(provider.slug);
    return provider;
  }
}
