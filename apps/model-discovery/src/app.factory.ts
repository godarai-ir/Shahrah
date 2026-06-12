import { INestApplication } from '@nestjs/common';
import { NestFactory } from '@nestjs/core';
import { AppModule } from './app.module';
import { applyCommonAppConfig, setupSwaggerDocs } from './app.setup';

export async function createConfiguredApp(): Promise<INestApplication> {
  const app = await NestFactory.create(AppModule);
  applyCommonAppConfig(app);
  setupSwaggerDocs(app);

  return app;
}
