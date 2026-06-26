"use client";

import { useMemo, useState } from "react";

import { extractPromptVariableNames, type PromptEditorMessage } from "./prompt-management-utils";
import {
  createEmptyBuilder,
  createMetadataRow,
  findProviderForModel,
  hydrateBuilderFromVersion,
  type BuilderState,
  type GatewayModel,
  type MetadataRow,
  type PromptVersion,
} from "./prompt-management-shared";

export function usePromptBuilderState(defaultSystemPrompt: string, models: GatewayModel[]) {
  const [builder, setBuilder] = useState<BuilderState>(() => createEmptyBuilder(defaultSystemPrompt));

  const detectedVariables = useMemo(
    () =>
      extractPromptVariableNames([
        builder.systemPrompt,
        ...builder.messages.map((message) => message.content),
        builder.composerText,
      ]),
    [builder.composerText, builder.messages, builder.systemPrompt],
  );

  const canSave =
    builder.name.trim() !== "" && builder.selectedModel.trim() !== "" && (builder.systemPrompt.trim() !== "" || builder.messages.length > 0);

  function resetBuilder() {
    setBuilder(createEmptyBuilder(defaultSystemPrompt));
  }

  function hydrateBuilder(version: PromptVersion) {
    const nextBuilder = hydrateBuilderFromVersion(version, defaultSystemPrompt, models);
    if (nextBuilder.selectedModel && !nextBuilder.selectedProvider) {
      nextBuilder.selectedProvider = findProviderForModel(nextBuilder.selectedModel, models);
    }
    setBuilder(nextBuilder);
  }

  function addComposerMessage() {
    if (!builder.composerText.trim()) return;
    setBuilder((current) => ({
      ...current,
      messages: [
        ...current.messages,
        {
          role: current.composerRole,
          content: current.composerText.trim(),
        },
      ],
      composerText: "",
    }));
  }

  function updateMessage(index: number, patch: Partial<PromptEditorMessage>) {
    setBuilder((current) => ({
      ...current,
      messages: current.messages.map((message, messageIndex) => (messageIndex === index ? { ...message, ...patch } : message)),
    }));
  }

  function removeMessage(index: number) {
    setBuilder((current) => ({
      ...current,
      messages: current.messages.filter((_, messageIndex) => messageIndex !== index),
    }));
  }

  function addMetadataRow() {
    setBuilder((current) => ({
      ...current,
      metadataRows: [...current.metadataRows, createMetadataRow()],
    }));
  }

  function updateMetadataRow(rowID: string, patch: Partial<MetadataRow>) {
    setBuilder((current) => ({
      ...current,
      metadataRows: current.metadataRows.map((row) => (row.id === rowID ? { ...row, ...patch } : row)),
    }));
  }

  function removeMetadataRow(rowID: string) {
    setBuilder((current) => ({
      ...current,
      metadataRows: current.metadataRows.filter((row) => row.id !== rowID),
    }));
  }

  return {
    builder,
    canSave,
    detectedVariables,
    setBuilder,
    resetBuilder,
    hydrateBuilder,
    addComposerMessage,
    updateMessage,
    removeMessage,
    addMetadataRow,
    updateMetadataRow,
    removeMetadataRow,
  };
}
