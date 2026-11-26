import { useEffect, useState } from "react";
import { api } from "../api/client";
import type { BookingDetail, Training } from "../api/types";
import { locale, t } from "../i18n";

type Action = "create" | "edit" | "cancel" | "delete" | "reset" | "wait" | "unwait";

export function ManageRoster({
  training,
  bookings,
  refresh,
  selectedBooking,
  onDone,
}: {
  training: Training;
  bookings: BookingDetail[];
  refresh: () => Promise<void>;
  selectedBooking?: number;
  onDone?: () => void;
}) {
  const [bookingId, setBookingId] = useState(selectedBooking ?? 0);
  const booking = bookings.find((b) => b.bookingId === bookingId);
  const past = Date.parse(training.date) <= Date.now();
  const count = past ? bookings.filter((b) => b.status !== "cancelled").length : training.bookedCount;
  const full = count >= training.maxQuantity;
  const waitInstead = !past && full;
  const [action, setAction] = useState<Action>(booking ? "edit" : waitInstead ? "wait" : "create");
  const [members, setMembers] = useState<Awaited<ReturnType<typeof api.sectionMembers>>>([]);
  const [sessions, setSessions] = useState<Training[]>([]);
  const [queue, setQueue] = useState<Awaited<ReturnType<typeof api.waitingMembers>>>([]);
  const [previous, setPrevious] = useState(0);
  const [person, setPerson] = useState(booking?.userId ?? 0);
  const [target, setTarget] = useState(training.id);
  const [error, setError] = useState("");
  const [busy, setBusy] = useState(false);
  const [confirm, setConfirm] = useState(false);
  useEffect(() => {
    let live = true;
    Promise.all([api.sectionMembers(), api.getTrainings(), past ? [] : api.waitingMembers(training.id)])
      .then(([m, s, q]) => {
        if (live) {
          setMembers(m ?? []);
          setSessions(s);
          setQueue(q ?? []);
        }
      })
      .catch((e) => {
        if (live) setError(String(e));
      });
    return () => {
      live = false;
    };
  }, [training.id, bookings, past]);
  const actions: Action[] = booking
    ? booking.status === "active"
      ? ["edit", "cancel", "delete"]
      : booking.status === "cancelled"
        ? ["delete"]
        : past
          ? ["edit", "reset", "delete"]
          : ["reset", "delete"]
    : waitInstead
      ? ["wait"]
      : ["create"];
  if (!booking && queue.length && selectedBooking === undefined) actions.push("unwait");
  const labels = {
    create: t("Добавить участника"),
    edit: t("Изменить запись"),
    cancel: t("Отменить запись"),
    delete: t("Удалить запись"),
    reset: t("Сбросить отметку посещаемости"),
    wait: t("Добавить в ожидание"),
    unwait: t("Убрать из ожидания"),
  };
  const method = actions.includes(action) ? action : actions[0];
  const needsPerson = ["create", "edit", "wait"].includes(method);
  const noPastCapacity = method === "create" && past && full;
  async function save() {
    if (method === "delete" && !confirm) {
      setConfirm(true);
      return;
    }
    setBusy(true);
    setError("");
    try {
      if (method === "cancel") await api.cancelBooking(bookingId);
      else
        await api.manageBooking({
          action: method,
          trainingId: training.id,
          bookingId,
          previousUserId: previous,
          userId: method === "unwait" ? previous : person,
          targetId: method === "edit" ? target : training.id,
        });
      await refresh();
      setConfirm(false);
      setBookingId(0);
      setPerson(0);
      setPrevious(0);
      setTarget(training.id);
      onDone?.();
    } catch (e) {
      setError(String(e));
    } finally {
      setBusy(false);
    }
  }
  return (
    <form
      className="sc-form"
      onSubmit={(e) => {
        e.preventDefault();
        void save();
      }}
    >
      {selectedBooking === undefined && (
        <label>
          {t("Запись")}
          <select
            value={bookingId}
            onChange={(e) => {
              const id = Number(e.target.value);
              setBookingId(id);
              setPerson(bookings.find((b) => b.bookingId === id)?.userId ?? 0);
              setTarget(training.id);
              setConfirm(false);
              setAction(id ? "edit" : waitInstead ? "wait" : "create");
            }}
          >
            <option value={0}>{t("Добавить новую запись")}</option>
            {bookings.map((b) => (
              <option key={b.bookingId} value={b.bookingId}>
                {b.firstName} {b.lastName} · #{b.bookingId}
              </option>
            ))}
          </select>
        </label>
      )}
      {actions.length > 1 && (
        <label>
          {t("Действие")}
          <select
            value={method}
            onChange={(e) => {
              setAction(e.target.value as Action);
              setConfirm(false);
            }}
          >
            {actions.map((a) => (
              <option key={a} value={a}>
                {labels[a]}
              </option>
            ))}
          </select>
        </label>
      )}
      {(method === "unwait" || (method === "wait" && selectedBooking === undefined)) && (
        <label>
          {t("Кто ожидает")}
          <select value={previous} onChange={(e) => setPrevious(Number(e.target.value))}>
            <option value={0}>{t("Новый участник")}</option>
            {queue.map((m) => (
              <option key={m.id} value={m.id}>
                {m.firstName} {m.lastName}
              </option>
            ))}
          </select>
        </label>
      )}
      {needsPerson && (
        <label>
          {t("Участник")}
          <select value={person} onChange={(e) => setPerson(Number(e.target.value))}>
            <option value={0}>{t("Выбери участника")}</option>
            {members
              .filter(
                (m) =>
                  ((method === "edit" && target !== training.id) ||
                    !bookings.some(
                      (b) =>
                        b.userId === m.id &&
                        b.bookingId !== bookingId &&
                        (past ? b.status !== "cancelled" : b.status === "active"),
                    )) &&
                  (method !== "wait" || !queue.some((w) => w.id === m.id)),
              )
              .map((m) => (
                <option key={m.id} value={m.id}>
                  {m.firstName} {m.lastName}
                </option>
              ))}
          </select>
        </label>
      )}
      {method === "edit" && (
        <label>
          {t("Тренировка")}
          <select value={target} onChange={(e) => setTarget(Number(e.target.value))}>
            {sessions
              .filter((s) => {
                if (s.id === training.id) return true;
                if (booking?.status !== "active") return false;
                return Date.parse(s.date) <= Date.now() ? past : s.isActive;
              })
              .map((s) => (
                <option key={s.id} value={s.id}>
                  {s.description} · {new Date(s.date).toLocaleString(locale)}
                </option>
              ))}
          </select>
        </label>
      )}
      {error && <p role="alert">{error}</p>}
      {noPastCapacity && <p role="status">{t("Мест нет")}</p>}
      <button
        className="btn primary"
        disabled={busy || noPastCapacity || (needsPerson && !person) || (method === "unwait" && !previous)}
      >
        {confirm ? t("Подтвердить удаление записи") : labels[method]}
      </button>
    </form>
  );
}
