import { randomBytes, scrypt as scryptCallback, timingSafeEqual } from 'crypto';
const N = 65_536;
const R = 8;
const P = 1;
const KEY_LENGTH = 64;
const MAX_MEMORY = 128 * 1024 * 1024;
const DUMMY_HASH = `scrypt$${N}$${R}$${P}$${Buffer.alloc(16).toString('base64url')}$${Buffer.alloc(KEY_LENGTH).toString('base64url')}`;

function derive(password: string, salt: Buffer, keyLength: number, n: number, r: number, p: number): Promise<Buffer> {
  return new Promise((resolve, reject) => {
    scryptCallback(password, salt, keyLength, { N: n, r, p, maxmem: MAX_MEMORY }, (error, value) => {
      if (error) reject(error);
      else resolve(value);
    });
  });
}

export async function hashPassword(password: string): Promise<string> {
  const salt = randomBytes(16);
  const derived = await derive(password, salt, KEY_LENGTH, N, R, P);
  return `scrypt$${N}$${R}$${P}$${salt.toString('base64url')}$${derived.toString('base64url')}`;
}

export async function verifyPassword(password: string, encoded = DUMMY_HASH): Promise<boolean> {
  try {
    const [algorithm, n, r, p, saltValue, hashValue] = encoded.split('$');
    if (algorithm !== 'scrypt' || !n || !r || !p || !saltValue || !hashValue) return false;
    const expected = Buffer.from(hashValue, 'base64url');
    const actual = await derive(
      password,
      Buffer.from(saltValue, 'base64url'),
      expected.length,
      Number(n),
      Number(r),
      Number(p),
    );
    return actual.length === expected.length && timingSafeEqual(actual, expected);
  } catch {
    return false;
  }
}
