import { Body, Controller, Post, Req, UseGuards } from '@nestjs/common';
import { ApiBearerAuth, ApiTags } from '@nestjs/swagger';
import { z } from 'zod';
import {
  auditActor,
  type AuthenticatedRequest,
} from '../common/authenticated-request';
import { Scopes } from '../common/decorators/scopes.decorator';
import { ScopesGuard } from '../common/guards/scopes.guard';
import { ServiceAccountGuard } from '../common/guards/service-account.guard';
import { PrismaService } from '../database/prisma.service';
import { IndexingService } from '../indexing/indexing.service';

const pricingRateSchema = z.object({
  key: z.string().optional(),
  usage: z.enum(['input', 'output', 'cache_read', 'cache_write']),
  unit: z.string().min(1),
  amount: z
    .union([
      z.string().regex(/^\d+(?:\.\d+)?$/),
      z.number().finite().nonnegative(),
    ])
    .transform(String),
  quantity: z.number().int().positive().default(1),
  serviceTier: z.enum(['standard', 'batch', 'priority']).default('standard'),
  resolution: z.string().optional(),
  tierFrom: z.number().int().nonnegative().optional(),
  tierTo: z.number().int().nonnegative().optional(),
  tierMode: z.string().optional(),
  qualifiers: z
    .record(z.string(), z.union([z.string(), z.number(), z.boolean()]))
    .optional(),
});

const pricingSchema = z.object({
  region: z.string().default('global'),
  currency: z.string().optional(),
  sourceUrl: z.string().optional(),
  effectiveFrom: z.string().datetime().optional(),
  effectiveTo: z.string().datetime().optional(),
  rates: z.array(pricingRateSchema),
});

const providerSchema = z.object({
  slug: z.string(),
  name: z.string(),
  status: z.string().optional(),
  websiteUrl: z.string().optional(),
  docsUrl: z.string().optional(),
  supportsByok: z.boolean().optional(),
});
const modelSchema = z.object({
  providerSlug: z.string(),
  canonicalId: z.string(),
  slug: z.string(),
  displayName: z.string(),
  status: z.string().optional(),
  lifecycleStatus: z.string().optional(),
  shortDescription: z.string().optional(),
  contextWindow: z.number().int().optional(),
  maxOutputTokens: z.number().int().optional(),
  openSource: z.boolean().optional(),
  pricingStatus: z.enum(['unknown', 'priced', 'free']).optional(),
  mode: z.array(z.string()).optional(),
  categorySlug: z.string().optional(),
  inputModalities: z.array(z.string()).optional(),
  outputModalities: z.array(z.string()).optional(),
  supportedParameters: z.array(z.string()).optional(),
  capabilities: z.array(z.string()).optional(),
  regions: z.array(z.string()).optional(),
  sources: z.array(z.string()).optional(),
  pricing: z.union([pricingSchema, z.array(pricingSchema)]).optional(),
});
const importSchema = z.object({
  providers: z.array(providerSchema),
  models: z.array(modelSchema),
});

@ApiTags('import')
@ApiBearerAuth()
@UseGuards(ServiceAccountGuard, ScopesGuard)
@Scopes('import:execute')
@Controller({ path: 'admin/import', version: '1' })
export class ImportController {
  constructor(
    private readonly prisma: PrismaService,
    private readonly indexing: IndexingService,
  ) {}
  private async ensureCapability(slug: string) {
    return this.prisma.capability.upsert({
      where: { slug },
      create: { slug, name: slug },
      update: {},
    });
  }

  private async ensureModality(slug: string) {
    return this.prisma.modality.upsert({
      where: { slug },
      create: { slug, name: slug },
      update: {},
    });
  }

  private async ensureSupportedParameter(slug: string) {
    return this.prisma.supportedParameter.upsert({
      where: { slug },
      create: { slug, name: slug },
      update: {},
    });
  }

  private async ensureRegion(slug: string) {
    return this.prisma.region.upsert({
      where: { slug },
      create: { slug, name: slug },
      update: {},
    });
  }

  @Post('catalog-json')
  async importCatalog(
    @Body() body: unknown,
    @Req() request: AuthenticatedRequest,
  ) {
    const parsed = importSchema.safeParse(body);
    if (!parsed.success) {
      return {
        ok: false,
        errors: parsed.error.issues.map((i) => ({
          path: i.path.join('.'),
          message: i.message,
        })),
      };
    }

    for (const p of parsed.data.providers) {
      await this.prisma.provider.upsert({
        where: { slug: p.slug },
        create: { ...p, status: (p.status as any) || 'active' },
        update: { ...p, status: (p.status as any) || undefined } as any,
      });
    }

    for (const row of parsed.data.models) {
      const provider = await this.prisma.provider.findUniqueOrThrow({
        where: { slug: row.providerSlug },
      });
      let categoryId: string | undefined;
      if (row.categorySlug) {
        const cat = await this.prisma.category.upsert({
          where: { slug: row.categorySlug },
          create: { slug: row.categorySlug, name: row.categorySlug },
          update: {},
        });
        categoryId = cat.id;
      }

      const model = await this.prisma.model.upsert({
        where: { canonicalId: row.canonicalId },
        create: {
          providerId: provider.id,
          canonicalId: row.canonicalId,
          slug: row.slug,
          displayName: row.displayName,
          status: (row.status as any) || 'active',
          lifecycleStatus: (row.lifecycleStatus as any) || 'published',
          shortDescription: row.shortDescription,
          contextWindow: row.contextWindow,
          maxOutputTokens: row.maxOutputTokens,
          openSource: row.openSource ?? false,
          pricingStatus: this.resolvePricingStatus(row),
          mode: row.mode?.length ? row.mode : ['unknown'],
          categoryId,
        },
        update: {
          providerId: provider.id,
          slug: row.slug,
          displayName: row.displayName,
          status: (row.status as any) || undefined,
          lifecycleStatus: (row.lifecycleStatus as any) || undefined,
          shortDescription: row.shortDescription,
          contextWindow: row.contextWindow,
          maxOutputTokens: row.maxOutputTokens,
          openSource: row.openSource,
          pricingStatus:
            row.pricing !== undefined || row.pricingStatus !== undefined
              ? this.resolvePricingStatus(row)
              : undefined,
          mode: row.mode?.length ? row.mode : undefined,
          categoryId,
        },
      });

      await this.prisma.$transaction([
        this.prisma.modelPricing.deleteMany({ where: { modelId: model.id } }),
        this.prisma.modelCapability.deleteMany({
          where: { modelId: model.id },
        }),
        this.prisma.modelInputModality.deleteMany({
          where: { modelId: model.id },
        }),
        this.prisma.modelOutputModality.deleteMany({
          where: { modelId: model.id },
        }),
        this.prisma.modelSupportedParameter.deleteMany({
          where: { modelId: model.id },
        }),
        this.prisma.modelRegion.deleteMany({ where: { modelId: model.id } }),
        this.prisma.sourceReference.deleteMany({
          where: { modelId: model.id },
        }),
      ]);

      for (const c of row.capabilities || []) {
        const capability = await this.ensureCapability(c);
        await this.prisma.modelCapability.create({
          data: { modelId: model.id, capabilityId: capability.id },
        });
      }

      for (const m of row.inputModalities || []) {
        const modality = await this.ensureModality(m);
        await this.prisma.modelInputModality.create({
          data: { modelId: model.id, modalityId: modality.id },
        });
      }

      for (const m of row.outputModalities || []) {
        const modality = await this.ensureModality(m);
        await this.prisma.modelOutputModality.create({
          data: { modelId: model.id, modalityId: modality.id },
        });
      }

      for (const p of row.supportedParameters || []) {
        const param = await this.ensureSupportedParameter(p);
        await this.prisma.modelSupportedParameter.create({
          data: { modelId: model.id, supportedParameterId: param.id },
        });
      }

      for (const r of row.regions || []) {
        const region = await this.ensureRegion(r);
        await this.prisma.modelRegion.create({
          data: { modelId: model.id, regionId: region.id },
        });
      }

      for (const src of row.sources || []) {
        await this.prisma.sourceReference.create({
          data: {
            modelId: model.id,
            title: 'Source',
            url: src,
            sourceType: 'import_json',
          },
        });
      }

      if (row.pricing) {
        const pricingRows = Array.isArray(row.pricing)
          ? row.pricing
          : [row.pricing];
        for (const pricing of pricingRows) {
          const region = await this.ensureRegion(
            this.normalizeRegion(pricing.region),
          );

          await this.prisma.modelPricing.create({
            data: {
              modelId: model.id,
              regionId: region.id,
              currency: pricing.currency || 'USD',
              sourceUrl: pricing.sourceUrl,
              effectiveFrom: pricing.effectiveFrom
                ? new Date(pricing.effectiveFrom)
                : new Date(),
              effectiveTo: pricing.effectiveTo
                ? new Date(pricing.effectiveTo)
                : undefined,
              rates: {
                create: pricing.rates.map((rate) => ({
                  rateKey: rate.key || this.rateKey(rate),
                  usage: rate.usage,
                  unit: rate.unit,
                  amount: rate.amount,
                  quantity: rate.quantity,
                  serviceTier: rate.serviceTier,
                  resolution: rate.resolution,
                  tierFrom: rate.tierFrom,
                  tierTo: rate.tierTo,
                  tierMode: rate.tierMode,
                  qualifiers: rate.qualifiers as any,
                })),
              },
            },
          });
        }
      }

      await this.prisma.modelChangelog.create({
        data: {
          modelId: model.id,
          changeType: 'import',
          title: 'Imported via JSON',
          createdBy: request.user.actorId ?? request.user.serviceAccountName,
        },
      });
      await this.prisma.auditLog.create({
        data: {
          entityType: 'model',
          entityId: model.id,
          action: 'import',
          afterJson: model as any,
          ...auditActor(request),
        },
      });
    }

    await this.indexing.enqueueFullReindex();
    return {
      ok: true,
      importedProviders: parsed.data.providers.length,
      importedModels: parsed.data.models.length,
    };
  }

  private resolvePricingStatus(
    row: z.infer<typeof modelSchema>,
  ): 'unknown' | 'priced' | 'free' {
    if (row.pricingStatus === 'free') return 'free';
    if (!row.pricing) return row.pricingStatus ?? 'unknown';

    const profiles = Array.isArray(row.pricing) ? row.pricing : [row.pricing];
    const hasPositiveRate = profiles.some((profile) =>
      profile.rates.some((rate) => {
        const amount = Number(rate.amount);
        return Number.isFinite(amount) && amount > 0;
      }),
    );
    return hasPositiveRate ? 'priced' : 'unknown';
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

  private rateKey(rate: {
    usage: string;
    unit: string;
    serviceTier: string;
    resolution?: string;
    tierFrom?: number;
  }) {
    return [
      rate.usage,
      rate.unit,
      rate.serviceTier,
      `resolution=${rate.resolution ?? 'default'}`,
      `tier=${rate.tierFrom ?? 'base'}`,
    ].join(':');
  }
}
