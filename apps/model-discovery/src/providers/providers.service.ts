import { Injectable } from '@nestjs/common';
import { randomUUID } from 'crypto';
import { mkdirSync, writeFileSync } from 'fs';
import { extname, join } from 'path';
import { PrismaService } from '../database/prisma.service';
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

  upsert(dto: UpsertProviderDto) {
    return this.prisma.provider.upsert({
      where: { slug: dto.slug },
      create: dto as any,
      update: dto as any,
    });
  }

  updateEnabled(slug: string, enabled: boolean) {
    return this.prisma.provider.update({
      where: { slug },
      data: { enabled },
    });
  }

  upsertConnectionProfile(dto: UpsertProviderConnectionProfileDto) {
    return this.prisma.provider.update({
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
  }

  getConnectionProfile(providerSlug: string) {
    return this.prisma.provider.findUnique({
      where: { slug: providerSlug },
      include: { connectionProfile: true },
    });
  }

  saveProviderLogo(providerSlug: string, file: any) {
    const uploadsDir = join(process.cwd(), 'uploads', 'provider-logos');
    mkdirSync(uploadsDir, { recursive: true });
    const ext = extname(file.originalname || '').toLowerCase() || '.png';
    const fileName = `${providerSlug}-${randomUUID()}${ext}`;
    const absPath = join(uploadsDir, fileName);
    writeFileSync(absPath, file.buffer);
    const logoUrl = `/api/v1/providers/assets/provider-logos/${fileName}`;
    return this.prisma.provider.update({
      where: { slug: providerSlug },
      data: { logoUrl },
    });
  }
}
