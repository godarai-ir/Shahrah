import {
  BadRequestException,
  Body,
  Controller,
  Get,
  Param,
  ParseBoolPipe,
  Post,
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
import { Roles } from '../common/decorators/roles.decorator';
import { Role } from '../common/enums/roles.enum';
import { JwtAuthGuard } from '../common/guards/jwt-auth.guard';
import { RolesGuard } from '../common/guards/roles.guard';
import { UpsertModelDto } from './dto/upsert-model.dto';
import { UpsertModelSeriesDto } from './dto/upsert-model-series.dto';

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
    const modelSeries = dto.modelSeriesSlug
      ? await this.prisma.modelSeries.findUnique({
          where: { slug: dto.modelSeriesSlug },
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
        enabled: dto.enabled ?? true,
        lifecycleStatus: (dto.lifecycleStatus as any) || 'draft',
        modelSeriesId: modelSeries?.id,
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
        enabled: dto.enabled,
        lifecycleStatus: (dto.lifecycleStatus as any) || undefined,
        modelSeriesId: modelSeries?.id,
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

  @Get()
  listAll() {
    return this.prisma.model.findMany({
      orderBy: [{ providerId: 'asc' }, { displayName: 'asc' }],
      include: { provider: true, modelSeries: true },
      take: 500,
    });
  }

  @Post(':canonicalId/enabled')
  async setModelEnabled(
    @Param('canonicalId') canonicalId: string,
    @Body('enabled', ParseBoolPipe) enabled: boolean,
  ) {
    return this.prisma.model.update({
      where: { canonicalId: decodeURIComponent(canonicalId) },
      data: { enabled },
    });
  }

  @Get('series')
  listSeries() {
    return this.prisma.modelSeries.findMany({
      orderBy: { name: 'asc' },
    });
  }

  @Post('series/upsert')
  upsertSeries(@Body() body: UpsertModelSeriesDto) {
    return this.prisma.modelSeries.upsert({
      where: { slug: body.slug },
      create: body,
      update: { name: body.name },
    });
  }

  @Post('series/:slug/logo')
  @UseInterceptors(FileInterceptor('file'))
  uploadSeriesLogo(
    @Param('slug') slug: string,
    @UploadedFile() file?: any,
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
    return this.prisma.modelSeries.update({
      where: { slug },
      data: { logoUrl },
    });
  }
}
