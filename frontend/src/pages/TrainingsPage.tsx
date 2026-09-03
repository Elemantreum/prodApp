import { useCallback, useEffect, useState } from "react";
import { api, ApiError } from "../api/client";
import type { Booking, Training } from "../api/types";
import { useAuth } from "../auth/AuthContext";
import { Layout } from "../components/Layout";
import { Spinner } from "../components/Spinner";
import { EmptyState } from "../components/EmptyState";
import { ErrorBanner } from "../components/ErrorBanner";
import { TrainingCard } from "../components/TrainingCard";

export function TrainingsPage() {
  const { claims } = useAuth();
  const [trainings, setTrainings] = useState<Training[] | null>(null);
  const [myBookings, setMyBookings] = useState<Booking[]>([]);
  const [error, setError] = useState<string | null>(null);
  const [signingUpId, setSigningUpId] = useState<number | null>(null);

  const load = useCallback(async () => {
    if (!claims) return;
    setError(null);
    try {
      const [trainingsRes, bookingsRes] = await Promise.all([
        api.getTrainings(),
        api.getBookingsByUserId(claims.user_id),
      ]);
      setTrainings(trainingsRes);
      setMyBookings(bookingsRes);
    } catch (e) {
      setError(e instanceof ApiError ? e.message : "Не удалось загрузить тренировки");
    }
  }, [claims]);

  useEffect(() => {
    load();
  }, [load]);

  async function handleSignUp(trainingId: number) {
    setSigningUpId(trainingId);
    setError(null);
    try {
      const booking = await api.signUp(trainingId);
      setMyBookings((prev) => [...prev, booking]);
      // Оптимистично подвинуть счётчик на карточке, не дожидаясь рефетча.
      setTrainings((prev) =>
        prev ? prev.map((t) => (t.id === trainingId ? { ...t, bookedCount: t.bookedCount + 1 } : t)) : prev,
      );
    } catch (e) {
      // 403 здесь может означать только одно: CreateBooking на бэке отказывает
      // только по причине "не одобрен админом" — так что подменяем сырой
      // "forbidden: ..." с сервера на понятный текст, а не показываем как есть.
      if (e instanceof ApiError && e.status === 403) {
        setError("Вы не одобрены администратором. Дождитесь подтверждения, чтобы записываться на тренировки.");
      } else {
        setError(e instanceof ApiError ? e.message : "Не удалось записаться");
      }
    } finally {
      setSigningUpId(null);
    }
  }

  const activeBookedTrainingIds = new Set(
    myBookings.filter((b) => b.status === "active").map((b) => b.trainingId),
  );

  return (
    <Layout title="Тренировки">
      {error && <ErrorBanner message={error} onRetry={load} />}

      {trainings === null && !error && (
        <div className="centered">
          <Spinner />
        </div>
      )}

      {trainings?.length === 0 && <EmptyState title="Пока нет тренировок" hint="Загляни попозже" />}

      {trainings && trainings.length > 0 && (
        <div className="list">
          {trainings
            .filter((t) => t.isActive)
            .map((t) => (
              <TrainingCard
                key={t.id}
                training={t}
                onSignUp={handleSignUp}
                signingUp={signingUpId === t.id}
                alreadyBooked={activeBookedTrainingIds.has(t.id)}
                onTrainingUpdated={(updated) =>
                  setTrainings((prev) => prev?.map((x) => (x.id === updated.id ? updated : x)) ?? prev)
                }
              />
            ))}
        </div>
      )}
    </Layout>
  );
}
