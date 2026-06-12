import type { INestApplication } from '@nestjs/common';
import { DocumentBuilder, SwaggerModule } from '@nestjs/swagger';

export function buildOpenAPIDocument(app: INestApplication) {
  const config = new DocumentBuilder()
    .setTitle('Model Discovery API')
    .setDescription('Public and admin APIs for provider discovery, model catalog, and display settings.')
    .setVersion('1.0.0')
    .addBearerAuth()
    .build();

  return SwaggerModule.createDocument(app, config);
}

