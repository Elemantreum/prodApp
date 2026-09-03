interface Props {
  booked: number;
  max: number;
}

// Зелёный (свободно) -> жёлтый -> красный (мест почти/совсем нет), по
// заполненности. HSL-хью линейно едет от 140 (зелёный) к 0 (красный).
function occupancyColor(ratio: number): string {
  const hue = 140 - 140 * Math.min(ratio, 1);
  return `hsl(${hue}, 70%, 45%)`;
}

/** Сегментированная полоса заполненности тренировки: сегментов ровно
 *  max (сколько мест всего), закрашено — booked. Цвет всех закрашенных
 *  сегментов общий, зависит от общей заполненности. */
export function OccupancyBar({ booked, max }: Props) {
  const safeMax = Math.max(max, 1);
  const ratio = Math.min(booked / safeMax, 1);
  const color = occupancyColor(ratio);

  const segments = Array.from({ length: safeMax }, (_, i) => i < booked);

  return (
    <div className="occupancy">
      <div className="occupancy__bar">
        {segments.map((filled, i) => (
          <span
            key={i}
            className="occupancy__segment"
            style={filled ? { backgroundColor: color } : undefined}
          />
        ))}
      </div>
      <div className="occupancy__label">
        {booked}/{max}
      </div>
    </div>
  );
}
