"use client";

import * as React from "react";
import { json } from "@codemirror/lang-json";
import CodeMirror, { EditorView } from "@uiw/react-codemirror";

import { cn } from "@/lib/utils";

const jsonViewerTheme = EditorView.theme({
  "&": {
    width: "100%",
    direction: "ltr",
    border: "1px solid var(--code-border)",
    borderRadius: "var(--radius-lg)",
    backgroundColor: "var(--code)",
    color: "var(--code-foreground)",
    fontFamily: "var(--font-mono)",
    fontSize: "var(--font-size-sm)",
    boxShadow: "var(--shadow-xs)",
  },
  "&.cm-focused": {
    outline: "none",
    borderColor: "var(--input-focus-border)",
    boxShadow: "0 0 0 3px color-mix(in oklab, var(--ring) 20%, transparent)",
  },
  ".cm-scroller": {
    fontFamily: "inherit",
    lineHeight: "1.75rem",
    scrollbarColor: "color-mix(in oklab, var(--code-muted-foreground) 24%, transparent) transparent",
  },
  ".cm-content": {
    padding: "var(--space-3) 0",
    caretColor: "transparent",
    textAlign: "left",
  },
  ".cm-content:focus-visible": {
    outline: "none",
  },
  ".cm-line": {
    padding: "0 var(--space-3)",
  },
  ".cm-gutters": {
    borderInlineEnd: "1px solid var(--code-border)",
    backgroundColor: "var(--code)",
    color: "var(--code-muted-foreground)",
  },
  ".cm-lineNumbers .cm-gutterElement": {
    minWidth: "var(--space-8)",
    padding: "0 var(--space-2)",
  },
  ".cm-foldGutter .cm-gutterElement": {
    color: "var(--code-muted-foreground)",
  },
  ".cm-selectionBackground, &.cm-focused .cm-selectionBackground, .cm-content ::selection": {
    backgroundColor: "var(--primary-soft)",
  },
  ".cm-matchingBracket": {
    backgroundColor: "var(--primary-soft)",
    color: "var(--primary-soft-foreground)",
    outline: "1px solid var(--primary-subtle-border)",
  },
  ".cm-panels, .cm-tooltip": {
    borderColor: "var(--code-border)",
    backgroundColor: "var(--card)",
    color: "var(--card-foreground)",
  },
});

export type JsonViewerProps = Omit<
  React.ComponentProps<typeof CodeMirror>,
  "basicSetup" | "dir" | "editable" | "extensions" | "indentWithTab" | "onChange" | "readOnly" | "theme" | "value"
> & {
  value: unknown;
  containerClassName?: string;
};

export function JsonViewer({
  value,
  containerClassName,
  className,
  maxHeight = "min(34rem, 60dvh)",
  id,
  "aria-label": ariaLabel,
  "aria-labelledby": ariaLabelledBy,
  "aria-describedby": ariaDescribedBy,
  ...props
}: JsonViewerProps) {
  const accessibilityExtension = React.useMemo(
    () => EditorView.contentAttributes.of(compactAttributes({
      id,
      dir: "ltr",
      spellcheck: "false",
      autocorrect: "off",
      autocapitalize: "off",
      "aria-label": ariaLabel,
      "aria-labelledby": ariaLabelledBy,
      "aria-describedby": ariaDescribedBy,
      "aria-readonly": true,
      "aria-multiline": true,
      tabindex: 0,
    })),
    [ariaDescribedBy, ariaLabel, ariaLabelledBy, id],
  );
  const extensions = React.useMemo(
    () => [json(), EditorView.lineWrapping, jsonViewerTheme, accessibilityExtension],
    [accessibilityExtension],
  );

  return (
    <div className={cn("min-w-0", containerClassName)}>
      <CodeMirror
        {...props}
        value={serializeJson(value)}
        readOnly
        editable={false}
        theme="none"
        indentWithTab={false}
        maxHeight={maxHeight}
        basicSetup={{
          lineNumbers: true,
          highlightActiveLineGutter: false,
          highlightSpecialChars: true,
          history: false,
          foldGutter: true,
          drawSelection: true,
          dropCursor: false,
          allowMultipleSelections: false,
          indentOnInput: false,
          syntaxHighlighting: true,
          bracketMatching: true,
          closeBrackets: false,
          autocompletion: false,
          highlightActiveLine: false,
          highlightSelectionMatches: true,
          searchKeymap: true,
          historyKeymap: false,
          closeBracketsKeymap: false,
          lintKeymap: false,
        }}
        extensions={extensions}
        className={cn("min-h-0", className)}
        data-slot="json-viewer"
      />
    </div>
  );
}

export function serializeJson(value: unknown): string {
  try {
    return JSON.stringify(value ?? {}, null, 2) ?? String(value);
  } catch {
    return String(value);
  }
}

function compactAttributes(attributes: Record<string, string | number | boolean | undefined>): Record<string, string> {
  return Object.fromEntries(
    Object.entries(attributes)
      .filter(([, value]) => value !== undefined)
      .map(([key, value]) => [key, String(value)]),
  );
}
