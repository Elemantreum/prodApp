import { useState } from "react";
import { api } from "../api/client";
import { useSection } from "../auth/Sections";
import { BookingCard } from "../components/BookingCard";
import { ErrorBanner } from "../components/ErrorBanner";
import { Layout } from "../components/Layout";
import { Spinner } from "../components/Spinner";
import { t } from "../i18n";
import { type PersonalSession, usePersonalSessions } from "./personalSessions";

export function MyBookingsPage() {
  const { current } = useSection();
  const [all, setAll] = useState(true);
  const { sessions, error, reload } = usePersonalSessions(all);
  const [saving, setSaving] = useState(false);
  const [actionError, setActionError] = useState("");

  async function cancel(session: PersonalSession) {
    setSaving(true);
    setActionError("");
    try {
      if (session.status === "waiting") await api.setWaiting(session.training.id, false, session.section.id);
      else if (session.bookingId !== undefined) await api.cancelBooking(session.bookingId, session.section.id);
      reload();
    } catch (reason) {
      setActionError(reason instanceof Error ? reason.message : t("Не удалось отменить запись"));
    } finally {
      setSaving(false);
    }
  }
  return (
    <Layout title={t("Мои записи")}>
      <section className="card personal-panel">
        <select
          className="btn personal-scope"
          aria-label={t("Секция")}
          value={all ? "all" : "current"}
          onChange={(e) => {
            setActionError("");
            setAll(e.target.value === "all");
          }}
        >
          <option value="all">{t("Все мои секции")}</option>
          <option value="current">{current.name}</option>
        </select>
        {error && <ErrorBanner message={error} onRetry={reload} />}
        {actionError && <ErrorBanner message={actionError} />}
        {!error && !sessions && <Spinner />}
        {sessions?.length === 0 && <p className="muted">{t("У вас пока нет записей.")}</p>}
        {sessions?.map((session) => (
          <BookingCard
            key={session.bookingId ?? `waiting-${session.training.id}`}
            session={session}
            busy={saving}
            onCancel={() => cancel(session)}
          />
        ))}
      </section>
    </Layout>
  );
}
