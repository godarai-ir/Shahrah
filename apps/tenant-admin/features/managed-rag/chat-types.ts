import type { UIMessage } from "ai";

import type { Citation } from "@/lib/contracts/managed-rag";

export type ManagedRAGChatData = {
  status: { stage: "retrieving" | "reranking" };
  result: {
    citations: Citation[];
    refusal: boolean;
    actualModel: string;
    requestID: string;
    traceID: string;
    sessionID: string;
  };
};

export type ManagedRAGUIMessage = UIMessage<unknown, ManagedRAGChatData>;
