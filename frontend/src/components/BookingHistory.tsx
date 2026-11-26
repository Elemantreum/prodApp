import { useEffect, useState } from "react";
import { Link } from "react-router-dom";
import { api } from "../api/client";
import type { BookingStatus, HistoryItem, Training } from "../api/types";
import { t } from "../i18n";
import { formatDateTime, isoToLocalDateTimeInput } from "../utils/date";
import { AdminDialog, AdminSearch, useAdminCreate } from "./AdminUI";
import { ErrorBanner } from "./ErrorBanner";
import { useSection } from "../auth/Sections";

const labels: Record<BookingStatus, string> = {
  active: "Записан",
  cancelled: "Запись отменена",
  completed: "Был",
  not_completed: "Не был",
};

function BookingEditor({ item, close, saved }: { item?: HistoryItem; close: () => void; saved: () => void }) {
  const [people, setPeople] = useState<Awaited<ReturnType<typeof api.sectionMembers>>>([]);
  const [sessions, setSessions] = useState<Training[]>([]);
  const [person, setPerson] = useState(item?.userId ?? 0);
  const [target, setTarget] = useState(item?.trainingId ?? 0);
  const [error, setError] = useState("");
  const [confirm, setConfirm] = useState(false);
  const [busy, setBusy] = useState(false);
  const adding = !item;
  const editable = adding || item.status === "active";
  useEffect(() => {
    let active = true;
    if (editable)
      Promise.all([api.sectionMembers(), api.getTrainings()])
        .then(([members, trainings]) => {
          if (active) {
            setPeople(members);
            const now = Date.now();
            setSessions(trainings.filter((s) => (Date.parse(s.date) <= now ? adding : s.isActive)));
          }
        })
        .catch((e) => {
          if (active) setError(String(e));
        });
    return () => {
      active = false;
    };
  }, [adding, editable]);
  async function run(change: () => Promise<unknown>) {
    setBusy(true);
    setError("");
    try {
      await change();
      saved();
    } catch (e) {
      setError(String(e));
    } finally {
      setBusy(false);
    }
  }
  return (
    <AdminDialog title={item ? t("Редактировать запись") : t("Добавить запись")} busy={busy} onClose={close}>
      {item && (
        <p>
          {item.name} · {item.description}
          <small>
            {formatDateTime(item.date)} · {t(labels[item.status])}
          </small>
        </p>
      )}
      {editable && (
        <form
          onSubmit={(e) => {
            e.preventDefault();
            void run(() =>
              api.manageBooking({
                action: item ? "edit" : "create",
                trainingId: item?.trainingId ?? target,
                targetId: target,
                userId: person,
                bookingId: item?.id,
              }),
            );
          }}
        >
          <label>
            {t("Участник")}
            <select required value={person || ""} onChange={(e) => setPerson(Number(e.target.value))}>
              <option value="">{t("Выбери участника")}</option>
              {item && !people.some((p) => p.id === item.userId) && (
                <option value={item.userId}>{item.name}</option>
              )}
              {people.map((p) => (
                <option key={p.id} value={p.id}>
                  {p.firstName} {p.lastName}
                </option>
              ))}
            </select>
          </label>
          <label>
            {t("Тренировка")}
            <select required value={target || ""} onChange={(e) => setTarget(Number(e.target.value))}>
              <option value="">{t("Выберите занятие")}</option>
              {item && !sessions.some((s) => s.id === item.trainingId) && (
                <option value={item.trainingId}>
                  {item.description} · {formatDateTime(item.date)}
                </option>
              )}
              {sessions.map((s) => (
                <option key={s.id} value={s.id}>
                  {s.description} · {formatDateTime(s.date)}
                </option>
              ))}
            </select>
          </label>
          <button
            className="btn primary"
            disabled={busy || !target || !person || (target === item?.trainingId && person === item?.userId)}
          >
            {t("Сохранить")}
          </button>
        </form>
      )}
      {item && (
        <div className="admin-actions">
          {item.status === "active" && (
            <button className="btn" disabled={busy} onClick={() => run(() => api.cancelBooking(item.id))}>
              {t("Отменить запись")}
            </button>
          )}
          {item.status !== "cancelled" &&
            Date.parse(item.date) <= Date.now() &&
            [true, false].map((attended) => (
              <button
                className="btn"
                key={String(attended)}
                disabled={busy || item.status === (attended ? "completed" : "not_completed")}
                onClick={() => run(() => api.setAttendance(item.id, attended))}
              >
                {attended ? t("Был") : t("Не был")}
              </button>
            ))}
          {["completed", "not_completed"].includes(item.status) && (
            <button
              className="btn"
              disabled={busy}
              onClick={() =>
                run(() =>
                  api.manageBooking({
                    action: "reset",
                    trainingId: item.trainingId,
                    bookingId: item.id,
                  }),
                )
              }
            >
              {t("Сбросить отметку посещаемости")}
            </button>
          )}
          <button
            className="btn danger"
            disabled={busy}
            onClick={() =>
              confirm
                ? run(() =>
                    api.manageBooking({
                      action: "delete",
                      trainingId: item.trainingId,
                      bookingId: item.id,
                    }),
                  )
                : setConfirm(true)
            }
          >
            {confirm ? t("Подтвердить удаление записи") : t("Удалить запись")}
          </button>
        </div>
      )}
      {error && <ErrorBanner message={error} />}
    </AdminDialog>
  );
}

export function BookingHistory({ userId, attendance = false }: { userId?: number; attendance?: boolean }) {
  const { current, directory, choose } = useSection();
  const [items, setItems] = useState<HistoryItem[]>();
  const [error, setError] = useState("");
  const [query, setQuery] = useState("");
  const [open, setOpen] = useAdminCreate();
  const [selected, setSelected] = useState<HistoryItem>();
  const [version, setVersion] = useState(0);
  useEffect(() => {
    let active = true;
    setError("");
    api
      .history(userId)
      .then((rows) => {
        if (active) setItems(rows ?? []);
      })
      .catch((e) => {
        if (active) setError(String(e));
      });
    return () => {
      active = false;
    };
  }, [current.id, userId, version]);
  const visible = items?.filter(
    (b) =>
      (!attendance || b.status === "completed" || b.status === "not_completed") &&
      [b.name, b.description, b.section].join(" ").toLowerCase().includes(query.trim().toLowerCase()),
  );
  function close() {
    setSelected(undefined);
    setOpen(false);
  }
  return (
    <section className="card">
      {attendance ? <h2>{t("Отметки посещаемости")}</h2> : <AdminSearch value={query} onChange={setQuery} />}
      {error && <ErrorBanner message={error} onRetry={() => setVersion((v) => v + 1)} />}
      {!items && !error && <p role="status">{t("Загрузка…")}</p>}
      {visible?.length === 0 && (
        <p>{attendance ? t("Пока нет отметок посещаемости.") : t("Записей пока нет")}</p>
      )}
      {visible?.map((b) => (
        <div key={b.id} className="row">
          <span className="avatar" aria-hidden="true">
            {b.name.trim()[0]}
          </span>
          <div className="grow">
            <b>{b.name}</b>
            <small>
              <Link
                to={`/?session=${b.trainingId}&day=${isoToLocalDateTimeInput(b.date).slice(0, 10)}`}
                onClick={() => {
                  if (userId && b.section !== current.name) {
                    const section = directory.sections.find((s) => s.name === b.section);
                    if (section) choose(section.id);
                  }
                }}
              >
                {b.description} · {formatDateTime(b.date)}
              </Link>
            </small>
            <small>
              {b.section} · {t(labels[b.status])}
            </small>
          </div>
          {!userId && (
            <button className="btn" onClick={() => setSelected(b)}>
              {t("Изменить")}
            </button>
          )}
        </div>
      ))}
      {!userId && (selected || (open && !attendance)) && (
        <BookingEditor
          item={selected}
          close={close}
          saved={() => {
            close();
            setVersion((v) => v + 1);
          }}
        />
      )}
    </section>
  );
}
export function AdminBookingsPage() {
  return <BookingHistory />;
}
export function AdminAttendancePage() {
  return <BookingHistory attendance />;
}
