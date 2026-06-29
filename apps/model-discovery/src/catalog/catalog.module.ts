import { Module } from '@nestjs/common';
import { IndexingModule } from '../indexing/indexing.module';
import { CatalogAdminController } from './catalog-admin.controller';
import { CatalogController } from './catalog.controller';
import { CatalogService } from './catalog.service';

@Module({
  imports: [IndexingModule],
  controllers: [CatalogController, CatalogAdminController],
  providers: [CatalogService],
  exports: [CatalogService],
})
export class CatalogModule {}
