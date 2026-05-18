import { Body, Controller, Get, Put, UseGuards } from '@nestjs/common';
import { ApiBearerAuth, ApiTags } from '@nestjs/swagger';
import { Roles } from '../common/decorators/roles.decorator';
import { Role } from '../common/enums/roles.enum';
import { JwtAuthGuard } from '../common/guards/jwt-auth.guard';
import { RolesGuard } from '../common/guards/roles.guard';
import { UpdateDisplaySettingsDto } from './dto/update-display-settings.dto';
import { DisplaySettingsService } from './display-settings.service';

@ApiTags('settings')
@Controller({ path: 'settings/display', version: '1' })
export class DisplaySettingsController {
  constructor(private readonly service: DisplaySettingsService) {}

  @Get()
  getDisplaySettings() {
    return this.service.getDisplaySettings();
  }

  @ApiBearerAuth()
  @UseGuards(JwtAuthGuard, RolesGuard)
  @Roles(Role.admin)
  @Put()
  updateDisplaySettings(@Body() body: UpdateDisplaySettingsDto) {
    return this.service.updateDisplaySettings(body);
  }
}
