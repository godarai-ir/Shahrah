import {
  BadRequestException,
  Body,
  Controller,
  Get,
  Param,
  ParseBoolPipe,
  Post,
  Put,
  Query,
  Req,
  UploadedFile,
  UseGuards,
  UseInterceptors,
} from '@nestjs/common';
import { ApiBearerAuth, ApiTags } from '@nestjs/swagger';
import { FileInterceptor } from '@nestjs/platform-express';
import { randomUUID } from 'crypto';
import { mkdirSync, writeFileSync } from 'fs';
import { extname, join } from 'path';
import { PrismaService } from '../database/prisma.service';
import {
  auditActor,
  type AuthenticatedRequest,
} from '../common/authenticated-request';
import { Scopes } from '../common/decorators/scopes.decorator';
import { ScopesGuard } from '../common/guards/scopes.guard';
import { ServiceAccountGuard } from '../common/guards/service-account.guard';
import { IndexingService } from '../indexing/indexing.service';
import {
  resolvePricingTier,
  UpdateModelPricingDto,
  UpsertModelDto,
} from './dto/upsert-model.dto';
import { UpsertModelSeriesDto } from './dto/upsert-model-series.dto';
import { Decimal } from '@prisma/client-runtime-utils';
import { serializePricingProfiles } from '../pricing/pricing-response';
import { evaluatePublicationReadiness } from '../publication/publication-policy';

@ApiTags('admin')
@ApiBearerAuth()
@UseGuards(ServiceAccountGuard, ScopesGuard)
@Controller({ path: 'admin/models', version: '1' })
export class CatalogAdminController {
  constructor(
    private readonly prisma: PrismaService,
    private readonly indexing: IndexingService,
  ) {}

  @Post('upsert')
  @Scopes('catalog:write')
  async upsert(
    @Body() dto: UpsertModelDto,
    @Req() request: AuthenticatedRequest,
  ) {
    if (dto.lifecycleStatus === 'published') {
      throw new BadRequestException(
        'Use the dedicated publish action after pricing validation.',
      );
    }
    this.validatePricingProfiles(dto);
    const provider = await this.prisma.provider.findUniqueOrThrow({
      where: { slug: dto.providerSlug },
    });
    const category = dto.categorySlug
      ? await this.prisma.category.findUnique({
          where: { slug: dto.categorySlug },
        })
      : null;
    const modelSeries = dto.modelSeriesSlug
      ? await this.prisma.modelSeries.findUnique({
          where: { slug: dto.modelSeriesSlug },
        })
      : null;

    const before = await this.prisma.model.findUnique({
      where: { canonicalId: dto.canonicalId },
    });
    const model = await this.prisma.$transaction(async (tx) => {
      const model = await tx.model.upsert({
        where: { canonicalId: dto.canonicalId },
        create: {
          providerId: provider.id,
          canonicalId: dto.canonicalId,
          slug: dto.slug,
          displayName: dto.displayName,
          shortDescription: dto.shortDescription,
          longDescription: dto.longDescription,
          status: (dto.status as any) || 'active',
          enabled: dto.enabled ?? true,
          lifecycleStatus: (dto.lifecycleStatus as any) || 'draft',
          modelSeriesId: modelSeries?.id,
          categoryId: category?.id,
          contextWindow: dto.contextWindow,
          maxOutputTokens: dto.maxOutputTokens,
          openSource: dto.openSource ?? false,
          pricingStatus:
            dto.pricingProfiles !== undefined ||
            dto.pricingStatus !== undefined
              ? this.resolvePricingStatus(dto)
              : undefined,
          pricingAuthority:
            dto.pricingProfiles !== undefined ? 'admin' : 'import',
          mode: dto.mode?.length ? dto.mode : ['unknown'],
        },
        update: {
          providerId: provider.id,
          slug: dto.slug,
          displayName: dto.displayName,
          shortDescription: dto.shortDescription,
          longDescription: dto.longDescription,
          status: (dto.status as any) || undefined,
          enabled: dto.enabled,
          lifecycleStatus: (dto.lifecycleStatus as any) || undefined,
          modelSeriesId: modelSeries?.id,
          categoryId: category?.id,
          contextWindow: dto.contextWindow,
          maxOutputTokens: dto.maxOutputTokens,
          openSource: dto.openSource,
          pricingStatus: this.resolvePricingStatus(dto),
          pricingAuthority:
            dto.pricingProfiles !== undefined ? 'admin' : undefined,
          mode: dto.mode?.length ? dto.mode : undefined,
        },
      });

      if (dto.pricingProfiles) {
        await tx.modelPricing.deleteMany({ where: { modelId: model.id } });
        for (const profile of dto.pricingProfiles) {
          const regionSlug = this.normalizeRegion(profile.region);
          const region = await tx.region.upsert({
            where: { slug: regionSlug },
            create: { slug: regionSlug, name: regionSlug },
            update: {},
          });
          await tx.modelPricing.create({
            data: {
              modelId: model.id,
              regionId: region.id,
              currency: profile.currency || 'USD',
              sourceUrl: profile.sourceUrl,
              effectiveFrom: profile.effectiveFrom
                ? new Date(profile.effectiveFrom)
                : new Date(),
              effectiveTo: profile.effectiveTo
                ? new Date(profile.effectiveTo)
                : undefined,
              rates: {
                create: profile.rates.map((rate) => {
                  const amount = new Decimal(rate.amount);
                  const tier = resolvePricingTier(rate);
                  if (!amount.isFinite() || amount.isNegative()) {
                    throw new BadRequestException(
                      `Invalid amount for ${rate.key ?? rate.unit}`,
                    );
                  }
                  return {
                    rateKey: rate.key || this.rateKey(rate, tier.from),
                    usage: rate.usage as any,
                    unit: rate.unit,
                    amount: rate.amount,
                    quantity: rate.quantity,
                    serviceTier: (rate.serviceTier ?? 'standard') as any,
                    resolution: rate.resolution,
                    tierFrom: tier.from,
                    tierTo: tier.to,
                    tierMode: tier.mode,
                    qualifiers: rate.qualifiers as any,
                  };
                }),
              },
            },
          });
        }
      }

      await tx.auditLog.create({
        data: {
          entityType: 'model',
          entityId: model.id,
          action: before ? 'update' : 'create',
          beforeJson: before as any,
          afterJson: model as any,
          ...auditActor(request),
        },
      });

      return model;
    });
    await this.indexing.enqueueModelIndex(model.id);
    return model;
  }

  private resolvePricingStatus(dto: UpsertModelDto) {
    if (dto.pricingStatus === 'free') return 'free' as const;
    if (!dto.pricingProfiles) {
      return (dto.pricingStatus as 'unknown' | 'priced' | undefined) ??
        'unknown';
    }
    const hasPositiveRate = dto.pricingProfiles.some((profile) =>
      profile.rates.some((rate) => {
        try {
          return new Decimal(rate.amount).greaterThan(0);
        } catch {
          return false;
        }
      }),
    );
    return hasPositiveRate ? ('priced' as const) : ('unknown' as const);
  }

  private validatePricingProfiles(dto: UpsertModelDto) {
    for (const profile of dto.pricingProfiles ?? []) {
      const effectiveFrom = profile.effectiveFrom
        ? new Date(profile.effectiveFrom)
        : undefined;
      const effectiveTo = profile.effectiveTo
        ? new Date(profile.effectiveTo)
        : undefined;
      if (
        effectiveFrom &&
        effectiveTo &&
        effectiveTo.getTime() <= effectiveFrom.getTime()
      ) {
        throw new BadRequestException(
          `Pricing profile ${profile.region} has an invalid effective interval`,
        );
      }

      const keys = new Set<string>();
      for (const rate of profile.rates) {
        const tier = resolvePricingTier(rate);
        const key = rate.key || this.rateKey(rate, tier.from);
        if (keys.has(key)) {
          throw new BadRequestException(
            `Duplicate pricing rate key ${key} in ${profile.region}`,
          );
        }
        keys.add(key);
        if (
          tier.from !== undefined &&
          tier.to !== undefined &&
          tier.to <= tier.from
        ) {
          throw new BadRequestException(
            `Invalid tier bounds for pricing rate ${key}`,
          );
        }
        if (
          dto.pricingStatus === 'free' &&
          new Decimal(rate.amount).greaterThan(0)
        ) {
          throw new BadRequestException(
            'A free model cannot contain a positive pricing rate',
          );
        }
      }
    }
  }

  private normalizeRegion(region: string) {
    const normalized = region.trim().toLowerCase();
    if (!normalized || normalized === '*' || normalized === 'global') {
      return 'global';
    }
    return normalized
      .replace(/[^a-z0-9]+/g, '-')
      .replace(/^-+|-+$/g, '');
  }

  private rateKey(
    rate: {
      usage: string;
      unit: string;
      serviceTier?: string;
      resolution?: string;
    },
    tierFrom?: number,
  ) {
    return [
      rate.usage,
      rate.unit,
      rate.serviceTier ?? 'standard',
      `resolution=${rate.resolution ?? 'default'}`,
      `tier=${tierFrom ?? 'base'}`,
    ].join(':');
  }

  @Get('audit-log')
  @Scopes('catalog:read')
  audit() {
    return this.prisma.auditLog.findMany({
      orderBy: { createdAt: 'desc' },
      take: 200,
    });
  }

  @Get()
  @Scopes('catalog:read')
  async listAll(
    @Query('page') rawPage?: string,
    @Query('limit') rawLimit?: string,
    @Query('q') q?: string,
    @Query('provider') provider?: string,
    @Query('mode') mode?: string,
    @Query('lifecycle') lifecycle?: string,
    @Query('readiness') readiness?: string,
  ) {
    const page = Math.max(1, Number(rawPage) || 1);
    const limit = Math.min(200, Math.max(1, Number(rawLimit) || 50));
    const where: any = {
      provider: provider ? { slug: provider } : undefined,
      mode: mode ? { has: mode } : undefined,
      lifecycleStatus: lifecycle || undefined,
      OR: q
        ? [
            { canonicalId: { contains: q, mode: 'insensitive' } },
            { displayName: { contains: q, mode: 'insensitive' } },
          ]
        : undefined,
    };
    const rows = await this.prisma.model.findMany({
      where,
      orderBy: [{ providerId: 'asc' }, { displayName: 'asc' }],
      include: {
        provider: true,
        modelSeries: true,
        pricing: {
          include: { region: true, rates: { orderBy: { rateKey: 'asc' } } },
          orderBy: [{ region: { slug: 'asc' } }, { effectiveFrom: 'desc' }],
        },
      },
    });
    const mapped = rows.map((row) => this.adminModel(row));
    const filtered = readiness
      ? mapped.filter((row) => row.publication.state === readiness)
      : mapped;
    const data = filtered.slice((page - 1) * limit, page * limit);
    return {
      data,
      pagination: {
        page,
        limit,
        total: filtered.length,
        totalPages: Math.max(1, Math.ceil(filtered.length / limit)),
      },
      summary: mapped.reduce<Record<string, number>>((acc, row) => {
        acc[row.publication.state] =
          (acc[row.publication.state] ?? 0) + 1;
        return acc;
      }, {}),
      providerSummary: filtered.reduce<Record<string, number>>((acc, row) => {
        acc[row.provider.name] = (acc[row.provider.name] ?? 0) + 1;
        return acc;
      }, {}),
    };
  }

  @Get('detail/:canonicalId')
  @Scopes('catalog:read')
  async detail(@Param('canonicalId') canonicalId: string) {
    const row = await this.prisma.model.findUniqueOrThrow({
      where: { canonicalId: decodeURIComponent(canonicalId) },
      include: {
        provider: true,
        modelSeries: true,
        pricing: {
          include: { region: true, rates: { orderBy: { rateKey: 'asc' } } },
          orderBy: [{ region: { slug: 'asc' } }, { effectiveFrom: 'desc' }],
        },
      },
    });
    return this.adminModel(row);
  }

  @Put(':canonicalId/pricing')
  @Scopes('catalog:write')
  async updatePricing(
    @Param('canonicalId') canonicalId: string,
    @Body() body: UpdateModelPricingDto,
    @Req() request: AuthenticatedRequest,
  ) {
    const decodedId = decodeURIComponent(canonicalId);
    const before = await this.prisma.model.findUniqueOrThrow({
      where: { canonicalId: decodedId },
      include: {
        pricing: { include: { region: true, rates: true } },
      },
    });
    if (before.lifecycleStatus === 'published') {
      throw new BadRequestException(
        'Unpublish the model before editing public pricing.',
      );
    }
    const dto = {
      providerSlug: '',
      canonicalId: decodedId,
      slug: before.slug,
      displayName: before.displayName,
      pricingStatus: body.pricingStatus,
      pricingProfiles: body.pricingProfiles ?? [],
    } satisfies UpsertModelDto;
    this.validatePricingProfiles(dto);
    const pricingStatus = this.resolvePricingStatus(dto);

    const model = await this.prisma.$transaction(async (tx) => {
      await tx.modelPricing.deleteMany({ where: { modelId: before.id } });
      for (const profile of dto.pricingProfiles ?? []) {
        const regionSlug = this.normalizeRegion(profile.region);
        const region = await tx.region.upsert({
          where: { slug: regionSlug },
          create: { slug: regionSlug, name: regionSlug },
          update: {},
        });
        await tx.modelPricing.create({
          data: {
            modelId: before.id,
            regionId: region.id,
            currency: profile.currency || 'USD',
            sourceUrl: profile.sourceUrl,
            effectiveFrom: profile.effectiveFrom
              ? new Date(profile.effectiveFrom)
              : new Date(),
            effectiveTo: profile.effectiveTo
              ? new Date(profile.effectiveTo)
              : undefined,
            rates: {
              create: profile.rates.map((rate) => {
                const tier = resolvePricingTier(rate);
                return {
                  rateKey: rate.key || this.rateKey(rate, tier.from),
                  usage: rate.usage as any,
                  unit: rate.unit,
                  amount: rate.amount,
                  quantity: rate.quantity,
                  serviceTier: (rate.serviceTier ?? 'standard') as any,
                  resolution: rate.resolution,
                  tierFrom: tier.from,
                  tierTo: tier.to,
                  tierMode: tier.mode,
                  qualifiers: rate.qualifiers as any,
                };
              }),
            },
          },
        });
      }
      const candidate = await tx.model.update({
        where: { id: before.id },
        data: {
          pricingStatus,
          pricingAuthority: 'admin',
          lifecycleStatus: 'draft',
        },
      });
      const withPricing = await tx.model.findUniqueOrThrow({
        where: { id: before.id },
        include: {
          provider: true,
          modelSeries: true,
          pricing: { include: { region: true, rates: true } },
        },
      });
      const publication = this.publication(withPricing);
      const updated = publication.eligible
        ? await tx.model.update({
            where: { id: candidate.id },
            data: { lifecycleStatus: 'review' },
          })
        : candidate;
      await tx.auditLog.create({
        data: {
          entityType: 'model',
          entityId: before.id,
          action: 'pricing_update',
          beforeJson: before as any,
          afterJson: { ...updated, publication } as any,
          ...auditActor(request),
        },
      });
      return updated;
    });
    await this.indexing.enqueueModelIndex(model.id);
    return this.detail(encodeURIComponent(decodedId));
  }

  @Post(':canonicalId/publish')
  @Scopes('catalog:write')
  async publish(
    @Param('canonicalId') canonicalId: string,
    @Req() request: AuthenticatedRequest,
  ) {
    return this.setPublicationStatus(
      decodeURIComponent(canonicalId),
      'published',
      request,
    );
  }

  @Post(':canonicalId/unpublish')
  @Scopes('catalog:write')
  async unpublish(
    @Param('canonicalId') canonicalId: string,
    @Req() request: AuthenticatedRequest,
  ) {
    return this.setPublicationStatus(
      decodeURIComponent(canonicalId),
      'draft',
      request,
    );
  }

  @Post(':canonicalId/pricing-authority/import')
  @Scopes('catalog:write')
  async releasePricingAuthority(
    @Param('canonicalId') canonicalId: string,
    @Req() request: AuthenticatedRequest,
  ) {
    const decodedId = decodeURIComponent(canonicalId);
    const model = await this.prisma.$transaction(async (tx) => {
      const before = await tx.model.findUniqueOrThrow({
        where: { canonicalId: decodedId },
      });
      const updated = await tx.model.update({
        where: { id: before.id },
        data: {
          pricingAuthority: 'import',
          lifecycleStatus: 'draft',
        },
      });
      await tx.auditLog.create({
        data: {
          entityType: 'model',
          entityId: before.id,
          action: 'release_pricing_authority',
          beforeJson: before as any,
          afterJson: updated as any,
          ...auditActor(request),
        },
      });
      return updated;
    });
    await this.indexing.enqueueModelIndex(model.id);
    return this.detail(encodeURIComponent(decodedId));
  }

  private async setPublicationStatus(
    canonicalId: string,
    lifecycleStatus: 'draft' | 'published',
    request: AuthenticatedRequest,
  ) {
    const model = await this.prisma.$transaction(async (tx) => {
      const before = await tx.model.findUniqueOrThrow({
        where: { canonicalId },
        include: {
          provider: true,
          modelSeries: true,
          pricing: { include: { region: true, rates: true } },
        },
      });
      const publication = this.publication(before);
      if (lifecycleStatus === 'published' && !publication.eligible) {
        throw new BadRequestException({
          code: 'model_not_ready_for_publication',
          blockers: publication.blockers,
        });
      }
      const updated = await tx.model.update({
        where: { id: before.id },
        data: { lifecycleStatus },
      });
      await tx.auditLog.create({
        data: {
          entityType: 'model',
          entityId: before.id,
          action:
            lifecycleStatus === 'published' ? 'publish' : 'unpublish',
          beforeJson: before as any,
          afterJson: { ...updated, publication } as any,
          ...auditActor(request),
        },
      });
      return updated;
    });
    await this.indexing.enqueueModelIndex(model.id);
    return this.detail(encodeURIComponent(canonicalId));
  }

  private adminModel(row: any) {
    const pricingProfiles = serializePricingProfiles(row.pricing ?? []);
    const publication = evaluatePublicationReadiness({
      mode: row.mode,
      pricingStatus: row.pricingStatus,
      pricingProfiles,
    });
    return {
      ...row,
      pricing: undefined,
      pricingProfiles,
      publication: {
        ...publication,
        state:
          row.lifecycleStatus === 'published'
            ? 'published'
            : row.lifecycleStatus === 'deprecated' ||
                row.lifecycleStatus === 'archived'
              ? 'retired'
              : publication.state,
      },
    };
  }

  private publication(row: any) {
    return evaluatePublicationReadiness({
      mode: row.mode,
      pricingStatus: row.pricingStatus,
      pricingProfiles: serializePricingProfiles(row.pricing ?? []),
    });
  }

  @Post(':canonicalId/enabled')
  @Scopes('catalog:write')
  async setModelEnabled(
    @Param('canonicalId') canonicalId: string,
    @Body('enabled', ParseBoolPipe) enabled: boolean,
    @Req() request: AuthenticatedRequest,
  ) {
    const decodedId = decodeURIComponent(canonicalId);
    const model = await this.prisma.$transaction(async (tx) => {
      const before = await tx.model.findUniqueOrThrow({
        where: { canonicalId: decodedId },
      });
      const model = await tx.model.update({
        where: { canonicalId: decodedId },
        data: { enabled },
      });
      await tx.auditLog.create({
        data: {
          entityType: 'model',
          entityId: model.id,
          action: 'set_enabled',
          beforeJson: before as any,
          afterJson: model as any,
          ...auditActor(request),
        },
      });
      return model;
    });
    await this.indexing.enqueueModelIndex(model.id);
    return model;
  }

  @Get('series')
  @Scopes('catalog:read')
  listSeries() {
    return this.prisma.modelSeries.findMany({
      orderBy: { name: 'asc' },
    });
  }

  @Post('series/upsert')
  @Scopes('catalog:write')
  upsertSeries(
    @Body() body: UpsertModelSeriesDto,
    @Req() request: AuthenticatedRequest,
  ) {
    return this.prisma
      .$transaction(async (tx) => {
        const before = await tx.modelSeries.findUnique({
          where: { slug: body.slug },
        });
        const series = await tx.modelSeries.upsert({
          where: { slug: body.slug },
          create: body,
          update: { name: body.name },
        });
        await tx.auditLog.create({
          data: {
            entityType: 'model_series',
            entityId: series.id,
            action: before ? 'update' : 'create',
            beforeJson: before as any,
            afterJson: series as any,
            ...auditActor(request),
          },
        });
        return series;
      })
      .then(async (series) => {
        await this.indexing.enqueueModelSeriesReindex(series.slug);
        return series;
      });
  }

  @Post('series/:slug/logo')
  @Scopes('catalog:write')
  @UseInterceptors(FileInterceptor('file'))
  async uploadSeriesLogo(
    @Param('slug') slug: string,
    @UploadedFile() file: any,
    @Req() request: AuthenticatedRequest,
  ) {
    if (!file) throw new BadRequestException('file is required');
    if (!file.mimetype.startsWith('image/')) {
      throw new BadRequestException('only image files are allowed');
    }
    if (file.size > 2 * 1024 * 1024) {
      throw new BadRequestException('file size must be <= 2MB');
    }
    const uploadsDir = join(process.cwd(), 'uploads', 'model-series-logos');
    mkdirSync(uploadsDir, { recursive: true });
    const ext = extname(file.originalname || '').toLowerCase() || '.png';
    const fileName = `${slug}-${randomUUID()}${ext}`;
    writeFileSync(join(uploadsDir, fileName), file.buffer);
    const logoUrl = `/api/v1/providers/assets/model-series-logos/${fileName}`;
    return this.prisma
      .$transaction(async (tx) => {
        const series = await tx.modelSeries.update({
          where: { slug },
          data: { logoUrl },
        });
        await tx.auditLog.create({
          data: {
            entityType: 'model_series',
            entityId: series.id,
            action: 'logo_update',
            afterJson: series as any,
            ...auditActor(request),
          },
        });
        return series;
      })
      .then(async (series) => {
        await this.indexing.enqueueModelSeriesReindex(series.slug);
        return series;
      });
  }
}
