import { useCallback, useEffect, useState } from "react";
import { api, ApiError } from "../api/client";
import type { User } from "../api/types";
import { useAuth } from "../auth/AuthContext";
import { Spinner } from "../components/Spinner";
import { EmptyState } from "../components/EmptyState";
import { ErrorBanner } from "../components/ErrorBanner";
import { UserCard } from "../components/UserCard";

type Action = "approve" | "admin" | "delete";

export function AdminUsersPage() {
  const { claims } = useAuth();
  const [users, setUsers] = useState<User[] | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [busy, setBusy] = useState<{ userId: number; action: Action } | null>(null);

  const load = useCallback(async () => {
    setError(null);
    try {
      setUsers(await api.getUsers());
    } catch (e) {
      setError(e instanceof ApiError ? e.message : "Не удалось загрузить пользователей");
    }
  }, []);

  useEffect(() => {
    load();
  }, [load]);

  async function runAction(user: User, action: Action, fn: () => Promise<User | true>) {
    setBusy({ userId: user.id, action });
    setError(null);
    try {
      const result = await fn();
      if (result !== true) {
        setUsers((prev) => prev?.map((u) => (u.id === result.id ? result : u)) ?? null);
      } else {
        setUsers((prev) => prev?.filter((u) => u.id !== user.id) ?? null);
      }
    } catch (e) {
      setError(e instanceof ApiError ? e.message : "Действие не удалось");
    } finally {
      setBusy(null);
    }
  }

  function handleToggleApproved(user: User) {
    runAction(user, "approve", () => (user.isApproved ? api.unapproveUser(user.id) : api.approveUser(user.id)));
  }

  function handleToggleAdmin(user: User) {
    runAction(user, "admin", () => (user.isAdmin ? api.removeAdmin(user.id) : api.makeAdmin(user.id)));
  }

  function handleDelete(user: User) {
    // UserCard already asks for a second click before calling this — удаление
    // отменит все брони юзера, но повторный native window.confirm() здесь не
    // нужен (и не работает надёжно в Telegram/VK webview).
    runAction(user, "delete", () => api.deleteUser(user.id));
  }

  async function handleRename(user: User, firstName: string, lastName: string) {
    setError(null);
    try {
      const updated = await api.updateProfile(firstName, lastName, user.id);
      setUsers((prev) => prev?.map((u) => (u.id === updated.id ? updated : u)) ?? null);
    } catch (e) {
      throw new Error(e instanceof ApiError ? e.message : "Не удалось сохранить имя");
    }
  }

  return (
    <>
      {error && <ErrorBanner message={error} onRetry={load} />}

      {users === null && !error && (
        <div className="centered">
          <Spinner />
        </div>
      )}

      {users?.length === 0 && <EmptyState title="Пользователей пока нет" />}

      {users && users.length > 0 && (
        <div className="list">
          {users.map((u) => (
            <UserCard
              key={u.id}
              user={u}
              isSelf={u.id === claims?.user_id}
              busyAction={busy?.userId === u.id ? busy.action : null}
              onToggleApproved={handleToggleApproved}
              onToggleAdmin={handleToggleAdmin}
              onDelete={handleDelete}
              onRename={handleRename}
            />
          ))}
        </div>
      )}
    </>
  );
}
