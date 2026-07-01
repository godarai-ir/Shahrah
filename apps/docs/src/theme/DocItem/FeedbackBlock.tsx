import {useEffect, useState} from 'react';
import {ThumbsDown, ThumbsUp} from 'lucide-react';

type Vote = 'like' | 'dislike';

const STORAGE_PREFIX = 'godarai-docs-feedback:';

export default function FeedbackBlock() {
  const [selected, setSelected] = useState<Vote | null>(null);
  const [message, setMessage] = useState<string>('');

  useEffect(() => {
    if (typeof window === 'undefined') return;
    const key = getStorageKey();
    const existing = window.localStorage.getItem(key) as Vote | null;
    setSelected(existing === 'like' || existing === 'dislike' ? existing : null);
  }, []);

  function submitVote(next: Vote) {
    setSelected(next);
    setMessage('ممنون از بازخورد شما');
    if (typeof window === 'undefined') return;
    window.localStorage.setItem(getStorageKey(), next);
  }

  return (
    <section className="doc-feedback">
      <div className="doc-feedback__inner">
        <div className="doc-feedback__text">
          <p className="doc-feedback__title">آیا این صفحه مفید بود؟</p>
          <p className="doc-feedback__description">با انتخاب یکی از گزینه‌ها به بهتر شدن مستندات کمک می‌کنید.</p>
        </div>

        <div className="doc-feedback__actions" role="group" aria-label="بازخورد صفحه">
          <button
            type="button"
            className={`doc-feedback__button ${selected === 'like' ? 'doc-feedback__button--active' : ''}`}
            onClick={() => submitVote('like')}>
            <span className="doc-feedback__button-icon" aria-hidden="true">
              <ThumbsUp size={16} strokeWidth={2.2} />
            </span>
            <span>بله</span>
          </button>
          <button
            type="button"
            className={`doc-feedback__button ${selected === 'dislike' ? 'doc-feedback__button--active' : ''}`}
            onClick={() => submitVote('dislike')}>
            <span className="doc-feedback__button-icon" aria-hidden="true">
              <ThumbsDown size={16} strokeWidth={2.2} />
            </span>
            <span>خیر</span>
          </button>
        </div>
      </div>
      {message ? <p className="doc-feedback__message">{message}</p> : null}
    </section>
  );
}

function getStorageKey() {
  const pathname = typeof window === 'undefined' ? '' : window.location.pathname;
  return `${STORAGE_PREFIX}${pathname}`;
}
