import { Module } from '@nestjs/common';
import { IndexingModule } from '../indexing/indexing.module';
import { ImportController } from './import.controller';

@Module({ imports: [IndexingModule], controllers: [ImportController] })
export class ImportModule {}
