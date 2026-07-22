import messages from '../messages/fa.json';

declare module 'next-intl' {
  interface AppConfig {
    Locale: 'fa';
    Messages: typeof messages;
  }
}
