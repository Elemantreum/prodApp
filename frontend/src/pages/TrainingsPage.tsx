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
  const [cancellingId, setCancellingId] = useState<number | null>(null);

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
      setTrainings((prev) =>
        prev ? prev.map((t) => (t.id === trainingId ? { ...t, bookedCount: t.bookedCount + 1 } : t)) : prev,
      );
    } catch (e) {
      if (e instanceof ApiError && e.status === 403) {
        setError("Вы не одобрены администратором. Дождитесь подтверждения, чтобы записываться на тренировки.");
      } else {
        setError(e instanceof ApiError ? e.message : "Не удалось записаться");
      }
    } finally {
      setSigningUpId(null);
    }
  }

  async function handleCancel(bookingId: number) {
    setCancellingId(bookingId);
    setError(null);
    try {
      const cancelled = myBookings.find((b) => b.id === bookingId);
      await api.cancelBooking(bookingId);
      setMyBookings((prev) => prev.map((b) => (b.id === bookingId ? { ...b, status: "cancelled" } : b)));
      if (cancelled) {
        setTrainings((prev) =>
          prev
            ? prev.map((t) =>
                t.id === cancelled.trainingId ? { ...t, bookedCount: Math.max(0, t.bookedCount - 1) } : t,
              )
            : prev,
        );
      }
    } catch (e) {
      setError(e instanceof ApiError ? e.message : "Не удалось отписаться");
    } finally {
      setCancellingId(null);
    }
  }

  const activeBookingIdByTrainingId = new Map(
    myBookings.filter((b) => b.status === "active").map((b) => [b.trainingId, b.id]),
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
          {trainings.map((t) => {
            const myBookingId = activeBookingIdByTrainingId.get(t.id);
            return (
              <TrainingCard
                key={t.id}
                training={t}
                onSignUp={handleSignUp}
                signingUp={signingUpId === t.id}
                myBookingId={myBookingId}
                onCancel={handleCancel}
                cancelling={myBookingId != null && cancellingId === myBookingId}
                onTrainingUpdated={(updated) =>
                  setTrainings((prev) => prev?.map((x) => (x.id === updated.id ? updated : x)) ?? prev)
                }
                onTrainingDeleted={(id) => setTrainings((prev) => prev?.filter((x) => x.id !== id) ?? prev)}
              />
            );
          })}
        </div>
      )}
    </Layout>
  );
}
