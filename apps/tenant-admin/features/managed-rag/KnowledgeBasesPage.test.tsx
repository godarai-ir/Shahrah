import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";

import KnowledgeBasesPage from "./KnowledgeBasesPage";

const apiMocks = vi.hoisted(() => ({
  listKnowledgeBases: vi.fn(),
  listDocuments: vi.fn(),
  createKnowledgeBase: vi.fn(),
  createTextDocument: vi.fn(),
  uploadDocument: vi.fn(),
}));

vi.mock("next-intl", () => ({
  useLocale: () => "fa",
  useTranslations: () => (key: string) => key,
}));

vi.mock("@/features/managed-rag/api", () => apiMocks);

const knowledgeBase = {
  id: "11111111-1111-4111-8111-111111111111",
  name: "پایگاه آزمون",
  description: "",
  rag_profile_revision: "rag-v3-fa-bge-m3",
  document_count: 1,
  ready_document_count: 0,
  created_at: "2026-08-19T00:00:00Z",
  updated_at: "2026-08-19T00:00:00Z",
};

const queuedDocument = {
  id: "22222222-2222-4222-8222-222222222222",
  knowledge_base_id: knowledgeBase.id,
  title: "سند آزمون",
  source_type: "file" as const,
  media_type: "application/pdf",
  status: "queued" as const,
  size_bytes: 100,
  chunk_count: 0,
  detected_language: "",
  created_at: "2026-08-19T00:00:00Z",
  updated_at: "2026-08-19T00:00:00Z",
};

afterEach(() => {
  vi.clearAllMocks();
});

describe("KnowledgeBasesPage live document status", () => {
  it("polls active ingestion and renders the next status without a page reload", async () => {
    apiMocks.listKnowledgeBases.mockResolvedValue({ items: [knowledgeBase], total: 1, limit: 100, offset: 0 });
    apiMocks.listDocuments
      .mockResolvedValueOnce({ items: [queuedDocument], total: 1, limit: 100, offset: 0 })
      .mockResolvedValue({
        items: [{ ...queuedDocument, status: "ready", chunk_count: 2, detected_language: "fa" }],
        total: 1,
        limit: 100,
        offset: 0,
      });

    render(<KnowledgeBasesPage />);

    await screen.findByText("پایگاه آزمون");
    fireEvent.click(screen.getByRole("button", { name: "documents.view" }));
    expect(await screen.findByText("documents.status.queued")).toBeInTheDocument();

    await waitFor(
      () => expect(screen.getByText("documents.status.ready")).toBeInTheDocument(),
      { timeout: 4_000 },
    );

    expect(apiMocks.listDocuments).toHaveBeenCalledTimes(2);
    expect(screen.getByText("2")).toBeInTheDocument();
  });
});
