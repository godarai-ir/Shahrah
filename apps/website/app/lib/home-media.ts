// Media slots of the home page. Each slot renders a labelled placeholder until
// `src` is set. To fill a slot, drop the file under `public/home/` and set
// `src` (and `poster` for video) to its public path, for example
// `/home/hero-demo.mp4`. Nothing else in the page has to change.

export type MediaKind = 'video' | 'gif' | 'image';

export type MediaAspect = 'video' | '4/3' | '3/2' | '5/6' | 'square' | 'strip' | 'fill';

export type MediaSpec = {
  kind: MediaKind;
  aspect: MediaAspect;
  /** What the media should show; displayed inside the placeholder. */
  title: string;
  /** Format, ratio and length guidance for whoever produces the file. */
  spec: string;
  /** Alternative text for images and the accessible label for video. */
  alt: string;
  src?: string;
  poster?: string;
};

export const homeMedia = {
  heroDemo: {
    kind: 'video',
    title: 'تور کوتاه محصول',
    aspect: 'fill',
    spec: 'ویدیو یا گیف بزرگ و افقی، ۲۰ تا ۴۰ ثانیه، MP4، بدون صدا. از ساختن یک دستیار تا گرفتن جواب، با زیرنویس فارسی',
    alt: 'تور کوتاه محصول گدارAI: ساختن یک دستیار و اتصال آن به ابزارهای سازمان',
  },
  logos: {
    kind: 'image',
    aspect: 'strip',
    title: 'لوگوی مشتریان',
    spec: 'نوار لوگوی خاکستری تک‌رنگ، SVG، حداکثر ۱۰ لوگو',
    alt: 'لوگوی سازمان‌هایی که از گدارAI استفاده می‌کنند',
  },
  quotePortrait: {
    kind: 'image',
    aspect: '5/6',
    title: 'عکس نقل‌قول‌کننده',
    spec: 'عکس عمودی ۵:۶ از مشتری یا تیمش، همراه نام، سمت و شرکت',
    alt: 'تصویر یکی از مشتریان گدارAI',
  },
} satisfies Record<string, MediaSpec>;

export type HomeMediaId = keyof typeof homeMedia;
