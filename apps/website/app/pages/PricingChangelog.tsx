import Link from "next/link";

export type PublicPricingChangelog = {
  rate_cards: {
    version: number;
    label: string;
    status: "scheduled" | "published" | "superseded" | string;
    effective_from: string;
    changelog_fa: string;
    changelog_en: string;
  }[];
  credit_prices: {
    price_toman_per_credit: number;
    status: "active" | "superseded" | string;
    effective_from: string;
    plan_effective_from: string;
  }[];
};

const dateFormatter = new Intl.DateTimeFormat("fa-IR-u-ca-persian", { dateStyle: "long" });
const numberFormatter = new Intl.NumberFormat("fa-IR");

function formatDate(value: string) {
  const date = new Date(value);
  return Number.isNaN(date.getTime()) ? "—" : dateFormatter.format(date);
}

const rateCardStatus: Record<string, string> = {
  scheduled: "زمان‌بندی‌شده",
  published: "فعال",
  superseded: "قبلی",
};

const priceStatus: Record<string, string> = {
  active: "فعال",
  superseded: "قبلی",
};

export function PricingChangelog({ changelog }: { changelog: PublicPricingChangelog | null }) {
  return (
    <main className="bg-site-ground px-4 py-10 sm:px-6 lg:py-20">
      <div className="mx-auto flex max-w-3xl flex-col gap-10">
        <header className="flex flex-col gap-3">
          <h1 className="m-0 text-balance text-h2 font-extrabold text-site-ink lg:text-h1">تغییرات قیمت و نرخ‌نامه</h1>
          <p className="m-0 text-lead font-light text-site-ink-soft">
            هر تغییر در قیمت هر کردیت یا نرخ‌نامه‌ی مدل‌ها همراه با تاریخ اعمال آن در این صفحه ثبت می‌شود. کردیت‌هایی که
            قبلاً خریده‌اید هیچ‌وقت ارزش مصرفشان را از دست نمی‌دهند. افزایش قیمت صندلی‌ها و افزایش نرخ مدل‌ها دست‌کم ۳۰ روز
            پیش از اعمال اعلام می‌شود؛ کاهش‌ها فوراً اعمال می‌شوند.
          </p>
          <Link href="/pricing" className="w-fit rounded-full text-label font-medium text-site-accent-text underline-offset-4 hover:underline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-site-focus">بازگشت به قیمت‌گذاری</Link>
        </header>

        {changelog === null ? (
          <p role="status" className="rounded-card border border-site-line bg-site-surface p-4 text-site-ink-soft">
            تاریخچه‌ی قیمت‌ها در حال حاضر در دسترس نیست. کمی بعد دوباره سر بزنید.
          </p>
        ) : (
          <>
            <section aria-labelledby="credit-prices" className="flex flex-col gap-4">
              <h2 id="credit-prices" className="m-0 text-h3 font-semibold text-site-ink">قیمت هر کردیت</h2>
              <ol className="flex flex-col divide-y divide-site-line rounded-card border border-site-line bg-site-surface">
                {changelog.credit_prices.map((price) => (
                  <li key={price.effective_from} className="flex flex-col gap-1 p-4">
                    <div className="flex flex-wrap items-center justify-between gap-2">
                      <span className="text-lead font-semibold text-site-ink">
                        {numberFormatter.format(price.price_toman_per_credit)} تومان
                      </span>
                      <span className="rounded-full border border-site-line px-3 py-0.5 text-caption text-site-ink-soft">
                        {priceStatus[price.status] ?? price.status}
                      </span>
                    </div>
                    <p className="m-0 text-label text-site-ink-soft">
                      بسته‌های کردیت از {formatDate(price.effective_from)} · صندلی‌ها از دوره‌ی بعد از {formatDate(price.plan_effective_from)}
                    </p>
                  </li>
                ))}
                {changelog.credit_prices.length === 0 ? <li className="p-4 text-site-ink-soft">هنوز تغییری ثبت نشده است.</li> : null}
              </ol>
            </section>

            <section aria-labelledby="rate-cards" className="flex flex-col gap-4">
              <h2 id="rate-cards" className="m-0 text-h3 font-semibold text-site-ink">نسخه‌های نرخ‌نامه</h2>
              <ol className="flex flex-col divide-y divide-site-line rounded-card border border-site-line bg-site-surface">
                {changelog.rate_cards.map((card) => (
                  <li key={card.version} className="flex flex-col gap-1 p-4">
                    <div className="flex flex-wrap items-center justify-between gap-2">
                      <span className="font-semibold text-site-ink" dir="ltr">{card.label}</span>
                      <span className="rounded-full border border-site-line px-3 py-0.5 text-caption text-site-ink-soft">
                        {rateCardStatus[card.status] ?? card.status}
                      </span>
                    </div>
                    <p className="m-0 text-label text-site-ink-soft">اعمال از {formatDate(card.effective_from)}</p>
                    {card.changelog_fa ? <p className="m-0 text-body text-site-ink">{card.changelog_fa}</p> : null}
                  </li>
                ))}
                {changelog.rate_cards.length === 0 ? <li className="p-4 text-site-ink-soft">هنوز نسخه‌ای منتشر نشده است.</li> : null}
              </ol>
            </section>
          </>
        )}
      </div>
    </main>
  );
}
