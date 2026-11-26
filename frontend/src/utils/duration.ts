import { locale } from "../i18n";
const NS_PER_MINUTE = 60_000_000_000;
const NS_PER_DAY = 86_400_000_000_000;

export function formatDurationNs(ns: number): string {
  const totalMinutes = Math.round(ns / NS_PER_MINUTE);
  const hours = Math.floor(totalMinutes / 60);
  const minutes = totalMinutes % 60;
  const min = locale === "en" ? "min" : "мин";
  const hr = locale === "en" ? "h" : "ч";
  if (hours === 0) return `${minutes} ${min}`;
  if (minutes === 0) return `${hours} ${hr}`;
  return `${hours} ${hr} ${minutes} ${min}`;
}

export function minutesToNs(minutes: number): number {
  return minutes * NS_PER_MINUTE;
}

export function nsToMinutes(ns: number): number {
  return Math.round(ns / NS_PER_MINUTE);
}

export function daysToNs(days: number): number {
  return days * NS_PER_DAY;
}

export function nsToDays(ns: number): number {
  return Math.round(ns / NS_PER_DAY);
}
