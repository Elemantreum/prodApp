import { AdminDialog, AdminSearch, useAdminCreate } from "../components/AdminUI";
import { t as translate } from "../i18n";
import { useCallback, useEffect, useState } from "react";
import type { FormEvent } from "react";
import { api, ApiError } from "../api/client";
import type { RecurringTraining, Training } from "../api/types";
import { Button } from "../components/Button";
import { ErrorBanner } from "../components/ErrorBanner";
import { RecurringRuleRow } from "../components/RecurringRuleRow";
import { Spinner } from "../components/Spinner";
import { VenueSelect } from "../components/VenueSelect";
import { localDateTimeInputToIso } from "../utils/date";
import { daysToNs, minutesToNs } from "../utils/duration";

export function AdminCreateTrainingPage({ onCreated }: { onCreated?: (training: Training) => void }) {
  const [venue, setVenue] = useState("");
  const [description, setDescription] = useState("");
  const [details, setDetails] = useState("");
  const [date, setDate] = useState("");
  const [maxQuantity, setMaxQuantity] = useState(10);
  const [durationMinutes, setDurationMinutes] = useState(60);
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState(false);

  async function handleSubmit(e: FormEvent) {
    e.preventDefault();
    if (!description.trim() || !date) {
      setError(translate("Заполни название и дату"));
      return;
    }

    setSubmitting(true);
    setError(null);
    setSuccess(false);
    try {
      const created = await api.createTraining({
        description: description.trim(),
        details: details.trim(),
        venue,
        date: localDateTimeInputToIso(date),
        maxQuantity,
        duration: minutesToNs(durationMinutes),
      });
      onCreated?.(created);
      setSuccess(true);
      setDescription("");
      setDetails("");
      setVenue("");
      setDate("");
      setMaxQuantity(10);
      setDurationMinutes(60);
    } catch (e) {
      setError(e instanceof ApiError ? e.message : translate("Не удалось создать тренировку"));
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <div className={onCreated ? undefined : "card"}>
      {error && <ErrorBanner message={error} />}
      {success && <div className="success-banner">{translate("Тренировка создана")}</div>}

      <form className="form" onSubmit={handleSubmit}>
        <VenueSelect value={venue} onChange={setVenue} />
        <label className="form__field">
          <span>{translate("Название")}</span>
          <input
            type="text"
            value={description}
            onChange={(e) => setDescription(e.target.value)}
            placeholder={translate("Йога для начинающих")}
            required
          />
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

        <Button type="submit" loading={submitting}>
          {" "}
          {translate("Создать")}{" "}
        </Button>
      </form>
    </div>
  );
}

export function AdminSeriesPage() {
  const [rVenue, setRVenue] = useState("");
  const [open, setOpen] = useAdminCreate();
  const [query, setQuery] = useState("");
  const [rDescription, setRDescription] = useState("");
  const [rDetails, setRDetails] = useState("");
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
      setRulesError(e instanceof ApiError ? e.message : translate("Не удалось загрузить список"));
    }
  }, []);

  useEffect(() => {
    loadRules();
  }, [loadRules]);

  async function handleRecurringSubmit(e: FormEvent) {
    e.preventDefault();
    if (!rDescription.trim() || !rAnchorDate) {
      setRError(translate("Заполни название и дату первого занятия"));
      return;
    }

    setRSubmitting(true);
    setRError(null);
    setRSuccess(false);
    try {
      await api.createRecurringTraining({
        venue: rVenue,
        description: rDescription.trim(),
        details: rDetails.trim(),
        anchorDate: localDateTimeInputToIso(rAnchorDate),
        maxQuantity: rMaxQuantity,
        duration: minutesToNs(rDurationMinutes),
        leadTime: daysToNs(rLeadDays),
      });
      setRSuccess(true);
      setOpen(false);
      setRDescription("");
      setRDetails("");
      setRVenue("");
      setRAnchorDate("");
      setRMaxQuantity(10);
      setRDurationMinutes(60);
      setRLeadDays(28);
      loadRules();
    } catch (e) {
      setRError(e instanceof ApiError ? e.message : translate("Не удалось создать регулярную тренировку"));
    } finally {
      setRSubmitting(false);
    }
  }

  const visible = (rules ?? []).filter((rule) =>
    rule.description.toLowerCase().includes(query.trim().toLowerCase()),
  );
  return (
    <section className="card">
      <AdminSearch value={query} onChange={setQuery} />
      {rSuccess && <p role="status">{translate("Серия создана")}</p>}
      {open && (
        <AdminDialog title={translate("Создать серию")} busy={rSubmitting} onClose={() => setOpen(false)}>
          <p className="card__meta">
            {" "}
            {translate(
              "Указанная ниже дата — первое занятие. Дальше оно само повторяется каждую неделю в то же время, пока правило не остановлено.",
            )}{" "}
          </p>

          <form className="form" onSubmit={handleRecurringSubmit}>
            <VenueSelect value={rVenue} onChange={setRVenue} />
            <label className="form__field">
              <span>{translate("Название")}</span>
              <input
                type="text"
                value={rDescription}
                onChange={(e) => setRDescription(e.target.value)}
                placeholder={translate("Йога для начинающих")}
                required
              />
            </label>

            <label className="form__field">
              <span>{translate("Описание (необязательно)")}</span>
              <textarea rows={3} value={rDetails} onChange={(e) => setRDetails(e.target.value)} />
            </label>

            <label className="form__field">
              <span>{translate("Дата и время первого занятия")}</span>
              <input
                type="datetime-local"
                value={rAnchorDate}
                onChange={(e) => setRAnchorDate(e.target.value)}
                required
              />
            </label>

            <label className="form__field">
              <span>{translate("Мест")}</span>
              <input
                type="number"
                min={1}
                value={rMaxQuantity}
                onChange={(e) => setRMaxQuantity(Number(e.target.value))}
                required
              />
            </label>

            <label className="form__field">
              <span>{translate("Длительность, мин")}</span>
              <input
                type="number"
                min={1}
                value={rDurationMinutes}
                onChange={(e) => setRDurationMinutes(Number(e.target.value))}
                required
              />
            </label>

            <label className="form__field">
              <span>{translate("Создавать за сколько дней до занятия")}</span>
              <input
                type="number"
                min={0}
                value={rLeadDays}
                onChange={(e) => setRLeadDays(Number(e.target.value))}
                required
              />
            </label>

            {rError && <ErrorBanner message={rError} />}
            <Button type="submit" loading={rSubmitting}>
              {" "}
              {translate("Создать регулярную тренировку")}{" "}
            </Button>
          </form>
        </AdminDialog>
      )}
      {rules === null && !rulesError && (
        <div className="centered">
          <Spinner size={20} />
        </div>
      )}
      {rulesError && <ErrorBanner message={rulesError} onRetry={loadRules} />}
      {rules && visible.length === 0 && (
        <div className="participants__empty">{translate("Регулярных тренировок пока нет")}</div>
      )}
      {rules && rules.length > 0 && (
        <ul className="attendance-list">
          {visible.map((r) => (
            <RecurringRuleRow
              key={r.id}
              rule={r}
              onDeleted={() => setRules((prev) => prev?.filter((item) => item.id !== r.id) ?? null)}
              onUpdated={(updated) =>
                setRules((prev) => prev?.map((x) => (x.id === updated.id ? updated : x)) ?? prev)
              }
            />
          ))}
        </ul>
      )}
    </section>
  );
}
