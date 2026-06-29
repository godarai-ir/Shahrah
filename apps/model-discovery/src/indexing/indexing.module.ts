import { BullModule } from '@nestjs/bullmq';
import { Module } from '@nestjs/common';
import { IndexingController } from './indexing.controller';
import { IndexingProcessor } from './indexing.processor';
import { IndexingService, INDEXING_QUEUE } from './indexing.service';

@Module({
  imports: [BullModule.registerQueue({ name: INDEXING_QUEUE })],
  controllers: [IndexingController],
  providers: [IndexingService, IndexingProcessor],
  exports: [IndexingService],
})
export class IndexingModule {}
