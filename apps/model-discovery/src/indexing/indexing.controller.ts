import { Body, Controller, Post, UseGuards } from '@nestjs/common';
import { ApiBearerAuth, ApiTags } from '@nestjs/swagger';
import { Roles } from '../common/decorators/roles.decorator';
import { Role } from '../common/enums/roles.enum';
import { JwtAuthGuard } from '../common/guards/jwt-auth.guard';
import { RolesGuard } from '../common/guards/roles.guard';
import { IndexingService } from './indexing.service';

class IndexOneDto {
  modelId!: string;
}

@ApiTags('indexing')
@ApiBearerAuth()
@UseGuards(JwtAuthGuard, RolesGuard)
@Roles(Role.admin)
@Controller({ path: 'admin/indexing', version: '1' })
export class IndexingController {
  constructor(private readonly indexing: IndexingService) {}

  @Post('rebuild') rebuild() {
    return this.indexing.enqueueFullReindex();
  }
  @Post('model') one(@Body() dto: IndexOneDto) {
    return this.indexing.enqueueModelIndex(dto.modelId);
  }
}
