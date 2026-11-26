import type { Training } from "../api/types";
import { t } from "../i18n";
import { OccupancyBar } from "./OccupancyBar";
import { clock } from "./scheduleDates";

export function TrainingCard({
  training,
  onOpen,
}: {
  training: Training;
  onOpen: (training: Training) => void;
}) {
  const start = Date.parse(training.date);
  const status =
    start <= Date.now()
      ? "Занятие завершено"
      : !training.isActive
        ? "Неактивна"
        : training.bookedCount >= training.maxQuantity
          ? "Мест нет · Лист ожидания"
          : "Есть места";

  return (
    <button className="sc-agenda-row" type="button" onClick={() => onOpen(training)}>
      <span className="sc-times">
        {clock.format(start)}
        <small>{clock.format(start + training.duration / 1e6)}</small>
      </span>
      <span className="sc-grow">
        <b>{training.description}</b>
        <small>{[training.venue, t(status)].filter(Boolean).join(" · ")}</small>
        <OccupancyBar booked={training.bookedCount} max={training.maxQuantity} />
      </span>
      <span aria-hidden="true">›</span>
    </button>
  );
}
