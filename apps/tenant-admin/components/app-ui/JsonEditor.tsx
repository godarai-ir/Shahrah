"use client";

import * as React from "react";
import { json, jsonParseLinter } from "@codemirror/lang-json";
import { linter } from "@codemirror/lint";
import CodeMirror, {
  Decoration,
  type DecorationSet,
  EditorView,
  type Extension,
  MatchDecorator,
  ViewPlugin,
  type ViewUpdate,
} from "@uiw/react-codemirror";
import stringify from "json-stringify-pretty-compact";

import { getVariableColorStyle } from "@/components/app-ui/VariableTextarea";
import { Field, FieldDescription, FieldError, FieldLabel } from "@/components/ui/field";
import { cn } from "@/lib/utils";

const VARIABLE_PATTERN = /\{\{\s*([A-Za-z_][A-Za-z0-9_]*)\s*\}\}/g;

const jsonEditorTheme = EditorView.theme({
  "&": {
    minHeight: "inherit",
    width: "100%",
    direction: "ltr",
    border: "1px solid var(--input)",
    borderRadius: "var(--radius-md)",
    backgroundColor: "var(--input-background)",
    color: "var(--foreground)",
    boxShadow: "var(--shadow-xs)",
    fontFamily: "var(--font-mono)",
    fontSize: "var(--font-size-xs)",
    transition: "border-color var(--duration-fast) var(--ease-out), box-shadow var(--duration-fast) var(--ease-out)",
  },
  "&.cm-focused": {
    outline: "none",
    borderColor: "var(--input-focus-border)",
    boxShadow: "0 0 0 3px color-mix(in oklab, var(--ring) 50%, transparent)",
  },
  ".cm-scroller": {
    minHeight: "inherit",
    fontFamily: "inherit",
    lineHeight: "1.5rem",
    scrollbarColor: "color-mix(in oklab, var(--muted-foreground) 24%, transparent) transparent",
  },
  ".cm-content": {
    minHeight: "inherit",
    padding: "var(--space-2) 0",
    caretColor: "var(--foreground)",
    textAlign: "left",
  },
  ".cm-content:focus-visible": {
    outline: "none",
    boxShadow: "none",
  },
  ".cm-line": {
    padding: "0 var(--space-3)",
  },
  ".cm-gutters": {
    borderInlineEnd: "1px solid var(--border)",
    backgroundColor: "var(--muted)",
    color: "var(--code-muted-foreground)",
  },
  ".cm-lineNumbers .cm-gutterElement": {
    minWidth: "var(--space-6)",
    padding: "0 var(--space-2)",
  },
  ".cm-activeLine, .cm-activeLineGutter": {
    backgroundColor: "var(--primary-soft)",
  },
  ".cm-selectionBackground, &.cm-focused .cm-selectionBackground, .cm-content ::selection": {
    backgroundColor: "var(--primary-soft)",
  },
  ".cm-cursor, .cm-dropCursor": {
    borderLeftColor: "var(--foreground)",
  },
  ".cm-matchingBracket": {
    backgroundColor: "var(--primary-soft)",
    color: "var(--primary-soft-foreground)",
    outline: "1px solid var(--primary-subtle-border)",
  },
  ".cm-lintRange-error": {
    backgroundImage: "none",
    borderBottom: "2px wavy var(--destructive)",
  },
  ".cm-tooltip": {
    border: "1px solid var(--border)",
    borderRadius: "var(--radius-md)",
    backgroundColor: "var(--card)",
    color: "var(--card-foreground)",
    boxShadow: "var(--shadow-md)",
  },
  ".cm-diagnostic-error": {
    borderInlineStartColor: "var(--destructive)",
  },
});

const jsonEditorInvalidTheme = EditorView.theme({
  "&": {
    borderColor: "var(--destructive)",
  },
  "&.cm-focused": {
    borderColor: "var(--destructive)",
    boxShadow: "0 0 0 3px color-mix(in oklab, var(--destructive) 20%, transparent)",
  },
});

const jsonEditorDisabledTheme = EditorView.theme({
  "&": {
    opacity: "0.5",
  },
  ".cm-content": {
    cursor: "not-allowed",
  },
});

export type JsonEditorProps = Omit<
  React.ComponentProps<typeof CodeMirror>,
  "basicSetup" | "dir" | "editable" | "extensions" | "indentWithTab" | "onChange" | "theme" | "value"
> & {
  value: string;
  onValueChange: (value: string) => void;
  variables?: readonly string[];
  containerClassName?: string;
  disabled?: boolean;
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
  className,
  disabled = false,
  readOnly = false,
  formatOnBlur = true,
  id,
  onBlur,
  "aria-label": ariaLabel,
  "aria-labelledby": ariaLabelledBy,
  "aria-describedby": ariaDescribedBy,
  "aria-invalid": ariaInvalid,
  "aria-required": ariaRequired,
  ...props
}: JsonEditorProps) {
  const variableExtension = React.useMemo(() => createVariableHighlightExtension(variables), [variables]);
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
      "aria-invalid": ariaInvalid,
      "aria-required": ariaRequired,
      "aria-readonly": readOnly || disabled,
      "aria-disabled": disabled,
      "aria-multiline": true,
      tabindex: disabled ? -1 : 0,
    })),
    [ariaDescribedBy, ariaInvalid, ariaLabel, ariaLabelledBy, ariaRequired, disabled, id, readOnly],
  );
  const extensions = React.useMemo(
    () => [
      json(),
      linter(jsonParseLinter(), { delay: 250 }),
      EditorView.lineWrapping,
      jsonEditorTheme,
      accessibilityExtension,
      variableExtension,
      ...(isAriaInvalid(ariaInvalid) ? [jsonEditorInvalidTheme] : []),
      ...(disabled ? [jsonEditorDisabledTheme] : []),
    ],
    [accessibilityExtension, ariaInvalid, disabled, variableExtension],
  );

  return (
    <div className={cn("min-w-0 rounded-md", disabled && "cursor-not-allowed", containerClassName)}>
      <CodeMirror
        {...props}
        value={value}
        readOnly={readOnly || disabled}
        editable={!disabled}
        theme="none"
        indentWithTab={false}
        basicSetup={{
          lineNumbers: true,
          highlightActiveLineGutter: true,
          highlightSpecialChars: true,
          history: true,
          foldGutter: false,
          drawSelection: true,
          dropCursor: true,
          allowMultipleSelections: true,
          indentOnInput: true,
          syntaxHighlighting: true,
          bracketMatching: true,
          closeBrackets: true,
          autocompletion: false,
          highlightActiveLine: true,
          highlightSelectionMatches: true,
          searchKeymap: true,
          historyKeymap: true,
          closeBracketsKeymap: true,
          lintKeymap: true,
        }}
        extensions={extensions}
        onChange={onValueChange}
        onBlur={(event) => {
          if (event.currentTarget.contains(event.relatedTarget as Node | null)) return;
          if (formatOnBlur && !disabled && !readOnly) {
            const formatted = formatJson(value);
            if (formatted !== value) onValueChange(formatted);
          }
          onBlur?.(event);
        }}
        className={cn("min-h-40", className)}
        data-slot="json-editor"
      />
    </div>
  );
}

export function JsonEditorField({
  label,
  description,
  error,
  fieldClassName,
  hideLabel,
  id,
  disabled,
  "aria-describedby": ariaDescribedBy,
  "aria-labelledby": ariaLabelledBy,
  ...props
}: JsonEditorFieldProps) {
  const generatedID = React.useId();
  const controlID = id ?? generatedID;
  const labelID = `${controlID}-label`;
  const descriptionID = description ? `${controlID}-description` : undefined;
  const errorID = error ? `${controlID}-error` : undefined;
  const describedBy = [ariaDescribedBy, descriptionID, errorID].filter(Boolean).join(" ") || undefined;
  const labelledBy = [ariaLabelledBy, labelID].filter(Boolean).join(" ");

  return (
    <Field data-invalid={Boolean(error) || undefined} data-disabled={disabled || undefined} className={fieldClassName}>
      <FieldLabel id={labelID} htmlFor={controlID} className={cn(hideLabel && "sr-only")}>{label}</FieldLabel>
      <JsonEditor
        {...props}
        id={controlID}
        disabled={disabled}
        aria-labelledby={labelledBy}
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

function createVariableHighlightExtension(variables: readonly string[]): Extension {
  if (variables.length === 0) return [];

  const allowedVariables = new Set(variables);
  const matcher = new MatchDecorator({
    regexp: new RegExp(VARIABLE_PATTERN.source, "g"),
    decoration: (match) => {
      const name = match[1]?.trim() ?? "";
      if (!allowedVariables.has(name)) return null;
      const hue = getVariableColorStyle(name, variables)["--variable-hue"];
      return Decoration.mark({
        class: "variable-placeholder font-semibold",
        attributes: { style: `--variable-hue: ${hue}` },
      });
    },
  });

  return ViewPlugin.fromClass(
    class {
      decorations: DecorationSet;

      constructor(view: EditorView) {
        this.decorations = matcher.createDeco(view);
      }

      update(update: ViewUpdate) {
        this.decorations = matcher.updateDeco(update, this.decorations);
      }
    },
    { decorations: (instance) => instance.decorations },
  );
}

function compactAttributes(attributes: Record<string, React.AriaAttributes[keyof React.AriaAttributes] | string | number | boolean | undefined>): Record<string, string> {
  return Object.fromEntries(
    Object.entries(attributes)
      .filter(([, value]) => value !== undefined)
      .map(([key, value]) => [key, String(value)]),
  );
}

function isAriaInvalid(value: React.AriaAttributes["aria-invalid"]): boolean {
  return value === true || value === "true" || value === "grammar" || value === "spelling";
}
