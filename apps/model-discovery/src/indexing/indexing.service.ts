import { InjectQueue } from '@nestjs/bullmq';
import { Injectable } from '@nestjs/common';
import { Queue } from 'bullmq';
export const INDEXING_QUEUE = 'indexing';
@Injectable()
export class IndexingService {
  constructor(@InjectQueue(INDEXING_QUEUE) private readonly queue: Queue) {}
  enqueue(modelId: string) {
    return this.queue.add('index-model', { modelId });
  }
}
