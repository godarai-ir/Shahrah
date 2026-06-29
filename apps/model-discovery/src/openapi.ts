import type { INestApplication } from '@nestjs/common';
import { DocumentBuilder, SwaggerModule } from '@nestjs/swagger';

export function buildOpenAPIDocument(app: INestApplication) {
  const config = new DocumentBuilder()
    .setTitle('Model Discovery API')
    .setDescription(
      'Public catalog APIs and service-account-protected administration APIs for model discovery.',
    )
    .setVersion('1.0.0')
    .addBearerAuth({
      type: 'http',
      scheme: 'bearer',
      bearerFormat: 'ServiceKey',
      description:
        'Opaque service credential in the form sa_<key-id>.<secret>.',
    })
    .build();

  return SwaggerModule.createDocument(app, config);
}
