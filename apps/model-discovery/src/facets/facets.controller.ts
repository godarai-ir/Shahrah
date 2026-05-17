import { Controller, Get } from '@nestjs/common';
import { ApiTags } from '@nestjs/swagger';
import { FacetsService } from './facets.service';

@ApiTags('facets')
@Controller({ version: '1' })
export class FacetsController {
  constructor(private readonly facets: FacetsService) {}

  @Get('facets') all() {
    return this.facets.all();
  }
  @Get('categories') categories() {
    return this.facets.categories();
  }
  @Get('capabilities') capabilities() {
    return this.facets.capabilities();
  }
  @Get('supported-parameters') supportedParameters() {
    return this.facets.supportedParameters();
  }
  @Get('modalities') modalities() {
    return this.facets.modalities();
  }
  @Get('regions') regions() {
    return this.facets.regions();
  }
}
