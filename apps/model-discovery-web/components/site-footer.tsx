import Link from 'next/link';
import { useTranslations } from 'next-intl';

import { GitHubIcon } from '@/components/github-icon';
import { GodarAILogo } from '@/components/godarai-logo';
import { SOCIAL_LINKS } from '@/lib/social-links';

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
              <GodarAILogo
                decorative
                className="h-9 w-auto max-w-12 shrink-0 object-contain"
                sizes="3rem"
              />
              <span className="text-xl font-bold text-foreground">{t('brand')}</span>
            </div>
            <p className="mb-3 max-w-2xl text-xs leading-6 text-muted-foreground md:text-sm">
              {t('description')}
            </p>
            <div className="flex items-center gap-3">
              <a
                href={SOCIAL_LINKS.linkedin}
                target="_blank"
                rel="noopener noreferrer"
                aria-label={t('linkedinAria')}
                className={footerLinkClass}
              >
                <svg className="size-4 md:size-5" fill="currentColor" viewBox="0 0 24 24" aria-hidden="true">
                  <path d="M20.45 20.45h-3.56v-5.57c0-1.33-.03-3.04-1.85-3.04-1.85 0-2.14 1.45-2.14 2.94v5.67H9.34V8.98h3.42v1.57h.05c.48-.9 1.64-1.85 3.37-1.85 3.61 0 4.27 2.37 4.27 5.46v6.29ZM5.32 7.41a2.06 2.06 0 1 1 0-4.12 2.06 2.06 0 0 1 0 4.12Zm1.78 13.04H3.54V8.98H7.1v11.47Z" />
                </svg>
              </a>
              <a
                href={SOCIAL_LINKS.telegram}
                target="_blank"
                rel="noopener noreferrer"
                aria-label={t('telegramAria')}
                className={footerLinkClass}
              >
                <svg className="size-4 md:size-5" fill="currentColor" viewBox="0 0 24 24" aria-hidden="true">
                  <path d="M21.94 4.66c.26-1.21-.44-1.69-1.21-1.4L2.4 10.33c-1.25.49-1.23 1.19-.21 1.5l4.71 1.47 10.91-6.88c.52-.31.99-.14.6.21l-8.84 7.98-.33 4.67c.48 0 .69-.22.96-.48l2.31-2.25 4.8 3.55c.89.49 1.52.24 1.75-.82l2.88-14.62Z" />
                </svg>
              </a>
              <a
                href={SOCIAL_LINKS.github}
                target="_blank"
                rel="noopener noreferrer"
                aria-label={t('githubAria')}
                className={footerLinkClass}
              >
                <GitHubIcon className="size-4 md:size-5" aria-hidden="true" />
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
              <a href={SOCIAL_LINKS.github} target="_blank" rel="noopener noreferrer" className={footerLinkClass}>
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
