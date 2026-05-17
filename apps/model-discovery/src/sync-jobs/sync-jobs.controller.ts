import { Body, Controller, Post, UseGuards } from '@nestjs/common';
import { ApiBearerAuth, ApiTags } from '@nestjs/swagger';
import { Roles } from '../common/decorators/roles.decorator';
import { Role } from '../common/enums/roles.enum';
import { JwtAuthGuard } from '../common/guards/jwt-auth.guard';
import { RolesGuard } from '../common/guards/roles.guard';
import { SyncJobsService } from './sync-jobs.service';
class SyncDto {
  providerSlug!: string;
}
@ApiTags('sync-jobs')
@ApiBearerAuth()
@UseGuards(JwtAuthGuard, RolesGuard)
@Roles(Role.Admin)
@Controller({ path: 'admin/sync-jobs', version: '1' })
export class SyncJobsController {
  constructor(private readonly service: SyncJobsService) {}
  @Post('providers') enqueue(@Body() dto: SyncDto) {
    return this.service.enqueue(dto.providerSlug);
  }
}
