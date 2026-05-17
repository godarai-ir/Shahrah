import { Processor, WorkerHost } from '@nestjs/bullmq';
import { Job } from 'bullmq';
import { INDEXING_QUEUE, IndexingService } from './indexing.service';

@Processor(INDEXING_QUEUE)
export class IndexingProcessor extends WorkerHost {
  constructor(private readonly indexing: IndexingService) {
    super();
  }
  async process(job: Job<{ modelId?: string }>) {
    if (job.name === 'full-reindex') return this.indexing.reindexAll();
    if (job.name === 'index-model' && job.data.modelId)
      return this.indexing.indexOneModel(job.data.modelId);
    return null;
  }
}
