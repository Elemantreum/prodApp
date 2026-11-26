import { t } from "../i18n";
import { formatDateTime } from "../utils/date";
import { useOpenSession, type PersonalSession } from "../pages/personalSessions";
import "../styles/personal.css";

const labels = {
  active: t("Вы записаны"),
  waiting: t("В листе ожидания"),
  completed: t("Был"),
  not_completed: t("Не был"),
  cancelled: t("Запись отменена"),
};

export function BookingCard({
  session,
  onCancel,
  busy,
}: {
  session: PersonalSession;
  onCancel?: () => void;
  busy?: boolean;
}) {
  const open = useOpenSession();
  return (
    <div className="row personal-session">
      <button type="button" className="personal-session-open" onClick={() => open(session)}>
        <span className="personal-detail">
          <strong>{session.training.description}</strong>
          <small>
            {session.section.name} ·{" "}
            <time dateTime={session.training.date}>{formatDateTime(session.training.date)}</time>
          </small>
        </span>
        <span className="personal-mark">{labels[session.status]}</span>
      </button>
      {onCancel && (
        <button className="btn" type="button" disabled={busy} onClick={onCancel}>
          {session.status === "waiting" ? t("Выйти") : t("Отменить")}
        </button>
      )}
    </div>
  );
}
