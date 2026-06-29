import { Body, Controller, Get, Put, Req, UseGuards } from '@nestjs/common';
import { ApiBearerAuth, ApiTags } from '@nestjs/swagger';
import { Scopes } from '../common/decorators/scopes.decorator';
import { ScopesGuard } from '../common/guards/scopes.guard';
import { ServiceAccountGuard } from '../common/guards/service-account.guard';
import { UpdateDisplaySettingsDto } from './dto/update-display-settings.dto';
import { DisplaySettingsService } from './display-settings.service';
import {
  auditActor,
  type AuthenticatedRequest,
} from '../common/authenticated-request';
import { PrismaService } from '../database/prisma.service';

@ApiTags('settings')
@Controller({ path: 'settings/display', version: '1' })
export class DisplaySettingsController {
  constructor(
    private readonly service: DisplaySettingsService,
    private readonly prisma: PrismaService,
  ) {}

  @Get()
  getDisplaySettings() {
    return this.service.getDisplaySettings();
  }

  @ApiBearerAuth()
  @UseGuards(ServiceAccountGuard, ScopesGuard)
  @Scopes('display-settings:write')
  @Put()
  async updateDisplaySettings(
    @Body() body: UpdateDisplaySettingsDto,
    @Req() request: AuthenticatedRequest,
  ) {
    const before = await this.service.getDisplaySettings();
    const result = await this.service.updateDisplaySettings(body);
    await this.prisma.auditLog.create({
      data: {
        entityType: 'display_settings',
        entityId: 'default',
        action: 'update',
        beforeJson: before as any,
        afterJson: result as any,
        ...auditActor(request),
      },
    });
    return result;
  }
}
