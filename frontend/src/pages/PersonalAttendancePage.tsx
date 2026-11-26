import { BookingCard } from "../components/BookingCard";
import { ErrorBanner } from "../components/ErrorBanner";
import { Layout } from "../components/Layout";
import { Spinner } from "../components/Spinner";
import { t } from "../i18n";
import { usePersonalSessions } from "./personalSessions";

export function PersonalAttendancePage() {
  const { sessions, error, reload } = usePersonalSessions(false, true);
  return (
    <Layout title={t("Посещаемость")}>
      <section className="card personal-panel">
        <h2>{t("Отметки посещаемости")}</h2>
        {error && <ErrorBanner message={error} onRetry={reload} />}
        {!error && !sessions && <Spinner />}
        {sessions?.length === 0 && <p className="muted">{t("Пока нет отметок посещаемости.")}</p>}
        {sessions?.map((session) => (
          <BookingCard key={session.bookingId} session={session} />
        ))}
      </section>
    </Layout>
  );
}
