import 'reflect-metadata';
import { mkdirSync, writeFileSync } from 'node:fs';
import { join, resolve } from 'node:path';
import { NestFactory } from '@nestjs/core';
import { stringify } from 'yaml';
import { applyCommonAppConfig } from '../src/app.setup';
import { OpenAPIAppModule } from '../src/openapi.module';
import { buildOpenAPIDocument } from '../src/openapi';

async function main() {
  const app = await NestFactory.create(OpenAPIAppModule, { logger: false });
  applyCommonAppConfig(app);
  const document = buildOpenAPIDocument(app);
  const outDir = resolve(process.cwd(), 'openapi');

  mkdirSync(outDir, { recursive: true });
  writeFileSync(join(outDir, 'model-discovery.openapi.json'), JSON.stringify(document, null, 2) + '\n', 'utf8');
  writeFileSync(join(outDir, 'model-discovery.openapi.yaml'), stringify(document), 'utf8');

  await app.close();
}

void main();
