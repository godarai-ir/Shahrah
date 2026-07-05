"use client";

import { Highlight, themes } from "prism-react-renderer";
import { Prism } from "prismjs";
import "prismjs/components/prism-bash";
import "prismjs/components/prism-javascript";
import "prismjs/components/prism-json";
import "prismjs/components/prism-markdown";
import "prismjs/components/prism-python";

import { cn } from "@/lib/utils";

export function CodeBlock(props: {
  code: string;
  language: "bash" | "javascript" | "json" | "markdown" | "python";
  className?: string;
  highlightedTerms?: string[];
  maxHeightClassName?: string;
  preClassName?: string;
}) {
  const highlightedTerms = props.highlightedTerms?.filter(Boolean) ?? [];

  return (
    <Highlight
      code={props.code}
      language={props.language}
      prism={Prism}
      theme={themes.github}
    >
      {({ className, style, tokens, getLineProps, getTokenProps }) => (
        <div
          className={cn(
            "overflow-hidden rounded-2xl border border-code-border bg-code text-code-foreground shadow-sm",
            props.className,
          )}
        >
          <pre
            dir="ltr"
            className={cn(
              "custom-scrollbar overflow-auto px-4 py-4 text-left text-[13px] leading-7 [scrollbar-gutter:stable_both-edges] sm:text-sm sm:leading-7",
              props.maxHeightClassName ?? "max-h-[min(22rem,42dvh)]",
              props.preClassName,
              className,
            )}
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
                  <span
                    className="table-cell w-8 select-none pe-3 text-right text-code-muted-foreground sm:w-10 sm:pe-4"
                  >
                    {lineIndex + 1}
                  </span>
                  <span className="table-cell w-full break-words">
                    {line.length === 0
                      ? "\n"
                      : line.map((token, tokenIndex) => {
                          const tokenProps = getTokenProps({ token });
                          return (
                            <span key={`${lineIndex}-${tokenIndex}`} {...tokenProps}>
                              {renderHighlightedToken(token.content, highlightedTerms)}
                            </span>
                          );
                        })}
                  </span>
                </div>
              );
            })}
          </pre>
        </div>
      )}
    </Highlight>
  );
}

function renderHighlightedToken(content: string, highlightedTerms: string[]) {
  if (highlightedTerms.length === 0) return content;

  const matchingTerm = highlightedTerms.find((term) => content.includes(term));
  if (!matchingTerm) return content;

  return content.split(matchingTerm).map((part, index, parts) => (
    <FragmentWithHighlight
      key={`${matchingTerm}-${index}-${part}`}
      highlight={index < parts.length - 1 ? matchingTerm : null}
      text={part}
    />
  ));
}

function FragmentWithHighlight(props: {
  highlight: string | null;
  text: string;
}) {
  return (
    <>
      {props.text}
      {props.highlight ? (
        <mark className="bg-transparent font-semibold text-code-foreground">
          {props.highlight}
        </mark>
      ) : null}
    </>
  );
}
