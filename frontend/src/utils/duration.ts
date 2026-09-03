const NS_PER_MINUTE = 60_000_000_000;
const NS_PER_DAY = 86_400_000_000_000;

/** duration comes from the API as Go time.Duration — nanoseconds. */
export function formatDurationNs(ns: number): string {
  const totalMinutes = Math.round(ns / NS_PER_MINUTE);
  const hours = Math.floor(totalMinutes / 60);
  const minutes = totalMinutes % 60;
  if (hours === 0) return `${minutes} мин`;
  if (minutes === 0) return `${hours} ч`;
  return `${hours} ч ${minutes} мин`;
}

export function minutesToNs(minutes: number): number {
  return minutes * NS_PER_MINUTE;
}

export function nsToMinutes(ns: number): number {
  return Math.round(ns / NS_PER_MINUTE);
}

/** leadTime у регулярных тренировок удобнее задавать в днях, а не минутах. */
export function daysToNs(days: number): number {
  return days * NS_PER_DAY;
}

export function nsToDays(ns: number): number {
  return Math.round(ns / NS_PER_DAY);
}
