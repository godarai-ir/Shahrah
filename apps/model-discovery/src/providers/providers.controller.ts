import {
  BadRequestException,
  Body,
  Controller,
  Get,
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
import { UpsertProviderConnectionProfileDto } from './dto/upsert-provider-connection-profile.dto';

@ApiTags('providers')
@Controller({ path: 'providers', version: '1' })
export class ProvidersController {
  constructor(private readonly providers: ProvidersService) {}

  @Get()
  list() {
    return this.providers.list();
  }

  @Get(':slug')
  detail(@Param('slug') slug: string) {
    return this.providers.detail(slug);
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
  @Post('admin/connection-profile/upsert')
  upsertConnectionProfile(
    @Body() body: UpsertProviderConnectionProfileDto,
    @Req() request: AuthenticatedRequest,
  ) {
    return this.providers.upsertConnectionProfile(body, request.user);
  }

  @ApiBearerAuth()
  @UseGuards(ServiceAccountGuard, ScopesGuard)
  @Scopes('catalog:read')
  @Get('admin/:slug/connection-profile')
  getConnectionProfile(@Param('slug') slug: string) {
    return this.providers.getConnectionProfile(slug);
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
    if (!file.mimetype.startsWith('image/')) {
      throw new BadRequestException('only image files are allowed');
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
    const base = resolve(process.cwd(), 'uploads', bucket);
    const filePath = resolve(join(base, fileName));
    if (!filePath.startsWith(base) || !existsSync(filePath)) {
      return res.status(404).send('Not found');
    }
    return res.sendFile(filePath);
  }
}
