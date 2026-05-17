import { Global, Module } from '@nestjs/common';
import { ConfigService } from '@nestjs/config';
import Typesense from 'typesense';

export const TYPESENSE_CLIENT = 'TYPESENSE_CLIENT';

@Global()
@Module({
  providers: [
    {
      provide: TYPESENSE_CLIENT,
      inject: [ConfigService],
      useFactory: (c: ConfigService) =>
        new Typesense.Client({
          nodes: [
            {
              host: c.get<string>('TYPESENSE_HOST', '127.0.0.1'),
              port: Number(c.get('TYPESENSE_PORT', 8108)),
              protocol: c.get<'http' | 'https'>('TYPESENSE_PROTOCOL', 'http'),
            },
          ],
          apiKey: c.get<string>('TYPESENSE_API_KEY', 'xyz'),
        }),
    },
  ],
  exports: [TYPESENSE_CLIENT],
})
export class TypesenseModule {}
