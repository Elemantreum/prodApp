import { occupancyColor } from "../utils/occupancy";
import "./schedule.css";

export function OccupancyBar({
  booked,
  max,
  meterOnly = false,
}: {
  booked: number;
  max: number;
  meterOnly?: boolean;
}) {
  return (
    <span className="sc-meter">
      <span className="sc-segments" aria-hidden="true">
        {Array.from({ length: Math.min(120, Math.max(1, max)) }, (_, i) => (
          <i
            key={i}
            style={
              i * Math.max(1, max / 120) < booked ? { background: occupancyColor(booked, max) } : undefined
            }
          />
        ))}
      </span>
      {!meterOnly && (
        <small>
          {booked}/{max}
        </small>
      )}
    </span>
  );
}
