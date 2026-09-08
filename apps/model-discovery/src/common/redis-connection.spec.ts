import { ConfigService } from '@nestjs/config';
import {
  buildRedisConnectionOptions,
  parseSentinelAddrs,
} from './redis-connection';

function configServiceFrom(
  env: Record<string, string | undefined>,
): ConfigService {
  return {
    get: <T>(key: string, defaultValue?: T) =>
      (env[key] as unknown as T) ?? defaultValue,
  } as unknown as ConfigService;
}

describe('parseSentinelAddrs', () => {
  it('returns an empty list for undefined/empty input', () => {
    expect(parseSentinelAddrs(undefined)).toEqual([]);
    expect(parseSentinelAddrs('')).toEqual([]);
  });

  it('parses a comma-separated host:port list, trimming whitespace', () => {
    expect(
      parseSentinelAddrs(
        'redis-node-0.example:26379, redis-node-1.example:26379,redis-node-2.example:26379',
      ),
    ).toEqual([
      { host: 'redis-node-0.example', port: 26379 },
      { host: 'redis-node-1.example', port: 26379 },
      { host: 'redis-node-2.example', port: 26379 },
    ]);
  });

  it('throws on a malformed entry', () => {
    expect(() => parseSentinelAddrs('redis-node-0.example')).toThrow(
      /Invalid REDIS_SENTINEL_ADDRS/,
    );
  });
});

describe('buildRedisConnectionOptions', () => {
  it('falls back to a direct REDIS_HOST/REDIS_PORT connection when no sentinels are configured', () => {
    const cfg = configServiceFrom({
      REDIS_HOST: 'redis.example.local',
      REDIS_PORT: '6380',
      REDIS_PASSWORD: 'secret',
    });

    const options = buildRedisConnectionOptions(cfg);

    expect(options).toMatchObject({
      host: 'redis.example.local',
      port: 6380,
      password: 'secret',
    });
    expect(options.sentinels).toBeUndefined();
  });

  it('uses Sentinel when REDIS_SENTINEL_ADDRS is configured', () => {
    const cfg = configServiceFrom({
      REDIS_SENTINEL_ADDRS: 'sentinel-0.example:26379,sentinel-1.example:26379',
      REDIS_MASTER_NAME: 'mymaster',
      REDIS_PASSWORD: 'redis-pass',
      REDIS_SENTINEL_PASSWORD: 'sentinel-pass',
    });

    const options = buildRedisConnectionOptions(cfg);

    expect(options).toMatchObject({
      sentinels: [
        { host: 'sentinel-0.example', port: 26379 },
        { host: 'sentinel-1.example', port: 26379 },
      ],
      name: 'mymaster',
      password: 'redis-pass',
      sentinelPassword: 'sentinel-pass',
    });
    expect(options.host).toBeUndefined();
  });

  it('defaults sentinelPassword to the redis password when unset', () => {
    const cfg = configServiceFrom({
      REDIS_SENTINEL_ADDRS: 'sentinel-0.example:26379',
      REDIS_PASSWORD: 'shared-secret',
    });

    const options = buildRedisConnectionOptions(cfg);

    expect(options.password).toBe('shared-secret');
    expect(options.sentinelPassword).toBe('shared-secret');
    expect(options.name).toBe('mymaster');
  });
});
