import { Inject, Injectable, Logger } from '@nestjs/common';
import { ConfigService } from '@nestjs/config';
import { TYPESENSE_CLIENT } from '../typesense/typesense.module';
import { SearchModelsDto } from './dto/search-models.dto';

type TypesenseSearchClient = {
  collections: (name: string) => {
    documents: () => {
      search: (params: {
        q: string;
        query_by: string;
        filter_by?: string;
      }) => Promise<unknown>;
    };
  };
};

@Injectable()
export class SearchService {
  private readonly logger = new Logger(SearchService.name);
  constructor(
    @Inject(TYPESENSE_CLIENT) private readonly client: TypesenseSearchClient,
    private readonly cfg: ConfigService,
  ) {}
  async searchModels(dto: SearchModelsDto): Promise<unknown> {
    try {
      const collection = this.cfg.get<string>(
        'TYPESENSE_COLLECTION_MODELS',
        'models',
      );
      return await this.client.collections(collection).documents().search({
        q: dto.q,
        query_by: 'name,slug,description',
        filter_by: dto.filterBy,
      });
    } catch (e) {
      this.logger.warn((e as Error).message);
      return { hits: [], found: 0 };
    }
  }
}
