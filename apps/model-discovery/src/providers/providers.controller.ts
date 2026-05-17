import {
  Body,
  Controller,
  Delete,
  Get,
  Param,
  Patch,
  Post,
  UseGuards,
} from '@nestjs/common';
import { ApiBearerAuth, ApiTags } from '@nestjs/swagger';
import { Roles } from '../common/decorators/roles.decorator';
import { Role } from '../common/enums/roles.enum';
import { JwtAuthGuard } from '../common/guards/jwt-auth.guard';
import { RolesGuard } from '../common/guards/roles.guard';
import { CreateProviderDto } from './dto/create-provider.dto';
import { UpdateProviderDto } from './dto/update-provider.dto';
import { ProvidersService } from './providers.service';

@ApiTags('providers')
@Controller({ path: 'providers', version: '1' })
export class ProvidersController {
  constructor(private readonly service: ProvidersService) {}
  @Get('public') findAll() {
    return this.service.findAll();
  }
  @ApiBearerAuth()
  @UseGuards(JwtAuthGuard, RolesGuard)
  @Roles(Role.Admin)
  @Post('admin')
  create(@Body() dto: CreateProviderDto) {
    return this.service.create(dto);
  }
  @ApiBearerAuth()
  @UseGuards(JwtAuthGuard, RolesGuard)
  @Roles(Role.Admin)
  @Patch('admin/:id')
  update(@Param('id') id: string, @Body() dto: UpdateProviderDto) {
    return this.service.update(id, dto);
  }
  @ApiBearerAuth()
  @UseGuards(JwtAuthGuard, RolesGuard)
  @Roles(Role.Admin)
  @Delete('admin/:id')
  remove(@Param('id') id: string) {
    return this.service.remove(id);
  }
}
