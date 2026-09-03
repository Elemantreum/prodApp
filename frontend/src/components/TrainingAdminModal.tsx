import { useCallback, useEffect, useState } from "react";
import type { FormEvent } from "react";
import { api, ApiError } from "../api/client";
import type { BookingDetail, BookingStatus, Training } from "../api/types";
import { isoToLocalDateTimeInput, localDateTimeInputToIso } from "../utils/date";
import { minutesToNs, nsToMinutes } from "../utils/duration";
import { Button } from "./Button";
import { ErrorBanner } from "./ErrorBanner";
import { Spinner } from "./Spinner";

const STATUS_LABEL: Record<BookingStatus, string> = {
  active: "Активна",
  cancelled: "Отменена",
  completed: "Пришёл",
  not_completed: "Не пришёл",
};

interface Props {
  training: Training;
  onClose: () => void;
  onUpdated: (training: Training) => void;
}

/** Всплывающее окно администратора на карточке тренировки: редактирование
 *  самой тренировки сверху, отметка посещаемости по каждой записи снизу.
 *  Открывается только для админов (см. TrainingCard). */
export function TrainingAdminModal({ training, onClose, onUpdated }: Props) {
  const [description, setDescription] = useState(training.description);
  const [date, setDate] = useState(isoToLocalDateTimeInput(training.date));
  const [maxQuantity, setMaxQuantity] = useState(training.maxQuantity);
  const [durationMinutes, setDurationMinutes] = useState(nsToMinutes(training.duration));
  const [isActive, setIsActive] = useState(training.isActive);
  const [saving, setSaving] = useState(false);
  const [saveError, setSaveError] = useState<string | null>(null);
  const [saved, setSaved] = useState(false);

  const [bookings, setBookings] = useState<BookingDetail[] | null>(null);
  const [bookingsError, setBookingsError] = useState<string | null>(null);
  const [pendingId, setPendingId] = useState<number | null>(null);

  const loadBookings = useCallback(async () => {
    setBookingsError(null);
    try {
      setBookings(await api.getTrainingBookings(training.id));
    } catch (e) {
      setBookingsError(e instanceof ApiError ? e.message : "Не удалось загрузить список записей");
    }
  }, [training.id]);

  useEffect(() => {
    loadBookings();
  }, [loadBookings]);

  async function handleSave(e: FormEvent) {
    e.preventDefault();
    setSaving(true);
    setSaveError(null);
    setSaved(false);
    try {
      const updated = await api.updateTraining(training.id, {
        description: description.trim(),
        date: localDateTimeInputToIso(date),
        maxQuantity,
        duration: minutesToNs(durationMinutes),
        isActive,
      });
      onUpdated(updated);
      setSaved(true);
    } catch (e) {
      setSaveError(e instanceof ApiError ? e.message : "Не удалось сохранить");
    } finally {
      setSaving(false);
    }
  }

  async function handleAttendance(bookingId: number, attended: boolean) {
    setPendingId(bookingId);
    setBookingsError(null);
    try {
      const updated = await api.setAttendance(bookingId, attended);
      setBookings((prev) => prev?.map((b) => (b.bookingId === bookingId ? { ...b, status: updated.status } : b)) ?? null);
    } catch (e) {
      setBookingsError(e instanceof ApiError ? e.message : "Не удалось отметить посещаемость");
    } finally {
      setPendingId(null);
    }
  }

  return (
    <div className="modal-overlay" onClick={onClose}>
      <div className="modal-panel" onClick={(e) => e.stopPropagation()}>
        <div className="modal-header">
          <h2>Управление тренировкой</h2>
          <button className="modal-close" onClick={onClose} aria-label="Закрыть">
            ×
          </button>
        </div>

        <div className="modal-section">
          <h3>Данные тренировки</h3>
          <form className="form" onSubmit={handleSave}>
            <label className="form__field">
              <span>Описание</span>
              <input type="text" value={description} onChange={(e) => setDescription(e.target.value)} required />
            </label>
            <label className="form__field">
              <span>Дата и время</span>
              <input type="datetime-local" value={date} onChange={(e) => setDate(e.target.value)} required />
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
            <label className="form__checkbox">
              <input type="checkbox" checked={isActive} onChange={(e) => setIsActive(e.target.checked)} />
              <span>Тренировка активна (видна в списке)</span>
            </label>
            {saveError && <ErrorBanner message={saveError} />}
            {saved && <div className="success-banner">Сохранено</div>}
            <Button type="submit" loading={saving}>
              Сохранить
            </Button>
          </form>
        </div>

        <div className="modal-section">
          <h3>Посещаемость</h3>
          {bookings === null && !bookingsError && (
            <div className="centered">
              <Spinner size={20} />
            </div>
          )}
          {bookingsError && <ErrorBanner message={bookingsError} onRetry={loadBookings} />}
          {bookings?.length === 0 && <div className="participants__empty">Пока никто не записался</div>}
          {bookings && bookings.length > 0 && (
            <ul className="attendance-list">
              {bookings.map((b) => {
                const fullName = [b.firstName, b.lastName].filter(Boolean).join(" ");
                const busy = pendingId === b.bookingId;
                return (
                  <li key={b.bookingId} className="attendance-row">
                    <div className="attendance-row__info">
                      <div className="participants__name">{fullName || "Без имени"}</div>
                      <div className="participants__username">{b.username}</div>
                      <span className={`badge badge--${b.status}`}>{STATUS_LABEL[b.status]}</span>
                    </div>
                    {b.status !== "cancelled" && (
                      <div className="attendance-row__actions">
                        <Button
                          variant={b.status === "completed" ? "primary" : "secondary"}
                          loading={busy}
                          disabled={busy}
                          onClick={() => handleAttendance(b.bookingId, true)}
                        >
                          Пришёл
                        </Button>
                        <Button
                          variant={b.status === "not_completed" ? "danger" : "secondary"}
                          loading={busy}
                          disabled={busy}
                          onClick={() => handleAttendance(b.bookingId, false)}
                        >
                          Не пришёл
                        </Button>
                      </div>
                    )}
                  </li>
                );
              })}
            </ul>
          )}
        </div>
      </div>
    </div>
  );
}
