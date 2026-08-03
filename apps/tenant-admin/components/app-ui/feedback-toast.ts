"use client"

import type * as React from "react"
import { toast, type ExternalToast } from "sonner"

export type FeedbackToastKind = "success" | "error" | "warning" | "info"

export type FeedbackToast = {
  kind: FeedbackToastKind
  title: React.ReactNode
  description?: React.ReactNode
  options?: Omit<ExternalToast, "description">
}

export function showFeedbackToast({
  kind,
  title,
  description,
  options,
}: FeedbackToast): string | number {
  return toast[kind](title, { ...options, description })
}

export function dismissFeedbackToast(id?: string | number): string | number {
  return toast.dismiss(id)
}
