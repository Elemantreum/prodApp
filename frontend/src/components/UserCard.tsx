import { useState } from "react";
import type { FormEvent } from "react";
import type { User } from "../api/types";
import { Button } from "./Button";

const PLATFORM_LABEL: Record<string, string> = {
  telegram: "Telegram",
  vk: "VK",
};

interface Props {
  user: User;
  isSelf: boolean;
  busyAction: string | null;
  onToggleApproved: (user: User) => void;
  onToggleAdmin: (user: User) => void;
  onDelete: (user: User) => void;
  onRename: (user: User, firstName: string, lastName: string) => Promise<void>;
}

export function UserCard({ user, isSelf, busyAction, onToggleApproved, onToggleAdmin, onDelete, onRename }: Props) {
  const displayName = [user.firstName, user.lastName].filter(Boolean).join(" ") || user.username;

  // window.confirm() is unreliable inside Telegram/VK webviews (silently
  // no-ops or throws in a lot of client versions) — a plain click-twice
  // confirm built into the UI works everywhere instead.
  const [confirmingDelete, setConfirmingDelete] = useState(false);

  const [editing, setEditing] = useState(false);
  const [firstName, setFirstName] = useState(user.firstName);
  const [lastName, setLastName] = useState(user.lastName);
  const [renaming, setRenaming] = useState(false);
  const [renameError, setRenameError] = useState<string | null>(null);

  function handleDeleteClick() {
    if (confirmingDelete) {
      setConfirmingDelete(false);
      onDelete(user);
    } else {
      setConfirmingDelete(true);
    }
  }

  function startEditing() {
    setFirstName(user.firstName);
    setLastName(user.lastName);
    setRenameError(null);
    setEditing(true);
  }

  async function handleRenameSubmit(e: FormEvent) {
    e.preventDefault();
    setRenaming(true);
    setRenameError(null);
    try {
      await onRename(user, firstName.trim(), lastName.trim());
      setEditing(false);
    } catch (e) {
      setRenameError(e instanceof Error ? e.message : "Не удалось сохранить");
    } finally {
      setRenaming(false);
    }
  }

  return (
    <div className="user-card">
      <div className="user-card__info">
        <div className="card__title">
          {displayName} {isSelf && <span className="user-card__you">(вы)</span>}
        </div>
        <div className="card__meta">{user.username}</div>
        <div className="card__meta">
          Платформы:{" "}
          {user.identities.length === 0
            ? "нет данных"
            : user.identities.map((i) => PLATFORM_LABEL[i.provider] ?? i.provider).join(", ")}
        </div>
        <div className="user-card__badges">
          <span className={`badge ${user.isApproved ? "badge--active" : "badge--not_completed"}`}>
            {user.isApproved ? "Одобрен" : "Не одобрен"}
          </span>
          {user.isAdmin && <span className="badge badge--admin">Админ</span>}
        </div>
      </div>

      {editing ? (
        <form className="form" onSubmit={handleRenameSubmit}>
          <label className="form__field">
            <span>Имя</span>
            <input type="text" value={firstName} onChange={(e) => setFirstName(e.target.value)} />
          </label>
          <label className="form__field">
            <span>Фамилия</span>
            <input type="text" value={lastName} onChange={(e) => setLastName(e.target.value)} />
          </label>
          {renameError && <div className="participants__error">{renameError}</div>}
          <div className="user-card__actions-row">
            <Button variant="secondary" type="button" disabled={renaming} onClick={() => setEditing(false)}>
              Отмена
            </Button>
            <Button type="submit" loading={renaming}>
              Сохранить
            </Button>
          </div>
        </form>
      ) : (
        <div className="user-card__actions">
          <Button variant="secondary" disabled={busyAction !== null} onClick={startEditing}>
            Изменить имя
          </Button>
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
          {confirmingDelete ? (
            <div className="user-card__actions-row">
              <Button variant="secondary" disabled={busyAction !== null} onClick={() => setConfirmingDelete(false)}>
                Отмена
              </Button>
              <Button
                variant="danger"
                loading={busyAction === "delete"}
                disabled={busyAction !== null || isSelf}
                onClick={handleDeleteClick}
              >
                Точно удалить?
              </Button>
            </div>
          ) : (
            <Button
              variant="danger"
              loading={busyAction === "delete"}
              disabled={busyAction !== null || isSelf}
              onClick={handleDeleteClick}
            >
              Удалить
            </Button>
          )}
        </div>
      )}
    </div>
  );
}
