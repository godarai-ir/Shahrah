import { Body, Controller, Post, UseGuards } from '@nestjs/common';
import { ApiBearerAuth, ApiTags } from '@nestjs/swagger';
import { z } from 'zod';
import { Roles } from '../common/decorators/roles.decorator';
import { Role } from '../common/enums/roles.enum';
import { JwtAuthGuard } from '../common/guards/jwt-auth.guard';
import { RolesGuard } from '../common/guards/roles.guard';
import { PrismaService } from '../database/prisma.service';
import { IndexingService } from '../indexing/indexing.service';

const pricingSchema = z.object({
  currency: z.string().optional(),
  inputPricePerMillionTokens: z.number().optional(),
  outputPricePerMillionTokens: z.number().optional(),
  cachedInputPricePerMillionTokens: z.number().optional(),
  imageInputPrice: z.number().optional(),
  imageOutputPrice: z.number().optional(),
  audioInputPrice: z.number().optional(),
  audioOutputPrice: z.number().optional(),
  requestPrice: z.number().optional(),
  region: z.string().optional(),
  sourceUrl: z.string().optional(),
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
  free: z.boolean().optional(),
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
@UseGuards(JwtAuthGuard, RolesGuard)
@Roles(Role.admin)
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
      create: { slug, name: slug, directionType: 'both' },
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
  async importCatalog(@Body() body: unknown) {
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
          free: row.free ?? false,
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
          free: row.free,
          categoryId,
        },
      });

      await this.prisma.$transaction([
        this.prisma.modelPricing.deleteMany({ where: { modelId: model.id } }),
        this.prisma.modelCapability.deleteMany({ where: { modelId: model.id } }),
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
        this.prisma.sourceReference.deleteMany({ where: { modelId: model.id } }),
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
          let regionId: string | undefined;
          if (pricing.region) {
            const region = await this.ensureRegion(pricing.region);
            regionId = region.id;
          }

          await this.prisma.modelPricing.create({
            data: {
              modelId: model.id,
              regionId,
              currency: pricing.currency || 'USD',
              inputPricePerMillionTokens: pricing.inputPricePerMillionTokens,
              outputPricePerMillionTokens:
                pricing.outputPricePerMillionTokens,
              cachedInputPricePerMillionTokens:
                pricing.cachedInputPricePerMillionTokens,
              imageInputPrice: pricing.imageInputPrice,
              imageOutputPrice: pricing.imageOutputPrice,
              audioInputPrice: pricing.audioInputPrice,
              audioOutputPrice: pricing.audioOutputPrice,
              requestPrice: pricing.requestPrice,
              sourceUrl: pricing.sourceUrl,
              effectiveFrom: new Date(),
            },
          });
        }
      }

      await this.prisma.modelChangelog.create({
        data: {
          modelId: model.id,
          changeType: 'import',
          title: 'Imported via JSON',
          createdBy: 'admin',
        },
      });
      await this.prisma.auditLog.create({
        data: {
          entityType: 'model',
          entityId: model.id,
          action: 'import',
          afterJson: model as any,
          actorId: 'admin',
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
}
