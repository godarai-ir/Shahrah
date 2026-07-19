import { BadRequestException } from '@nestjs/common';
import { createHash } from 'crypto';
import sharp from 'sharp';

export const PROVIDER_LOGO_SIZE = 64;
export const PROVIDER_LOGO_CONTENT_TYPE = 'image/png';
export const PROVIDER_LOGO_CACHE_CONTROL =
  'public, max-age=31536000, immutable';

const ALLOWED_INPUT_FORMATS = new Set(['png', 'jpeg', 'webp']);
const ALLOWED_MIME_TYPES = new Set(['image/png', 'image/jpeg', 'image/webp']);

export type ProcessedProviderLogo = {
  data: Buffer;
  contentHash: string;
  contentType: typeof PROVIDER_LOGO_CONTENT_TYPE;
  width: typeof PROVIDER_LOGO_SIZE;
  height: typeof PROVIDER_LOGO_SIZE;
  sizeBytes: number;
};

export function isAllowedProviderLogoMimeType(mimeType?: string): boolean {
  return ALLOWED_MIME_TYPES.has((mimeType ?? '').trim().toLowerCase());
}

export async function processProviderLogoImage(
  input: Buffer,
): Promise<ProcessedProviderLogo> {
  let metadata: sharp.Metadata;
  try {
    metadata = await sharp(input, {
      animated: false,
      failOn: 'error',
      limitInputPixels: 16_777_216,
    }).metadata();
  } catch {
    throw new BadRequestException('invalid image file');
  }

  if (!metadata.format || !ALLOWED_INPUT_FORMATS.has(metadata.format)) {
    throw new BadRequestException(
      'only PNG, JPEG, and WebP images are allowed',
    );
  }

  const data = await sharp(input, {
    animated: false,
    failOn: 'error',
    limitInputPixels: 16_777_216,
  })
    .resize(PROVIDER_LOGO_SIZE, PROVIDER_LOGO_SIZE, {
      fit: 'contain',
      background: { r: 0, g: 0, b: 0, alpha: 0 },
    })
    .png({ compressionLevel: 9, adaptiveFiltering: true })
    .toBuffer();

  return {
    data,
    contentHash: createHash('sha256').update(data).digest('hex'),
    contentType: PROVIDER_LOGO_CONTENT_TYPE,
    width: PROVIDER_LOGO_SIZE,
    height: PROVIDER_LOGO_SIZE,
    sizeBytes: data.byteLength,
  };
}
