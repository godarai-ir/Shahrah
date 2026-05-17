import { Body, Controller, Post, UseGuards } from '@nestjs/common';
import { ApiBearerAuth, ApiTags } from '@nestjs/swagger';
import { z } from 'zod';
import { Roles } from '../common/decorators/roles.decorator';
import { Role } from '../common/enums/roles.enum';
import { JwtAuthGuard } from '../common/guards/jwt-auth.guard';
import { RolesGuard } from '../common/guards/roles.guard';
import { PrismaService } from '../database/prisma.service';
import { IndexingService } from '../indexing/indexing.service';

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
  pricing: z
    .object({
      currency: z.string().optional(),
      inputPricePerMillionTokens: z.number().optional(),
      outputPricePerMillionTokens: z.number().optional(),
      cachedInputPricePerMillionTokens: z.number().optional(),
      sourceUrl: z.string().optional(),
    })
    .optional(),
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

      if (row.pricing) {
        await this.prisma.modelPricing.create({
          data: {
            modelId: model.id,
            currency: row.pricing.currency || 'USD',
            inputPricePerMillionTokens: row.pricing.inputPricePerMillionTokens,
            outputPricePerMillionTokens:
              row.pricing.outputPricePerMillionTokens,
            cachedInputPricePerMillionTokens:
              row.pricing.cachedInputPricePerMillionTokens,
            sourceUrl: row.pricing.sourceUrl,
            effectiveFrom: new Date(),
          },
        });
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
