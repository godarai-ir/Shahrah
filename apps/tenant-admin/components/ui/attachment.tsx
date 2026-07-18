"use client";

import * as React from "react";
import { Slot } from "@radix-ui/react-slot";
import { cva, type VariantProps } from "class-variance-authority";

import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";

type AttachmentState = "idle" | "uploading" | "processing" | "error" | "done";
type AttachmentSize = "default" | "sm" | "xs";
type AttachmentOrientation = "horizontal" | "vertical";
type AttachmentMediaVariant = "icon" | "image";

type AttachmentContextValue = {
  orientation: AttachmentOrientation;
  size: AttachmentSize;
  state: AttachmentState;
};

const AttachmentContext = React.createContext<AttachmentContextValue>({
  orientation: "horizontal",
  size: "default",
  state: "done",
});

const attachmentVariants = cva(
  "relative flex rounded-xl border border-border bg-card text-card-foreground transition-colors",
  {
    variants: {
      orientation: {
        horizontal: "items-center gap-3 p-3",
        vertical: "flex-col gap-3 p-3",
      },
      size: {
        default: "",
        sm: "gap-2.5 p-2.5",
        xs: "gap-2 p-2",
      },
      state: {
        idle: "",
        uploading: "",
        processing: "",
        error: "border-destructive/40 bg-destructive/5",
        done: "",
      },
    },
    defaultVariants: {
      orientation: "horizontal",
      size: "default",
      state: "done",
    },
  },
);

const attachmentMediaVariants = cva(
  "flex shrink-0 items-center justify-center overflow-hidden rounded-lg border border-border bg-muted text-muted-foreground [&_img]:size-full [&_img]:object-cover",
  {
    variants: {
      variant: {
        icon: "",
        image: "bg-accent",
      },
      orientation: {
        horizontal: "",
        vertical: "w-full",
      },
      size: {
        default: "",
        sm: "",
        xs: "",
      },
    },
    compoundVariants: [
      { orientation: "horizontal", size: "default", className: "size-14" },
      { orientation: "horizontal", size: "sm", className: "size-12" },
      { orientation: "horizontal", size: "xs", className: "size-10" },
      { orientation: "vertical", size: "default", variant: "image", className: "aspect-[4/3] min-h-36" },
      { orientation: "vertical", size: "sm", variant: "image", className: "aspect-[4/3] min-h-28" },
      { orientation: "vertical", size: "xs", variant: "image", className: "aspect-[4/3] min-h-24" },
      { orientation: "vertical", size: "default", variant: "icon", className: "h-24" },
      { orientation: "vertical", size: "sm", variant: "icon", className: "h-20" },
      { orientation: "vertical", size: "xs", variant: "icon", className: "h-16" },
    ],
    defaultVariants: {
      variant: "icon",
    },
  },
);

const attachmentTitleVariants = cva(
  "truncate font-medium text-foreground",
  {
    variants: {
      size: {
        default: "text-sm",
        sm: "text-sm",
        xs: "text-xs",
      },
      state: {
        idle: "",
        uploading: "animate-pulse",
        processing: "animate-pulse",
        error: "",
        done: "",
      },
    },
    defaultVariants: {
      size: "default",
      state: "done",
    },
  },
);

const attachmentDescriptionVariants = cva("text-muted-foreground", {
  variants: {
    size: {
      default: "text-xs",
      sm: "text-xs",
      xs: "text-xs",
    },
    state: {
      idle: "",
      uploading: "",
      processing: "",
      error: "text-destructive",
      done: "",
    },
  },
  defaultVariants: {
    size: "default",
    state: "done",
  },
});

function Attachment({
  className,
  orientation = "horizontal",
  size = "default",
  state = "done",
  ...props
}: React.ComponentProps<"div"> &
  VariantProps<typeof attachmentVariants> & {
    state?: AttachmentState;
    size?: AttachmentSize;
    orientation?: AttachmentOrientation;
  }) {
  return (
    <AttachmentContext.Provider value={{ orientation, size, state }}>
      <div
        data-slot="attachment"
        className={cn(attachmentVariants({ orientation, size, state, className }))}
        {...props}
      />
    </AttachmentContext.Provider>
  );
}

function AttachmentMedia({
  className,
  variant = "icon",
  ...props
}: React.ComponentProps<"div"> & {
  variant?: AttachmentMediaVariant;
}) {
  const { orientation, size } = React.useContext(AttachmentContext);

  return (
    <div
      data-slot="attachment-media"
      className={cn(attachmentMediaVariants({ variant, orientation, size, className }))}
      {...props}
    />
  );
}

function AttachmentContent({
  className,
  ...props
}: React.ComponentProps<"div">) {
  return (
    <div
      data-slot="attachment-content"
      className={cn("flex min-w-0 flex-1 flex-col gap-1", className)}
      {...props}
    />
  );
}

function AttachmentTitle({
  className,
  ...props
}: React.ComponentProps<"div">) {
  const { size, state } = React.useContext(AttachmentContext);

  return (
    <div
      data-slot="attachment-title"
      className={cn(attachmentTitleVariants({ size, state, className }))}
      {...props}
    />
  );
}

function AttachmentDescription({
  className,
  ...props
}: React.ComponentProps<"div">) {
  const { size, state } = React.useContext(AttachmentContext);

  return (
    <div
      data-slot="attachment-description"
      className={cn(attachmentDescriptionVariants({ size, state, className }))}
      {...props}
    />
  );
}

function AttachmentActions({
  className,
  ...props
}: React.ComponentProps<"div">) {
  const { orientation } = React.useContext(AttachmentContext);

  return (
    <div
      data-slot="attachment-actions"
      className={cn(
        "relative z-10 flex shrink-0 items-center gap-1",
        orientation === "vertical" ? "self-end" : "",
        className,
      )}
      {...props}
    />
  );
}

function AttachmentAction({
  className,
  size = "icon-xs",
  variant = "ghost",
  ...props
}: React.ComponentProps<typeof Button>) {
  return (
    <Button
      data-slot="attachment-action"
      type="button"
      size={size}
      variant={variant}
      className={cn("shrink-0", className)}
      {...props}
    />
  );
}

function AttachmentTrigger({
  className,
  asChild = false,
  ...props
}: React.ComponentProps<"button"> & {
  asChild?: boolean;
}) {
  const Comp = asChild ? Slot : "button";

  return (
    <Comp
      data-slot="attachment-trigger"
      className={cn("absolute inset-0 rounded-xl", className)}
      {...props}
    />
  );
}

export {
  Attachment,
  AttachmentAction,
  AttachmentActions,
  AttachmentContent,
  AttachmentDescription,
  AttachmentMedia,
  AttachmentTitle,
  AttachmentTrigger,
};
