import type { AbstractIntlMessages } from "next-intl";

export function pickMessages(
  messages: AbstractIntlMessages,
  namespaces: readonly string[],
): AbstractIntlMessages {
  return Object.fromEntries(
    namespaces.flatMap((namespace) =>
      namespace in messages ? [[namespace, messages[namespace]]] : [],
    ),
  ) as AbstractIntlMessages;
}
