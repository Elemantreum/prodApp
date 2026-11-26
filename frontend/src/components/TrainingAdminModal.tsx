import * as Dialog from "@radix-ui/react-dialog";
import { useCallback, useEffect, useRef, useState } from "react";
import { api } from "../api/client";
import type { Booking, BookingDetail, Participant, Training } from "../api/types";
import { useAuth } from "../auth/AuthContext";
import { useProfile } from "../auth/Profile";
import { useSection } from "../auth/Sections";
import { useScheduleUpdates } from "../hooks/useScheduleUpdates";
import { locale, t } from "../i18n";
import { clock, fullDay } from "./scheduleDates";
import { ContactLinks } from "./ContactLinks";
import { ErrorBanner } from "./ErrorBanner";
import { ManageRoster } from "./ManageRoster";
import { TrainingEditor } from "./TrainingEditor";
import { OccupancyBar } from "./OccupancyBar";
import { Spinner } from "./Spinner";

const participantNames = new Intl.Collator(locale, { sensitivity: "base" });

function compareParticipants(a: Participant, b: Participant) {
  const aSurname = a.lastName.trim();
  const bSurname = b.lastName.trim();
  if (Boolean(aSurname) !== Boolean(bSurname)) return aSurname ? -1 : 1;
  return participantNames.compare(aSurname, bSurname) || participantNames.compare(a.firstName, b.firstName);
}

interface Props {
  training: Training;
  onClose: () => void;
  onUpdated: (training: Training) => void;
  onDeleted: (id: number) => void;
}

export function TrainingAdminModal({ training, onClose, onUpdated, onDeleted }: Props) {
  const { current } = useSection();
  const { profile } = useProfile();
  const telegramLinked = profile.identities?.some((identity) => identity.provider === "telegram");
  const { claims, token } = useAuth();
  const user = claims?.user_id;
  const start = Date.parse(training.date);
  const past = start <= Date.now();
  const [people, setPeople] = useState<(Participant | BookingDetail)[] | null>(null);
  const [bookings, setBookings] = useState<BookingDetail[]>([]);
  const [own, setOwn] = useState<Booking>();
  const [waiting, setWaiting] = useState(false);
  const [queue, setQueue] = useState<Awaited<ReturnType<typeof api.waitingMembers>> | null>(null);
  const [queueError, setQueueError] = useState("");
  const [actionError, setActionError] = useState("");
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");
  const [editor, setEditor] = useState<"training" | number | null>(null);
  const version = useRef(0);
  const refresh = useCallback(async () => {
    if (!user) return;
    const request = ++version.current;
    const loadQueue = async () => {
      if (past || !current.admin) return;
      try {
        const list = await api.waitingMembers(training.id);
        if (request === version.current) {
          setQueue(list ?? []);
          setQueueError("");
        }
      } catch (e) {
        if (request === version.current) setQueueError(String(e));
      }
    };
    const loadRoster = async () => {
      try {
        const [list, mine, ids] = await Promise.all([
          current.admin ? api.getTrainingBookings(training.id) : api.getTrainingParticipants(training.id),
          api.getBookingsByUserId(user),
          api.getWaitlist(),
        ]);
        if (request !== version.current) return;
        const details = list.filter((b): b is BookingDetail => "bookingId" in b);
        setBookings(details);
        setPeople(list.filter((b) => !("status" in b) || b.status !== "cancelled").sort(compareParticipants));
        setOwn(mine.find((b) => b.trainingId === training.id && b.status !== "cancelled"));
        setWaiting(ids.includes(training.id));
        setError("");
      } catch (e) {
        if (request === version.current) setError(String(e));
      }
    };
    await Promise.all([loadQueue(), loadRoster()]);
  }, [current.admin, training.id, past, user]);
  useEffect(() => {
    void refresh();
    return () => {
      version.current++;
    };
  }, [refresh, training.bookedCount]);
  useScheduleUpdates(token, refresh);
  async function apply(change: () => Promise<unknown>) {
    setBusy(true);
    setActionError("");
    try {
      await change();
      await refresh();
      onUpdated(await api.getTraining(training.id));
    } catch (e) {
      setActionError(e instanceof Error ? t(e.message) : t("Не удалось сохранить"));
      await refresh();
    } finally {
      setBusy(false);
    }
  }
  const showCapacity = !past || current.admin;
  const count = past ? (people?.length ?? training.bookedCount) : training.bookedCount;
  const full = count >= training.maxQuantity;
  const booked = own?.id;
  const inQueue = waiting;
  const button = booked
    ? t("Отменить запись")
    : !full
      ? t("Записаться")
      : inQueue
        ? t("Выйти из ожидания")
        : t("Встать в лист ожидания");
  function act() {
    void apply(() =>
      booked
        ? api.cancelBooking(booked)
        : full
          ? api.setWaiting(training.id, !inQueue)
          : api.signUp(training.id),
    );
  }
  const hint = past
    ? own?.status === "completed"
      ? t("Вы были на тренировке")
      : own?.status === "not_completed"
        ? t("Отмечена неявка")
        : booked
          ? t("Вы были записаны")
          : ""
    : booked
      ? t("Вы записаны")
      : inQueue
        ? full
          ? t("Вы в листе ожидания")
          : t("Освободилось место. Можно записаться.")
        : full && telegramLinked
          ? t("Когда появится место, бот напишет в Telegram.")
          : "";
  const selected = bookings.find((b) => b.bookingId === editor);
  return (
    <Dialog.Root
      open
      onOpenChange={(open) => {
        if (!open) onClose();
      }}
    >
      <Dialog.Portal>
        <Dialog.Overlay className="sc-overlay" />
        <Dialog.Content className="sc-dialog" aria-describedby={undefined}>
          <header className="sc-heading">
            <div className="sc-title">
              <Dialog.Title>{training.description}</Dialog.Title>
              {current.admin && (
                <button aria-label={t("Изменить тренировку")} onClick={() => setEditor("training")}>
                  ⋯
                </button>
              )}
              <Dialog.Close aria-label={t("Закрыть")}>×</Dialog.Close>
            </div>
            <div className="sc-meta">
              {current.name} · {fullDay.format(new Date(training.date))}
            </div>
            <b>
              {[clock.format(start) + "–" + clock.format(start + training.duration / 1e6), training.venue]
                .filter(Boolean)
                .join(" · ")}
            </b>
          </header>
          <div className="sc-scroll">
            {training.details?.trim() && <p className="sc-details">{training.details}</p>}
            <div className="sc-section-head">
              <b>
                {t("Кто записался")}
                {showCapacity && ` · ${people?.length ?? count}`}
              </b>
              {current.admin && (
                <button
                  disabled={people === null || busy || (past && full)}
                  title={past && full ? t("Мест нет") : undefined}
                  onClick={() => setEditor(0)}
                >
                  + {t("Добавить")}
                </button>
              )}
            </div>
            {error && <ErrorBanner message={error} onRetry={refresh} />}
            {people === null && !error && <Spinner />}
            {people?.length === 0 && (
              <p className="sc-meta">
                {t(
                  past && !current.admin
                    ? "Состав завершённой тренировки недоступен"
                    : "Пока никто не записался",
                )}
              </p>
            )}
            <ul className="sc-people">
              {people?.map((p) => {
                const name = [p.firstName, p.lastName].filter(Boolean).join(" ") || t("Без имени");
                return (
                  <li key={"bookingId" in p ? p.bookingId : p.userId}>
                    <span className="sc-avatar" aria-hidden="true">
                      {[p.firstName, p.lastName].map((n) => n?.[0]).join("")}
                    </span>
                    <span className="sc-name">
                      {name}
                      {p.userId === user && " " + t("(вы)")}
                      <ContactLinks identities={p.identities ?? []} />
                    </span>
                    {"bookingId" in p && current.admin && (
                      <>
                        {past && (
                          <span className="sc-marks">
                            {([true, false] as const).map((attended) => {
                              const pressed = p.status === (attended ? "completed" : "not_completed");
                              return (
                                <button
                                  key={String(attended)}
                                  aria-pressed={pressed}
                                  disabled={busy}
                                  onClick={() =>
                                    void apply(() =>
                                      pressed
                                        ? api.manageBooking({
                                            action: "reset",
                                            trainingId: training.id,
                                            bookingId: p.bookingId,
                                          })
                                        : api.setAttendance(p.bookingId, attended),
                                    )
                                  }
                                >
                                  {attended ? t("Был") : t("Не был")}
                                </button>
                              );
                            })}
                          </span>
                        )}
                        <button
                          aria-label={t("Изменить запись") + " · " + name}
                          disabled={busy}
                          onClick={() => setEditor(p.bookingId)}
                        >
                          ⋯
                        </button>
                      </>
                    )}
                  </li>
                );
              })}
            </ul>
            {!past && current.admin && (
              <details>
                <summary>
                  {t("В ожидании")}
                  {queue && " · " + queue.length}
                </summary>
                {queueError && <ErrorBanner message={queueError} onRetry={refresh} />}
                {queue === null && !queueError && <Spinner />}
                <ul className="sc-people">
                  {queue?.map((p) => (
                    <li key={p.id}>
                      <span className="sc-name">
                        {p.firstName} {p.lastName}
                        {p.id === user && " " + t("(вы)")}
                      </span>
                      {current.admin && (
                        <button
                          disabled={busy}
                          onClick={() =>
                            void apply(() =>
                              api.manageBooking({ action: "unwait", trainingId: training.id, userId: p.id }),
                            )
                          }
                        >
                          {t("Убрать")}
                        </button>
                      )}
                    </li>
                  ))}
                </ul>
              </details>
            )}
          </div>
          <footer className="sc-footer">
            {actionError && <ErrorBanner message={actionError} />}
            {showCapacity && <OccupancyBar booked={count} max={training.maxQuantity} meterOnly />}
            <small>
              {past
                ? t("Занятие завершено")
                : !training.isActive
                  ? t("Неактивна")
                  : full
                    ? t("Мест нет")
                    : t("Свободно мест:") + " " + (training.maxQuantity - count)}
              {showCapacity && ` · ${count} ${t("из")} ${training.maxQuantity}`}
            </small>
            {!past && (
              <button
                className={"btn" + (!booked && (!full || !inQueue) ? " primary" : "")}
                disabled={busy || people === null || (!training.isActive && !booked && !inQueue)}
                onClick={act}
              >
                {button}
              </button>
            )}
            {hint && <small>{hint}</small>}
          </footer>
          {current.admin && editor !== null && (
            <Dialog.Root
              open
              onOpenChange={(open) => {
                if (!open) setEditor(null);
              }}
            >
              <Dialog.Portal>
                <Dialog.Overlay className="sc-overlay sc-nested" />
                <Dialog.Content className="sc-editor" aria-describedby={undefined}>
                  <Dialog.Title>
                    {editor === "training"
                      ? t("Изменить тренировку")
                      : selected
                        ? [selected.firstName, selected.lastName].join(" ")
                        : !past && full
                          ? t("Добавить в лист ожидания")
                          : t("Добавить участника")}
                  </Dialog.Title>
                  {editor === "training" ? (
                    <TrainingEditor
                      training={training}
                      onUpdated={(item) => {
                        onUpdated(item);
                        setEditor(null);
                      }}
                      onDeleted={onDeleted}
                    />
                  ) : (
                    <ManageRoster
                      training={training}
                      bookings={bookings}
                      selectedBooking={editor}
                      onDone={() => setEditor(null)}
                      refresh={async () => {
                        await refresh();
                        onUpdated(await api.getTraining(training.id));
                      }}
                    />
                  )}
                  <Dialog.Close className="btn">{t("Назад")}</Dialog.Close>
                </Dialog.Content>
              </Dialog.Portal>
            </Dialog.Root>
          )}
        </Dialog.Content>
      </Dialog.Portal>
    </Dialog.Root>
  );
}
