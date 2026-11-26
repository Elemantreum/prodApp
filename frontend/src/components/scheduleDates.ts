import { locale } from "../i18n";

export function dayKey(date: Date) {
  return [
    date.getFullYear(),
    String(date.getMonth() + 1).padStart(2, "0"),
    String(date.getDate()).padStart(2, "0"),
  ].join("-");
}

export function shiftDay(date: Date, offset: number) {
  return new Date(date.getFullYear(), date.getMonth(), date.getDate() + offset, 12);
}

export const fullDay = new Intl.DateTimeFormat(locale, { weekday: "long", day: "numeric", month: "long" });
export const clock = new Intl.DateTimeFormat(locale, { hour: "2-digit", minute: "2-digit", hourCycle: "h23" });
