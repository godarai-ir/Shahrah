import { Controller, Get, Param, Query } from '@nestjs/common';
import { ApiTags } from '@nestjs/swagger';
import { ApiOkResponse } from '@nestjs/swagger';
import { CatalogService } from './catalog.service';
import { ModelsQueryDto } from './dto/models-query.dto';
import { RuntimeSnapshotDto } from './dto/runtime-snapshot.dto';

@ApiTags('models')
@Controller({ path: 'models', version: '1' })
export class CatalogController {
  constructor(private readonly catalog: CatalogService) {}

  @Get()
  list(@Query() query: ModelsQueryDto) {
    return this.catalog.listModels(query);
  }

  @Get('compare')
  compare(@Query('ids') ids: string) {
    return this.catalog.compare(
      (ids || '')
        .split(',')
        .map((s) => s.trim())
        .filter(Boolean),
    );
  }

  @Get('recent')
  recent() {
    return this.catalog.recent();
  }

  @Get('free')
  free() {
    return this.catalog.freeModels();
  }

  @Get('deprecated')
  deprecated() {
    return this.catalog.deprecatedModels();
  }

  @Get('runtime-snapshot/:providerSlug')
  @ApiOkResponse({ type: RuntimeSnapshotDto })
  runtimeSnapshot(@Param('providerSlug') providerSlug: string) {
    return this.catalog.runtimeSnapshot(providerSlug);
  }

  @Get(':canonicalId')
  getOne(@Param('canonicalId') canonicalId: string) {
    return this.catalog.getModelByCanonicalId(decodeURIComponent(canonicalId));
  }
}
