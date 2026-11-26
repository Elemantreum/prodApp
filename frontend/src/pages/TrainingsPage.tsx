import { useCallback, useEffect, useRef, useState } from "react";
import { useSearchParams } from "react-router-dom";
import { api } from "../api/client";
import type { Training } from "../api/types";
import { useAuth } from "../auth/AuthContext";
import { useSection } from "../auth/Sections";
import { AdminDialog, useAdminCreate } from "../components/AdminUI";
import { AdminCreateTrainingPage } from "./AdminCreateTrainingPage";
import { Layout } from "../components/Layout";
import { ErrorBanner } from "../components/ErrorBanner";
import { Spinner } from "../components/Spinner";
import { TrainingCard } from "../components/TrainingCard";
import { TrainingAdminModal } from "../components/TrainingAdminModal";
import { ScheduleCalendar } from "../components/ScheduleCalendar";
import { dayKey, fullDay } from "../components/scheduleDates";
import { useScheduleUpdates } from "../hooks/useScheduleUpdates";
import { t } from "../i18n";

export function TrainingsPage() {
  const [params, setParams] = useSearchParams();
  const [creating, setCreating] = useAdminCreate();
  const { current } = useSection();
  const { token, claims } = useAuth();
  const [trainings, setTrainings] = useState<Training[] | null>(null);
  const [linked, setLinked] = useState<Training | null>(null);
  const [error, setError] = useState("");
  const [linkError, setLinkError] = useState("");
  const latest = useRef(0);
  const session = Number(params.get("session"));
  const load = useCallback(async () => {
    const request = ++latest.current;
    try {
      const items = await api.getTrainings();
      if (request === latest.current) {
        setTrainings(items);
        setError("");
      }
    } catch (e) {
      if (request === latest.current) setError(String(e));
    }
  }, []);
  useEffect(() => {
    void load();
    return () => {
      latest.current++;
    };
  }, [load]);
  useScheduleUpdates(token, load);
  const listed = trainings?.find((item) => item.id === session);
  const needsLookup = session > 0 && trainings !== null && !listed;
  const userId = claims?.user_id;
  useEffect(() => {
    if (!needsLookup) return;
    let live = true;
    setLinkError("");
    async function resolve() {
      try {
        const item = await api.getTraining(session, current.id);
        if (!current.admin && !item.isActive) {
          const own = userId ? await api.getBookingsByUserId(userId, current.id) : [];
          if (!own.some((booking) => booking.trainingId === session && booking.status !== "cancelled")) {
            throw new Error(t("Тренировка недоступна"));
          }
        }
        if (live) setLinked(item);
      } catch (error) {
        if (live) setLinkError(error instanceof Error ? t(error.message) : t("Тренировка недоступна"));
      }
    }
    void resolve();
    return () => {
      live = false;
    };
  }, [session, needsLookup, current.id, current.admin, userId]);
  const opened = listed ?? (linked?.id === session ? linked : null);
  const rawDay = params.get("day");
  const parsed = new Date(rawDay + "T12:00:00");
  const selected = rawDay && dayKey(parsed) === rawDay ? parsed : opened ? new Date(opened.date) : new Date();
  function navigate(date: Date, id?: number, replace = false) {
    const next = new URLSearchParams(params);
    next.set("day", dayKey(date));
    next.delete("new");
    if (id) next.set("session", String(id));
    else next.delete("session");
    setParams(next, { replace });
  }
  function open(item: Training) {
    navigate(new Date(item.date), item.id);
  }
  const visible = (trainings ?? []).filter((item) => current.admin || item.isActive);
  const daily = visible
    .filter((item) => dayKey(new Date(item.date)) === dayKey(selected))
    .sort((a, b) => Date.parse(a.date) - Date.parse(b.date));
  return (
    <Layout
      title={t("Расписание")}
      action={
        current.admin && (
          <button className="btn primary" onClick={() => setCreating(true)}>
            + {t("Создать занятие")}
          </button>
        )
      }
    >
      {error && <ErrorBanner message={error} onRetry={load} />}
      {needsLookup && linkError && <ErrorBanner message={linkError} />}
      {!trainings && !error && <Spinner />}
      {trainings && (
        <>
          <ScheduleCalendar trainings={visible} selected={selected} onSelect={navigate} onOpen={open} />
          <section className="card sc-agenda">
            <h2>{fullDay.format(selected)}</h2>
            {daily.length ? (
              daily.map((item) => <TrainingCard key={item.id} training={item} onOpen={open} />)
            ) : (
              <p className="muted">{t("В этот день тренировок нет")}</p>
            )}
          </section>
        </>
      )}
      {creating && current.admin && (
        <AdminDialog title={t("Создать занятие")} onClose={() => setCreating(false)}>
          <AdminCreateTrainingPage
            onCreated={(item) => {
              setParams((previous) => {
                const next = new URLSearchParams(previous);
                next.delete("new");
                next.set("day", dayKey(new Date(item.date)));
                return next;
              });
              void load();
            }}
          />
        </AdminDialog>
      )}
      {opened && (
        <TrainingAdminModal
          key={opened.id}
          training={opened}
          onClose={() => navigate(selected, undefined, true)}
          onUpdated={(item) => {
            setLinked(item);
            setTrainings((items) => items?.map((s) => (s.id === item.id ? item : s)) ?? null);
          }}
          onDeleted={(id) => {
            setTrainings((items) => items?.filter((item) => item.id !== id) ?? null);
            navigate(selected, undefined, true);
            void load();
          }}
        />
      )}
    </Layout>
  );
}
