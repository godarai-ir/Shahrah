import { z } from 'zod';

const thinkingModeSchema = z.enum(['adaptive', 'manual', 'disabled']);
const supportedToolSchema = z.enum([
  'function_calling',
  'web_search',
  'web_search_premium',
  'web_fetch',
  'file_search',
  'document_library',
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
  'google_search',
  'google_maps',
  'url_context',
  'x_search',
  'attachment_search',
  'collections_search',
]);
const endpointSchema = z.enum([
  'messages',
  'count_tokens',
  'message_batches',
  'files',
  'models',
  'language_models',
  'chat_completions',
  'deferred_chat_completions',
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
  'image_generation_models',
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
  'batch_requests',
  'batch_results',
  'file_public_urls',
  'generate_content',
  'stream_generate_content',
  'batch_generate_content',
  'batch_embed_contents',
  'cached_contents',
  'generate_answer',
  'embed_content',
  'imagen_generate',
  'fim_completions',
  'chat_moderations',
  'ocr',
  'agents',
]);

export const modelRuntimeProfileSchema = z
  .object({
    version: z.number().int().positive(),
    aliases: z.array(z.string().min(1)).default([]),
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
          z.enum([
            'unknown',
            'text',
            'image',
            'pdf',
            'document',
            'audio',
            'video',
            'code',
          ]),
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
            'document',
            'moderation_result',
          ]),
        )
        .min(1),
    }),
    parameters: z
      .partialRecord(endpointSchema, z.array(z.string().min(1)))
      .default({}),
    requestControls: z
      .object({
        serviceTierValues: z
          .array(
            z.enum([
              'auto',
              'standard_only',
              'default',
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
    requestHeaders: z
      .object({
        allowed: z.array(z.string().min(1)).default([]),
        byEndpoint: z
          .partialRecord(endpointSchema, z.array(z.string().min(1)))
          .default({}),
      })
      .default({ allowed: [], byEndpoint: {} }),
    contextPricing: z
      .object({
        longContextThresholdTokens: z.number().int().positive().optional(),
        appliesToAllTokens: z.boolean().default(true),
      })
      .default({ appliesToAllTokens: true }),
    batch: z
      .object({
        eligible: z.boolean().default(false),
        discounted: z.boolean().default(false),
        supportedEndpoints: z.array(endpointSchema).default([]),
        maxRequestBytes: z.number().int().positive().optional(),
      })
      .default({
        eligible: false,
        discounted: false,
        supportedEndpoints: [],
      }),
    deferredCompletions: z
      .object({
        eligible: z.boolean().default(false),
        retentionSeconds: z.number().int().positive().optional(),
        maximumRetrievals: z.number().int().positive().optional(),
      })
      .default({ eligible: false }),
    responseResources: z
      .object({
        store: z.boolean().default(false),
        retrieve: z.boolean().default(false),
        delete: z.boolean().default(false),
        compact: z.boolean().default(false),
        countInputTokens: z.boolean().default(false),
      })
      .default({
        store: false,
        retrieve: false,
        delete: false,
        compact: false,
        countInputTokens: false,
      }),
    imageControls: z
      .object({
        aspectRatios: z.array(z.string().min(1)).default([]),
        resolutions: z.array(z.string().min(1)).default([]),
        responseFormats: z.array(z.string().min(1)).default([]),
        maximumOutputs: z.number().int().positive().optional(),
        maximumReferenceImages: z.number().int().positive().optional(),
        editTransport: z.enum(['json', 'multipart']).optional(),
        filesInput: z.boolean().default(false),
        storageOptions: z.boolean().default(false),
        streaming: z.boolean().default(false),
      })
      .default({
        aspectRatios: [],
        resolutions: [],
        responseFormats: [],
        filesInput: false,
        storageOptions: false,
        streaming: false,
      }),
    usageAccounting: z
      .object({
        requiredResponseFields: z
          .array(
            z.enum([
              'cost_in_usd_ticks',
              'num_server_side_tools_used',
              'service_tier',
            ]),
          )
          .default([]),
        exactProviderCost: z.boolean().default(false),
      })
      .default({ requiredResponseFields: [], exactProviderCost: false }),
    unsupportedParameters: z.array(z.string().min(1)).default([]),
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
          tier: z.number().int().min(1).max(5).optional(),
          usageTier: z.enum(['free', 'tier_1', 'tier_2', 'tier_3']).optional(),
          scope: z.enum(['project']).default('project'),
          dynamic: z.boolean().default(false),
          requestsPerMinute: z.number().int().nonnegative().optional(),
          requestsPerDay: z.number().int().nonnegative().optional(),
          tokensPerMinute: z.number().int().nonnegative().optional(),
          tokensPerDay: z.number().int().nonnegative().optional(),
          imagesPerMinute: z.number().int().nonnegative().optional(),
          batchQueueTokens: z.number().int().nonnegative().optional(),
          spendUsdPerTenMinutes: z.number().nonnegative().optional(),
          concurrentBatchRequests: z.number().int().nonnegative().optional(),
          inputFileBytes: z.number().int().nonnegative().optional(),
          fileStorageBytes: z.number().int().nonnegative().optional(),
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
    const usageTiers = new Set<string>();
    for (const [index, limits] of profile.referenceRateLimits.entries()) {
      if (limits.tier === undefined && limits.usageTier === undefined) {
        context.addIssue({
          code: 'custom',
          path: ['referenceRateLimits', index],
          message: 'tier or usageTier is required',
        });
      }
      if (limits.usageTier && usageTiers.has(limits.usageTier)) {
        context.addIssue({
          code: 'custom',
          path: ['referenceRateLimits', index, 'usageTier'],
          message: `duplicate usage tier ${limits.usageTier}`,
        });
      }
      if (limits.usageTier) usageTiers.add(limits.usageTier);
    }
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
    if (new Set(profile.aliases).size !== profile.aliases.length) {
      context.addIssue({
        code: 'custom',
        path: ['aliases'],
        message: 'aliases must be unique',
      });
    }
    if (
      profile.batch.eligible &&
      profile.batch.supportedEndpoints.length === 0
    ) {
      context.addIssue({
        code: 'custom',
        path: ['batch', 'supportedEndpoints'],
        message: 'eligible batch profiles require supported endpoints',
      });
    }
    for (const endpoint of profile.batch.supportedEndpoints) {
      if (!profile.endpoints.includes(endpoint)) {
        context.addIssue({
          code: 'custom',
          path: ['batch', 'supportedEndpoints'],
          message: `${endpoint} must be included in endpoints`,
        });
      }
    }
    for (const endpoint of Object.keys(profile.requestHeaders.byEndpoint)) {
      if (
        !profile.endpoints.includes(
          endpoint as (typeof profile.endpoints)[number],
        )
      ) {
        context.addIssue({
          code: 'custom',
          path: ['requestHeaders', 'byEndpoint', endpoint],
          message: `${endpoint} must be included in endpoints`,
        });
      }
    }
  });

export type ModelRuntimeProfile = z.infer<typeof modelRuntimeProfileSchema>;
