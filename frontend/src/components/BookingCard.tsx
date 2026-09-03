import type { Booking, BookingStatus, Training } from "../api/types";
import { formatDateTime } from "../utils/date";
import { Button } from "./Button";

const STATUS_LABEL: Record<BookingStatus, string> = {
  active: "Активна",
  cancelled: "Отменена",
  completed: "Завершена",
  not_completed: "Не пройдена",
};

interface Props {
  booking: Booking;
  training: Training | undefined;
  onCancel: (id: number) => void;
  cancelling: boolean;
}

export function BookingCard({ booking, training, onCancel, cancelling }: Props) {
  return (
    <div className="card">
      <div className="card__body">
        <div className="card__title">{training?.description ?? `Тренировка #${booking.trainingId}`}</div>
        {training && <div className="card__meta">{formatDateTime(training.date)}</div>}
        <span className={`badge badge--${booking.status}`}>{STATUS_LABEL[booking.status]}</span>
      </div>
      {booking.status === "active" && (
        <Button variant="danger" loading={cancelling} onClick={() => onCancel(booking.id)}>
          Отменить
        </Button>
      )}
    </div>
  );
}
