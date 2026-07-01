import type {SidebarsConfig} from '@docusaurus/plugin-content-docs';

const sidebars: SidebarsConfig = {
  tutorialSidebar: [
    'intro',
    {
      type: 'category',
      label: 'شروع کار',
      items: [
        'quick-start',
        'making-llm-requests-via-gateway',
        'authentication',
        'request-headers',
        'native-sdk-support',
      ],
    },
    {
      type: 'category',
      label: 'مفاهیم و معماری',
      items: [
        'feature-overview',
        'intro-to-gateway',
        'virtual-model',
        'cost-tracking',
        'observability-logs',
        'observability-logs-export',
        'observability-opentelemetry',
      ],
    },
    {
      type: 'category',
      label: 'API های پشتیبانی‌ شده',
      items: [
        'supported-apis/index',
        {
          type: 'category',
          label: 'متنی و مکالمه‌ای',
          items: [
            'supported-apis/chat-completions/overview',
            'supported-apis/chat-completions/multimodal',
            'supported-apis/chat-completions/tools',
            'supported-apis/chat-completions/structured-outputs',
            'supported-apis/chat-completions/caching-and-reasoning',
            'supported-apis/chat-completions/extended-thinking',
            'supported-apis/responses-api',
            'supported-apis/messages',
          ],
        },
        {
          type: 'category',
          label: 'بازیابی و رتبه‌بندی',
          items: [
            'supported-apis/embeddings',
            'supported-apis/rerank',
          ],
        },
        {
          type: 'category',
          label: 'تصویر',
          items: [
            'supported-apis/image-generation',
            'supported-apis/image-edit',
            'supported-apis/image-variation',
          ],
        },
        {
          type: 'category',
          label: 'صوت',
          items: [
            'supported-apis/speech-to-text',
            'supported-apis/text-to-speech',
            'supported-apis/audio-translation',
          ],
        },
        {
          type: 'category',
          label: 'کنترل و پردازش',
          items: [
            'supported-apis/moderation',
            'supported-apis/batch-api',
          ],
        },
      ],
    },
    {
      type: 'category',
      label: 'مدیریت دسترسی و هویت',
      items: [
        'guides/api-keys',
        'guides/api-key-rotation',
        'guides/user-management',
        'guides/team-management',
        'guides/tenant-team-account-model',
        'guides/virtual-account-management',
        'guides/roles-and-permissions',
      ],
    },
    {
      type: 'category',
      label: 'سیاست‌ها و عملیات',
      items: [
        'guardrails',
        'guardrails-capabilities',
        'guardrails-pii-redaction',
        'guides/caching',
        'guides/rate-limiting',
        'guides/budget-limiting',
      ],
    },
  ],
};

export default sidebars;
