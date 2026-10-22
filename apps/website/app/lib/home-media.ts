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

const videoSpec = 'ویدیو، بدون صدا، حلقه‌ای، ۸ تا ۱۲ ثانیه، MP4';
const gifSpec = 'گیف یا ویدیوی کوتاه حلقه‌ای، ۴ تا ۶ ثانیه';

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
  collabPhoto: {
    kind: 'image',
    aspect: '4/3',
    title: 'همکاری تیم و دستیار',
    spec: 'عکس یا گیف: چند همکار کنار هم و یک دستیار در حال کار، ۴:۳',
    alt: 'تیمی که همراه دستیار گدارAI کار می‌کند',
  },
  pillarContext: {
    kind: 'video',
    aspect: '4/3',
    title: 'دستیار با اطلاعات شرکت جواب می‌دهد',
    spec: `${videoSpec}. یک سؤال می‌پرسند و دستیار از CRM و اسناد جواب را پیدا می‌کند`,
    alt: 'دستیار از CRM و اسناد شرکت برای جواب دادن استفاده می‌کند',
  },
  pillarHuman: {
    kind: 'video',
    aspect: '4/3',
    title: 'پیش‌نویس و تأیید انسانی',
    spec: `${videoSpec}. دستیار پیش‌نویس می‌سازد و همکار تأیید یا ویرایش می‌کند`,
    alt: 'دستیار پیش‌نویس می‌سازد و یک همکار آن را تأیید می‌کند',
  },
  pillarModels: {
    kind: 'video',
    aspect: '4/3',
    title: 'جابه‌جایی بین مدل‌ها',
    spec: `${videoSpec}. انتخاب مدل برای یک دستیار و تعویض آن`,
    alt: 'تعویض مدل یک دستیار در گدارAI',
  },
  pillarControl: {
    kind: 'video',
    aspect: '4/3',
    title: 'سقف مصرف و گزارش رخدادها',
    spec: `${videoSpec}. تعیین سقف یک تیم و دیدن گزارش رخدادها`,
    alt: 'تعیین سقف مصرف تیم و مشاهده گزارش رخدادها',
  },
  teamSales: {
    kind: 'gif',
    aspect: '4/3',
    title: 'تیم فروش',
    spec: `${gifSpec}. خلاصه‌ی سابقه‌ی مشتری قبل از جلسه`,
    alt: 'دستیار فروش سابقه‌ی مشتری را خلاصه می‌کند',
  },
  teamSupport: {
    kind: 'gif',
    aspect: '4/3',
    title: 'تیم پشتیبانی',
    spec: `${gifSpec}. پیش‌نویس پاسخ یک تیکت`,
    alt: 'دستیار پشتیبانی پیش‌نویس پاسخ تیکت را می‌نویسد',
  },
  teamFinance: {
    kind: 'gif',
    aspect: '4/3',
    title: 'تیم مالی',
    spec: `${gifSpec}. تطبیق فاکتور و سفارش`,
    alt: 'دستیار مالی فاکتورها را با سفارش‌ها تطبیق می‌دهد',
  },
  teamLegal: {
    kind: 'gif',
    aspect: '4/3',
    title: 'تیم حقوقی',
    spec: `${gifSpec}. علامت‌زدن بندهای یک قرارداد`,
    alt: 'دستیار حقوقی بندهای قرارداد را علامت می‌زند',
  },
  teamHr: {
    kind: 'gif',
    aspect: '4/3',
    title: 'منابع انسانی',
    spec: `${gifSpec}. جواب یک سؤال درباره‌ی مرخصی`,
    alt: 'دستیار منابع انسانی به سؤال درباره‌ی مرخصی جواب می‌دهد',
  },
  teamEngineering: {
    kind: 'gif',
    aspect: '4/3',
    title: 'تیم فنی',
    spec: `${gifSpec}. ربط دادن تسک Jira با تغییر Git`,
    alt: 'دستیار فنی تسک Jira را با تغییر Git مرتبط می‌کند',
  },
  // Slots of the `/workplace` page.
  tourVideo: {
    kind: 'video',
    aspect: 'video',
    title: 'تور کوتاه محصول',
    spec: 'ویدیو ۱۶:۹، ۲ تا ۴ دقیقه، با زیرنویس فارسی. بدون صدا برای پیش‌نمایش حلقه‌ای و نسخه‌ی کامل با صدا برای پخش',
    alt: 'تور محصول گدارAI از ساختن فضای کاری تا اولین دستیار',
  },
  suiteChat: {
    kind: 'image',
    aspect: 'fill',
    title: 'گفتگو',
    spec: 'تصویر محصول: فضای گفتگوی سازمان با انتخاب مدل، قاب عمودی',
    alt: 'فضای گفتگوی گدارAI',
  },
  suiteAgents: {
    kind: 'image',
    aspect: 'fill',
    title: 'دستیارها',
    spec: 'تصویر محصول: صفحه‌ی ساخت یک دستیار با دستورالعمل و اتصال‌ها',
    alt: 'صفحه‌ی ساخت دستیار در گدارAI',
  },
  suiteConnections: {
    kind: 'image',
    aspect: 'fill',
    title: 'اتصال‌ها',
    spec: 'تصویر محصول: فهرست اتصال‌ها (CRM، پیام‌رسان‌ها، Jira، Notion، Git، MCP)',
    alt: 'فهرست اتصال‌های گدارAI',
  },
  suiteApi: {
    kind: 'image',
    aspect: 'fill',
    title: 'درگاه API',
    spec: 'تصویر: تکه‌کد اتصال با API سازگار با OpenAI',
    alt: 'نمونه‌کد اتصال به درگاه گدارAI',
  },
  suiteControl: {
    kind: 'image',
    aspect: 'fill',
    title: 'کنترل',
    spec: 'تصویر محصول: سقف مصرف تیم‌ها و گزارش رخدادها',
    alt: 'صفحه‌ی کنترل مصرف و گزارش رخدادها',
  },
  valModels: {
    kind: 'gif',
    src: '/home/val-models.gif',
    aspect: 'fill',
    title: 'انتخاب مدل',
    spec: 'تصویر یا گیف: فهرست مدل‌ها و جابه‌جایی بین آن‌ها',
    alt: 'انتخاب مدل در گدارAI',
  },
  valSecurity: {
    kind: 'image',
    aspect: 'fill',
    title: 'امنیت',
    spec: 'تصویر: گاردریل در حال گرفتن داده‌ی حساس',
    alt: 'گاردریل گدارAI داده‌ی حساس را می‌گیرد',
  },
  valCustom: {
    kind: 'image',
    aspect: 'fill',
    title: 'شخصی‌سازی',
    spec: 'تصویر: تنظیم دستورالعمل و دسترسی‌های یک دستیار',
    alt: 'تنظیم دستیار اختصاصی هر تیم',
  },
  valEnterprise: {
    kind: 'image',
    aspect: 'fill',
    title: 'آماده‌ی سازمان',
    spec: 'تصویر: سقف مصرف تیم‌ها و گزارش مالی تومانی',
    alt: 'مدیریت مصرف و گزارش مالی سازمان',
  },
  valDeploy: {
    kind: 'image',
    aspect: 'fill',
    title: 'استقرار',
    spec: 'تصویر: دو گزینه‌ی ابر گدارAI و سرور خودتان',
    alt: 'گزینه‌های استقرار گدارAI',
  },
  valConnect: {
    kind: 'image',
    aspect: 'fill',
    title: 'اتصال به ابزارها',
    spec: 'تصویر: نشان ابزارهای وصل‌شده کنار هم',
    alt: 'ابزارهای متصل به گدارAI',
  },
  guideHandbook: {
    kind: 'image',
    aspect: 'fill',
    title: 'راهنمای شروع',
    spec: 'تصویر جلد یا صفحه‌ی راهنمای شروع، قاب عمودی',
    alt: 'راهنمای شروع گدارAI',
  },
  guideTour: {
    kind: 'video',
    aspect: 'fill',
    title: 'گدارAI در عمل',
    spec: 'ویدیوی حلقه‌ای از محصول، ۱۰ تا ۱۵ ثانیه',
    alt: 'پیش‌نمایش محصول گدارAI',
  },
  storyOne: {
    kind: 'image',
    aspect: 'fill',
    title: 'تجربه‌ی مشتری ۱',
    spec: 'عکس یا ویدیوی کوتاه از مشتری، قاب عمودی',
    alt: 'تجربه‌ی یکی از مشتریان گدارAI',
  },
  storyTwo: {
    kind: 'image',
    aspect: 'fill',
    title: 'تجربه‌ی مشتری ۲',
    spec: 'عکس یا ویدیوی کوتاه از مشتری، قاب عمودی',
    alt: 'تجربه‌ی یکی از مشتریان گدارAI',
  },
  storyThree: {
    kind: 'image',
    aspect: 'fill',
    title: 'تجربه‌ی مشتری ۳',
    spec: 'عکس یا ویدیوی کوتاه از مشتری، قاب عمودی',
    alt: 'تجربه‌ی یکی از مشتریان گدارAI',
  },
} satisfies Record<string, MediaSpec>;

export type HomeMediaId = keyof typeof homeMedia;
