import Link from 'next/link';
import { ArrowLeft, BadgeCheck, CalendarClock, Check, ChevronDown, CreditCard, Layers, PowerOff, Users } from 'lucide-react';

type SeatDefinition = {
  slug: 'free' | 'pro' | 'max';
  name: string;
  description: string;
  credits: number;
  creditsLabel: string;
  features: string[];
  highlighted?: boolean;
};

const seatDefinitions: SeatDefinition[] = [
  {
    slug: 'free',
    name: 'رایگان',
    description: 'برای شروع کار و آشنایی اعضای تیم با GodarChat',
    credits: 500,
    creditsLabel: 'اعتبار یک‌باره برای هر عضو',
    features: ['تا ۵ عضو با صندلی رایگان در هر سازمان', 'همه‌ی امکانات پلتفرم', 'اعتبار شخصی فقط برای خود عضو', 'هر عضو فقط یک‌بار صندلی رایگان می‌گیرد'],
  },
  {
    slug: 'pro',
    name: 'Pro',
    description: 'برای اعضایی که هر روز با هوش مصنوعی کار می‌کنند',
    credits: 2000,
    creditsLabel: 'اعتبار شخصی در هر ماه',
    features: ['همه‌ی امکانات پلتفرم', 'اعتبار هر ماه از نو داده می‌شود', 'پس از اعتبار شخصی، مصرف از استخر سازمان تا سقف تعیین‌شده', 'ارتقای فوری به Max با اعتبار کامل ماه'],
    highlighted: true,
  },
  {
    slug: 'max',
    name: 'Max',
    description: 'برای تحلیل‌گران، تولید محتوا و مصرف سنگین',
    credits: 8000,
    creditsLabel: 'اعتبار شخصی در هر ماه',
    features: ['همه‌ی امکانات پلتفرم', 'چهار برابر اعتبار Pro با قیمت کمتر برای هر اعتبار', 'پس از اعتبار شخصی، مصرف از استخر سازمان تا سقف تعیین‌شده', 'کاهش به Pro در پایان دوره'],
  },
];

const principles = [
  { icon: BadgeCheck, title: 'همه‌ی امکانات برای همه', text: 'پلن و قابلیت قفل‌شده وجود ندارد؛ هر سازمان از روز اول به همه‌ی امکانات گدارAI دسترسی دارد.' },
  { icon: CreditCard, title: 'پرداخت پیش از استفاده', text: 'صندلی‌ها و بسته‌های اعتبار از درگاه پرداخت و پیش از استفاده پرداخت می‌شوند. کارت شما ذخیره و خودکار شارژ نمی‌شود.' },
  { icon: CalendarClock, title: 'بدون تمدید خودکار', text: 'اشتراک خودکار تمدید نمی‌شود. از ۷ روز مانده به پایان دوره تا روز آخر می‌توانید آن را تمدید کنید.' },
  { icon: PowerOff, title: 'پایان اشتراک، فقط صندلی‌های پولی', text: 'اگر اشتراک تمدید نشود فقط صندلی‌های پولی غیرفعال می‌شوند؛ صندلی‌های رایگان، اعتبار بسته‌ها و API کار می‌کنند.' },
] as const;

const subscriptionSteps = [
  { title: 'خرید اولین صندلی پولی', text: 'اشتراک همان پرداخت صندلی‌های پولی است. با خرید اولین صندلی Pro یا Max اشتراک شروع می‌شود و دوره‌ی آن، ماهانه یا سالانه، همان‌جا انتخاب می‌شود.' },
  { title: 'افزودن صندلی در میانه‌ی دوره', text: 'صندلی‌هایی که بعداً می‌خرید هم‌دوره با اشتراک‌اند و فقط به نسبت روزهای باقی‌مانده‌ی دوره پرداخت می‌شوند. صندلی بلافاصله پس از پرداخت فعال می‌شود.' },
  { title: 'تمدید در هفته‌ی آخر', text: 'از ۷ روز مانده به پایان دوره تا روز آخر، تمدید در پنل مدیریت سازمان فعال است. پرداخت از درگاه انجام می‌شود و هیچ مبلغی خودکار کسر نمی‌شود.' },
  { title: 'اگر تمدید نکنید', text: 'در پایان دوره فقط صندلی‌های پولی غیرفعال می‌شوند و اعضای آن‌ها بدون صندلی می‌مانند. صندلی‌های رایگان، اعتبار بسته‌ها و کلیدهای API مثل قبل کار می‌کنند.' },
] as const;

const pricingFaqs = [
  {
    question: 'پلن‌های قبلی (توسعه‌دهنده، تیم، رشد و سازمانی) چه شدند؟',
    answer: 'پلن‌ها حذف شده‌اند و همه‌ی امکانات برای همه‌ی سازمان‌ها در دسترس است. هزینه فقط از دو جا می‌آید: صندلی‌های پولی اعضا و بسته‌های اعتبار برای استخر سازمان. اشتراک‌های پولی فعلی بلافاصله به مدل جدید منتقل شده‌اند.',
  },
  {
    question: 'دوره‌ی آزمایشی دارید؟',
    answer: 'دوره‌ی آزمایشی جداگانه نداریم، چون نیازی به آن نیست: تا ۵ عضو هر سازمان صندلی رایگان با ۵۰۰ اعتبار می‌گیرند و می‌توانند بدون پرداخت با همه‌ی امکانات کار کنند.',
  },
  {
    question: 'قیمت صندلی Pro و Max چطور حساب می‌شود؟',
    answer: 'قیمت هر صندلی برابر است با اعتبار ماهانه‌ی آن × قیمت فهرست فعال هر اعتبار، منهای تخفیف همان صندلی. صورتحساب سالانه تخفیف بیشتری دارد ولی اعتبار همچنان ماه‌به‌ماه داده می‌شود. مالیات بر ارزش افزوده جداگانه اضافه می‌شود. قیمت هر اعتبار و تاریخچه‌ی تغییراتش در صفحه‌ی تغییرات قیمت منتشر می‌شود.',
  },
  {
    question: 'استخر سازمان و بسته‌های اعتبار برای چیست؟',
    answer: 'استخر سازمان اعتبار مشترک سازمان است و فقط با خرید بسته‌های اعتبار پر می‌شود. مصرف API، کلیدهای دسترسی، triggerها و گاردریل‌ها از همین استخر کم می‌شود. اعضای Pro و Max پس از تمام شدن اعتبار شخصی‌شان می‌توانند تا سقفی که مدیر تعیین می‌کند از استخر هم استفاده کنند.',
  },
  {
    question: 'اگر عضوی را حذف کنیم، صندلی پولی او چه می‌شود؟',
    answer: 'صندلی به انبار صندلی‌های سازمان برمی‌گردد و تا پایان دوره‌ی پرداخت‌شده به عضو دیگری قابل تخصیص است. صندلی‌ها بازپرداخت ریالی ندارند.',
  },
  {
    question: 'یادم برود تمدید کنم چه می‌شود؟',
    answer: 'از ۷ روز مانده به پایان دوره در پنل یادآوری می‌بینید. اگر تمدید نکنید فقط صندلی‌های پولی غیرفعال می‌شوند؛ هر وقت دوباره صندلی بخرید، اشتراک تازه‌ای با دوره‌ی جدید شروع می‌شود.',
  },
  {
    question: 'برای پرسش‌های فنی یا استقرار اختصاصی از کجا کمک بگیریم؟',
    answer: 'از صفحه‌ی تماس با ما یا ایمیل support@godarai.ir با تیم گدارAI در ارتباط باشید. برای استقرار در محل یا ابر خصوصی، بررسی نیازها به‌صورت اختصاصی انجام می‌شود.',
  },
] as const;

/**
 * Public seat pricing. Seat prices are not hard-coded: they follow the active credit list price,
 * which the page receives from the public pricing changelog when it is available.
 */
export function SeatPricing({ creditPriceToman }: { creditPriceToman: number | null }) {
  return (
    <main dir="rtl" className="overflow-hidden text-right">
      <section className="px-4 pb-16 pt-20 sm:px-6 sm:pt-24 lg:px-8">
        <div className="mx-auto max-w-7xl">
          <header className="mx-auto flex max-w-3xl flex-col items-center gap-5 text-center">
            <div className="inline-flex items-center gap-2 rounded-full border border-border bg-muted px-3 py-1.5 text-sm font-semibold text-foreground">
              <Users aria-hidden="true" className="size-4" />
              صندلی‌ها و بسته‌های اعتبار
            </div>
            <h1 className="text-4xl font-bold leading-[1.35] tracking-tight text-foreground sm:text-5xl">
              برای هر عضو یک صندلی، برای API استخر سازمان
            </h1>
            <p className="max-w-2xl text-base leading-8 text-muted-foreground sm:text-lg">
              پلنی برای انتخاب وجود ندارد. هر عضو با صندلی خودش اعتبار شخصی برای GodarChat دارد و مصرف API از استخر مشترک سازمان پرداخت می‌شود که با بسته‌های اعتبار پر می‌شود.
            </p>
          </header>

          <div className="mt-12 grid items-stretch gap-5 md:grid-cols-3">
            {seatDefinitions.map((seat) => (
              <SeatCard key={seat.slug} seat={seat} creditPriceToman={creditPriceToman} />
            ))}
          </div>

          <p className="mt-5 text-center text-sm leading-7 text-muted-foreground">
            {creditPriceToman === null
              ? 'قیمت فهرست هر اعتبار موقتاً در دسترس نیست؛ قیمت دقیق هر صندلی هنگام خرید در پنل مدیریت سازمان نمایش داده می‌شود.'
              : `قیمت فهرست فعلی هر اعتبار ${formatNumber(creditPriceToman)} تومان است. قیمت نهایی هر صندلی با تخفیف همان صندلی و مالیات، هنگام خرید در پنل مدیریت سازمان نمایش داده می‌شود.`}
          </p>
        </div>
      </section>

      <section aria-labelledby="pricing-principles" className="border-y border-border bg-muted/40 px-4 py-16 sm:px-6 lg:px-8">
        <div className="mx-auto max-w-7xl">
          <h2 id="pricing-principles" className="sr-only">قواعد پرداخت</h2>
          <div className="grid gap-5 sm:grid-cols-2 lg:grid-cols-4">
            {principles.map(({ icon: Icon, title, text }) => (
              <div key={title} className="flex flex-col gap-3 rounded-2xl border border-border bg-card p-6">
                <span className="grid size-10 place-items-center rounded-xl bg-muted text-foreground">
                  <Icon aria-hidden="true" className="size-5" />
                </span>
                <h3 className="font-bold text-foreground">{title}</h3>
                <p className="text-sm leading-7 text-muted-foreground">{text}</p>
              </div>
            ))}
          </div>
        </div>
      </section>

      <section aria-labelledby="credit-packs" className="px-4 py-16 sm:px-6 sm:py-20 lg:px-8">
        <div className="mx-auto grid max-w-7xl gap-10 lg:grid-cols-2 lg:items-center">
          <div className="flex flex-col gap-4">
            <p className="text-sm font-semibold text-muted-foreground">بسته‌های اعتبار</p>
            <h2 id="credit-packs" className="text-3xl font-bold leading-[1.45] text-foreground sm:text-4xl">استخر مشترک سازمان برای API</h2>
            <p className="leading-8 text-muted-foreground">
              بسته‌های اعتبار استخر مشترک سازمان را پر می‌کنند. مصرف API، کلیدهای دسترسی، triggerها و گاردریل‌ها فقط از این استخر کم می‌شود و هرگز اعتبار شخصی اعضا را مصرف نمی‌کند.
            </p>
          </div>
          <ul className="flex flex-col gap-4 rounded-2xl border border-border bg-card p-6 text-sm leading-7 text-foreground">
            {[
              'قیمت هر بسته از قیمت فهرست اعتبار ساخته می‌شود و بسته‌های بزرگ‌تر قیمت کمتری برای هر اعتبار دارند.',
              'خرید بسته به صندلی پولی یا اشتراک وابسته نیست و پیش از استفاده از درگاه پرداخت می‌شود.',
              'اعتبار بسته‌ها با پایان اشتراک صندلی‌ها از بین نمی‌رود و API به کار خود ادامه می‌دهد.',
              'اعضای Pro و Max پس از اعتبار شخصی، تا سقفی که مدیر تعیین می‌کند از استخر استفاده می‌کنند.',
            ].map((item) => (
              <li key={item} className="flex items-start gap-2.5">
                <Check aria-hidden="true" className="mt-1 size-4 shrink-0 text-muted-foreground" />
                <span>{item}</span>
              </li>
            ))}
          </ul>
        </div>
      </section>

      <section aria-labelledby="subscription" className="border-y border-border bg-muted/40 px-4 py-16 sm:px-6 sm:py-20 lg:px-8">
        <div className="mx-auto max-w-7xl">
          <div className="max-w-2xl">
            <p className="text-sm font-semibold text-muted-foreground">اشتراک</p>
            <h2 id="subscription" className="mt-2 text-3xl font-bold text-foreground sm:text-4xl">اشتراک یعنی پرداخت صندلی‌های پولی</h2>
            <p className="mt-3 leading-7 text-muted-foreground">تا وقتی صندلی پولی نخریده‌اید اشتراکی ندارید و چیزی پرداخت نمی‌کنید.</p>
          </div>
          <ol className="mt-8 grid gap-5 md:grid-cols-2 lg:grid-cols-4">
            {subscriptionSteps.map((step, index) => (
              <li key={step.title} className="flex flex-col gap-3 rounded-2xl border border-border bg-card p-6">
                <span className="grid size-8 place-items-center rounded-full bg-primary text-sm font-bold text-primary-foreground">{formatNumber(index + 1)}</span>
                <h3 className="font-bold text-foreground">{step.title}</h3>
                <p className="text-sm leading-7 text-muted-foreground">{step.text}</p>
              </li>
            ))}
          </ol>
        </div>
      </section>

      <section aria-labelledby="pricing-faq" className="px-4 py-16 sm:px-6 sm:py-20 lg:px-8">
        <div className="mx-auto grid max-w-7xl gap-10 lg:grid-cols-[0.72fr_1.28fr] lg:gap-16">
          <div className="lg:sticky lg:top-24 lg:self-start">
            <p className="text-sm font-semibold text-muted-foreground">پرسش‌های پرتکرار</p>
            <h2 id="pricing-faq" className="mt-2 text-3xl font-bold leading-[1.45] text-foreground sm:text-4xl">پیش از خرید صندلی بدانید</h2>
            <Link href="/contact" className="mt-6 inline-flex min-h-11 items-center gap-2 text-sm font-semibold text-foreground hover:underline">
              پاسخ پرسشتان را پیدا نکردید؟
              <ArrowLeft aria-hidden="true" className="size-4" />
            </Link>
          </div>

          <div className="overflow-hidden rounded-2xl border border-border bg-card px-5 sm:px-7">
            {pricingFaqs.map((item, index) => (
              <details key={item.question} className="group border-b border-border last:border-b-0" open={index === 0}>
                <summary className="flex min-h-20 cursor-pointer list-none items-center justify-between gap-5 py-5 text-right font-bold leading-7 text-foreground marker:content-none [&::-webkit-details-marker]:hidden">
                  <span>{item.question}</span>
                  <span className="grid size-8 shrink-0 place-items-center rounded-full bg-muted text-muted-foreground">
                    <ChevronDown aria-hidden="true" className="size-4 transition-transform duration-200 group-open:rotate-180" />
                  </span>
                </summary>
                <p className="max-w-3xl pb-6 pl-12 text-sm leading-8 text-muted-foreground sm:text-base">{item.answer}</p>
              </details>
            ))}
          </div>
        </div>
      </section>

      <section className="px-4 pb-20 sm:px-6 lg:px-8">
        <div className="mx-auto flex max-w-5xl flex-col items-center gap-6 rounded-3xl border border-border bg-muted px-6 py-10 text-center sm:px-10">
          <Layers aria-hidden="true" className="size-8 text-foreground" />
          <h2 className="text-2xl font-bold text-foreground sm:text-3xl">با صندلی‌های رایگان شروع کنید</h2>
          <p className="max-w-2xl leading-7 text-muted-foreground">سازمان را بسازید، تا ۵ عضو را رایگان دعوت کنید و هر وقت لازم شد صندلی پولی یا بسته‌ی اعتبار بخرید.</p>
          <Link href="/onboarding" className="inline-flex min-h-12 items-center justify-center gap-2 rounded-xl bg-primary px-6 text-sm font-semibold text-primary-foreground transition-opacity hover:opacity-90">
            ساخت سازمان رایگان
            <ArrowLeft aria-hidden="true" className="size-4" />
          </Link>
          <Link href="/pricing/changelog" className="text-sm font-medium text-foreground hover:underline">
            تاریخچه‌ی تغییرات قیمت و نرخ‌نامه
          </Link>
        </div>
      </section>
    </main>
  );
}

function SeatCard({ seat, creditPriceToman }: { seat: SeatDefinition; creditPriceToman: number | null }) {
  const isFree = seat.slug === 'free';

  return (
    <article className={`relative flex flex-col rounded-2xl border bg-card p-6 ${seat.highlighted ? 'border-primary ring-2 ring-primary/15' : 'border-border'}`}>
      {seat.highlighted ? (
        <span className="absolute -top-3 right-5 rounded-full bg-primary px-3 py-1 text-xs font-bold text-primary-foreground">پرکاربردترین</span>
      ) : null}
      <h2 className="text-xl font-bold text-foreground">{seat.name}</h2>
      <p className="mt-2 min-h-14 text-sm leading-7 text-muted-foreground">{seat.description}</p>

      <div className="mt-5 border-b border-border pb-5">
        <div className="flex flex-wrap items-baseline gap-x-2 gap-y-1">
          <span className="text-3xl font-black tracking-tight text-foreground">{formatNumber(seat.credits)}</span>
          <span className="text-sm font-medium text-muted-foreground">{seat.creditsLabel}</span>
        </div>
        <p className="mt-3 text-sm leading-7 text-muted-foreground">
          {isFree
            ? 'رایگان، بدون نیاز به پرداخت'
            : creditPriceToman === null
              ? 'قیمت ماهانه = اعتبار ماهانه × قیمت هر اعتبار، منهای تخفیف صندلی'
              : `پیش از تخفیف صندلی: ${formatNumber(seat.credits * creditPriceToman)} تومان در ماه`}
        </p>
      </div>

      <ul className="mt-5 flex flex-1 flex-col gap-3 text-sm leading-6 text-foreground">
        {seat.features.map((feature) => (
          <li key={feature} className="flex items-start gap-2.5">
            <Check aria-hidden="true" className="mt-0.5 size-4 shrink-0 text-muted-foreground" />
            <span>{feature}</span>
          </li>
        ))}
      </ul>

      <Link href="/onboarding" className={`mt-6 inline-flex min-h-12 items-center justify-center gap-2 rounded-xl px-4 text-sm font-semibold transition-opacity hover:opacity-90 ${seat.highlighted ? 'bg-primary text-primary-foreground' : 'bg-muted text-foreground'}`}>
        {isFree ? 'شروع رایگان' : 'شروع و خرید صندلی در پنل'}
        <ArrowLeft aria-hidden="true" className="size-4" />
      </Link>
    </article>
  );
}

function formatNumber(value: number) {
  return new Intl.NumberFormat('fa-IR').format(value);
}
