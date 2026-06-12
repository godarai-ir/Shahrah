import { INestApplication, ValidationPipe, VersioningType } from '@nestjs/common';
import { SwaggerModule } from '@nestjs/swagger';
import { buildOpenAPIDocument } from './openapi';

export function applyCommonAppConfig(app: INestApplication): void {
  const corsOrigins = (process.env.CORS_ORIGIN ?? 'http://localhost:3000')
    .split(',')
    .map((origin) => origin.trim())
    .filter(Boolean);

  app.enableCors({
    origin: corsOrigins,
    methods: ['GET', 'HEAD', 'PUT', 'PATCH', 'POST', 'DELETE', 'OPTIONS'],
    allowedHeaders: ['Content-Type', 'Authorization'],
  });
  app.setGlobalPrefix('api');
  app.enableVersioning({ type: VersioningType.URI, defaultVersion: '1' });
  app.useGlobalPipes(
    new ValidationPipe({
      whitelist: true,
      transform: true,
      forbidNonWhitelisted: true,
    }),
  );
}

export function setupSwaggerDocs(app: INestApplication): void {
  const document = buildOpenAPIDocument(app);
  SwaggerModule.setup('docs', app, document, {
    jsonDocumentUrl: 'openapi.json',
    yamlDocumentUrl: 'openapi.yaml',
  });
}

