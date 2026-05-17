import { BullModule } from '@nestjs/bullmq';
import { Module } from '@nestjs/common';
import { CatalogModule } from '../catalog/catalog.module';
import { IndexingController } from './indexing.controller';
import { IndexingProcessor } from './indexing.processor';
import { IndexingService, INDEXING_QUEUE } from './indexing.service';

@Module({
  imports: [BullModule.registerQueue({ name: INDEXING_QUEUE }), CatalogModule],
  controllers: [IndexingController],
  providers: [IndexingService, IndexingProcessor],
  exports: [IndexingService],
})
export class IndexingModule {}
