import { useState } from "react";
import type { FormEvent } from "react";
import { api, ApiError } from "../api/client";
import { Button } from "../components/Button";
import { ErrorBanner } from "../components/ErrorBanner";
import { localDateTimeInputToIso } from "../utils/date";
import { minutesToNs } from "../utils/duration";

export function AdminCreateTrainingPage() {
  const [description, setDescription] = useState("");
  const [date, setDate] = useState("");
  const [maxQuantity, setMaxQuantity] = useState(10);
  const [durationMinutes, setDurationMinutes] = useState(60);
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState(false);

  async function handleSubmit(e: FormEvent) {
    e.preventDefault();
    if (!description.trim() || !date) {
      setError("Заполни описание и дату");
      return;
    }

    setSubmitting(true);
    setError(null);
    setSuccess(false);
    try {
      await api.createTraining({
        description: description.trim(),
        date: localDateTimeInputToIso(date),
        maxQuantity,
        duration: minutesToNs(durationMinutes),
      });
      setSuccess(true);
      setDescription("");
      setDate("");
      setMaxQuantity(10);
      setDurationMinutes(60);
    } catch (e) {
      setError(e instanceof ApiError ? e.message : "Не удалось создать тренировку");
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <>
      {error && <ErrorBanner message={error} />}
      {success && <div className="success-banner">Тренировка создана</div>}

      <form className="form" onSubmit={handleSubmit}>
        <label className="form__field">
          <span>Описание</span>
          <input
            type="text"
            value={description}
            onChange={(e) => setDescription(e.target.value)}
            placeholder="Йога для начинающих"
            required
          />
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

        <Button type="submit" loading={submitting}>
          Создать
        </Button>
      </form>
    </>
  );
}
