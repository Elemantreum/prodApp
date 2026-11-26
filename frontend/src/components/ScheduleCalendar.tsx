import { useSearchParams } from "react-router-dom";
import type { Training } from "../api/types";
import { locale, t } from "../i18n";
import { occupancyColor } from "../utils/occupancy";
import { clock, dayKey, fullDay, shiftDay } from "./scheduleDates";
import "./schedule.css";

interface Props {
  trainings: Training[];
  selected: Date;
  onSelect: (date: Date) => void;
  onOpen?: (training: Training) => void;
}

export function ScheduleCalendar({ trainings, selected, onSelect, onOpen }: Props) {
  const [params, setParams] = useSearchParams();
  const monthly = params.get("view") !== "list";
  const first = monthly ? new Date(selected.getFullYear(), selected.getMonth(), 1) : selected;
  const start = shiftDay(first, -((first.getDay() + 6) % 7));
  const end = new Date(selected.getFullYear(), selected.getMonth() + 1, 0);
  const days = monthly ? 7 * Math.ceil((((first.getDay() + 6) % 7) + end.getDate()) / 7) : 7;
  const weekdays = [t("Пн"), t("Вт"), t("Ср"), t("Чт"), t("Пт"), t("Сб"), t("Вс")];
  const byDay = new Map<string, Training[]>();
  for (const item of [...trainings].sort((a, b) => Date.parse(a.date) - Date.parse(b.date))) {
    const key = dayKey(new Date(item.date));
    if (!byDay.has(key)) byDay.set(key, []);
    byDay.get(key)!.push(item);
  }
  function move(direction: number) {
    if (!monthly) return onSelect(shiftDay(selected, direction * 7));
    const last = new Date(selected.getFullYear(), selected.getMonth() + direction + 1, 0);
    onSelect(new Date(last.getFullYear(), last.getMonth(), Math.min(selected.getDate(), last.getDate())));
  }
  function view(month: boolean) {
    const next = new URLSearchParams(params);
    if (month) next.delete("view");
    else next.set("view", "list");
    setParams(next);
  }
  function dot(item: Training) {
    return (
      <i
        aria-hidden="true"
        key={item.id}
        className="sc-dot"
        style={{ background: occupancyColor(item.bookedCount, item.maxQuantity) }}
      />
    );
  }
  return (
    <section className="sc-calendar" aria-label={t("Выбор дня тренировки")}>
      <div className="sc-tools">
        <div className="sc-switch">
          <button aria-pressed={monthly} onClick={() => view(true)}>
            {t("Календарь")}
          </button>
          <button aria-pressed={!monthly} onClick={() => view(false)}>
            {t("Список на день")}
          </button>
        </div>
        <div className="sc-period">
          <button aria-label={t("Предыдущий период")} onClick={() => move(-1)}>
            ‹
          </button>
          <strong>
            {monthly
              ? selected.toLocaleDateString(locale, { month: "long", year: "numeric" })
              : start.toLocaleDateString(locale, { day: "numeric", month: "short" }) +
                " – " +
                shiftDay(start, 6).toLocaleDateString(locale, {
                  day: "numeric",
                  month: "short",
                  year: "numeric",
                })}
          </strong>
          <button aria-label={t("Следующий период")} onClick={() => move(1)}>
            ›
          </button>
        </div>
      </div>
      <div className={monthly ? "sc-month" : "sc-week"}>
        {monthly &&
          weekdays.map((day) => (
            <span className="sc-weekday" key={day}>
              {day}
            </span>
          ))}
        {Array.from({ length: days }, (_, i) => {
          const date = shiftDay(start, i);
          const key = dayKey(date);
          const entries = byDay.get(key) ?? [];
          const chosen = key === dayKey(selected);
          const label = fullDay.format(date) + t(", тренировок:") + " " + entries.length;
          if (!monthly)
            return (
              <button key={key} aria-label={label} aria-pressed={chosen} onClick={() => onSelect(date)}>
                {weekdays[i]}
                <b>{date.getDate()}</b>
              </button>
            );
          return (
            <div key={key} className="sc-cell" data-other={date.getMonth() !== selected.getMonth()}>
              <button
                className="sc-date"
                aria-label={label}
                aria-pressed={chosen}
                onClick={() => onSelect(date)}
              >
                {date.getDate()}
                <span className="sc-dots">{entries.map(dot)}</span>
              </button>
              {entries.map((item) => (
                <button
                  key={item.id}
                  className="sc-chip"
                  onClick={() => (onOpen ? onOpen(item) : onSelect(date))}
                >
                  {dot(item)} <b>{clock.format(new Date(item.date))}</b> {item.description}
                </button>
              ))}
            </div>
          );
        })}
      </div>
    </section>
  );
}
