import {
  Body,
  Controller,
  Delete,
  Get,
  Param,
  Patch,
  Post,
  Query,
  UseGuards,
} from '@nestjs/common';
import { ApiBearerAuth, ApiTags } from '@nestjs/swagger';
import { Roles } from '../common/decorators/roles.decorator';
import { Role } from '../common/enums/roles.enum';
import { JwtAuthGuard } from '../common/guards/jwt-auth.guard';
import { RolesGuard } from '../common/guards/roles.guard';
import { CreateModelDto } from './dto/create-model.dto';
import { QueryModelDto } from './dto/query-model.dto';
import { UpdateModelDto } from './dto/update-model.dto';
import { ModelsService } from './models.service';

@ApiTags('models')
@Controller({ path: 'models', version: '1' })
export class ModelsController {
  constructor(private readonly service: ModelsService) {}
  @Get('public') findAll(@Query() query: QueryModelDto) {
    return this.service.findAll(query);
  }
  @Get('public/:id') findOne(@Param('id') id: string) {
    return this.service.findOne(id);
  }
  @ApiBearerAuth()
  @UseGuards(JwtAuthGuard, RolesGuard)
  @Roles(Role.Admin)
  @Post('admin')
  create(@Body() dto: CreateModelDto) {
    return this.service.create(dto);
  }
  @ApiBearerAuth()
  @UseGuards(JwtAuthGuard, RolesGuard)
  @Roles(Role.Admin)
  @Patch('admin/:id')
  update(@Param('id') id: string, @Body() dto: UpdateModelDto) {
    return this.service.update(id, dto);
  }
  @ApiBearerAuth()
  @UseGuards(JwtAuthGuard, RolesGuard)
  @Roles(Role.Admin)
  @Patch('admin/:id/publish')
  publish(@Param('id') id: string) {
    return this.service.publish(id);
  }
  @ApiBearerAuth()
  @UseGuards(JwtAuthGuard, RolesGuard)
  @Roles(Role.Admin)
  @Patch('admin/:id/deprecate')
  deprecate(@Param('id') id: string) {
    return this.service.deprecate(id);
  }
  @ApiBearerAuth()
  @UseGuards(JwtAuthGuard, RolesGuard)
  @Roles(Role.Admin)
  @Delete('admin/:id')
  remove(@Param('id') id: string) {
    return this.service.remove(id);
  }
}
