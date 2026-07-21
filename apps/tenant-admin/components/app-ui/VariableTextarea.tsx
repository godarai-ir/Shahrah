"use client";

import * as React from "react";

import { Field, FieldDescription, FieldError, FieldLabel } from "@/components/ui/field";
import { Textarea } from "@/components/ui/textarea";
import { cn } from "@/lib/utils";

const VARIABLE_PATTERN = /\{\{\s*([A-Za-z_][A-Za-z0-9_]*)\s*\}\}/g;
const VARIABLE_HUE_START = 264;
const VARIABLE_HUE_STEP = 137.508;

export type VariableColorStyle = React.CSSProperties & { "--variable-hue": string };

export type VariableTextareaProps = React.ComponentProps<typeof Textarea> & {
  variables: readonly string[];
  containerClassName?: string;
  overlayClassName?: string;
};

export type VariableTextareaFieldProps = VariableTextareaProps & {
  label: React.ReactNode;
  description?: React.ReactNode;
  error?: React.ReactNode;
  fieldClassName?: string;
  hideLabel?: boolean;
};

export function VariableTextarea({ variables, containerClassName, overlayClassName, className, value, defaultValue, dir, disabled, onScroll, ...props }: VariableTextareaProps) {
  const [scrollPosition, setScrollPosition] = React.useState({ left: 0, top: 0 });
  const text = value === undefined ? String(defaultValue ?? "") : String(value ?? "");

  return (
    <div className={cn("relative rounded-md", disabled && "opacity-50", containerClassName)}>
      <Textarea
        {...props}
        value={value}
        defaultValue={defaultValue}
        dir={dir}
        disabled={disabled}
        onScroll={(event) => {
          setScrollPosition({ left: event.currentTarget.scrollLeft, top: event.currentTarget.scrollTop });
          onScroll?.(event);
        }}
        className={cn(
          "peer relative bg-transparent text-transparent caret-foreground selection:bg-primary/20 selection:text-transparent focus-visible:!outline-none focus-visible:!ring-0 focus-visible:!ring-offset-0 disabled:opacity-100",
          className,
        )}
      />
      <div
        aria-hidden="true"
        dir={dir}
        className={cn(
          "pointer-events-none absolute inset-0 overflow-hidden rounded-md px-3 py-2 text-base leading-6 text-foreground transition-[color,box-shadow] peer-focus-visible:ring-[3px] peer-focus-visible:ring-inset peer-focus-visible:ring-ring/50 md:text-sm",
          overlayClassName,
        )}
      >
        <div
          className="min-h-full whitespace-pre-wrap break-words"
          style={{ transform: `translate(${-scrollPosition.left}px, ${-scrollPosition.top}px)` }}
        >
          {renderVariableText(text, variables)}
        </div>
      </div>
    </div>
  );
}

export function VariableTextareaField({ label, description, error, fieldClassName, hideLabel, id, disabled, "aria-describedby": ariaDescribedBy, ...props }: VariableTextareaFieldProps) {
  const generatedID = React.useId();
  const controlID = id ?? generatedID;
  const descriptionID = description ? `${controlID}-description` : undefined;
  const errorID = error ? `${controlID}-error` : undefined;
  const describedBy = [ariaDescribedBy, descriptionID, errorID].filter(Boolean).join(" ") || undefined;

  return (
    <Field data-invalid={Boolean(error) || undefined} data-disabled={disabled || undefined} className={fieldClassName}>
      <FieldLabel htmlFor={controlID} className={cn(hideLabel && "sr-only")}>{label}</FieldLabel>
      <VariableTextarea
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

export function getVariableColorStyle(name: string, variables: readonly string[]): VariableColorStyle {
  const index = Math.max(0, variables.indexOf(name));
  const hue = (VARIABLE_HUE_START + index * VARIABLE_HUE_STEP) % 360;
  return { "--variable-hue": hue.toFixed(3) };
}

function renderVariableText(text: string, variables: readonly string[]): React.ReactNode[] {
  if (!text) return [];
  const allowedVariables = new Set(variables);
  const parts: React.ReactNode[] = [];
  let lastIndex = 0;

  for (const match of text.matchAll(VARIABLE_PATTERN)) {
    const placeholder = match[0];
    const name = match[1]?.trim() ?? "";
    const index = match.index ?? 0;
    if (index > lastIndex) parts.push(text.slice(lastIndex, index));
    parts.push(
      allowedVariables.has(name)
        ? <span key={`${index}-${placeholder}`} className="variable-placeholder font-semibold" style={getVariableColorStyle(name, variables)}>{placeholder}</span>
        : placeholder,
    );
    lastIndex = index + placeholder.length;
  }

  if (lastIndex < text.length) parts.push(text.slice(lastIndex));
  return parts;
}
