import {
  BadRequestException,
  Body,
  Controller,
  Get,
  GoneException,
  Param,
  ParseBoolPipe,
  Post,
  Req,
  Res,
  UploadedFile,
  UseGuards,
  UseInterceptors,
} from '@nestjs/common';
import { ApiBearerAuth, ApiTags } from '@nestjs/swagger';
import { FileInterceptor } from '@nestjs/platform-express';
import { existsSync } from 'fs';
import { join, resolve } from 'path';
import type { Response } from 'express';
import type { AuthenticatedRequest } from '../common/authenticated-request';
import { Scopes } from '../common/decorators/scopes.decorator';
import { ScopesGuard } from '../common/guards/scopes.guard';
import { ServiceAccountGuard } from '../common/guards/service-account.guard';
import { UpsertProviderDto } from './dto/upsert-provider.dto';
import { ProvidersService } from './providers.service';
import { UpsertProviderConnectionSchemaDto } from './dto/upsert-provider-connection-schema.dto';
import {
  isAllowedProviderLogoMimeType,
  PROVIDER_LOGO_CACHE_CONTROL,
} from './provider-logo-image';

@ApiTags('providers')
@Controller({ path: 'providers', version: '1' })
export class ProvidersController {
  constructor(private readonly providers: ProvidersService) {}

  @Get()
  list() {
    return this.providers.list();
  }

  @Get(':slug/models')
  listModels(@Param('slug') slug: string) {
    return this.providers.listModels(slug);
  }

  @Get(':slug/connection-schemas')
  listConnectionSchemas(@Param('slug') slug: string) {
    return this.providers.listConnectionSchemas(slug);
  }

  @ApiBearerAuth()
  @UseGuards(ServiceAccountGuard, ScopesGuard)
  @Scopes('catalog:read')
  @Get('admin/list')
  listAllForAdmin() {
    return this.providers.listAllForAdmin();
  }

  @ApiBearerAuth()
  @UseGuards(ServiceAccountGuard, ScopesGuard)
  @Scopes('provider:write')
  @Post('admin/upsert')
  upsert(
    @Body() body: UpsertProviderDto,
    @Req() request: AuthenticatedRequest,
  ) {
    return this.providers.upsert(body, request.user);
  }

  @ApiBearerAuth()
  @UseGuards(ServiceAccountGuard, ScopesGuard)
  @Scopes('provider:write')
  @Post('admin/enabled')
  setEnabled(
    @Body('slug') slug: string,
    @Body('enabled', ParseBoolPipe) enabled: boolean,
    @Req() request: AuthenticatedRequest,
  ) {
    return this.providers.updateEnabled(slug, enabled, request.user);
  }

  @ApiBearerAuth()
  @UseGuards(ServiceAccountGuard, ScopesGuard)
  @Scopes('provider:write')
  @Post('admin/connection-schemas/upsert')
  upsertConnectionSchema(
    @Body() body: UpsertProviderConnectionSchemaDto,
    @Req() request: AuthenticatedRequest,
  ) {
    return this.providers.upsertConnectionSchema(body, request.user);
  }

  @ApiBearerAuth()
  @UseGuards(ServiceAccountGuard, ScopesGuard)
  @Scopes('catalog:read')
  @Get('admin/:slug/connection-profile')
  getConnectionProfile(@Param('slug') slug: string) {
    throw new GoneException({
      error: {
        code: 'provider_connection_profile_retired',
        message:
          'Value-bearing provider connection profiles have been retired.',
        provider: slug,
      },
    });
  }

  @ApiBearerAuth()
  @UseGuards(ServiceAccountGuard, ScopesGuard)
  @Scopes('provider:write')
  @Post('admin/connection-profile/upsert')
  retiredConnectionProfileWrite() {
    throw new GoneException({
      error: {
        code: 'provider_connection_profile_retired',
        message:
          'Write credential values through the control-plane provider connection API.',
      },
    });
  }

  @Get(':slug')
  detail(@Param('slug') slug: string) {
    return this.providers.detail(slug);
  }

  @ApiBearerAuth()
  @UseGuards(ServiceAccountGuard, ScopesGuard)
  @Scopes('provider:write')
  @Post('admin/:slug/logo')
  @UseInterceptors(FileInterceptor('file'))
  uploadProviderLogo(
    @Param('slug') slug: string,
    @UploadedFile() file: any,
    @Req() request: AuthenticatedRequest,
  ) {
    if (!file) throw new BadRequestException('file is required');
    const rawMimeType = (file as { mimetype?: unknown }).mimetype;
    const mimeType = typeof rawMimeType === 'string' ? rawMimeType : undefined;
    if (!isAllowedProviderLogoMimeType(mimeType)) {
      throw new BadRequestException(
        'only PNG, JPEG, and WebP images are allowed',
      );
    }
    if (file.size > 2 * 1024 * 1024) {
      throw new BadRequestException('file size must be <= 2MB');
    }
    return this.providers.saveProviderLogo(slug, file, request.user);
  }

  @Get('assets/:bucket/:fileName')
  getAsset(
    @Param('bucket') bucket: string,
    @Param('fileName') fileName: string,
    @Res() res: Response,
  ) {
    if (bucket !== 'provider-logos' && bucket !== 'model-series-logos') {
      throw new BadRequestException('invalid bucket');
    }
    if (fileName.includes('..'))
      throw new BadRequestException('invalid file name');

    if (bucket === 'provider-logos') {
      return this.sendProviderLogoAsset(fileName, res);
    }

    const base = resolve(process.cwd(), 'uploads', bucket);
    const filePath = resolve(join(base, fileName));
    if (!filePath.startsWith(base) || !existsSync(filePath)) {
      return res.status(404).send('Not found');
    }
    res.setHeader('Cache-Control', PROVIDER_LOGO_CACHE_CONTROL);
    return res.sendFile(filePath);
  }

  private async sendProviderLogoAsset(fileName: string, res: Response) {
    const asset = await this.providers.getProviderLogoAsset(fileName);
    if (!asset) return res.status(404).send('Not found');

    res.setHeader('Content-Type', asset.contentType);
    res.setHeader('Cache-Control', PROVIDER_LOGO_CACHE_CONTROL);
    return res.status(200).send(asset.body);
  }
}
