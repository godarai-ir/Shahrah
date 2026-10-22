import Link from 'next/link';
import { footerColumns, visibleColumns } from '@/app/lib/navigation';
import { SOCIAL_LINKS } from '@/app/lib/social-links';
import { GodarAILogo } from './GodarAILogo';
import { TrustStrip } from './patterns';

const focusRing = 'focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-site-accent';
const linkClass = `rounded text-sm leading-7 text-site-ink-soft transition-colors hover:text-site-ink ${focusRing}`;
const socialClass = `flex size-11 items-center justify-center rounded-full text-site-ink-soft transition-colors hover:bg-site-panel-deep hover:text-site-ink ${focusRing}`;

export function Footer() {
  return (
    <footer className="border-t border-site-line bg-site-panel">
      <div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
        <div className="border-b border-site-line py-10">
          <TrustStrip />
        </div>

        <div className="grid grid-cols-2 gap-x-6 gap-y-10 py-12 sm:grid-cols-3 lg:grid-cols-6">
          <div className="col-span-2 flex flex-col gap-4 sm:col-span-3 lg:col-span-1">
            <div className="flex items-center gap-2">
              <GodarAILogo decorative className="h-9 w-auto max-w-12 shrink-0 object-contain" sizes="3rem" />
              <span className="text-xl font-bold text-site-ink">گدارAI</span>
            </div>
            <p className="m-0 max-w-xs text-sm leading-7 text-site-ink-soft">
              فضای کاری هوش مصنوعی و درگاه مدل‌های زبانی برای سازمان‌های ایرانی.
            </p>
            <div className="-ms-3 flex items-center">
              <a href={SOCIAL_LINKS.linkedin} target="_blank" rel="noopener noreferrer" aria-label="گدارAI در لینکدین" className={socialClass}>
                <svg className="size-5" fill="currentColor" viewBox="0 0 24 24" aria-hidden="true">
                  <path d="M20.45 20.45h-3.56v-5.57c0-1.33-.03-3.04-1.85-3.04-1.85 0-2.14 1.45-2.14 2.94v5.67H9.34V8.98h3.42v1.57h.05c.48-.9 1.64-1.85 3.37-1.85 3.61 0 4.27 2.37 4.27 5.46v6.29ZM5.32 7.41a2.06 2.06 0 1 1 0-4.12 2.06 2.06 0 0 1 0 4.12Zm1.78 13.04H3.54V8.98H7.1v11.47Z" />
                </svg>
              </a>
              <a href={SOCIAL_LINKS.telegram} target="_blank" rel="noopener noreferrer" aria-label="گدارAI در تلگرام" className={socialClass}>
                <svg className="size-5" fill="currentColor" viewBox="0 0 24 24" aria-hidden="true">
                  <path d="M21.94 4.66c.26-1.21-.44-1.69-1.21-1.4L2.4 10.33c-1.25.49-1.23 1.19-.21 1.5l4.71 1.47 10.91-6.88c.52-.31.99-.14.6.21l-8.84 7.98-.33 4.67c.48 0 .69-.22.96-.48l2.31-2.25 4.8 3.55c.89.49 1.52.24 1.75-.82l2.88-14.62Z" />
                </svg>
              </a>
              <a href={SOCIAL_LINKS.github} target="_blank" rel="noopener noreferrer" aria-label="گدارAI در گیت‌هاب" className={socialClass}>
                <svg className="size-5" fill="currentColor" viewBox="0 0 24 24" aria-hidden="true">
                  <path fillRule="evenodd" d="M12 2C6.477 2 2 6.484 2 12.017c0 4.425 2.865 8.18 6.839 9.504.5.092.682-.217.682-.483 0-.237-.008-.868-.013-1.703-2.782.605-3.369-1.343-3.369-1.343-.454-1.158-1.11-1.466-1.11-1.466-.908-.62.069-.608.069-.608 1.003.07 1.531 1.032 1.531 1.032.892 1.53 2.341 1.088 2.91.832.092-.647.35-1.088.636-1.338-2.22-.253-4.555-1.113-4.555-4.951 0-1.093.39-1.988 1.029-2.688-.103-.253-.446-1.272.098-2.65 0 0 .84-.27 2.75 1.026A9.564 9.564 0 0112 6.844c.85.004 1.705.115 2.504.337 1.909-1.296 2.747-1.027 2.747-1.027.546 1.379.202 2.398.1 2.651.64.7 1.028 1.595 1.028 2.688 0 3.848-2.339 4.695-4.566 4.943.359.309.678.92.678 1.855 0 1.338-.012 2.419-.012 2.747 0 .268.18.58.688.482A10.019 10.019 0 0022 12.017C22 6.484 17.522 2 12 2z" clipRule="evenodd" />
                </svg>
              </a>
            </div>
          </div>

          {visibleColumns(footerColumns).map((column) => (
            <div key={column.title} className="min-w-0">
              <h2 className="mb-3 text-sm font-semibold text-site-ink">{column.title}</h2>
              <ul className="m-0 flex list-none flex-col gap-1 p-0">
                {column.links.map((item) => (
                  <li key={item.href}>
                    {item.external ? (
                      <a href={item.href} className={linkClass}>
                        {item.label}
                      </a>
                    ) : (
                      <Link href={item.href} className={linkClass}>
                        {item.label}
                      </Link>
                    )}
                  </li>
                ))}
              </ul>
            </div>
          ))}
        </div>

        <div className="flex items-center justify-between gap-3 border-t border-site-line py-6">
          <p className="m-0 min-w-0 text-sm leading-6 text-site-ink-soft">© ۱۴۰۵ گدارAI. تمامی حقوق محفوظ است.</p>
          {/* The trust seal keeps a light surface in both themes so the issuer's image stays legible. */}
          <a
            referrerPolicy="origin"
            target="_blank"
            rel="noopener noreferrer"
            href="https://trustseal.enamad.ir/?id=737580&Code=Dc476GaakpMUDp6MG3lSrtPK8uL9943N"
            className={`inline-flex h-14 w-16 shrink-0 items-center justify-center overflow-hidden rounded-xl border border-site-line bg-site-night-ink p-1 text-xs text-site-night ${focusRing}`}
          >
            {/* eslint-disable-next-line @next/next/no-img-element */}
            <img
              referrerPolicy="origin"
              src="https://trustseal.enamad.ir/logo.aspx?id=737580&Code=Dc476GaakpMUDp6MG3lSrtPK8uL9943N"
              alt="نماد اعتماد"
              style={{ cursor: 'pointer' }}
              width={56}
              height={48}
              className="max-h-full w-auto object-contain"
              data-code="Dc476GaakpMUDp6MG3lSrtPK8uL9943N"
            />
          </a>
        </div>
      </div>
    </footer>
  );
}
