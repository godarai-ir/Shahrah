import { Injectable } from '@nestjs/common';
import { randomUUID } from 'crypto';
import { PrismaService } from '../database/prisma.service';
import {
  auditPrincipal,
  type ServicePrincipal,
} from '../common/authenticated-request';
import { IndexingService } from '../indexing/indexing.service';
import { UpsertProviderDto } from './dto/upsert-provider.dto';
import { validateConnectionSchemaFields } from './connection-schema-validation';
import { UpsertProviderConnectionSchemaDto } from './dto/upsert-provider-connection-schema.dto';
import { processProviderLogoImage } from './provider-logo-image';

type ProviderLogoAssetResult = {
  body: Buffer;
  contentType: string;
};

@Injectable()
export class ProvidersService {
  constructor(
    private readonly prisma: PrismaService,
    private readonly indexing: IndexingService,
  ) {}

  async list() {
    const providers = await this.prisma.provider.findMany({
      where: { enabled: true },
      orderBy: { name: 'asc' },
      include: {
        models: {
          where: { enabled: true },
          select: { id: true, mode: true },
        },
      },
    });

    return providers.map((provider) => {
      const safeProvider = { ...provider } as Record<string, unknown>;
      const models = provider.models;
      const modes = Array.from(
        new Set(models.flatMap((model) => model.mode)),
      ).sort();
      delete safeProvider.models;
      delete safeProvider.legacyConnectionProfile;
      delete safeProvider.connectionProfile;
      return { ...safeProvider, modelCount: models.length, modes };
    });
  }

  listAllForAdmin() {
    return this.prisma.provider.findMany({
      orderBy: { name: 'asc' },
      include: {
        models: { orderBy: { displayName: 'asc' } },
        connectionSchemas: {
          orderBy: [{ authMode: 'asc' }, { version: 'desc' }],
        },
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

  async upsertConnectionSchema(
    dto: UpsertProviderConnectionSchemaDto,
    principal: ServicePrincipal,
  ) {
    validateConnectionSchemaFields(dto.fields);
    return this.prisma.$transaction(async (tx) => {
      const provider = await tx.provider.findUniqueOrThrow({
        where: { slug: dto.providerSlug },
      });
      const schema = await tx.providerConnectionSchema.upsert({
        where: {
          providerId_authMode_version: {
            providerId: provider.id,
            authMode: dto.authMode,
            version: dto.version,
          },
        },
        create: {
          providerId: provider.id,
          authMode: dto.authMode,
          version: dto.version,
          runtimeReady: dto.runtimeReady,
          fieldsJson: dto.fields as any,
        },
        update: {
          runtimeReady: dto.runtimeReady,
          fieldsJson: dto.fields as any,
        },
      });
      await tx.auditLog.create({
        data: {
          entityType: 'provider_connection_schema',
          entityId: schema.id,
          action: 'upsert',
          afterJson: {
            providerSlug: dto.providerSlug,
            authMode: dto.authMode,
            version: dto.version,
            runtimeReady: dto.runtimeReady,
            fieldKeys: dto.fields.map((field) => String(field.key ?? '')),
          },
          ...auditPrincipal(principal),
        },
      });
      return schema;
    });
  }

  listConnectionSchemas(providerSlug: string) {
    return this.prisma.provider.findUnique({
      where: { slug: providerSlug },
      select: {
        slug: true,
        connectionSchemas: {
          orderBy: [{ authMode: 'asc' }, { version: 'desc' }],
        },
      },
    });
  }

  async saveProviderLogo(
    providerSlug: string,
    file: any,
    principal: ServicePrincipal,
  ) {
    const rawBuffer = (file as { buffer?: unknown }).buffer;
    const uploadBuffer = Buffer.isBuffer(rawBuffer)
      ? rawBuffer
      : Buffer.from(rawBuffer as Uint8Array);
    const processed = await processProviderLogoImage(uploadBuffer);
    const fileName = providerLogoFileName(providerSlug, processed.contentHash);
    const logoUrl = `/api/v1/providers/assets/provider-logos/${fileName}`;
    const provider = await this.prisma.$transaction(async (tx) => {
      const before = await tx.provider.findUniqueOrThrow({
        where: { slug: providerSlug },
      });
      await tx.providerLogoAsset.create({
        data: {
          providerId: before.id,
          fileName,
          contentHash: processed.contentHash,
          contentType: processed.contentType,
          width: processed.width,
          height: processed.height,
          sizeBytes: processed.sizeBytes,
          data: toPrismaBytes(processed.data),
        },
      });
      const provider = await tx.provider.update({
        where: { slug: providerSlug },
        data: { logoUrl },
      });
      await tx.auditLog.create({
        data: {
          entityType: 'provider',
          entityId: provider.id,
          action: 'logo_update',
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

  async getProviderLogoAsset(
    fileName: string,
  ): Promise<ProviderLogoAssetResult | null> {
    const asset = await this.prisma.providerLogoAsset.findUnique({
      where: { fileName },
      select: {
        data: true,
        contentType: true,
      },
    });
    if (!asset) return null;

    return {
      body: Buffer.from(asset.data),
      contentType: asset.contentType,
    };
  }
}

function providerLogoFileName(
  providerSlug: string,
  contentHash: string,
): string {
  const safeSlug =
    providerSlug
      .trim()
      .toLowerCase()
      .replace(/[^a-z0-9_-]+/g, '-')
      .replace(/^-+|-+$/g, '') || 'provider';
  return `${safeSlug}-${contentHash.slice(0, 16)}-${randomUUID()}.png`;
}

function toPrismaBytes(buffer: Buffer): Uint8Array<ArrayBuffer> {
  const arrayBuffer = buffer.buffer.slice(
    buffer.byteOffset,
    buffer.byteOffset + buffer.byteLength,
  ) as ArrayBuffer;
  return new Uint8Array(arrayBuffer);
}
