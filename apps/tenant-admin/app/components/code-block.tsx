"use client";

import { Highlight, themes } from "prism-react-renderer";
import { Prism } from "prismjs";
import "prismjs/components/prism-bash";
import "prismjs/components/prism-javascript";
import "prismjs/components/prism-json";
import "prismjs/components/prism-markdown";
import "prismjs/components/prism-python";

import { useTheme } from "@/components/theme-provider";
import { cn } from "@/lib/utils";

export function CodeBlock(props: {
  code: string;
  language: "bash" | "javascript" | "json" | "markdown" | "python";
  className?: string;
  maxHeightClassName?: string;
}) {
  const { theme } = useTheme();
  const prismTheme = theme === "dark" ? themes.nightOwl : themes.github;

  return (
    <Highlight
      code={props.code}
      language={props.language}
      prism={Prism}
      theme={prismTheme}
    >
      {({ className, style, tokens, getLineProps, getTokenProps }) => (
        <div
          className={cn(
            "overflow-hidden rounded-2xl border shadow-sm",
            theme === "dark"
              ? "border-slate-800 bg-slate-950 text-slate-100"
              : "border-slate-200 bg-slate-50 text-slate-900",
            props.className,
          )}
        >
          <pre
            dir="ltr"
            className={cn(
              "custom-scrollbar overflow-auto px-4 py-4 text-left text-[13px] leading-7 [scrollbar-gutter:stable_both-edges] sm:text-sm sm:leading-7",
              props.maxHeightClassName ?? "max-h-[min(22rem,42dvh)]",
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
                    className={cn(
                      "table-cell w-8 select-none pe-3 text-right sm:w-10 sm:pe-4",
                      theme === "dark" ? "text-slate-500" : "text-slate-400",
                    )}
                  >
                    {lineIndex + 1}
                  </span>
                  <span className="table-cell w-full break-words">
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
        </div>
      )}
    </Highlight>
  );
}
