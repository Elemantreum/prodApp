import { useState, type FormEvent } from "react";
import { t as translate } from "../i18n";
import { api, ApiError } from "../api/client";
import type { Training } from "../api/types";
import { isoToLocalDateTimeInput, localDateTimeInputToIso } from "../utils/date";
import { minutesToNs, nsToMinutes } from "../utils/duration";
import { Button } from "./Button";
import { ErrorBanner } from "./ErrorBanner";
import { VenueSelect } from "./VenueSelect";

export function TrainingEditor({
  training,
  onUpdated,
  onDeleted,
}: {
  training: Training;
  onUpdated: (item: Training) => void;
  onDeleted: (id: number) => void;
}) {
  const [venue, setVenue] = useState(training.venue ?? "");
  const [description, setDescription] = useState(training.description);
  const [details, setDetails] = useState(training.details ?? "");
  const [date, setDate] = useState(isoToLocalDateTimeInput(training.date));
  const [maxQuantity, setMaxQuantity] = useState(training.maxQuantity);
  const [durationMinutes, setDurationMinutes] = useState(nsToMinutes(training.duration));
  const [isActive, setIsActive] = useState(training.isActive);
  const [saving, setSaving] = useState(false);
  const [saveError, setSaveError] = useState<string | null>(null);
  const [saved, setSaved] = useState(false);

  const [confirmingDelete, setConfirmingDelete] = useState(false);
  const [deleting, setDeleting] = useState(false);
  const [deleteError, setDeleteError] = useState<string | null>(null);

  async function handleSave(e: FormEvent) {
    e.preventDefault();
    setSaving(true);
    setSaveError(null);
    setSaved(false);
    try {
      const updated = await api.updateTraining(training.id, {
        description: description.trim(),
        details: details.trim(),
        venue,
        date: localDateTimeInputToIso(date),
        maxQuantity,
        duration: minutesToNs(durationMinutes),
        isActive,
      });
      onUpdated(updated);
      setSaved(true);
    } catch (e) {
      setSaveError(e instanceof ApiError ? e.message : translate("Не удалось сохранить"));
    } finally {
      setSaving(false);
    }
  }

  async function handleDeleteClick() {
    if (!confirmingDelete) {
      setConfirmingDelete(true);
      return;
    }

    setDeleting(true);
    setDeleteError(null);
    try {
      await api.deleteTraining(training.id);
      onDeleted(training.id);
    } catch (e) {
      setDeleteError(e instanceof ApiError ? e.message : translate("Не удалось удалить"));
      setConfirmingDelete(false);
    } finally {
      setDeleting(false);
    }
  }

  return (
    <>
      <form className="form" onSubmit={handleSave}>
        <VenueSelect value={venue} onChange={setVenue} savedVenue={training.venue} />
        <label className="form__field">
          <span>{translate("Название")}</span>
          <input type="text" value={description} onChange={(e) => setDescription(e.target.value)} required />
        </label>
        <label className="form__field">
          <span>{translate("Описание (необязательно)")}</span>
          <textarea rows={3} value={details} onChange={(e) => setDetails(e.target.value)} />
        </label>
        <label className="form__field">
          <span>{translate("Дата и время")}</span>
          <input type="datetime-local" value={date} onChange={(e) => setDate(e.target.value)} required />
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
        <label className="form__checkbox">
          <input type="checkbox" checked={isActive} onChange={(e) => setIsActive(e.target.checked)} />
          <span>{translate("Тренировка активна (видна в списке)")}</span>
        </label>
        {saveError && <ErrorBanner message={saveError} />}
        {saved && <div className="success-banner">{translate("Сохранено")}</div>}
        <Button type="submit" loading={saving}>
          {" "}
          {translate("Сохранить")}{" "}
        </Button>
      </form>
      {deleteError && <ErrorBanner message={deleteError} />}
      {confirmingDelete ? (
        <div className="user-card__actions-row">
          <Button variant="secondary" disabled={deleting} onClick={() => setConfirmingDelete(false)}>
            {" "}
            {translate("Отмена")}{" "}
          </Button>
          <Button variant="danger" loading={deleting} onClick={handleDeleteClick}>
            {" "}
            {translate("Точно удалить?")}{" "}
          </Button>
        </div>
      ) : (
        <Button variant="danger" onClick={handleDeleteClick}>
          {" "}
          {translate("Удалить тренировку")}{" "}
        </Button>
      )}
    </>
  );
}
