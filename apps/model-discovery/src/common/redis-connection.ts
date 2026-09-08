import type { ConfigService } from '@nestjs/config';

/**
 * BullMQ/ioredis connection options. Mirrors the subset of ioredis'
 * `RedisOptions` used here so this module has no direct ioredis dependency.
 */
export type RedisConnectionOptions = {
  host?: string;
  port?: number;
  sentinels?: { host: string; port: number }[];
  name?: string;
  password?: string;
  sentinelPassword?: string;
  maxRetriesPerRequest: null;
  connectTimeout: number;
  retryStrategy: (attempt: number) => number;
};

const BASE_OPTIONS: Pick<
  RedisConnectionOptions,
  'maxRetriesPerRequest' | 'connectTimeout' | 'retryStrategy'
> = {
  // BullMQ workers are long-lived and must survive transient DNS or
  // Redis outages instead of exhausting ioredis' default 20 retries.
  maxRetriesPerRequest: null,
  connectTimeout: 10_000,
  retryStrategy: (attempt: number) => Math.min(attempt * 200, 5_000),
};

/**
 * Parses a comma-separated "host:port,host:port" list, e.g. the
 * REDIS_SENTINEL_ADDRS env var shared with the Go control-plane service.
 */
export function parseSentinelAddrs(
  raw: string | undefined,
): { host: string; port: number }[] {
  if (!raw) return [];

  return raw
    .split(',')
    .map((entry) => entry.trim())
    .filter(Boolean)
    .map((entry) => {
      const separatorIndex = entry.lastIndexOf(':');
      if (separatorIndex === -1) {
        throw new Error(
          `Invalid REDIS_SENTINEL_ADDRS entry (expected host:port): ${entry}`,
        );
      }
      const host = entry.slice(0, separatorIndex);
      const port = Number(entry.slice(separatorIndex + 1));
      if (!host || !Number.isFinite(port)) {
        throw new Error(
          `Invalid REDIS_SENTINEL_ADDRS entry (expected host:port): ${entry}`,
        );
      }
      return { host, port };
    });
}

/**
 * Builds the ioredis/BullMQ connection options for Redis. When
 * REDIS_SENTINEL_ADDRS is configured, connects through Sentinel and follows
 * master failover (matching the control-plane Go service); otherwise falls
 * back to a direct REDIS_HOST/REDIS_PORT connection for local development
 * and environments without Sentinel.
 */
export function buildRedisConnectionOptions(
  cfg: ConfigService,
): RedisConnectionOptions {
  const sentinels = parseSentinelAddrs(cfg.get<string>('REDIS_SENTINEL_ADDRS'));
  const password = cfg.get<string>('REDIS_PASSWORD') || undefined;

  if (sentinels.length > 0) {
    const masterName = cfg.get<string>('REDIS_MASTER_NAME', 'mymaster');
    return {
      ...BASE_OPTIONS,
      sentinels,
      name: masterName,
      password,
      sentinelPassword: cfg.get<string>('REDIS_SENTINEL_PASSWORD') || password,
    };
  }

  return {
    ...BASE_OPTIONS,
    host: cfg.get<string>('REDIS_HOST', '127.0.0.1'),
    port: Number(cfg.get<number>('REDIS_PORT', 6379)),
    password,
  };
}
