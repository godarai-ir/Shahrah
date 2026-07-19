import sharp from 'sharp';
import {
  isAllowedProviderLogoMimeType,
  processProviderLogoImage,
  PROVIDER_LOGO_CONTENT_TYPE,
  PROVIDER_LOGO_SIZE,
} from './provider-logo-image';

describe('provider logo image processing', () => {
  it('accepts PNG, JPEG, and WebP upload MIME types', () => {
    expect(isAllowedProviderLogoMimeType('image/png')).toBe(true);
    expect(isAllowedProviderLogoMimeType('image/jpeg')).toBe(true);
    expect(isAllowedProviderLogoMimeType('image/webp')).toBe(true);
    expect(isAllowedProviderLogoMimeType('image/gif')).toBe(false);
    expect(isAllowedProviderLogoMimeType('text/plain')).toBe(false);
  });

  it('resizes a non-square logo to a transparent 64x64 PNG without cropping', async () => {
    const input = await sharp({
      create: {
        width: 128,
        height: 32,
        channels: 4,
        background: { r: 12, g: 120, b: 220, alpha: 1 },
      },
    })
      .png()
      .toBuffer();

    const processed = await processProviderLogoImage(input);
    const metadata = await sharp(processed.data).metadata();

    expect(processed.contentType).toBe(PROVIDER_LOGO_CONTENT_TYPE);
    expect(processed.width).toBe(PROVIDER_LOGO_SIZE);
    expect(processed.height).toBe(PROVIDER_LOGO_SIZE);
    expect(processed.sizeBytes).toBe(processed.data.byteLength);
    expect(processed.contentHash).toMatch(/^[a-f0-9]{64}$/);
    expect(metadata.format).toBe('png');
    expect(metadata.width).toBe(PROVIDER_LOGO_SIZE);
    expect(metadata.height).toBe(PROVIDER_LOGO_SIZE);
    expect(metadata.hasAlpha).toBe(true);
  });

  it('rejects unsupported image formats even when they are valid images', async () => {
    const gif = Buffer.from(
      'R0lGODlhAQABAIAAAP///wAAACH5BAEAAAAALAAAAAABAAEAAAICRAEAOw==',
      'base64',
    );

    await expect(processProviderLogoImage(gif)).rejects.toThrow(
      'only PNG, JPEG, and WebP images are allowed',
    );
  });
});
