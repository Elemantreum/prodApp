import "../styles/personal.css";
import { t } from "../i18n";
import { useCallback, useEffect, useState } from "react";
import { api, ApiError } from "../api/client";
import type { SinnerStats } from "../api/types";
import { Layout } from "../components/Layout";
import { Spinner } from "../components/Spinner";
import { EmptyState } from "../components/EmptyState";
import { ErrorBanner } from "../components/ErrorBanner";

export function SinnersPage() {
  const [sinners, setSinners] = useState<SinnerStats[] | null>(null);
  const [error, setError] = useState<string | null>(null);

  const load = useCallback(async () => {
    setError(null);
    try {
      setSinners(await api.getAntiLeaderboard());
    } catch (e) {
      setError(e instanceof ApiError ? e.message : t("Не удалось загрузить список"));
    }
  }, []);

  useEffect(() => {
    load();
  }, [load]);

  return (
    <Layout title={t("Список грешников")}>
      <section className="card personal-panel">
        {error && <ErrorBanner message={error} onRetry={load} />}

        {sinners === null && !error && (
          <div className="centered">
            <Spinner />
          </div>
        )}

        {sinners?.length === 0 && (
          <EmptyState title={t("Грешников пока нет")} hint={t("Все ведут себя примерно")} />
        )}

        {sinners && sinners.length > 0 && (
          <ol className="personal-ranking">
            {sinners.map((s, i) => (
              <li className="row personal-section" key={s.userId}>
                <span className="personal-avatar">{i + 1}</span>
                <div className="personal-detail">
                  <strong>{[s.firstName, s.lastName].filter(Boolean).join(" ") || t("Без имени")}</strong>
                  <small>
                    {" "}
                    {t("Поздних отмен:")} {s.lateCancellations} {t("· Неявок:")} {s.noShows}
                  </small>
                </div>
                <span className="personal-mark">
                  {t("Штрафные баллы")}: {s.penaltyPoints}
                </span>
              </li>
            ))}
          </ol>
        )}
      </section>
    </Layout>
  );
}
