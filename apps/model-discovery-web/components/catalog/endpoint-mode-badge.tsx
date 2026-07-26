import {
  AudioLines,
  Bot,
  FileText,
  Image as ImageIcon,
  ListOrdered,
  MessageSquareText,
  Mic2,
  Orbit,
  ShieldCheck,
  Video,
  type LucideIcon,
} from "lucide-react";
import { Badge } from "@/components/ui/badge";
import { cn } from "@/lib/utils";

type EndpointModeMeta = {
  icon: LucideIcon;
  className: string;
};

const modeStyles = {
  primary:
    "border-primary-soft bg-primary-soft text-primary-soft-foreground",
  success:
    "border-success-soft bg-success-soft text-success-soft-foreground",
  warning:
    "border-warning-soft bg-warning-soft text-warning-soft-foreground",
  destructive:
    "border-destructive-soft bg-destructive-soft text-destructive-soft-foreground",
  accent: "border-border bg-accent text-accent-foreground",
} as const;

function getEndpointModeMeta(mode: string): EndpointModeMeta {
  const normalized = mode.trim().toLowerCase();

  if (normalized === "text") {
    return { icon: FileText, className: modeStyles.primary };
  }
  if (normalized.includes("image")) {
    return { icon: ImageIcon, className: modeStyles.success };
  }
  if (normalized.includes("audio") || normalized.includes("speech")) {
    return { icon: AudioLines, className: modeStyles.warning };
  }
  if (
    normalized.includes("transcription") ||
    normalized.includes("recognition")
  ) {
    return { icon: Mic2, className: modeStyles.warning };
  }
  if (normalized.includes("video")) {
    return { icon: Video, className: modeStyles.destructive };
  }
  if (normalized.includes("embedding")) {
    return { icon: Orbit, className: modeStyles.accent };
  }
  if (normalized.includes("rerank")) {
    return { icon: ListOrdered, className: modeStyles.accent };
  }
  if (normalized.includes("moderation")) {
    return { icon: ShieldCheck, className: modeStyles.destructive };
  }
  if (normalized.includes("chat") || normalized.includes("message")) {
    return { icon: MessageSquareText, className: modeStyles.primary };
  }

  return { icon: Bot, className: modeStyles.primary };
}

export function EndpointModeBadge({
  mode,
  className,
}: {
  mode: string;
  className?: string;
}) {
  const { icon: Icon, className: toneClassName } = getEndpointModeMeta(mode);

  return (
    <Badge
      className={cn("gap-1.5 font-medium", toneClassName, className)}
      dir="ltr"
      title={mode}
    >
      <Icon className="size-3.5" aria-hidden="true" />
      <span>{mode}</span>
    </Badge>
  );
}

export function EndpointModeBadgeList({
  modes,
  className,
  limit,
}: {
  modes: string[];
  className?: string;
  limit?: number;
}) {
  const uniqueModes = Array.from(new Set(modes.filter(Boolean)));
  const visibleModes =
    typeof limit === "number" ? uniqueModes.slice(0, limit) : uniqueModes;

  return (
    <div className={cn("flex flex-wrap gap-1", className)}>
      {visibleModes.map((mode) => (
        <EndpointModeBadge key={mode} mode={mode} />
      ))}
    </div>
  );
}
