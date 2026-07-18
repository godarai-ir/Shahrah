"use client";

import type React from "react";
import { Highlight, themes } from "prism-react-renderer";
import { Prism } from "prismjs";
import "prismjs/components/prism-bash";
import "prismjs/components/prism-json";
import "prismjs/components/prism-powershell";
import { Code2 } from "lucide-react";

import { CopyAction } from "@/components/app-ui";
import { Badge } from "@/components/ui/badge";
import { cn } from "@/lib/utils";
export function CodeExampleBlock(props: {
  code: string;
  language: "powershell" | "bash" | "json";
  shellLabel: string;
  className?: string;
  title?: React.ReactNode;
  headerActions?: React.ReactNode;
  copyLabel?: string;
  copiedLabel?: string;
  onCopied?: () => void;
  footer?: React.ReactNode;
}) {
  return (
    <div
      className={cn(
        "flex max-w-full min-w-0 flex-col gap-2 rounded-xl border border-border bg-card p-2.5 shadow-sm sm:gap-3 sm:rounded-2xl sm:p-4",
        props.className,
      )}
    >
      <div dir="rtl" className="flex min-w-0 flex-row flex-nowrap items-center justify-between gap-2 sm:gap-3">
        <div className="flex min-w-0 items-center gap-2 sm:gap-3">
          <span className="inline-flex size-8 shrink-0 items-center justify-center rounded-xl bg-primary/10 text-primary shadow-sm sm:size-10">
            <Code2 aria-hidden="true" />
          </span>
          <div className="min-w-0 truncate whitespace-nowrap text-sm font-semibold text-foreground sm:text-base">{props.title ?? props.shellLabel}</div>
        </div>
        {props.headerActions ?? (
          <div className="flex min-w-0 items-center gap-3">
            <Badge variant="outline" className="px-2.5 py-0.5 font-mono text-xs">
              {props.shellLabel}
            </Badge>
          </div>
        )}
      </div>
      <div className="surface-demo relative overflow-hidden rounded-xl border shadow-[0_14px_36px_-24px_var(--surface-demo)]">
        {props.copyLabel && props.copiedLabel ? (
          <CopyAction
            value={props.code}
            copyLabel={props.copyLabel}
            copiedLabel={props.copiedLabel}
            showLabel
            variant="outline"
            size="sm"
            className="absolute right-2 top-2 h-8 rounded-lg border-[var(--surface-demo-border)] bg-background/10 px-2.5 text-xs text-[var(--surface-demo-foreground)] shadow-sm backdrop-blur hover:bg-background/20 hover:text-[var(--surface-demo-foreground)] sm:right-3 sm:top-3 sm:h-9 sm:rounded-xl sm:px-3"
            onCopied={props.onCopied}
          />
        ) : null}
        <Highlight
          code={props.code}
          language={props.language}
          prism={Prism}
          theme={themes.nightOwl}
        >
          {({ className, style, tokens, getLineProps, getTokenProps }) => (
            <pre
              dir="ltr"
              className={cn("custom-scrollbar max-h-[min(18rem,38dvh)] min-w-0 overflow-auto px-3 py-3 pe-24 text-start font-mono text-xs leading-6 [scrollbar-gutter:stable_both-edges] sm:px-4 sm:py-4 sm:pe-32", className)}
              style={{
                ...style,
                margin: 0,
                background: "transparent",
              }}
            >
              {tokens.map((line, lineIndex) => {
                const lineProps = getLineProps({ line });
                return (
                  <div
                    key={`${props.language}-${lineIndex}`}
                    {...lineProps}
                    className={cn("table w-full table-fixed", lineProps.className)}
                  >
                    <span className="table-cell w-8 select-none border-e border-[var(--surface-demo-border)] pe-2 text-end text-[color-mix(in_oklab,var(--surface-demo-foreground)_62%,transparent)] sm:w-9 sm:pe-3">
                      {lineIndex + 1}
                    </span>
                    <span className="table-cell w-full ps-3">
                      {line.length === 0
                        ? "\n"
                        : line.map((token, tokenIndex) => {
                            const tokenProps = getTokenProps({ token });
                            return <span key={`${lineIndex}-${tokenIndex}`} {...tokenProps} />;
                          })}
                    </span>
                  </div>
                );
              })}
            </pre>
          )}
        </Highlight>
      </div>
      {props.footer ? <div>{props.footer}</div> : null}
    </div>
  );
}


