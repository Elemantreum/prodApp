import { useEffect, useState } from "react";
import { useSearchParams } from "react-router-dom";
import { api } from "../api/client";
import { useSection } from "../auth/Sections";
import { AdminDialog, AdminSearch, useAdminCreate } from "../components/AdminUI";
import { AdminInvites } from "../components/AdminInvites";
import { t } from "../i18n";

import type { User } from "../api/types";

type Person = Awaited<ReturnType<typeof api.sectionMembers>>[number];

export function AdminMembersPage({ administrators = false }: { administrators?: boolean }) {
  const { current, directory, choose } = useSection();
  const [params, setParams] = useSearchParams();
  const scope = params.get("scope") === "system" ? "system" : "section";
  const [accounts, setAccounts] = useState<User[]>();
  const systemScope = administrators && scope === "system";
  const [members, setMembers] = useState<Person[]>();
  const [people, setPeople] = useState<Person[]>([]);
  const [error, setError] = useState("");
  const [query, setQuery] = useState("");
  const [selected, setSelected] = useState<Person>();
  const [adding, setAdding] = useAdminCreate();
  const [id, setId] = useState("");
  const [confirm, setConfirm] = useState(false);
  const [busy, setBusy] = useState(false);
  const [version, setVersion] = useState(0);
  useEffect(() => {
    let active = true;
    if (systemScope) {
      api
        .getUsers()
        .then((users) => {
          if (active) setAccounts(users);
        })
        .catch((e) => {
          if (active) setError(String(e));
        });
      return () => {
        active = false;
      };
    }
    Promise.all([api.sectionMembers(), api.candidates()])
      .then(([rows, candidates]) => {
        if (active) {
          setMembers(rows);
          setPeople(candidates);
        }
      })
      .catch((e) => {
        if (active) setError(String(e));
      });
    return () => {
      active = false;
    };
  }, [current.id, version, systemScope]);
  function close() {
    setSelected(undefined);
    setAdding(false);
    setConfirm(false);
    setId("");
  }
  async function apply(admin: boolean, remove = false) {
    setBusy(true);
    setError("");
    try {
      await api.setMember(current.id, selected?.id ?? Number(id), admin, remove);
      setVersion((v) => v + 1);
      close();
    } catch (e) {
      setError(String(e));
    } finally {
      setBusy(false);
    }
  }
  const roster = systemScope ? accounts?.map((u) => ({ ...u, admin: u.isAdmin })) : members;
  const rows = roster?.filter(
    (p) =>
      (!administrators || p.admin) &&
      (p.firstName + " " + p.lastName).toLowerCase().includes(query.trim().toLowerCase()),
  );
  const existing = new Set(roster?.filter((p) => !administrators || p.admin).map((p) => p.id));
  const role = (p: Person) =>
    systemScope ? t("Администратор системы") : p.admin ? t("Администратор секции") : t("Участник секции");
  return (
    <section className="card">
      {administrators && (
        <label>
          {t("Права")}
          <select
            value={scope}
            onChange={(e) => {
              const next = new URLSearchParams(params);
              next.set("scope", e.target.value);
              next.delete("new");
              setParams(next);
              setError("");
            }}
          >
            <option value="section">{t("Администратор секции")}</option>
            <option value="system">{t("Администратор системы")}</option>
          </select>
        </label>
      )}
      {administrators && !systemScope && (
        <label>
          {t("Секция")}
          <select value={current.id} onChange={(e) => choose(Number(e.target.value))}>
            {directory.sections.map((s) => (
              <option key={s.id} value={s.id}>
                {s.name}
              </option>
            ))}
          </select>
        </label>
      )}
      <AdminSearch value={query} onChange={setQuery} />
      {error && <p role="alert">{error}</p>}
      {!roster && !error && <p role="status">{t("Загрузка…")}</p>}
      {rows?.map((p) => (
        <div className="row" key={p.id}>
          <span className="avatar">
            {p.firstName[0]}
            {p.lastName[0]}
          </span>
          <div className="grow">
            <b>
              {p.firstName} {p.lastName}
            </b>
            <small>{role(p)}</small>
          </div>
          {!systemScope && (
            <button
              className="btn"
              onClick={() => {
                setError("");
                setSelected(p);
              }}
            >
              {administrators ? t("Доступ") : t("Профиль")}
            </button>
          )}
        </div>
      ))}
      {rows?.length === 0 && <p>{t("Ничего не найдено")}</p>}
      {!systemScope && (adding || selected) && (
        <AdminDialog
          title={
            selected
              ? selected.firstName + " " + selected.lastName
              : administrators
                ? t("Назначить роль")
                : t("Добавить участника")
          }
          busy={busy}
          onClose={close}
        >
          <p>{systemScope ? t("Администратор системы") : current.name}</p>
          {selected ? (
            <>
              <p>{role(selected)}</p>
              <div className="admin-actions">
                {directory.systemAdmin && (
                  <button className="btn" disabled={busy} onClick={() => apply(!selected.admin)}>
                    {selected.admin ? t("Отозвать роль") : t("Назначить администратором")}
                  </button>
                )}
                {!administrators && (!selected.admin || directory.systemAdmin) && (
                  <button
                    className="btn danger"
                    disabled={busy}
                    onClick={() => (confirm ? apply(false, true) : setConfirm(true))}
                  >
                    {confirm ? t("Подтвердить") : t("Убрать доступ")}
                  </button>
                )}
              </div>
            </>
          ) : (
            <form
              onSubmit={(e) => {
                e.preventDefault();
                void apply(administrators);
              }}
            >
              <label>
                {t("Пользователь")}
                <select required value={id} onChange={(e) => setId(e.target.value)}>
                  <option value="">{t("Выбери участника")}</option>
                  {people
                    .filter((p) => !existing.has(p.id))
                    .map((p) => (
                      <option key={p.id} value={p.id}>
                        {p.firstName} {p.lastName}
                      </option>
                    ))}
                </select>
              </label>
              <button className="btn primary" disabled={!id || busy}>
                {t("Сохранить")}
              </button>
            </form>
          )}
          {error && <p role="alert">{error}</p>}
        </AdminDialog>
      )}
      {!administrators && <AdminInvites sectionId={current.id} />}
    </section>
  );
}
export function AdminAdministratorsPage() {
  return <AdminMembersPage administrators />;
}
