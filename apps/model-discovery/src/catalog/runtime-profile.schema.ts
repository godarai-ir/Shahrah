import { z } from 'zod';

const thinkingModeSchema = z.enum(['adaptive', 'manual', 'disabled']);
const endpointSchema = z.enum([
  'messages',
  'count_tokens',
  'message_batches',
  'files',
  'models',
  'chat_completions',
  'embeddings',
  'image_generations',
  'audio_speech',
  'videos',
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
    thinking: z.object({
      defaultMode: thinkingModeSchema,
      allowedModes: z.array(thinkingModeSchema).min(1),
      manualDeprecated: z.boolean().default(false),
      effortValues: z
        .array(z.enum(['none', 'minimal', 'low', 'medium', 'high', 'max']))
        .default([]),
    }),
    assistantPrefill: z.boolean(),
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
    limits: z.object({
      contextWindow: z.number().int().nonnegative(),
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
