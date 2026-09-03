import { useCallback, useEffect, useState } from "react";
import type { FormEvent } from "react";
import { api, ApiError } from "../api/client";
import type { RecurringTraining } from "../api/types";
import { Button } from "../components/Button";
import { ErrorBanner } from "../components/ErrorBanner";
import { RecurringRuleRow } from "../components/RecurringRuleRow";
import { Spinner } from "../components/Spinner";
import { localDateTimeInputToIso } from "../utils/date";
import { daysToNs, minutesToNs } from "../utils/duration";

export function AdminCreateTrainingPage() {
  const [description, setDescription] = useState("");
  const [date, setDate] = useState("");
  const [maxQuantity, setMaxQuantity] = useState(10);
  const [durationMinutes, setDurationMinutes] = useState(60);
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState(false);

  const [rDescription, setRDescription] = useState("");
  const [rAnchorDate, setRAnchorDate] = useState("");
  const [rMaxQuantity, setRMaxQuantity] = useState(10);
  const [rDurationMinutes, setRDurationMinutes] = useState(60);
  const [rLeadDays, setRLeadDays] = useState(28);
  const [rSubmitting, setRSubmitting] = useState(false);
  const [rError, setRError] = useState<string | null>(null);
  const [rSuccess, setRSuccess] = useState(false);

  const [rules, setRules] = useState<RecurringTraining[] | null>(null);
  const [rulesError, setRulesError] = useState<string | null>(null);

  const loadRules = useCallback(async () => {
    setRulesError(null);
    try {
      setRules(await api.getRecurringTrainings());
    } catch (e) {
      setRulesError(e instanceof ApiError ? e.message : "Не удалось загрузить список");
    }
  }, []);

  useEffect(() => {
    loadRules();
  }, [loadRules]);

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

  async function handleRecurringSubmit(e: FormEvent) {
    e.preventDefault();
    if (!rDescription.trim() || !rAnchorDate) {
      setRError("Заполни описание и дату первого занятия");
      return;
    }

    setRSubmitting(true);
    setRError(null);
    setRSuccess(false);
    try {
      await api.createRecurringTraining({
        description: rDescription.trim(),
        anchorDate: localDateTimeInputToIso(rAnchorDate),
        maxQuantity: rMaxQuantity,
        duration: minutesToNs(rDurationMinutes),
        leadTime: daysToNs(rLeadDays),
      });
      setRSuccess(true);
      setRDescription("");
      setRAnchorDate("");
      setRMaxQuantity(10);
      setRDurationMinutes(60);
      setRLeadDays(28);
      loadRules();
    } catch (e) {
      setRError(e instanceof ApiError ? e.message : "Не удалось создать регулярную тренировку");
    } finally {
      setRSubmitting(false);
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

      <div className="modal-section">
        <h3>Регулярная тренировка</h3>
        <p className="card__meta">
          Указанная ниже дата — первое занятие. Дальше оно само повторяется каждую неделю в то же время, пока правило
          не остановлено.
        </p>

        <form className="form" onSubmit={handleRecurringSubmit}>
          <label className="form__field">
            <span>Описание</span>
            <input
              type="text"
              value={rDescription}
              onChange={(e) => setRDescription(e.target.value)}
              placeholder="Йога для начинающих"
              required
            />
          </label>

          <label className="form__field">
            <span>Дата и время первого занятия</span>
            <input type="datetime-local" value={rAnchorDate} onChange={(e) => setRAnchorDate(e.target.value)} required />
          </label>

          <label className="form__field">
            <span>Мест</span>
            <input
              type="number"
              min={1}
              value={rMaxQuantity}
              onChange={(e) => setRMaxQuantity(Number(e.target.value))}
              required
            />
          </label>

          <label className="form__field">
            <span>Длительность, мин</span>
            <input
              type="number"
              min={1}
              value={rDurationMinutes}
              onChange={(e) => setRDurationMinutes(Number(e.target.value))}
              required
            />
          </label>

          <label className="form__field">
            <span>Создавать за сколько дней до занятия</span>
            <input type="number" min={0} value={rLeadDays} onChange={(e) => setRLeadDays(Number(e.target.value))} required />
          </label>

          {rError && <ErrorBanner message={rError} />}
          {rSuccess && <div className="success-banner">Готово — ближайшие занятия уже появились в списке тренировок</div>}
          <Button type="submit" loading={rSubmitting}>
            Создать регулярную тренировку
          </Button>
        </form>

        {rules === null && !rulesError && (
          <div className="centered">
            <Spinner size={20} />
          </div>
        )}
        {rulesError && <ErrorBanner message={rulesError} onRetry={loadRules} />}
        {rules?.length === 0 && <div className="participants__empty">Регулярных тренировок пока нет</div>}
        {rules && rules.length > 0 && (
          <ul className="attendance-list">
            {rules.map((r) => (
              <RecurringRuleRow
                key={r.id}
                rule={r}
                onUpdated={(updated) => setRules((prev) => prev?.map((x) => (x.id === updated.id ? updated : x)) ?? prev)}
              />
            ))}
          </ul>
        )}
      </div>
    </>
  );
}
