import { Body, Controller, Post, Req, UseGuards } from '@nestjs/common';
import {
  ApiBearerAuth,
  ApiBody,
  ApiOkResponse,
  ApiProperty,
  ApiTags,
} from '@nestjs/swagger';
import { Scopes } from '../common/decorators/scopes.decorator';
import { ScopesGuard } from '../common/guards/scopes.guard';
import { ServiceAccountGuard } from '../common/guards/service-account.guard';
import { IndexingService } from './indexing.service';
import { auditActor, type AuthenticatedRequest } from '../common/authenticated-request';
import { PrismaService } from '../database/prisma.service';

class IndexOneDto {
  @ApiProperty()
  modelId!: string;
}

class IndexingResponse {
  @ApiProperty()
  ok!: boolean;
}

@ApiTags('indexing')
@ApiBearerAuth()
@UseGuards(ServiceAccountGuard, ScopesGuard)
@Scopes('indexing:execute')
@Controller({ path: 'admin/indexing', version: '1' })
export class IndexingController {
  constructor(
    private readonly indexing: IndexingService,
    private readonly prisma: PrismaService,
  ) {}

  @Post('rebuild')
  @ApiOkResponse({ type: IndexingResponse })
  async rebuild(@Req() request: AuthenticatedRequest) {
    const result = await this.indexing.enqueueFullReindex();
    await this.prisma.auditLog.create({ data: {
      entityType: 'index', entityId: 'models', action: 'rebuild',
      ...auditActor(request),
    } });
    return result;
  }

  @Post('model')
  @ApiBody({ type: IndexOneDto })
  @ApiOkResponse({ type: IndexingResponse })
  async one(@Body() dto: IndexOneDto, @Req() request: AuthenticatedRequest) {
    const result = await this.indexing.enqueueModelIndex(dto.modelId);
    await this.prisma.auditLog.create({ data: {
      entityType: 'index', entityId: dto.modelId, action: 'reindex_model',
      ...auditActor(request),
    } });
    return result;
  }
}
