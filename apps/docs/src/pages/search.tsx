import {useMemo, useState} from 'react';
import {Search as SearchIcon} from 'lucide-react';
import Layout from '@theme/Layout';
import Link from '@docusaurus/Link';
import {usePluginData} from '@docusaurus/useGlobalData';
import {searchDocs} from '../search/searchDocs';
import type {SearchData} from '../search/types';

const localSearchPluginName = 'shahrah-local-search';

export default function SearchPage() {
  const [query, setQuery] = useState('');
  const searchData = usePluginData(localSearchPluginName) as SearchData | undefined;
  const entries = searchData?.entries ?? [];

  const search = useMemo(() => {
    return searchDocs(entries, query);
  }, [entries, query]);
  const {results, total} = search;

  const hasQuery = query.trim().length > 0;

  return (
    <Layout title="جستجو" description="جستجو در مستندات گدارAI">
      <main className="search-page" dir="rtl">
        <section className="search-page__hero">
          <p className="search-page__eyebrow">جستجو در مستندات</p>
          <p>
            عنوان صفحه، عبارت فنی یا بخشی از مفهوم را بنویسید.
          </p>
          <label className="search-page__field">
            <span className="search-page__input-wrap">
              <SearchIcon aria-hidden="true" size={19} strokeWidth={2} />
              <input
                className="search-page__input"
                value={query}
                onChange={event => setQuery(event.target.value)}
                placeholder="مثلا: مدل مجازی، كليد دسترسی، Chat Completions"
                aria-label="عبارت جستجو"
                dir="auto"
                autoComplete="off"
              />
            </span>
          </label>
        </section>

        <section className="search-page__summary" aria-live="polite">
          {hasQuery ? (
            <p>
              {results.length > 0
                ? `${results.length.toLocaleString('fa-IR')} از ${total.toLocaleString('fa-IR')} نتیجه`
                : 'نتیجه‌ای پیدا نشد.'}
            </p>
          ) : (
            <p>
              {entries.length.toLocaleString('fa-IR')} صفحه آماده جستجو است.
              برای شروع، یک عبارت بنویسید.
            </p>
          )}
        </section>

        <section className="search-page__results" aria-label="نتایج جستجو">
          {results.map(doc => (
            <Link key={doc.id} className="search-page__card" to={doc.path}>
              <div>
                <h2>{doc.title}</h2>
                {doc.excerpt ? <p>{doc.excerpt}</p> : null}
              </div>
            </Link>
          ))}
          {!hasQuery ? (
            <div className="search-page__empty-state">
              <SearchIcon aria-hidden="true" size={24} strokeWidth={2} />
              <h2>عبارت جستجو را وارد کنید</h2>
              <p>
                می‌توانید نام صفحه، قابلیت، خطا، API یا عبارت فنی موردنظر را
                جستجو کنید.
              </p>
            </div>
          ) : null}
          {hasQuery && results.length === 0 ? (
            <p className="search-page__empty">
              املای فارسی و انگلیسی عبارت را بررسی کنید یا یک واژه کوتاه‌تر
              بنویسید.
            </p>
          ) : null}
        </section>
      </main>
    </Layout>
  );
}
