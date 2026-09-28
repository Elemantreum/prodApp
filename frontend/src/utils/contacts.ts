import type { Identity } from "../api/types";

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
