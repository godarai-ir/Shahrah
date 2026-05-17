import { Controller, Get, Query } from '@nestjs/common';
import { ApiTags } from '@nestjs/swagger';
import { SearchModelsDto } from './dto/search-models.dto';
import { SearchService } from './search.service';
@ApiTags('search')
@Controller({ path: 'public/search', version: '1' })
export class SearchController {
  constructor(private readonly service: SearchService) {}
  @Get('models') searchModels(@Query() dto: SearchModelsDto) {
    return this.service.searchModels(dto);
  }
}
