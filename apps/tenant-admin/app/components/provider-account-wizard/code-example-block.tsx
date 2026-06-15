"use client";

import { Highlight, themes } from "prism-react-renderer";
import { Prism } from "prismjs";
import "prismjs/components/prism-bash";
import "prismjs/components/prism-json";
import "prismjs/components/prism-powershell";

import { Badge } from "@/components/ui/badge";
import { cn } from "@/lib/utils";
export function CodeExampleBlock(props: {
  code: string;
  language: "powershell" | "bash" | "json";
  shellLabel: string;
  className?: string;
}) {
  return (
    <div
      className={cn(
        "flex max-w-full min-w-0 flex-col overflow-hidden rounded-lg border border-slate-800 bg-slate-950 shadow-[0_20px_64px_-32px_rgba(15,23,42,0.85)]",
        props.className,
      )}
    >
      <div className="flex shrink-0 items-center justify-between border-b border-slate-800 bg-slate-900/80 px-3 py-2.5 text-slate-200 sm:px-4 sm:py-3">
        <div className="flex items-center gap-3">
          <div className="flex items-center gap-1.5">
            <span className="size-2 rounded-full bg-rose-400 sm:size-2.5" />
            <span className="size-2 rounded-full bg-amber-300 sm:size-2.5" />
            <span className="size-2 rounded-full bg-emerald-400 sm:size-2.5" />
          </div>
          <Badge variant="secondary" className="border border-slate-700 bg-slate-800 px-2.5 py-0.5 font-mono text-[11px] text-slate-200">
            {props.shellLabel}
          </Badge>
        </div>
      </div>
      <Highlight
        code={props.code}
        language={props.language}
        prism={Prism}
        theme={themes.nightOwl}
      >
        {({ className, style, tokens, getLineProps, getTokenProps }) => (
          <pre
            dir="ltr"
            className={cn("custom-scrollbar max-h-[min(22rem,42dvh)] min-w-0 overflow-auto p-3 text-left text-[12px] leading-6 sm:p-4 sm:text-[13px] sm:leading-7", className)}
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
                  <span className="table-cell w-8 select-none pe-3 text-right text-slate-500 sm:w-10 sm:pe-4">
                    {lineIndex + 1}
                  </span>
                  <span className="table-cell w-full">
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
  );
}


