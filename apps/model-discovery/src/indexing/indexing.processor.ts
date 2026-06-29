import { Processor, WorkerHost } from '@nestjs/bullmq';
import { Job } from 'bullmq';
import { INDEXING_QUEUE, IndexingService } from './indexing.service';

@Processor(INDEXING_QUEUE)
export class IndexingProcessor extends WorkerHost {
  constructor(private readonly indexing: IndexingService) {
    super();
  }
  async process(
    job: Job<{
      modelId?: string;
      providerSlug?: string;
      modelSeriesSlug?: string;
    }>,
  ) {
    if (job.name === 'full-reindex') return this.indexing.reindexAll();
    if (job.name === 'index-model' && job.data.modelId)
      return this.indexing.indexOneModel(job.data.modelId);
    if (
      job.name === 'index-provider-models' &&
      typeof job.data.providerSlug === 'string'
    ) {
      return this.indexing.reindexProviderModels(job.data.providerSlug);
    }
    if (
      job.name === 'index-model-series' &&
      typeof job.data.modelSeriesSlug === 'string'
    ) {
      return this.indexing.reindexModelSeries(job.data.modelSeriesSlug);
    }
    return null;
  }
}
