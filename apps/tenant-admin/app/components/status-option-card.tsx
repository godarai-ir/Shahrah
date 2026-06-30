"use client";

import type { LucideIcon } from "lucide-react";
import { CheckCircle2 } from "lucide-react";

import { cn } from "@/lib/utils";

type Tone = "neutral" | "info" | "warning" | "success";

const toneClasses: Record<Tone, string> = {
  neutral: "tone-neutral",
  info: "tone-info",
  warning: "tone-warning",
  success: "tone-success",
};

type StatusOptionCardProps = {
  badge: string;
  description: string;
  icon: LucideIcon;
  selected: boolean;
  title: string;
  tone: Tone;
  disabled?: boolean;
  onClick: () => void;
};

export function StatusOptionCard(props: StatusOptionCardProps) {
  const Icon = props.icon;

  return (
    <button
      type="button"
      onClick={props.onClick}
      disabled={props.disabled}
      aria-pressed={props.selected}
      className={cn(
        "group rounded-xl border p-4 text-right transition-all",
        "hover:shadow-sm disabled:cursor-not-allowed disabled:opacity-70",
        props.selected
          ? cn(toneClasses[props.tone], "ring-2 ring-primary/20")
          : "border-border bg-background hover:border-primary/30 hover:bg-muted/30",
      )}
    >
      <div className="flex items-start justify-between gap-3">
        <div
          className={cn(
            "flex size-10 shrink-0 items-center justify-center rounded-lg border transition-colors",
            props.selected ? toneClasses[props.tone] : "border-border bg-muted/40 text-muted-foreground",
          )}
        >
          <Icon className="size-5" />
        </div>
        <div className="flex min-w-0 items-center gap-2">
          {props.selected ? <CheckCircle2 className="size-4 text-primary" /> : null}
          <span
            className={cn(
              "rounded-full border px-2 py-0.5 text-[11px] font-medium transition-colors",
              props.selected ? "surface-primary-soft" : "border-transparent bg-muted text-muted-foreground",
            )}
          >
            {props.badge}
          </span>
        </div>
      </div>
      <div className="mt-3 space-y-1">
        <p className="text-sm font-semibold text-foreground">{props.title}</p>
        <p className="text-xs leading-6 text-muted-foreground">{props.description}</p>
      </div>
    </button>
  );
}
