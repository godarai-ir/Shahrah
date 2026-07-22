import {getRequestConfig} from 'next-intl/server';

export const DEFAULT_LOCALE = 'fa' as const;

export default getRequestConfig(async () => ({
  locale: DEFAULT_LOCALE,
  messages: (await import('../messages/fa.json')).default,
}));
