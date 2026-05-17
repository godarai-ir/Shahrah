import { Injectable } from '@nestjs/common';
import { PrismaService } from '../database/prisma.service';
import { UpsertProviderDto } from './dto/upsert-provider.dto';

@Injectable()
export class ProvidersService {
  constructor(private readonly prisma: PrismaService) {}

  list() {
    return this.prisma.provider.findMany({
      orderBy: { name: 'asc' },
      include: { models: true },
    });
  }

  detail(slug: string) {
    return this.prisma.provider.findUnique({
      where: { slug },
      include: {
        models: {
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
}
