import { Injectable } from '@nestjs/common';
import { randomUUID } from 'crypto';
import { mkdirSync, writeFileSync } from 'fs';
import { extname, join } from 'path';
import { PrismaService } from '../database/prisma.service';
import { auditPrincipal, type ServicePrincipal } from '../common/authenticated-request';
import { UpsertProviderDto } from './dto/upsert-provider.dto';
import { UpsertProviderConnectionProfileDto } from './dto/upsert-provider-connection-profile.dto';

@Injectable()
export class ProvidersService {
  constructor(private readonly prisma: PrismaService) {}

  list() {
    return this.prisma.provider.findMany({
      where: { enabled: true },
      orderBy: { name: 'asc' },
      include: { models: { where: { enabled: true } }, connectionProfile: true },
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

  async upsert(dto: UpsertProviderDto, principal: ServicePrincipal) {
    return this.prisma.$transaction(async (tx) => {
      const before = await tx.provider.findUnique({ where: { slug: dto.slug } });
      const provider = await tx.provider.upsert({
        where: { slug: dto.slug },
        create: dto as any,
        update: dto as any,
      });
      await tx.auditLog.create({ data: {
        entityType: 'provider', entityId: provider.id,
        action: before ? 'update' : 'create', beforeJson: before as any,
        afterJson: provider as any, ...auditPrincipal(principal),
      } });
      return provider;
    });
  }

  async updateEnabled(slug: string, enabled: boolean, principal: ServicePrincipal) {
    return this.prisma.$transaction(async (tx) => {
      const before = await tx.provider.findUniqueOrThrow({ where: { slug } });
      const provider = await tx.provider.update({ where: { slug }, data: { enabled } });
      await tx.auditLog.create({ data: {
        entityType: 'provider', entityId: provider.id, action: 'set_enabled',
        beforeJson: before as any, afterJson: provider as any, ...auditPrincipal(principal),
      } });
      return provider;
    });
  }

  async upsertConnectionProfile(dto: UpsertProviderConnectionProfileDto, principal: ServicePrincipal) {
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
      await tx.auditLog.create({ data: {
        entityType: 'provider_connection_profile', entityId: provider.id,
        action: 'upsert',
        afterJson: {
          providerType: dto.providerType,
          authMode: dto.authMode ?? null,
        },
        ...auditPrincipal(principal),
      } });
      return provider;
    });
  }

  getConnectionProfile(providerSlug: string) {
    return this.prisma.provider.findUnique({
      where: { slug: providerSlug },
      include: { connectionProfile: true },
    });
  }

  async saveProviderLogo(providerSlug: string, file: any, principal: ServicePrincipal) {
    const uploadsDir = join(process.cwd(), 'uploads', 'provider-logos');
    mkdirSync(uploadsDir, { recursive: true });
    const ext = extname(file.originalname || '').toLowerCase() || '.png';
    const fileName = `${providerSlug}-${randomUUID()}${ext}`;
    const absPath = join(uploadsDir, fileName);
    writeFileSync(absPath, file.buffer);
    const logoUrl = `/api/v1/providers/assets/provider-logos/${fileName}`;
    return this.prisma.$transaction(async (tx) => {
      const provider = await tx.provider.update({
        where: { slug: providerSlug },
        data: { logoUrl },
      });
      await tx.auditLog.create({ data: {
        entityType: 'provider', entityId: provider.id, action: 'logo_update',
        afterJson: provider as any, ...auditPrincipal(principal),
      } });
      return provider;
    });
  }
}
