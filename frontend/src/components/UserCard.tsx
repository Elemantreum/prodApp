import { t as translate } from "../i18n";
import { AdminDialog } from "./AdminUI";
import { BookingHistory } from "./BookingHistory";
import { useState } from "react";
import type { FormEvent } from "react";
import type { User } from "../api/types";
import { Button } from "./Button";
import { ContactLinks } from "./ContactLinks";

const PLATFORM_LABEL: Record<string, string> = {
  telegram: "Telegram",
  vk: "VK",
};

interface Props {
  user: User;
  isSelf: boolean;
  deleting: boolean;
  onDelete: (user: User) => void;
  onRename: (user: User, firstName: string, lastName: string) => Promise<void>;
}

export function UserCard({ user, isSelf, deleting, onDelete, onRename }: Props) {
  const [open, setOpen] = useState(false);
  const [history, setHistory] = useState(false);
  const displayName = [user.firstName, user.lastName].filter(Boolean).join(" ") || translate("Без имени");

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
      setRenameError(e instanceof Error ? e.message : translate("Не удалось сохранить"));
    } finally {
      setRenaming(false);
    }
  }

  return (
    <div className="row">
      <span className="avatar" aria-hidden="true">
        {user.firstName[0]}
        {user.lastName[0]}
      </span>
      <div className="grow">
        <b>{displayName}</b>
        <small>{user.isAdmin ? translate("Администратор системы") : translate("Профиль системы")}</small>
      </div>
      <button className="btn" onClick={() => setOpen(true)}>
        {translate("Профиль")}
      </button>
      {open && (
        <AdminDialog
          title={displayName}
          busy={renaming || deleting}
          onClose={() => {
            setOpen(false);
            setEditing(false);
            setHistory(false);
            setConfirmingDelete(false);
          }}
        >
          <div className="user-card__info">
            <div className="card__title">
              {displayName} {isSelf && <span className="user-card__you">{translate("(вы)")}</span>}
              <ContactLinks identities={user.identities} />
            </div>
            <div className="card__meta">
              {" "}
              {translate("Платформы:")}{" "}
              {user.identities.length === 0
                ? translate("нет данных")
                : user.identities.map((i) => PLATFORM_LABEL[i.provider] ?? i.provider).join(", ")}
            </div>
            {user.isAdmin && <span className="badge badge--admin">{translate("Админ")}</span>}
          </div>

          {editing ? (
            <form className="form" onSubmit={handleRenameSubmit}>
              <label className="form__field">
                <span>{translate("Имя")}</span>
                <input type="text" value={firstName} onChange={(e) => setFirstName(e.target.value)} />
              </label>
              <label className="form__field">
                <span>{translate("Фамилия")}</span>
                <input type="text" value={lastName} onChange={(e) => setLastName(e.target.value)} />
              </label>
              {renameError && <div className="participants__error">{renameError}</div>}
              <div className="user-card__actions-row">
                <Button variant="secondary" type="button" disabled={renaming} onClick={() => setEditing(false)}>
                  {" "}
                  {translate("Отмена")}{" "}
                </Button>
                <Button type="submit" loading={renaming}>
                  {" "}
                  {translate("Сохранить")}{" "}
                </Button>
              </div>
            </form>
          ) : (
            <div className="user-card__actions">
              <Button variant="secondary" disabled={deleting} onClick={startEditing}>
                {" "}
                {translate("Изменить имя")}{" "}
              </Button>
              <Button variant="secondary" onClick={() => setHistory((prev) => !prev)}>
                {history ? translate("Скрыть историю") : translate("История записей")}
              </Button>
              {history && <BookingHistory userId={user.id} />}
              {confirmingDelete ? (
                <div className="user-card__actions-row">
                  <Button
                    variant="secondary"
                    disabled={deleting}
                    onClick={() => setConfirmingDelete(false)}
                  >
                    {" "}
                    {translate("Отмена")}{" "}
                  </Button>
                  <Button
                    variant="danger"
                    loading={deleting}
                    disabled={isSelf}
                    onClick={handleDeleteClick}
                  >
                    {" "}
                    {translate("Точно удалить?")}{" "}
                  </Button>
                </div>
              ) : (
                <Button
                  variant="danger"
                  loading={deleting}
                  disabled={isSelf}
                  onClick={handleDeleteClick}
                >
                  {" "}
                  {translate("Удалить")}{" "}
                </Button>
              )}
            </div>
          )}
        </AdminDialog>
      )}
    </div>
  );
}
