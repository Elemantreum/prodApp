export function occupancyColor(booked: number, capacity: number): string {
  if (booked >= capacity) return "#d95360";
  const ratio = booked / Math.max(1, capacity);
  return `hsl(${145 - 111 * ratio}, 62%, 46%)`;
}
