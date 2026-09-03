import { useCallback, useEffect, useState } from "react";
import { api, ApiError } from "../api/client";
import type { Booking, Training } from "../api/types";
import { useAuth } from "../auth/AuthContext";
import { Layout } from "../components/Layout";
import { Spinner } from "../components/Spinner";
import { EmptyState } from "../components/EmptyState";
import { ErrorBanner } from "../components/ErrorBanner";
import { BookingCard } from "../components/BookingCard";

export function MyBookingsPage() {
  const { claims } = useAuth();
  const [bookings, setBookings] = useState<Booking[] | null>(null);
  const [trainingsById, setTrainingsById] = useState<Map<number, Training>>(new Map());
  const [error, setError] = useState<string | null>(null);
  const [cancellingId, setCancellingId] = useState<number | null>(null);

  const load = useCallback(async () => {
    if (!claims) return;
    setError(null);
    try {
      const [bookingsRes, trainingsRes] = await Promise.all([
        api.getBookingsByUserId(claims.user_id),
        api.getTrainings(),
      ]);
      setBookings(bookingsRes);
      setTrainingsById(new Map(trainingsRes.map((t) => [t.id, t])));
    } catch (e) {
      setError(e instanceof ApiError ? e.message : "Не удалось загрузить записи");
    }
  }, [claims]);

  useEffect(() => {
    load();
  }, [load]);

  async function handleCancel(bookingId: number) {
    setCancellingId(bookingId);
    setError(null);
    try {
      await api.cancelBooking(bookingId);
      setBookings((prev) =>
        prev?.map((b) => (b.id === bookingId ? { ...b, status: "cancelled", unbookingDate: new Date().toISOString() } : b)) ?? null,
      );
    } catch (e) {
      setError(e instanceof ApiError ? e.message : "Не удалось отменить запись");
    } finally {
      setCancellingId(null);
    }
  }

  const sorted = bookings
    ? [...bookings].sort((a, b) => new Date(b.bookingDate).getTime() - new Date(a.bookingDate).getTime())
    : null;

  return (
    <Layout title="Мои записи">
      {error && <ErrorBanner message={error} onRetry={load} />}

      {bookings === null && !error && (
        <div className="centered">
          <Spinner />
        </div>
      )}

      {sorted?.length === 0 && <EmptyState title="Пока нет записей" hint="Запишись на тренировку на первой вкладке" />}

      {sorted && sorted.length > 0 && (
        <div className="list">
          {sorted.map((b) => (
            <BookingCard
              key={b.id}
              booking={b}
              training={trainingsById.get(b.trainingId)}
              onCancel={handleCancel}
              cancelling={cancellingId === b.id}
            />
          ))}
        </div>
      )}
    </Layout>
  );
}
