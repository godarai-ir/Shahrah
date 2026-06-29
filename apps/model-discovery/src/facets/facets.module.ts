import { Module } from '@nestjs/common';
import { IndexingModule } from '../indexing/indexing.module';
import { FacetsController } from './facets.controller';
import { FacetsService } from './facets.service';

@Module({
  imports: [IndexingModule],
  controllers: [FacetsController],
  providers: [FacetsService],
})
export class FacetsModule {}
