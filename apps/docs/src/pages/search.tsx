import {useMemo, useState} from 'react';
import Layout from '@theme/Layout';
import Link from '@docusaurus/Link';

type DocEntry = {
  title: string;
  path: string;
  excerpt: string;
  keywords: string[];
};

const docs: DocEntry[] = [
  {
    title: 'معرفی گیت‌وی گدارAI',
    path: '/intro-to-gateway',
    excerpt: 'مرور مفاهیم gateway، provider account، virtual model، logs و policyها',
    keywords: ['gateway', 'provider', 'virtual model', 'logs', 'policy'],
  },
  {
    title: 'شروع سریع گدارAI',
    path: '/quick-start',
    excerpt: 'مراحل ثبت‌نام، اتصال provider، آزمایش مدل و گرفتن base URL و API key',
    keywords: ['شروع سریع', 'provider', 'api key', 'playground'],
  },
  {
    title: 'مدل مجازی',
    path: '/virtual-model',
    excerpt: 'alias پایدار، routing اولویتی یا وزنی، retry، fallback و کنترل دسترسی',
    keywords: ['alias', 'routing', 'fallback', 'retry', 'weighted', 'priority'],
  },
  {
    title: 'مقدمه',
    path: '/intro',
    excerpt: 'نمای کلی مستندات و معرفی بخش‌های اصلی محصول',
    keywords: ['intro', 'documentation', 'مستندات'],
  },
];

export default function SearchPage() {
  const [query, setQuery] = useState('');

  const results = useMemo(() => {
    const normalized = normalize(query);
    if (!normalized) return docs;
    return docs.filter((doc) => {
      const haystack = normalize([doc.title, doc.excerpt, ...doc.keywords].join(' '));
      return haystack.includes(normalized);
    });
  }, [query]);

  return (
    <Layout title="جستجو" description="جستجو در مستندات گدارAI">
      <main className="search-page" dir="rtl">
        <section className="search-page__hero">
          <p className="search-page__eyebrow">جستجو در مستندات</p>
          <h1>صفحه جستجو</h1>
          <p>نام صفحه، عبارت کلیدی یا موضوع موردنظر را بنویسید تا مستندات مرتبط را پیدا کنید.</p>
          <input
            className="search-page__input"
            value={query}
            onChange={(event) => setQuery(event.target.value)}
            placeholder="مثلا: مدل مجازی، شروع سریع، provider account"
            dir="rtl"
          />
        </section>

        <section className="search-page__results" aria-live="polite">
          {results.map((doc) => (
            <article key={doc.path} className="search-page__card">
              <h2>{doc.title}</h2>
              <p>{doc.excerpt}</p>
              <Link className="search-page__link" to={doc.path}>
                مشاهده صفحه
              </Link>
            </article>
          ))}
          {results.length === 0 ? <p className="search-page__empty">نتیجه‌ای پیدا نشد.</p> : null}
        </section>
      </main>
    </Layout>
  );
}

function normalize(value: string) {
  return value
    .toLowerCase()
    .replace(/\s+/g, ' ')
    .trim();
}
