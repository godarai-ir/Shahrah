import { z } from 'zod';

const thinkingModeSchema = z.enum(['adaptive', 'manual', 'disabled']);
const supportedToolSchema = z.enum([
  'function_calling',
  'web_search',
  'web_fetch',
  'file_search',
  'tool_search',
  'image_generation',
  'code_interpreter',
  'code_execution',
  'hosted_shell',
  'bash',
  'text_editor',
  'apply_patch',
  'skills',
  'computer_use',
  'mcp',
  'memory',
  'advisor',
  'programmatic_tool_calling',
]);
const endpointSchema = z.enum([
  'messages',
  'count_tokens',
  'message_batches',
  'files',
  'models',
  'chat_completions',
  'completions',
  'responses',
  'responses_input_tokens',
  'responses_compact',
  'conversations',
  'embeddings',
  'moderations',
  'image_generations',
  'image_edits',
  'image_variations',
  'audio_speech',
  'audio_transcriptions',
  'audio_translations',
  'realtime',
  'realtime_transcription',
  'realtime_translation',
  'videos',
  'uploads',
  'vector_stores',
  'containers',
  'interactions',
  'live',
  'batches',
  'generate_content',
  'generate_answer',
  'embed_content',
  'imagen_generate',
]);

export const modelRuntimeProfileSchema = z
  .object({
    version: z.number().int().positive(),
    endpoints: z.array(endpointSchema).min(1),
    transports: z.array(z.enum(['http', 'sse', 'websocket'])).default(['http']),
    lifecycle: z
      .enum(['active', 'preview', 'deprecated', 'retired', 'unavailable'])
      .default('active'),
    thinking: z.object({
      defaultMode: thinkingModeSchema,
      allowedModes: z.array(thinkingModeSchema).min(1),
      manualDeprecated: z.boolean().default(false),
      effortValues: z
        .array(
          z.enum(['none', 'minimal', 'low', 'medium', 'high', 'xhigh', 'max']),
        )
        .default([]),
      reasoningModes: z.array(z.enum(['standard', 'pro'])).default([]),
      contextValues: z
        .array(z.enum(['auto', 'current_turn', 'all_turns']))
        .default([]),
      alwaysEnabled: z.boolean().default(false),
      disabledMaxEffort: z
        .enum(['low', 'medium', 'high', 'xhigh', 'max'])
        .optional(),
    }),
    assistantPrefill: z.boolean(),
    fineTuning: z.boolean().default(false),
    predictedOutputs: z.boolean().default(false),
    sampling: z.object({
      parameters: z.array(z.enum(['temperature', 'top_p', 'top_k'])),
      disallowedWithThinking: z
        .array(z.enum(['temperature', 'top_p', 'top_k']))
        .default([]),
    }),
    tools: z.object({
      custom: z.boolean(),
      server: z.boolean().default(false),
      parallel: z.boolean(),
      structuredOutput: z.boolean(),
      supported: z.array(supportedToolSchema).default([]),
      preferredVersions: z
        .partialRecord(supportedToolSchema, z.string().min(1))
        .default({}),
    }),
    modalities: z.object({
      input: z
        .array(
          z.enum(['unknown', 'text', 'image', 'pdf', 'audio', 'video', 'code']),
        )
        .min(1),
      output: z
        .array(
          z.enum([
            'unknown',
            'text',
            'image',
            'audio',
            'video',
            'embedding',
            'code',
          ]),
        )
        .min(1),
    }),
    requestControls: z
      .object({
        serviceTierValues: z
          .array(
            z.enum([
              'auto',
              'standard_only',
              'standard',
              'batch',
              'flex',
              'priority',
              'fast',
            ]),
          )
          .default(['auto', 'standard_only']),
        inferenceGeoValues: z
          .array(z.enum(['global', 'us']))
          .default(['global']),
        speedValues: z.array(z.enum(['standard', 'fast'])).default([]),
        effectiveSpeed: z
          .record(z.string(), z.enum(['standard', 'fast']))
          .default({}),
      })
      .default({
        serviceTierValues: ['auto', 'standard_only'],
        inferenceGeoValues: ['global'],
        speedValues: [],
        effectiveSpeed: {},
      }),
    betaHeaders: z.array(z.string().min(1)).default([]),
    betaCapabilities: z
      .array(
        z.object({
          header: z.string().min(1),
          endpoints: z.array(endpointSchema).min(1),
          maxOutputTokens: z.number().int().positive().optional(),
        }),
      )
      .default([]),
    referenceRateLimits: z
      .array(
        z.object({
          tier: z.number().int().min(1).max(5),
          requestsPerMinute: z.number().int().nonnegative().optional(),
          requestsPerDay: z.number().int().nonnegative().optional(),
          tokensPerMinute: z.number().int().nonnegative().optional(),
          imagesPerMinute: z.number().int().nonnegative().optional(),
          batchQueueTokens: z.number().int().nonnegative().optional(),
        }),
      )
      .default([]),
    minimumRetentionDays: z.number().int().nonnegative().default(0),
    zdrEligible: z.boolean().default(false),
    limits: z.object({
      contextWindow: z.number().int().nonnegative(),
      maxInputTokens: z.number().int().nonnegative().default(0),
      maxInputCharacters: z.number().int().nonnegative().default(0),
      maxInputBytes: z.number().int().nonnegative().default(0),
      maxOutputTokens: z.number().int().nonnegative(),
      messagesBytes: z
        .number()
        .int()
        .positive()
        .default(32 * 1024 * 1024),
      batchBytes: z
        .number()
        .int()
        .positive()
        .default(256 * 1024 * 1024),
      fileBytes: z
        .number()
        .int()
        .positive()
        .default(500 * 1024 * 1024),
    }),
  })
  .strict()
  .superRefine((profile, context) => {
    if (!profile.thinking.allowedModes.includes(profile.thinking.defaultMode)) {
      context.addIssue({
        code: 'custom',
        path: ['thinking', 'defaultMode'],
        message: 'defaultMode must be included in allowedModes',
      });
    }
    if (
      profile.thinking.defaultMode === 'disabled' &&
      profile.thinking.effortValues.length > 0
    ) {
      context.addIssue({
        code: 'custom',
        path: ['thinking', 'effortValues'],
        message: 'effortValues require an enabled thinking mode',
      });
    }
    if (
      profile.thinking.alwaysEnabled &&
      (profile.thinking.allowedModes.length !== 1 ||
        profile.thinking.allowedModes[0] !== 'adaptive')
    ) {
      context.addIssue({
        code: 'custom',
        path: ['thinking', 'alwaysEnabled'],
        message: 'alwaysEnabled thinking must allow only adaptive mode',
      });
    }
    if (
      profile.thinking.disabledMaxEffort &&
      !profile.thinking.allowedModes.includes('disabled')
    ) {
      context.addIssue({
        code: 'custom',
        path: ['thinking', 'disabledMaxEffort'],
        message: 'disabledMaxEffort requires disabled thinking mode',
      });
    }
    for (const [requested, effective] of Object.entries(
      profile.requestControls.effectiveSpeed,
    )) {
      if (
        !profile.requestControls.speedValues.includes(
          requested as 'standard' | 'fast',
        )
      ) {
        context.addIssue({
          code: 'custom',
          path: ['requestControls', 'effectiveSpeed', requested],
          message: `${requested} must be included in speedValues`,
        });
      }
      if (effective === 'fast' && requested !== 'fast') {
        context.addIssue({
          code: 'custom',
          path: ['requestControls', 'effectiveSpeed', requested],
          message: 'standard requests cannot become fast implicitly',
        });
      }
    }
    for (const tool of Object.keys(profile.tools.preferredVersions)) {
      if (
        !profile.tools.supported.includes(
          tool as (typeof profile.tools.supported)[number],
        )
      ) {
        context.addIssue({
          code: 'custom',
          path: ['tools', 'preferredVersions', tool],
          message: `${tool} must be included in tools.supported`,
        });
      }
    }
    for (const capability of profile.betaCapabilities) {
      if (!profile.betaHeaders.includes(capability.header)) {
        context.addIssue({
          code: 'custom',
          path: ['betaCapabilities'],
          message: `${capability.header} must be included in betaHeaders`,
        });
      }
      for (const endpoint of capability.endpoints) {
        if (!profile.endpoints.includes(endpoint)) {
          context.addIssue({
            code: 'custom',
            path: ['betaCapabilities', capability.header, 'endpoints'],
            message: `${endpoint} must be included in endpoints`,
          });
        }
      }
    }
  });

export type ModelRuntimeProfile = z.infer<typeof modelRuntimeProfileSchema>;
