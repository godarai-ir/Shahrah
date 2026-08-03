import { describe, expect, it, vi } from "vitest"

const { dismiss, toastMethod } = vi.hoisted(() => ({
  toastMethod: vi.fn(() => "toast-id"),
  dismiss: vi.fn(() => "dismissed-id"),
}))

vi.mock("sonner", () => ({
  toast: {
    success: toastMethod,
    error: toastMethod,
    warning: toastMethod,
    info: toastMethod,
    dismiss,
  },
}))

import { dismissFeedbackToast, showFeedbackToast } from "./feedback-toast"

describe("feedback toast", () => {
  it.each(["success", "error", "warning", "info"] as const)(
    "routes %s feedback through the shared adapter",
    (kind) => {
      const id = showFeedbackToast({
        kind,
        title: "Operation result",
        description: "Resource details",
        options: { id: `feedback-${kind}` },
      })

      expect(id).toBe("toast-id")
      expect(toastMethod).toHaveBeenLastCalledWith("Operation result", {
        description: "Resource details",
        id: `feedback-${kind}`,
      })
    },
  )

  it("dismisses feedback through the shared adapter", () => {
    expect(dismissFeedbackToast("feedback-id")).toBe("dismissed-id")
    expect(dismiss).toHaveBeenCalledWith("feedback-id")
  })
})
