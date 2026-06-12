import { createConfiguredApp } from './app.factory';

async function bootstrap() {
  const app = await createConfiguredApp();
  await app.listen(process.env.PORT ?? 3002);
}

void bootstrap();
