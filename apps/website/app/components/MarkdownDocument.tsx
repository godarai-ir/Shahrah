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
          <h1 className="mb-6 text-2xl font-bold leading-9 text-foreground md:text-3xl">
            {children}
          </h1>
        ),
        h2: ({ children }) => (
          <h2 className="mb-4 mt-8 text-xl font-bold leading-8 text-foreground md:text-2xl">
            {children}
          </h2>
        ),
        h3: ({ children }) => (
          <h3 className="mb-3 mt-6 text-base font-semibold leading-7 text-foreground md:text-lg">
            {children}
          </h3>
        ),
        p: ({ children }) => (
          <p className="mb-5 text-sm leading-8 text-card-foreground md:text-base">
            {children}
          </p>
        ),
        ul: ({ children }) => (
          <ul className="mb-5 flex list-disc flex-col gap-2 pe-6 text-sm leading-8 text-card-foreground md:text-base">
            {children}
          </ul>
        ),
        ol: ({ children }) => (
          <ol className="mb-5 flex list-decimal flex-col gap-2 pe-6 text-sm leading-8 text-card-foreground md:text-base">
            {children}
          </ol>
        ),
        li: ({ children }) => <li className="ps-1">{children}</li>,
        strong: ({ children }) => (
          <strong className="font-semibold text-foreground">{children}</strong>
        ),
        a: ({ children, href }) => (
          <a
            href={href}
            className="font-medium text-primary underline underline-offset-4"
          >
            {children}
          </a>
        ),
        hr: () => <hr className="my-8 border-border" />,
        code: ({ children }) => (
          <code className="rounded-sm bg-muted px-1.5 py-0.5 font-mono text-sm text-foreground">
            {children}
          </code>
        ),
      }}
    >
      {content}
    </ReactMarkdown>
  );
}
