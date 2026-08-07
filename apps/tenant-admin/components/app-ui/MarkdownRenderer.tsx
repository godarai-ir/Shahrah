"use client";

import { Children, isValidElement, type ComponentPropsWithoutRef, type ReactNode } from "react";
import ReactMarkdown from "react-markdown";
import remarkGfm from "remark-gfm";

import { cn } from "@/lib/utils";

import { CodeBlock, type CodeLanguage } from "./CodeBlock";

export type MarkdownRendererProps = {
  content: string;
  className?: string;
};

export function MarkdownRenderer({ content, className }: MarkdownRendererProps) {
  return (
    <div className={cn("min-w-0 space-y-4 text-sm leading-7 text-card-foreground", className)} dir="auto">
      <ReactMarkdown
        remarkPlugins={[remarkGfm]}
        components={{
          h1: ({ children }) => <h1 className="text-2xl font-bold leading-9 text-foreground">{children}</h1>,
          h2: ({ children }) => <h2 className="text-xl font-bold leading-8 text-foreground">{children}</h2>,
          h3: ({ children }) => <h3 className="text-lg font-semibold leading-8 text-foreground">{children}</h3>,
          h4: ({ children }) => <h4 className="text-base font-semibold leading-7 text-foreground">{children}</h4>,
          h5: ({ children }) => <h5 className="font-semibold text-foreground">{children}</h5>,
          h6: ({ children }) => <h6 className="font-medium text-muted-foreground">{children}</h6>,
          p: ({ children }) => <p className="leading-7">{children}</p>,
          ul: ({ children }) => <ul className="list-disc space-y-2 ps-6 marker:text-muted-foreground">{children}</ul>,
          ol: ({ children }) => <ol className="list-decimal space-y-2 ps-6 marker:text-muted-foreground">{children}</ol>,
          li: ({ children }) => <li className="ps-1">{children}</li>,
          strong: ({ children }) => <strong className="font-semibold text-foreground">{children}</strong>,
          blockquote: ({ children }) => (
            <blockquote className="border-s-4 border-primary-subtle-border bg-muted/50 px-4 py-3 text-muted-foreground">
              {children}
            </blockquote>
          ),
          a: ({ children, href }) => {
            const external = href?.startsWith("http://") || href?.startsWith("https://");
            return (
              <a
                href={href}
                className="font-medium text-primary underline decoration-primary/40 underline-offset-4 hover:decoration-primary"
                target={external ? "_blank" : undefined}
                rel={external ? "noopener noreferrer" : undefined}
              >
                {children}
              </a>
            );
          },
          hr: () => <hr className="border-border" />,
          code: ({ children }) => (
            <code dir="ltr" className="rounded-sm bg-muted px-1.5 py-0.5 font-mono text-xs text-foreground">
              {children}
            </code>
          ),
          pre: MarkdownPre,
          table: ({ children }) => (
            <div className="custom-scrollbar overflow-x-auto rounded-lg border border-border">
              <table className="w-full border-collapse text-sm">{children}</table>
            </div>
          ),
          thead: ({ children }) => <thead className="bg-muted/60 text-foreground">{children}</thead>,
          tr: ({ children }) => <tr className="border-b border-border last:border-b-0">{children}</tr>,
          th: ({ children }) => <th className="px-3 py-2 text-start font-semibold">{children}</th>,
          td: ({ children }) => <td className="px-3 py-2 align-top">{children}</td>,
          del: ({ children }) => <del className="text-muted-foreground">{children}</del>,
          img: ({ alt, src }) => (
            // Markdown images may be remote provider output, so a native image preserves arbitrary dimensions.
            // eslint-disable-next-line @next/next/no-img-element
            <img src={src} alt={alt ?? ""} loading="lazy" referrerPolicy="no-referrer" className="h-auto max-w-full rounded-lg border border-border" />
          ),
          input: (inputProps) => <input {...inputProps} className="me-2 accent-primary" />,
        }}
      >
        {content}
      </ReactMarkdown>
    </div>
  );
}

function MarkdownPre({ children }: ComponentPropsWithoutRef<"pre">) {
  const elements = Children.toArray(children);
  const codeElement = elements.length === 1 ? elements[0] : null;

  if (isValidElement<{ children?: ReactNode; className?: string }>(codeElement)) {
    const code = String(codeElement.props.children ?? "").replace(/\n$/, "");
    const language = markdownCodeLanguage(codeElement.props.className);

    return <CodeBlock code={code} language={language} className="my-1" />;
  }

  return <pre className="custom-scrollbar overflow-x-auto rounded-lg border border-code-border bg-code p-4 font-mono text-code-foreground">{children}</pre>;
}

function markdownCodeLanguage(className?: string): CodeLanguage {
  const language = /(?:^|\s)language-([\w-]+)/.exec(className ?? "")?.[1]?.toLowerCase();
  const aliases: Record<string, CodeLanguage> = {
    bash: "bash",
    console: "bash",
    js: "javascript",
    javascript: "javascript",
    json: "json",
    markdown: "markdown",
    md: "markdown",
    powershell: "powershell",
    ps1: "powershell",
    py: "python",
    python: "python",
    shell: "bash",
    sh: "bash",
    ts: "typescript",
    typescript: "typescript",
  };

  return language ? aliases[language] ?? "text" : "text";
}
