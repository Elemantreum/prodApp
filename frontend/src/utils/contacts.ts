import type { Identity } from "../api/types";

/** Ссылка на профиль по привязанной платформе. VK строится по числовому
 *  externalId напрямую — работает всегда. Telegram требует @username — не у
 *  всех он есть (можно скрыть в настройках), для таких возвращаем null,
 *  ссылку показать нечем. */
export function contactUrl(identity: Identity): string | null {
  if (identity.provider === "vk") {
    return `https://vk.com/id${identity.externalId}`;
  }
  if (identity.provider === "telegram" && identity.username) {
    return `https://t.me/${identity.username}`;
  }
  return null;
}

export const CONTACT_LABEL: Record<string, string> = {
  telegram: "TG",
  vk: "VK",
};
