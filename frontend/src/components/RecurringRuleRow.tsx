import { useState } from "react";
import type { FormEvent } from "react";
import { api, ApiError } from "../api/client";
import type { RecurringTraining } from "../api/types";
import { formatDateTime, isoToLocalDateTimeInput, localDateTimeInputToIso } from "../utils/date";
import { daysToNs, minutesToNs, nsToDays, nsToMinutes } from "../utils/duration";
import { Button } from "./Button";
import { ErrorBanner } from "./ErrorBanner";

interface Props {
  rule: RecurringTraining;
  onUpdated: (rule: RecurringTraining) => void;
}

export function RecurringRuleRow({ rule, onUpdated }: Props) {
  const [editing, setEditing] = useState(false);
  const [description, setDescription] = useState(rule.description);
  const [anchorDate, setAnchorDate] = useState(isoToLocalDateTimeInput(rule.anchorDate));
  const [maxQuantity, setMaxQuantity] = useState(rule.maxQuantity);
  const [durationMinutes, setDurationMinutes] = useState(nsToMinutes(rule.duration));
  const [leadDays, setLeadDays] = useState(nsToDays(rule.leadTime));
  const [isActive, setIsActive] = useState(rule.isActive);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);

  function startEditing() {
    setDescription(rule.description);
    setAnchorDate(isoToLocalDateTimeInput(rule.anchorDate));
    setMaxQuantity(rule.maxQuantity);
    setDurationMinutes(nsToMinutes(rule.duration));
    setLeadDays(nsToDays(rule.leadTime));
    setIsActive(rule.isActive);
    setError(null);
    setEditing(true);
  }

  async function handleSave(e: FormEvent) {
    e.preventDefault();
    setSaving(true);
    setError(null);
    try {
      const updated = await api.updateRecurringTraining(rule.id, {
        description: description.trim(),
        anchorDate: localDateTimeInputToIso(anchorDate),
        maxQuantity,
        duration: minutesToNs(durationMinutes),
        leadTime: daysToNs(leadDays),
        isActive,
      });
      onUpdated(updated);
      setEditing(false);
    } catch (e) {
      setError(e instanceof ApiError ? e.message : "Не удалось сохранить");
    } finally {
      setSaving(false);
    }
  }

  if (editing) {
    return (
      <li className="attendance-row attendance-row--column">
        <form className="form" onSubmit={handleSave}>
          <label className="form__field">
            <span>Описание</span>
            <input type="text" value={description} onChange={(e) => setDescription(e.target.value)} required />
          </label>
          <label className="form__field">
            <span>Дата и время первого занятия</span>
            <input type="datetime-local" value={anchorDate} onChange={(e) => setAnchorDate(e.target.value)} required />
          </label>
          <label className="form__field">
            <span>Мест</span>
            <input
              type="number"
              min={1}
              value={maxQuantity}
              onChange={(e) => setMaxQuantity(Number(e.target.value))}
              required
            />
          </label>
          <label className="form__field">
            <span>Длительность, мин</span>
            <input
              type="number"
              min={1}
              value={durationMinutes}
              onChange={(e) => setDurationMinutes(Number(e.target.value))}
              required
            />
          </label>
          <label className="form__field">
            <span>Создавать за сколько дней до занятия</span>
            <input type="number" min={0} value={leadDays} onChange={(e) => setLeadDays(Number(e.target.value))} required />
          </label>
          <label className="form__checkbox">
            <input type="checkbox" checked={isActive} onChange={(e) => setIsActive(e.target.checked)} />
            <span>Правило активно</span>
          </label>
          {error && <ErrorBanner message={error} />}
          <div className="user-card__actions-row">
            <Button variant="secondary" type="button" disabled={saving} onClick={() => setEditing(false)}>
              Отмена
            </Button>
            <Button type="submit" loading={saving}>
              Сохранить
            </Button>
          </div>
        </form>
      </li>
    );
  }

  return (
    <li className="attendance-row">
      <div className="attendance-row__info">
        <div className="participants__name">{rule.description}</div>
        <div className="card__meta">С {formatDateTime(rule.anchorDate)}, каждую неделю</div>
        <div className="card__meta">Создаётся за {nsToDays(rule.leadTime)} дн. до занятия</div>
        <span className={`badge ${rule.isActive ? "badge--active" : "badge--cancelled"}`}>
          {rule.isActive ? "Активна" : "Остановлена"}
        </span>
      </div>
      <div className="attendance-row__actions">
        <Button variant="secondary" onClick={startEditing}>
          Изменить
        </Button>
      </div>
    </li>
  );
}
