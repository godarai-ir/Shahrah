"use client"

import * as React from "react"
import { Direction, Progress as ProgressPrimitive } from "radix-ui"

import { cn } from "@/lib/utils"

function Progress({
  className,
  dir,
  value,
  ...props
}: React.ComponentProps<typeof ProgressPrimitive.Root>) {
  const direction = Direction.useDirection(
    dir === "ltr" || dir === "rtl" ? dir : undefined
  )

  return (
    <ProgressPrimitive.Root
      data-slot="progress"
      dir={direction}
      className={cn(
        "relative h-2 w-full overflow-hidden rounded-full bg-primary/20",
        className
      )}
      {...props}
    >
      <ProgressPrimitive.Indicator
        data-slot="progress-indicator"
        className="h-full w-full flex-1 bg-primary transition-all"
        style={{
          transform: `scaleX(${(value || 0) / 100})`,
          transformOrigin: direction === "rtl" ? "right" : "left",
        }}
      />
    </ProgressPrimitive.Root>
  )
}

export { Progress }
