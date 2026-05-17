import { Body, Controller, Get, Post, UseGuards } from '@nestjs/common';
import { ApiBearerAuth, ApiTags } from '@nestjs/swagger';
import { PrismaService } from '../database/prisma.service';
import { Roles } from '../common/decorators/roles.decorator';
import { Role } from '../common/enums/roles.enum';
import { JwtAuthGuard } from '../common/guards/jwt-auth.guard';
import { RolesGuard } from '../common/guards/roles.guard';
import { UpsertModelDto } from './dto/upsert-model.dto';

@ApiTags('admin')
@ApiBearerAuth()
@UseGuards(JwtAuthGuard, RolesGuard)
@Roles(Role.admin)
@Controller({ path: 'admin/models', version: '1' })
export class CatalogAdminController {
  constructor(private readonly prisma: PrismaService) {}

  @Post('upsert')
  async upsert(@Body() dto: UpsertModelDto) {
    const provider = await this.prisma.provider.findUniqueOrThrow({
      where: { slug: dto.providerSlug },
    });
    const category = dto.categorySlug
      ? await this.prisma.category.findUnique({
          where: { slug: dto.categorySlug },
        })
      : null;

    const before = await this.prisma.model.findUnique({
      where: { canonicalId: dto.canonicalId },
    });
    const model = await this.prisma.model.upsert({
      where: { canonicalId: dto.canonicalId },
      create: {
        providerId: provider.id,
        canonicalId: dto.canonicalId,
        slug: dto.slug,
        displayName: dto.displayName,
        shortDescription: dto.shortDescription,
        longDescription: dto.longDescription,
        status: (dto.status as any) || 'active',
        lifecycleStatus: (dto.lifecycleStatus as any) || 'draft',
        categoryId: category?.id,
        contextWindow: dto.contextWindow,
        maxOutputTokens: dto.maxOutputTokens,
        openSource: dto.openSource ?? false,
        free: dto.free ?? false,
      },
      update: {
        providerId: provider.id,
        slug: dto.slug,
        displayName: dto.displayName,
        shortDescription: dto.shortDescription,
        longDescription: dto.longDescription,
        status: (dto.status as any) || undefined,
        lifecycleStatus: (dto.lifecycleStatus as any) || undefined,
        categoryId: category?.id,
        contextWindow: dto.contextWindow,
        maxOutputTokens: dto.maxOutputTokens,
        openSource: dto.openSource,
        free: dto.free,
      },
    });

    if (dto.pricing) {
      await this.prisma.modelPricing.create({
        data: {
          modelId: model.id,
          currency: dto.pricing.currency || 'USD',
          inputPricePerMillionTokens: dto.pricing.inputPricePerMillionTokens,
          outputPricePerMillionTokens: dto.pricing.outputPricePerMillionTokens,
          cachedInputPricePerMillionTokens:
            dto.pricing.cachedInputPricePerMillionTokens,
          sourceUrl: dto.pricing.sourceUrl,
          effectiveFrom: new Date(),
        },
      });
    }

    await this.prisma.auditLog.create({
      data: {
        entityType: 'model',
        entityId: model.id,
        action: before ? 'update' : 'create',
        beforeJson: before as any,
        afterJson: model as any,
        actorId: 'admin',
      },
    });

    return model;
  }

  @Get('audit-log')
  audit() {
    return this.prisma.auditLog.findMany({
      orderBy: { createdAt: 'desc' },
      take: 200,
    });
  }
}
