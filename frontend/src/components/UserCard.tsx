import type { User } from "../api/types";
import { Button } from "./Button";

interface Props {
  user: User;
  isSelf: boolean;
  busyAction: string | null;
  onToggleApproved: (user: User) => void;
  onToggleAdmin: (user: User) => void;
  onDelete: (user: User) => void;
}

export function UserCard({ user, isSelf, busyAction, onToggleApproved, onToggleAdmin, onDelete }: Props) {
  const displayName = [user.firstName, user.lastName].filter(Boolean).join(" ") || user.username;

  return (
    <div className="user-card">
      <div className="user-card__info">
        <div className="card__title">
          {displayName} {isSelf && <span className="user-card__you">(вы)</span>}
        </div>
        <div className="card__meta">{user.username}</div>
        <div className="user-card__badges">
          <span className={`badge ${user.isApproved ? "badge--active" : "badge--not_completed"}`}>
            {user.isApproved ? "Одобрен" : "Не одобрен"}
          </span>
          {user.isAdmin && <span className="badge badge--admin">Админ</span>}
        </div>
      </div>

      <div className="user-card__actions">
        <Button
          variant={user.isApproved ? "secondary" : "primary"}
          loading={busyAction === "approve"}
          disabled={busyAction !== null}
          onClick={() => onToggleApproved(user)}
        >
          {user.isApproved ? "Снять одобрение" : "Одобрить"}
        </Button>
        <Button
          variant="secondary"
          loading={busyAction === "admin"}
          disabled={busyAction !== null || (isSelf && user.isAdmin)}
          onClick={() => onToggleAdmin(user)}
        >
          {user.isAdmin ? "Забрать админку" : "Сделать админом"}
        </Button>
        <Button
          variant="danger"
          loading={busyAction === "delete"}
          disabled={busyAction !== null || isSelf}
          onClick={() => onDelete(user)}
        >
          Удалить
        </Button>
      </div>
    </div>
  );
}
