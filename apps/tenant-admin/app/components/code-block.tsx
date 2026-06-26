"use client";

import { Highlight, themes } from "prism-react-renderer";
import { Prism } from "prismjs";
import "prismjs/components/prism-bash";
import "prismjs/components/prism-json";
import "prismjs/components/prism-markdown";
import "prismjs/components/prism-python";

import { cn } from "@/lib/utils";

export function CodeBlock(props: {
  code: string;
  language: "bash" | "json" | "markdown" | "python";
  className?: string;
  maxHeightClassName?: string;
}) {
  return (
    <Highlight
      code={props.code}
      language={props.language}
      prism={Prism}
      theme={themes.nightOwl}
    >
      {({ className, style, tokens, getLineProps, getTokenProps }) => (
        <pre
          dir="ltr"
          className={cn(
            "custom-scrollbar overflow-auto rounded-xl border border-slate-800 bg-slate-950 p-4 text-left text-[13px] leading-7 sm:text-sm sm:leading-7",
            props.maxHeightClassName ?? "max-h-[min(22rem,42dvh)]",
            className,
            props.className,
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
                <span className="table-cell w-8 select-none pe-3 text-right text-slate-400 sm:w-10 sm:pe-4">
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
      )}
    </Highlight>
  );
}
