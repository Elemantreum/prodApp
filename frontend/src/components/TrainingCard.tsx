import { useState } from "react";
import { api, ApiError } from "../api/client";
import type { Participant, Training } from "../api/types";
import { useAuth } from "../auth/AuthContext";
import { formatDateTime } from "../utils/date";
import { formatDurationNs } from "../utils/duration";
import { Button } from "./Button";
import { ContactLinks } from "./ContactLinks";
import { OccupancyBar } from "./OccupancyBar";
import { Spinner } from "./Spinner";
import { TrainingAdminModal } from "./TrainingAdminModal";

interface Props {
  training: Training;
  onSignUp: (id: number) => void;
  signingUp: boolean;
  /** Id активной брони текущего юзера на эту тренировку, если есть —
   *  определяет режим кнопки: "Записаться" vs "Отписаться". */
  myBookingId?: number;
  onCancel: (bookingId: number) => void;
  cancelling: boolean;
  /** Опционально — только если родитель хочет получать актуальную
   *  тренировку после того, как админ отредактирует её в модалке. */
  onTrainingUpdated?: (training: Training) => void;
  /** Опционально — только если родитель хочет убрать карточку из списка
   *  после того, как админ удалит тренировку в модалке. */
  onTrainingDeleted?: (id: number) => void;
}

export function TrainingCard({
  training,
  onSignUp,
  signingUp,
  myBookingId,
  onCancel,
  cancelling,
  onTrainingUpdated,
  onTrainingDeleted,
}: Props) {
  const { claims } = useAuth();
  const isPast = new Date(training.date).getTime() < Date.now();
  const isFull = training.bookedCount >= training.maxQuantity;
  const [expanded, setExpanded] = useState(false);
  const [participants, setParticipants] = useState<Participant[] | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [adminOpen, setAdminOpen] = useState(false);

  async function toggleParticipants() {
    if (expanded) {
      setExpanded(false);
      return;
    }

    setExpanded(true);

    if (participants !== null) return; // уже загружали — не дёргаем API повторно

    setLoading(true);
    setError(null);
    try {
      setParticipants(await api.getTrainingParticipants(training.id));
    } catch (e) {
      setError(e instanceof ApiError ? e.message : "Не удалось загрузить список");
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="card training-card">
      <div className="training-card__row">
        <div className="card__body">
          <div className="card__title">
            {training.description}
            {training.recurringTrainingId != null && (
              <span className="training-card__recurring" title="Повторяется каждую неделю">
                🔁
              </span>
            )}
          </div>
          {/* Бэк отдаёт неактивные тренировки только админу (см. TrainingsPage) —
              значит если она тут и isActive=false, показывать её видит только он. */}
          {!training.isActive && <span className="badge badge--cancelled">Неактивна, скрыта от пользователей</span>}
          <div className="card__meta">{formatDateTime(training.date)}</div>
          <div className="card__meta">{formatDurationNs(training.duration)}</div>
          <OccupancyBar booked={training.bookedCount} max={training.maxQuantity} />
        </div>
        {myBookingId != null ? (
          <Button variant="danger" loading={cancelling} onClick={() => onCancel(myBookingId)}>
            Отписаться
          </Button>
        ) : (
          <Button
            variant="primary"
            disabled={isPast || isFull}
            loading={signingUp}
            onClick={() => onSignUp(training.id)}
          >
            {isPast ? "Тренировка прошла" : isFull ? "Нет мест" : "Записаться"}
          </Button>
        )}
      </div>

      <div className="training-card__toolbar">
        <button className="training-card__toggle" onClick={toggleParticipants}>
          {expanded ? "Скрыть участников" : "Кто записан"}
        </button>
        {claims?.is_admin && (
          <button className="training-card__toggle" onClick={() => setAdminOpen(true)}>
            Управление
          </button>
        )}
      </div>

      {adminOpen && (
        <TrainingAdminModal
          training={training}
          onClose={() => setAdminOpen(false)}
          onUpdated={(updated) => onTrainingUpdated?.(updated)}
          onDeleted={(id) => {
            setAdminOpen(false);
            onTrainingDeleted?.(id);
          }}
        />
      )}

      {expanded && (
        <div className="participants">
          {loading && (
            <div className="participants__loading">
              <Spinner size={16} />
            </div>
          )}
          {error && <div className="participants__error">{error}</div>}
          {participants?.length === 0 && <div className="participants__empty">Пока никто не записался</div>}
          {participants && participants.length > 0 && (
            <ul className="participants__list">
              {participants.map((p) => {
                const fullName = [p.firstName, p.lastName].filter(Boolean).join(" ");
                return (
                  <li key={p.userId}>
                    <div className="participants__name">
                      {fullName || "Без имени"}
                      <ContactLinks identities={p.identities} />
                    </div>
                  </li>
                );
              })}
            </ul>
          )}
        </div>
      )}
    </div>
  );
}
