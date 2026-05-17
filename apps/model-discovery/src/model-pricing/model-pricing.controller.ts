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
import { CreateModelPricingDto } from './dto/create-model-pricing.dto';
import { UpdateModelPricingDto } from './dto/update-model-pricing.dto';
import { ModelPricingService } from './model-pricing.service';

@ApiTags('model-pricing')
@Controller({ path: 'model-pricing', version: '1' })
export class ModelPricingController {
  constructor(private readonly service: ModelPricingService) {}
  @Get('public') findAll() {
    return this.service.findAll();
  }
  @ApiBearerAuth()
  @UseGuards(JwtAuthGuard, RolesGuard)
  @Roles(Role.Admin)
  @Post('admin')
  create(@Body() dto: CreateModelPricingDto) {
    return this.service.create(dto);
  }
  @ApiBearerAuth()
  @UseGuards(JwtAuthGuard, RolesGuard)
  @Roles(Role.Admin)
  @Patch('admin/:id')
  update(@Param('id') id: string, @Body() dto: UpdateModelPricingDto) {
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
