"use client"

import {
  CircleCheckIcon,
  InfoIcon,
  Loader2Icon,
  OctagonXIcon,
  TriangleAlertIcon,
} from "lucide-react"
import { Toaster as Sonner, type ToasterProps } from "sonner"

import { useTheme } from "@/components/theme-provider"

type AppToasterProps = ToasterProps & {
  closeButtonAriaLabel?: string
}

const toastClassNames: NonNullable<ToasterProps["toastOptions"]>["classNames"] = {
  toast: "border-border bg-popover text-popover-foreground shadow-lg",
  title: "text-sm font-semibold text-current",
  description: "text-sm text-current opacity-80",
  success: "border-success-soft-border bg-success-soft text-success-soft-foreground",
  error: "border-destructive-soft-border bg-destructive-soft text-destructive-soft-foreground",
  warning: "border-warning-soft-border bg-warning-soft text-warning-soft-foreground",
  info: "border-info-soft-border bg-info-soft text-info-soft-foreground",
  actionButton: "bg-primary text-primary-foreground",
  cancelButton: "bg-secondary text-secondary-foreground",
  closeButton: "border-border bg-background text-foreground",
}

const Toaster = ({
  closeButtonAriaLabel,
  toastOptions,
  ...props
}: AppToasterProps) => {
  const { resolvedTheme = "light" } = useTheme()

  return (
    <Sonner
      theme={resolvedTheme as ToasterProps["theme"]}
      className="toaster group"
      dir="auto"
      closeButton
      expand
      visibleToasts={4}
      icons={{
        success: <CircleCheckIcon className="size-4" />,
        info: <InfoIcon className="size-4" />,
        warning: <TriangleAlertIcon className="size-4" />,
        error: <OctagonXIcon className="size-4" />,
        loading: <Loader2Icon className="size-4 animate-spin" />,
      }}
      style={
        {
          "--normal-bg": "var(--popover)",
          "--normal-text": "var(--popover-foreground)",
          "--normal-border": "var(--border)",
          "--border-radius": "var(--radius)",
        } as React.CSSProperties
      }
      toastOptions={{
        duration: 5000,
        ...toastOptions,
        closeButtonAriaLabel:
          closeButtonAriaLabel ?? toastOptions?.closeButtonAriaLabel,
        classNames: {
          ...toastClassNames,
          ...toastOptions?.classNames,
        },
      }}
      {...props}
    />
  )
}

export { Toaster }
