"use client";

import * as React from "react";
import stringify from "json-stringify-pretty-compact";
import { Highlight, themes } from "prism-react-renderer";
import { Prism } from "prismjs";
import "prismjs/components/prism-json";

import { getVariableColorStyle } from "@/components/app-ui/VariableTextarea";
import { Field, FieldDescription, FieldError, FieldLabel } from "@/components/ui/field";
import { Textarea } from "@/components/ui/textarea";
import { cn } from "@/lib/utils";

const VARIABLE_PATTERN = /\{\{\s*([A-Za-z_][A-Za-z0-9_]*)\s*\}\}/g;

export type JsonEditorProps = Omit<React.ComponentProps<typeof Textarea>, "children" | "defaultValue" | "dir" | "onChange" | "value"> & {
  value: string;
  onValueChange: (value: string) => void;
  variables?: readonly string[];
  containerClassName?: string;
  overlayClassName?: string;
  formatOnBlur?: boolean;
};

export type JsonEditorFieldProps = JsonEditorProps & {
  label: React.ReactNode;
  description?: React.ReactNode;
  error?: React.ReactNode;
  fieldClassName?: string;
  hideLabel?: boolean;
};

export function JsonEditor({
  value,
  onValueChange,
  variables = [],
  containerClassName,
  overlayClassName,
  className,
  disabled,
  readOnly,
  formatOnBlur = true,
  onBlur,
  onScroll,
  ...props
}: JsonEditorProps) {
  const [scrollPosition, setScrollPosition] = React.useState({ left: 0, top: 0 });

  return (
    <div className={cn("relative rounded-md", disabled && "opacity-50", containerClassName)}>
      <Textarea
        {...props}
        value={value}
        dir="ltr"
        disabled={disabled}
        readOnly={readOnly}
        spellCheck={false}
        wrap="soft"
        onChange={(event) => onValueChange(event.currentTarget.value)}
        onBlur={(event) => {
          if (formatOnBlur && !disabled && !readOnly) {
            const formatted = formatJson(value);
            if (formatted !== value) onValueChange(formatted);
          }
          onBlur?.(event);
        }}
        onScroll={(event) => {
          setScrollPosition({ left: event.currentTarget.scrollLeft, top: event.currentTarget.scrollTop });
          onScroll?.(event);
        }}
        className={cn(
          "peer relative bg-transparent text-start font-mono text-xs leading-6 text-transparent caret-foreground selection:bg-primary/20 selection:text-transparent focus-visible:!outline-none focus-visible:!ring-0 focus-visible:!ring-offset-0 disabled:opacity-100",
          className,
        )}
      />
      <Highlight code={value} language="json" prism={Prism} theme={themes.github}>
        {({ className: highlightClassName, style, tokens, getLineProps, getTokenProps }) => (
          <pre
            aria-hidden="true"
            dir="ltr"
            className={cn(
              "pointer-events-none absolute inset-0 overflow-hidden rounded-md px-3 py-2 text-start font-mono text-xs leading-6 transition-[color,box-shadow] peer-focus-visible:ring-[3px] peer-focus-visible:ring-inset peer-focus-visible:ring-ring/50",
              highlightClassName,
              overlayClassName,
            )}
            style={{ ...style, margin: 0, background: "transparent" }}
          >
            <div style={{ transform: `translate(${-scrollPosition.left}px, ${-scrollPosition.top}px)` }}>
              {tokens.map((line, lineIndex) => {
                const lineProps = getLineProps({ line });
                return (
                  <div key={`json-${lineIndex}`} {...lineProps} className={cn("min-h-6 whitespace-pre-wrap break-words", lineProps.className)}>
                    {line.length === 0 ? "\n" : line.map((token, tokenIndex) => {
                      const tokenProps = getTokenProps({ token });
                      return (
                        <span key={`${lineIndex}-${tokenIndex}`} {...tokenProps}>
                          {renderJsonToken(token.content, variables)}
                        </span>
                      );
                    })}
                  </div>
                );
              })}
            </div>
          </pre>
        )}
      </Highlight>
    </div>
  );
}

export function JsonEditorField({ label, description, error, fieldClassName, hideLabel, id, disabled, "aria-describedby": ariaDescribedBy, ...props }: JsonEditorFieldProps) {
  const generatedID = React.useId();
  const controlID = id ?? generatedID;
  const descriptionID = description ? `${controlID}-description` : undefined;
  const errorID = error ? `${controlID}-error` : undefined;
  const describedBy = [ariaDescribedBy, descriptionID, errorID].filter(Boolean).join(" ") || undefined;

  return (
    <Field data-invalid={Boolean(error) || undefined} data-disabled={disabled || undefined} className={fieldClassName}>
      <FieldLabel htmlFor={controlID} className={cn(hideLabel && "sr-only")}>{label}</FieldLabel>
      <JsonEditor
        {...props}
        id={controlID}
        disabled={disabled}
        aria-invalid={Boolean(error) || undefined}
        aria-describedby={describedBy}
      />
      {description ? <FieldDescription id={descriptionID}>{description}</FieldDescription> : null}
      {error ? <FieldError id={errorID}>{error}</FieldError> : null}
    </Field>
  );
}

export function formatJson(value: string): string {
  try {
    return stringify(JSON.parse(value), { indent: 2, maxLength: 100 });
  } catch {
    return value;
  }
}

function renderJsonToken(content: string, variables: readonly string[]): React.ReactNode {
  if (!content || variables.length === 0) return content;
  const allowedVariables = new Set(variables);
  const parts: React.ReactNode[] = [];
  let lastIndex = 0;

  for (const match of content.matchAll(VARIABLE_PATTERN)) {
    const placeholder = match[0];
    const name = match[1]?.trim() ?? "";
    const index = match.index ?? 0;
    if (!allowedVariables.has(name)) continue;
    if (index > lastIndex) parts.push(content.slice(lastIndex, index));
    parts.push(<span key={`${index}-${placeholder}`} className="variable-placeholder font-semibold" style={getVariableColorStyle(name, variables)}>{placeholder}</span>);
    lastIndex = index + placeholder.length;
  }

  if (parts.length === 0) return content;
  if (lastIndex < content.length) parts.push(content.slice(lastIndex));
  return parts;
}
