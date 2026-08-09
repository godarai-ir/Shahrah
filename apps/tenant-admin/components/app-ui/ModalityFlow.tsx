"use client";

import type { ReactNode, SVGProps } from "react";
import {
  ArrowLeft,
  Braces,
  CircleHelp,
  Code2,
  ImageIcon,
  Text,
  Video,
  Volume2,
} from "lucide-react";

import { Badge } from "@/components/ui/badge";
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from "@/components/ui/tooltip";
import { cn } from "@/lib/utils";

export type ModalityFlowProps = {
  inputModalities?: readonly string[];
  outputModalities?: readonly string[];
  inputLabel: string;
  outputLabel: string;
  emptyLabel: string;
  getModalityLabel: (modality: string) => string;
  className?: string;
};

export function ModalityFlow({
  inputModalities,
  outputModalities,
  inputLabel,
  outputLabel,
  emptyLabel,
  getModalityLabel,
  className,
}: ModalityFlowProps) {
  const inputs = uniqueModalities(inputModalities);
  const outputs = uniqueModalities(outputModalities);

  return (
    <TooltipProvider>
      <div className={cn("flex items-center gap-2", className)} dir="rtl">
        <ModalityGroup
          modalities={inputs}
          groupLabel={inputLabel}
          emptyLabel={emptyLabel}
          getModalityLabel={getModalityLabel}
        />
        <ArrowLeft className="size-4 shrink-0 text-muted-foreground" aria-hidden="true" />
        <ModalityGroup
          modalities={outputs}
          groupLabel={outputLabel}
          emptyLabel={emptyLabel}
          getModalityLabel={getModalityLabel}
        />
      </div>
    </TooltipProvider>
  );
}

type ModalityGroupProps = {
  modalities: string[];
  groupLabel: string;
  emptyLabel: string;
  getModalityLabel: (modality: string) => string;
};

function ModalityGroup({ modalities, groupLabel, emptyLabel, getModalityLabel }: ModalityGroupProps) {
  if (modalities.length === 0) {
    return (
      <span className="text-xs text-muted-foreground" aria-label={`${groupLabel}: ${emptyLabel}`}>
        {emptyLabel}
      </span>
    );
  }

  return (
    <span className="flex flex-wrap items-center gap-1" aria-label={groupLabel}>
      {modalities.map((modality) => {
        const label = getModalityLabel(modality);
        const icon = modalityIcon(modality);
        const variant = modalityVariant(modality);
        const accessibleLabel = `${groupLabel}: ${label}`;

        return (
          <Tooltip key={modality}>
            <TooltipTrigger asChild>
              <Badge
                variant={variant}
                tabIndex={0}
                role="img"
                aria-label={accessibleLabel}
                className="size-7 p-0"
              >
                {icon}
              </Badge>
            </TooltipTrigger>
            <TooltipContent dir="auto">{accessibleLabel}</TooltipContent>
          </Tooltip>
        );
      })}
    </span>
  );
}

function uniqueModalities(modalities?: readonly string[]): string[] {
  return [...new Set((modalities ?? []).map((value) => value.trim()).filter(Boolean))];
}

type ModalityVariant = "default" | "secondary" | "destructive" | "success" | "warning" | "info";

function modalityIcon(modality: string): ReactNode {
  const normalized = modality.toLocaleLowerCase();
  if (normalized.includes("image") || normalized.includes("vision")) return <ImageIcon className="size-4" aria-hidden="true" />;
  if (normalized.includes("audio") || normalized.includes("speech")) return <Volume2 className="size-4" aria-hidden="true" />;
  if (normalized.includes("video")) return <Video className="size-4" aria-hidden="true" />;
  if (normalized.includes("pdf")) return <PdfIcon className="size-4" aria-hidden="true" />;
  if (normalized.includes("json")) return <Braces className="size-4" aria-hidden="true" />;
  if (normalized.includes("code")) return <Code2 className="size-4" aria-hidden="true" />;
  if (normalized.includes("text")) return <Text className="size-4" aria-hidden="true" />;

  return <CircleHelp className="size-4" aria-hidden="true" />;
}

function modalityVariant(modality: string): ModalityVariant {
  const normalized = modality.toLocaleLowerCase();
  if (normalized.includes("pdf")) return "destructive";
  if (normalized.includes("audio") || normalized.includes("speech")) return "success";
  if (normalized.includes("video")) return "warning";
  if (normalized.includes("image") || normalized.includes("vision")) return "info";
  if (normalized.includes("text")) return "default";

  return "secondary";
}

function PdfIcon(props: SVGProps<SVGSVGElement>) {
  return (
    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.75" strokeLinecap="round" strokeLinejoin="round" focusable="false" {...props}>
      <path d="M6 2.75h8l4 4v14.5H6z" />
      <path d="M14 2.75v4h4" />
      <text x="12" y="16.25" fill="currentColor" stroke="none" fontSize="5.5" fontWeight="700" textAnchor="middle">PDF</text>
    </svg>
  );
}
