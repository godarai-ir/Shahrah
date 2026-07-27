import { type ComponentType, type SVGProps } from "react";
import {
  AudioLines,
  ArrowLeft,
  Code2,
  FileText,
  Image as ImageIcon,
  Video,
} from "lucide-react";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "@/components/ui/tooltip";
import { cn } from "@/lib/utils";

type ModalityIcon = ComponentType<SVGProps<SVGSVGElement>>;
type ModalityMeta = { icon: ModalityIcon; className: string };

function getModalityMeta(value: string): ModalityMeta {
  const key = value.toLowerCase();
  if (key.includes("pdf")) return { icon: PdfFileIcon, className: "border-border bg-accent text-accent-foreground" };
  if (key.includes("text")) return { icon: FileText, className: "border-primary-soft bg-primary-soft text-primary-soft-foreground" };
  if (key.includes("image")) return { icon: ImageIcon, className: "border-border bg-accent text-accent-foreground" };
  if (key.includes("audio") || key.includes("speech") || key.includes("voice")) return { icon: AudioLines, className: "border-warning-soft bg-warning-soft text-warning-soft-foreground" };
  if (key.includes("video")) return { icon: Video, className: "border-destructive-soft bg-destructive-soft text-destructive-soft-foreground" };
  if (key.includes("embedding")) return { icon: Code2, className: "border-success-soft bg-success-soft text-success-soft-foreground" };
  return { icon: Code2, className: "border-border bg-muted text-muted-foreground" };
}

function PdfFileIcon({ className, ...props }: SVGProps<SVGSVGElement>) {
  return (
    <svg
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="1.8"
      strokeLinecap="round"
      strokeLinejoin="round"
      className={className}
      {...props}
    >
      <path d="M6 2.75h8l4 4v14.5H6z" />
      <path d="M14 2.75v4h4" />
      <text
        x="7.25"
        y="17.25"
        fill="currentColor"
        stroke="none"
        fontSize="5.5"
        fontWeight="700"
        letterSpacing="-.25"
      >
        PDF
      </text>
    </svg>
  );
}

export function ModalityIconList({ values, emptyLabel }: { values: string[]; emptyLabel: string }) {
  if (values.length === 0) return <p className="text-sm text-muted-foreground">{emptyLabel}</p>;

  return (
    <div className="flex flex-wrap gap-2">
      {values.map((value) => {
        const { icon: Icon, className } = getModalityMeta(value);
        return (
          <span key={value} className={cn("inline-flex items-center gap-1.5 rounded-full border px-2.5 py-1 text-xs", className)}>
            <Icon className="size-3.5" aria-hidden="true" />
            <bdi dir="auto">{value}</bdi>
          </span>
        );
      })}
    </div>
  );
}

function ModalityIconGroup({
  values,
  label,
  emptyLabel,
}: {
  values: string[];
  label: string;
  emptyLabel: string;
}) {
  if (values.length === 0) {
    return <span className="text-sm text-muted-foreground">{emptyLabel}</span>;
  }

  return (
    <div className="flex flex-wrap items-center gap-1" role="list" aria-label={label}>
      {values.map((value) => {
        const { icon: Icon, className } = getModalityMeta(value);
        return (
          <Tooltip key={value}>
            <TooltipTrigger asChild>
              <span
                role="listitem"
                tabIndex={0}
                className={cn(
                  "inline-flex size-7 items-center justify-center rounded-md border outline-none focus-visible:ring-2 focus-visible:ring-ring",
                  className,
                )}
              >
                <Icon className="size-4" aria-hidden="true" />
                <span className="sr-only">{value}</span>
              </span>
            </TooltipTrigger>
            <TooltipContent>{label}: {value}</TooltipContent>
          </Tooltip>
        );
      })}
    </div>
  );
}

export function ModalityFlow({
  inputValues,
  outputValues,
  inputLabel,
  outputLabel,
  emptyLabel,
}: {
  inputValues: string[];
  outputValues: string[];
  inputLabel: string;
  outputLabel: string;
  emptyLabel: string;
}) {
  return (
    <TooltipProvider>
      <div className="flex min-w-0 items-center gap-2" dir="rtl">
        <ModalityIconGroup values={inputValues} label={inputLabel} emptyLabel={emptyLabel} />
        <ArrowLeft className="size-4 shrink-0 text-muted-foreground" aria-hidden="true" />
        <ModalityIconGroup values={outputValues} label={outputLabel} emptyLabel={emptyLabel} />
      </div>
    </TooltipProvider>
  );
}
