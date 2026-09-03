import { isTelegram } from "./telegram";
import { isVk } from "./vk";

export type Platform = "telegram" | "vk" | "browser";

export function detectPlatform(): Platform {
  if (isTelegram()) return "telegram";
  if (isVk()) return "vk";
  return "browser";
}
