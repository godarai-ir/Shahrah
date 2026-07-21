"use client";

import * as React from "react";
import { ArrowDown, ArrowUp, Copy, GripVertical, Trash2 } from "lucide-react";

import { AppSelectField } from "@/components/app-ui/AppSelectField";
import { AppTextField } from "@/components/app-ui/AppTextField";
import { Button } from "@/components/ui/button";
import { Field, FieldError } from "@/components/ui/field";
import { Textarea } from "@/components/ui/textarea";
import { cn } from "@/lib/utils";

export type PromptMessageRole = "system" | "user" | "assistant" | "tool";
export type PromptMessageRoleOption = { value: PromptMessageRole; label: string };

export type PromptMessageEditorProps = {
  role: PromptMessageRole;
  content: string;
  roleOptions: PromptMessageRoleOption[];
  roleLabel: string;
  contentLabel: string;
  duplicateLabel: string;
  removeLabel: string;
  dragLabel: string;
  moveUpLabel: string;
  moveDownLabel: string;
  toolCallId?: string;
  toolCallIdLabel: string;
  error?: string;
  disableMoveUp?: boolean;
  disableMoveDown?: boolean;
  readOnly?: boolean;
  onRoleChange: (role: PromptMessageRole) => void;
  onContentChange: (content: string) => void;
  onToolCallIdChange?: (value: string) => void;
  onDuplicate: () => void;
  onRemove: () => void;
  onMoveUp: () => void;
  onMoveDown: () => void;
  onDragStart?: React.DragEventHandler<HTMLSpanElement>;
  onDragEnd?: React.DragEventHandler<HTMLSpanElement>;
  onDragOver?: React.DragEventHandler<HTMLDivElement>;
  onDrop?: React.DragEventHandler<HTMLDivElement>;
  className?: string;
};

export function PromptMessageEditor(props: PromptMessageEditorProps) {
  const errorID = React.useId();
  return (
    <div className={cn("group/message flex flex-col gap-2 border-b border-border py-3 last:border-b-0", props.className)} onDragOver={props.onDragOver} onDrop={props.onDrop}>
      <div className="flex flex-wrap items-center gap-2">
        <span draggable={!props.readOnly} title={props.dragLabel} aria-hidden="true" className={cn("shrink-0 text-muted-foreground", !props.readOnly && "cursor-grab active:cursor-grabbing")} onDragStart={props.onDragStart} onDragEnd={props.onDragEnd}>
          <GripVertical className="size-4" />
        </span>
        <AppSelectField
          value={props.role}
          options={props.roleOptions}
          ariaLabel={props.roleLabel}
          size="sm"
          triggerClassName="w-auto min-w-28"
          disabled={props.readOnly}
          onValueChange={(value) => props.onRoleChange(value as PromptMessageRole)}
        />
        <div className="ms-auto flex items-center gap-1">
          <Button type="button" variant="ghost" size="icon-sm" aria-label={props.moveUpLabel} title={props.moveUpLabel} disabled={props.readOnly || props.disableMoveUp} onClick={props.onMoveUp}><ArrowUp data-icon="inline-start" /></Button>
          <Button type="button" variant="ghost" size="icon-sm" aria-label={props.moveDownLabel} title={props.moveDownLabel} disabled={props.readOnly || props.disableMoveDown} onClick={props.onMoveDown}><ArrowDown data-icon="inline-start" /></Button>
          <Button type="button" variant="ghost" size="icon-sm" aria-label={props.duplicateLabel} title={props.duplicateLabel} disabled={props.readOnly} onClick={props.onDuplicate}><Copy data-icon="inline-start" /></Button>
          <Button type="button" variant="ghost" size="icon-sm" aria-label={props.removeLabel} title={props.removeLabel} disabled={props.readOnly} onClick={props.onRemove}><Trash2 data-icon="inline-start" /></Button>
        </div>
      </div>
      {props.role === "tool" && props.onToolCallIdChange ? (
        <AppTextField
          label={props.toolCallIdLabel}
          dir="ltr"
          value={props.toolCallId ?? ""}
          readOnly={props.readOnly}
          className="h-8 font-mono text-sm"
          onChange={(event) => props.onToolCallIdChange?.(event.target.value)}
        />
      ) : null}
      <Field data-invalid={Boolean(props.error) || undefined}>
        <Textarea
          value={props.content}
          onChange={(event) => props.onContentChange(event.target.value)}
          placeholder={props.contentLabel}
          aria-label={props.contentLabel}
          aria-invalid={Boolean(props.error) || undefined}
          aria-describedby={props.error ? errorID : undefined}
          readOnly={props.readOnly}
          dir="auto"
          className="min-h-24 resize-y border-0 bg-muted/30 shadow-none focus-visible:ring-1"
        />
        {props.error ? <FieldError id={errorID}>{props.error}</FieldError> : null}
      </Field>
    </div>
  );
}
