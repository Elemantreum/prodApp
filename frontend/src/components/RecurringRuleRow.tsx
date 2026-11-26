import { t as translate } from "../i18n";
import { useState } from "react";
import type { FormEvent } from "react";
import { api, ApiError } from "../api/client";
import type { RecurringTraining } from "../api/types";
import { formatDateTime, isoToLocalDateTimeInput, localDateTimeInputToIso } from "../utils/date";
import { daysToNs, minutesToNs, nsToDays, nsToMinutes } from "../utils/duration";
import { AdminDialog } from "./AdminUI";
import { Button } from "./Button";
import { ErrorBanner } from "./ErrorBanner";
import { VenueSelect } from "./VenueSelect";

interface Props {
  rule: RecurringTraining;
  onDeleted: () => void;
  onUpdated: (rule: RecurringTraining) => void;
}

export function RecurringRuleRow({ rule, onUpdated, onDeleted }: Props) {
  const [confirmDelete, setConfirmDelete] = useState(false);
  const [editing, setEditing] = useState(false);
  const [description, setDescription] = useState(rule.description);
  const [details, setDetails] = useState(rule.details ?? "");
  const [venue, setVenue] = useState(rule.venue ?? "");
  const [anchorDate, setAnchorDate] = useState(isoToLocalDateTimeInput(rule.anchorDate));
  const [maxQuantity, setMaxQuantity] = useState(rule.maxQuantity);
  const [durationMinutes, setDurationMinutes] = useState(nsToMinutes(rule.duration));
  const [leadDays, setLeadDays] = useState(nsToDays(rule.leadTime));
  const [isActive, setIsActive] = useState(rule.isActive);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);

  function startEditing() {
    setConfirmDelete(false);
    setDescription(rule.description);
    setDetails(rule.details ?? "");
    setVenue(rule.venue ?? "");
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
        venue,
        description: description.trim(),
        details: details.trim(),
        anchorDate: localDateTimeInputToIso(anchorDate),
        maxQuantity,
        duration: minutesToNs(durationMinutes),
        leadTime: daysToNs(leadDays),
        isActive,
      });
      onUpdated(updated);
      setEditing(false);
    } catch (e) {
      setError(e instanceof ApiError ? e.message : translate("Не удалось сохранить"));
    } finally {
      setSaving(false);
    }
  }

  return (
    <li className="row">
      <span className="avatar" aria-hidden="true">
        ↻
      </span>
      <div className="grow">
        <b>{rule.description}</b>
        <small>
          {formatDateTime(rule.anchorDate)} · {translate("Еженедельно")} · {rule.maxQuantity}{" "}
          {translate("мест")}
          {rule.venue && ` · ${rule.venue}`}
        </small>
        {!rule.isActive && <small>{translate("Остановлена")}</small>}
      </div>
      <Button variant="secondary" onClick={startEditing}>
        {translate("Изменить")}
      </Button>
      {editing && (
        <AdminDialog title={translate("Редактировать серию")} busy={saving} onClose={() => setEditing(false)}>
          <form className="form" onSubmit={handleSave}>
            <VenueSelect value={venue} onChange={setVenue} savedVenue={rule.venue} />
            <label className="form__field">
              <span>{translate("Название")}</span>
              <input
                type="text"
                value={description}
                onChange={(e) => setDescription(e.target.value)}
                required
              />
            </label>
            <label className="form__field">
              <span>{translate("Описание (необязательно)")}</span>
              <textarea rows={3} value={details} onChange={(e) => setDetails(e.target.value)} />
            </label>
            <label className="form__field">
              <span>{translate("Дата и время первого занятия")}</span>
              <input
                type="datetime-local"
                value={anchorDate}
                onChange={(e) => setAnchorDate(e.target.value)}
                required
              />
            </label>
            <label className="form__field">
              <span>{translate("Мест")}</span>
              <input
                type="number"
                min={1}
                value={maxQuantity}
                onChange={(e) => setMaxQuantity(Number(e.target.value))}
                required
              />
            </label>
            <label className="form__field">
              <span>{translate("Длительность, мин")}</span>
              <input
                type="number"
                min={1}
                value={durationMinutes}
                onChange={(e) => setDurationMinutes(Number(e.target.value))}
                required
              />
            </label>
            <label className="form__field">
              <span>{translate("Создавать за сколько дней до занятия")}</span>
              <input
                type="number"
                min={0}
                value={leadDays}
                onChange={(e) => setLeadDays(Number(e.target.value))}
                required
              />
            </label>
            <label className="form__checkbox">
              <input type="checkbox" checked={isActive} onChange={(e) => setIsActive(e.target.checked)} />
              <span>{translate("Правило активно")}</span>
            </label>
            {error && <ErrorBanner message={error} />}
            <div className="user-card__actions-row">
              <Button variant="secondary" type="button" disabled={saving} onClick={() => setEditing(false)}>
                {" "}
                {translate("Отмена")}{" "}
              </Button>
              <Button type="submit" loading={saving}>
                {" "}
                {translate("Сохранить")}{" "}
              </Button>
            </div>
          </form>
          <Button
            variant="danger"
            loading={saving}
            onClick={async () => {
              if (!confirmDelete) {
                setConfirmDelete(true);
                return;
              }
              setSaving(true);
              try {
                await api.deleteRule(rule.id);
                onDeleted();
              } catch (e) {
                setError(String(e));
                setConfirmDelete(false);
              } finally {
                setSaving(false);
              }
            }}
          >
            {confirmDelete ? translate("Подтвердить удаление серии") : translate("Удалить серию")}
          </Button>
        </AdminDialog>
      )}
    </li>
  );
}
