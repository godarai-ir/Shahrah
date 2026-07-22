import Image from 'next/image';
import Link from 'next/link';
import { useTranslations } from 'next-intl';

const footerLinkClass =
  'text-muted-foreground transition-colors hover:text-primary focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-primary';

export function SiteFooter() {
  const t = useTranslations('Footer');

  return (
    <footer className="border-t border-border bg-muted/40 py-8 text-start">
      <div className="mx-auto w-full max-w-7xl px-4 sm:px-6 lg:px-8">
        <div className="mb-6 grid grid-cols-3 gap-x-4 gap-y-6 md:grid-cols-4 md:gap-6">
          <div className="col-span-3 md:col-span-1">
            <div className="mb-3 flex items-center gap-2">
              <Image
                src="/catalog/mini_logo.svg"
                alt=""
                aria-hidden="true"
                width={36}
                height={36}
                className="size-9 shrink-0 rounded-lg"
              />
              <span className="text-xl font-bold text-foreground">{t('brand')}</span>
            </div>
            <p className="mb-3 max-w-2xl text-xs leading-6 text-muted-foreground md:text-sm">
              {t('description')}
            </p>
            <div className="flex items-center gap-3">
              <a
                href="https://x.com/r_delpak"
                target="_blank"
                rel="noopener noreferrer"
                aria-label={t('xAria')}
                className={footerLinkClass}
              >
                <svg className="size-4 md:size-5" fill="currentColor" viewBox="0 0 24 24" aria-hidden="true">
                  <path d="M8.29 20.251c7.547 0 11.675-6.253 11.675-11.675 0-.178 0-.355-.012-.53A8.348 8.348 0 0022 5.92a8.19 8.19 0 01-2.357.646 4.118 4.118 0 001.804-2.27 8.224 8.224 0 01-2.605.996 4.107 4.107 0 00-6.993 3.743 11.65 11.65 0 01-8.457-4.287 4.106 4.106 0 001.27 5.477A4.072 4.072 0 012.8 9.713v.052a4.105 4.105 0 003.292 4.022 4.095 4.095 0 01-1.853.07 4.108 4.108 0 003.834 2.85A8.233 8.233 0 012 18.407a11.616 11.616 0 006.29 1.84" />
                </svg>
              </a>
              <a
                href="https://github.com/delpak/shahrah/"
                target="_blank"
                rel="noopener noreferrer"
                aria-label={t('githubAria')}
                className={footerLinkClass}
              >
                <svg className="size-4 md:size-5" fill="currentColor" viewBox="0 0 24 24" aria-hidden="true">
                  <path
                    fillRule="evenodd"
                    d="M12 2C6.477 2 2 6.484 2 12.017c0 4.425 2.865 8.18 6.839 9.504.5.092.682-.217.682-.483 0-.237-.008-.868-.013-1.703-2.782.605-3.369-1.343-3.369-1.343-.454-1.158-1.11-1.466-1.11-1.466-.908-.62.069-.608.069-.608 1.003.07 1.531 1.032 1.531 1.032.892 1.53 2.341 1.088 2.91.832.092-.647.35-1.088.636-1.338-2.22-.253-4.555-1.113-4.555-4.951 0-1.093.39-1.988 1.029-2.688-.103-.253-.446-1.272.098-2.65 0 0 .84-.27 2.75 1.026A9.564 9.564 0 0112 6.844c.85.004 1.705.115 2.504.337 1.909-1.296 2.747-1.027 2.747-1.027.546 1.379.202 2.398.1 2.651.64.7 1.028 1.595 1.028 2.688 0 3.848-2.339 4.695-4.566 4.943.359.309.678.92.678 1.855 0 1.338-.012 2.419-.012 2.747 0 .268.18.58.688.482A10.019 10.019 0 0022 12.017C22 6.484 17.522 2 12 2z"
                    clipRule="evenodd"
                  />
                </svg>
              </a>
            </div>
          </div>

          <FooterSection title={t('product')}>
            <li><Link href="/" className={footerLinkClass}>{t('overview')}</Link></li>
            <li><a href="/features/observability" className={footerLinkClass}>{t('observability')}</a></li>
            <li><a href="/features/ai-gateway" className={footerLinkClass}>{t('aiGateway')}</a></li>
            <li><a href="/features/guardrails" className={footerLinkClass}>{t('guardrails')}</a></li>
            <li><a href="/features/prompt-engineering" className={footerLinkClass}>{t('promptRegistry')}</a></li>
            <li><a href="/features/security" className={footerLinkClass}>{t('security')}</a></li>
            <li><a href="/pricing" className={footerLinkClass}>{t('plans')}</a></li>
          </FooterSection>

          <FooterSection title={t('resources')}>
            <li><a href="/docs" className={footerLinkClass}>{t('docs')}</a></li>
            <li><a href="/blog" className={footerLinkClass}>{t('blog')}</a></li>
            <li><Link href="/" className={footerLinkClass}>{t('catalog')}</Link></li>
            <li>
              <a href="https://github.com/delpak/shahrah/" target="_blank" rel="noopener noreferrer" className={footerLinkClass}>
                {t('github')}
              </a>
            </li>
          </FooterSection>

          <FooterSection title={t('company')}>
            <li><a href="/about" className={footerLinkClass}>{t('about')}</a></li>
            <li><a href="/contact" className={footerLinkClass}>{t('contact')}</a></li>
            <li><a href="/terms" className={footerLinkClass}>{t('terms')}</a></li>
            <li><a href="/privacy" className={footerLinkClass}>{t('privacy')}</a></li>
          </FooterSection>
        </div>

        <div className="grid grid-cols-[1fr_auto] items-center gap-3 border-t border-border pt-4 md:pt-6">
          <p className="min-w-0 text-xs leading-5 text-muted-foreground md:text-sm">
            {t('copyright')}
          </p>
          <a
            referrerPolicy="origin"
            target="_blank"
            rel="noopener noreferrer"
            href="https://trustseal.enamad.ir/?id=737580&Code=Dc476GaakpMUDp6MG3lSrtPK8uL9943N"
            aria-label={t('trustSealAria')}
            className="inline-flex h-11 w-14 shrink-0 items-center justify-center overflow-hidden rounded-lg border border-border bg-background p-1 md:h-14 md:w-16"
          >
            {/* eslint-disable-next-line @next/next/no-img-element */}
            <img
              referrerPolicy="origin"
              src="https://trustseal.enamad.ir/logo.aspx?id=737580&Code=Dc476GaakpMUDp6MG3lSrtPK8uL9943N"
              alt={t('trustSealAlt')}
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

function FooterSection({ title, children }: { title: string; children: React.ReactNode }) {
  return (
    <div className="min-w-0">
      <h2 className="mb-2 text-sm font-bold text-foreground md:mb-3 md:text-base">{title}</h2>
      <ul className="space-y-1.5 text-xs md:space-y-2 md:text-sm">{children}</ul>
    </div>
  );
}
