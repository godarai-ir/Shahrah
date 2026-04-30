// Script to download Vazirmatn font files for self-hosting
// Run: node scripts/download-fonts.js

const https = require('https');
const fs = require('fs');
const path = require('path');

const FONT_DIR = path.join(__dirname, '..', 'apps', 'website', 'public', 'fonts', 'vazirmatn');

const FONTS = [
  { weight: '100', file: 'Vazirmatn-Thin.woff2' },
  { weight: '200', file: 'Vazirmatn-ExtraLight.woff2' },
  { weight: '300', file: 'Vazirmatn-Light.woff2' },
  { weight: '400', file: 'Vazirmatn-Regular.woff2' },
  { weight: '500', file: 'Vazirmatn-Medium.woff2' },
  { weight: '600', file: 'Vazirmatn-SemiBold.woff2' },
  { weight: '700', file: 'Vazirmatn-Bold.woff2' },
  { weight: '800', file: 'Vazirmatn-ExtraBold.woff2' },
  { weight: '900', file: 'Vazirmatn-Black.woff2' },
];

const CDN_BASE = 'https://cdn.jsdelivr.net/gh/rastikerdar/vazirmatn@v33.003/fonts/webfonts/';

function downloadFile(url, dest) {
  return new Promise((resolve, reject) => {
    https.get(url, (res) => {
      if (res.statusCode === 302 || res.statusCode === 301) {
        downloadFile(res.headers.location, dest).then(resolve, reject);
        return;
      }
      if (res.statusCode !== 200) {
        reject(new Error(`Failed to download ${url}: ${res.statusCode}`));
        return;
      }
      const file = fs.createWriteStream(dest);
      res.pipe(file);
      file.on('finish', () => {
        file.close();
        resolve();
      });
    }).on('error', reject);
  });
}

async function main() {
  if (!fs.existsSync(FONT_DIR)) {
    fs.mkdirSync(FONT_DIR, { recursive: true });
  }

  for (const font of FONTS) {
    const url = CDN_BASE + font.file;
    const dest = path.join(FONT_DIR, font.file);
    
    if (fs.existsSync(dest)) {
      console.log(`Skipping ${font.file} (already exists)`);
      continue;
    }

    console.log(`Downloading ${font.file}...`);
    try {
      await downloadFile(url, dest);
      console.log(`Downloaded ${font.file}`);
    } catch (err) {
      console.error(`Failed to download ${font.file}: ${err.message}`);
    }
  }

  console.log('Done!');
}

main();
