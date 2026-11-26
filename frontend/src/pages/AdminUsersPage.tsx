import { t as translate } from "../i18n";
import { useSearchParams } from "react-router-dom";
import { useCallback, useEffect, useState } from "react";
import { api, ApiError } from "../api/client";
import type { User } from "../api/types";
import { useAuth } from "../auth/AuthContext";
import { Spinner } from "../components/Spinner";
import { EmptyState } from "../components/EmptyState";
import { ErrorBanner } from "../components/ErrorBanner";
import { UserCard } from "../components/UserCard";

export function AdminUsersPage() {
  const { claims } = useAuth();
  const [users, setUsers] = useState<User[] | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [deletingId, setDeletingId] = useState<number | null>(null);
  const [params, setParams] = useSearchParams();
  const search = params.get("q") ?? "";
  const [query, setQuery] = useState(search);
  useEffect(() => setQuery(search), [search]);
  useEffect(() => {
    if (query === search) return;
    const timer = setTimeout(
      () =>
        setParams(
          (prev) => {
            const next = new URLSearchParams(prev);
            if (query.trim()) next.set("q", query.trim());
            else next.delete("q");
            next.delete("page");
            return next;
          },
          { replace: true },
        ),
      250,
    );
    return () => clearTimeout(timer);
  }, [query, search, setParams]);

  const load = useCallback(async () => {
    setError(null);
    try {
      setUsers(await api.getUsers());
    } catch (e) {
      setError(e instanceof ApiError ? e.message : translate("Не удалось загрузить пользователей"));
    }
  }, []);

  useEffect(() => {
    load();
  }, [load]);

  async function handleDelete(user: User) {
    setDeletingId(user.id);
    setError(null);
    try {
      await api.deleteUser(user.id);
      setUsers((prev) => prev?.filter((u) => u.id !== user.id) ?? null);
    } catch (e) {
      setError(e instanceof ApiError ? e.message : translate("Действие не удалось"));
    } finally {
      setDeletingId(null);
    }
  }

  async function handleRename(user: User, firstName: string, lastName: string) {
    setError(null);
    try {
      const updated = await api.updateProfile(firstName, lastName, user.id);
      setUsers((prev) => prev?.map((u) => (u.id === updated.id ? updated : u)) ?? null);
    } catch (e) {
      throw new Error(e instanceof ApiError ? e.message : translate("Не удалось сохранить имя"));
    }
  }

  const normalizedQuery = search.trim().toLowerCase();
  const filteredUsers = users?.filter((u) => {
    if (!normalizedQuery) return true;
    const fullName = `${u.firstName} ${u.lastName}`.toLowerCase();
    return fullName.includes(normalizedQuery);
  });

  const pages = Math.ceil((filteredUsers?.length ?? 0) / 20);
  const page = Math.min(pages || 1, Math.max(1, Number.parseInt(params.get("page") ?? "1", 10) || 1));
  function navigatePage(n: number) {
    const next = new URLSearchParams(params);
    next.set("page", String(n));
    setParams(next);
  }
  return (
    <section className="card">
      {error && <ErrorBanner message={error} onRetry={load} />}

      {users === null && !error && (
        <div className="centered">
          <Spinner />
        </div>
      )}

      {users?.length === 0 && <EmptyState title={translate("Пользователей пока нет")} />}

      {users && users.length > 0 && (
        <>
          <input
            type="text"
            aria-label={translate("Поиск")}
            className="field"
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            placeholder={translate("Поиск по имени…")}
          />

          {filteredUsers?.length === 0 && (
            <EmptyState title={translate("Никого не нашлось")} hint={translate("Попробуй другой запрос")} />
          )}

          {pages > 1 && (
            <nav aria-label={translate("Пользователи")}>
              <button disabled={page <= 1} onClick={() => navigatePage(page - 1)}>
                ‹
              </button>{" "}
              {page} / {pages}{" "}
              <button disabled={page >= pages} onClick={() => navigatePage(page + 1)}>
                ›
              </button>
            </nav>
          )}
          {filteredUsers && filteredUsers.length > 0 && (
            <div className="admin-users">
              {filteredUsers.slice((page - 1) * 20, page * 20).map((u) => (
                <UserCard
                  key={u.id}
                  user={u}
                  isSelf={u.id === claims?.user_id}
                  deleting={deletingId === u.id}
                  onDelete={handleDelete}
                  onRename={handleRename}
                />
              ))}
            </div>
          )}
        </>
      )}
    </section>
  );
}
