import ReactMarkdown from "react-markdown";
import remarkGfm from "remark-gfm";

type MarkdownDocumentProps = {
  content: string;
};

export function MarkdownDocument({ content }: MarkdownDocumentProps) {
  return (
    <ReactMarkdown
      remarkPlugins={[remarkGfm]}
      components={{
        h1: ({ children }) => (
          <h1 className="mb-6 text-h2 font-bold text-site-ink">
            {children}
          </h1>
        ),
        h2: ({ children }) => (
          <h2 className="mb-4 mt-10 text-h3 font-bold text-site-ink">
            {children}
          </h2>
        ),
        h3: ({ children }) => (
          <h3 className="mb-3 mt-8 text-lead font-semibold text-site-ink">
            {children}
          </h3>
        ),
        p: ({ children }) => (
          <p className="mb-5 text-body leading-8 text-site-ink">
            {children}
          </p>
        ),
        ul: ({ children }) => (
          <ul className="mb-5 flex list-disc flex-col gap-2 pe-6 text-body leading-8 text-site-ink">
            {children}
          </ul>
        ),
        ol: ({ children }) => (
          <ol className="mb-5 flex list-decimal flex-col gap-2 pe-6 text-body leading-8 text-site-ink">
            {children}
          </ol>
        ),
        li: ({ children }) => <li className="ps-1">{children}</li>,
        strong: ({ children }) => (
          <strong className="font-semibold text-site-ink">{children}</strong>
        ),
        a: ({ children, href }) => (
          <a
            href={href}
            className="rounded-field font-medium text-site-accent-text underline underline-offset-4 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-site-focus"
          >
            {children}
          </a>
        ),
        hr: () => <hr className="my-10 border-site-line" />,
        code: ({ children }) => (
          <code className="rounded-field bg-site-panel px-1.5 py-0.5 font-mono text-label text-site-ink">
            {children}
          </code>
        ),
      }}
    >
      {content}
    </ReactMarkdown>
  );
}
