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
      setError(e instanceof ApiError ? e.message : "Не удалось загрузить список");
    }
  }, []);

  useEffect(() => {
    load();
  }, [load]);

  return (
    <Layout title="Список грешников">
      {error && <ErrorBanner message={error} onRetry={load} />}

      {sinners === null && !error && (
        <div className="centered">
          <Spinner />
        </div>
      )}

      {sinners?.length === 0 && <EmptyState title="Грешников пока нет" hint="Все ведут себя примерно" />}

      {sinners && sinners.length > 0 && (
        <div className="list">
          {sinners.map((s, i) => (
            <div className="card" key={s.userId}>
              <div className="card__body">
                <div className="card__title">
                  {i + 1}. {[s.firstName, s.lastName].filter(Boolean).join(" ") || "Без имени"}
                </div>
                <div className="card__meta">
                  Поздних отмен: {s.lateCancellations} · Неявок: {s.noShows}
                </div>
              </div>
              <span className="badge badge--cancelled">{s.penaltyPoints} балл(а)</span>
            </div>
          ))}
        </div>
      )}
    </Layout>
  );
}
