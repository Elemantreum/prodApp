import { useEffect, useState, type FormEvent } from "react";
import { api } from "../api/client";
import type { SectionInvite } from "../api/types";
import { locale, t } from "../i18n";
import { AdminDialog } from "./AdminUI";
import { ErrorBanner } from "./ErrorBanner";

const MAX_USES = 2_147_483_647;

function browserInviteUrl(code: string) {
  const url = new URL("/invite", window.location.origin);
  url.hash = new URLSearchParams({ invite: code }).toString();
  return url.href;
}

async function copyText(value: string) {
  try {
    await navigator.clipboard.writeText(value);
    return true;
  } catch {
    const focused = document.activeElement;
    const field = document.createElement("textarea");
    field.value = value;
    field.readOnly = true;
    field.style.position = "fixed";
    field.style.opacity = "0";
    document.body.appendChild(field);
    field.select();
    try {
      return document.execCommand("copy");
    } catch {
      return false;
    } finally {
      field.remove();
      if (focused instanceof HTMLElement) focused.focus({ preventScroll: true });
    }
  }
}

export function AdminInvites({ sectionId }: { sectionId: number }) {
  const [invites, setInvites] = useState<SectionInvite[]>();
  const [loadError, setLoadError] = useState(false);
  const [attempt, setAttempt] = useState(0);
  const [creating, setCreating] = useState(false);
  const [expiresAt, setExpiresAt] = useState("");
  const [maxUses, setMaxUses] = useState("1");
  const [revoking, setRevoking] = useState<SectionInvite>();
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");
  const [message, setMessage] = useState("");
  const [manualCopy, setManualCopy] = useState<{ value: string; label: string }>();

  useEffect(() => {
    let active = true;
    api
      .sectionInvites(sectionId)
      .then((rows) => {
        if (active) setInvites(rows);
      })
      .catch(() => {
        if (active) setLoadError(true);
      });
    return () => {
      active = false;
    };
  }, [sectionId, attempt]);

  async function create(event: FormEvent) {
    event.preventDefault();
    if (busy) return;
    const expiry = new Date(expiresAt);
    const limit = Number(maxUses);
    if (!Number.isFinite(expiry.getTime()) || expiry.getTime() <= Date.now()) {
      setError(t("Укажите дату и время в будущем."));
      return;
    }
    if (!Number.isInteger(limit) || limit < 1 || limit > MAX_USES) {
      setError(t("Укажите целое число использований от 1 до 2 147 483 647."));
      return;
    }
    setBusy(true);
    setError("");
    try {
      const invite = await api.createSectionInvite(sectionId, { expiresAt: expiry.toISOString(), maxUses: limit });
      setInvites((rows) => [invite, ...(rows ?? [])]);
      setCreating(false);
      setMessage(t("Приглашение создано. Скопируйте ссылку или код перед уходом со страницы."));
    } catch {
      setError(t("Не удалось создать приглашение."));
    } finally {
      setBusy(false);
    }
  }

  async function revoke(invite: SectionInvite) {
    if (busy) return;
    setBusy(true);
    setError("");
    try {
      await api.revokeSectionInvite(sectionId, invite.id);
      const revokedAt = new Date().toISOString();
      setInvites((rows) => rows?.map((row) => (row.id === invite.id ? { ...row, revokedAt } : row)));
      setRevoking(undefined);
      setMessage(t("Приглашение отозвано."));
    } catch {
      setError(t("Не удалось отозвать приглашение."));
    } finally {
      setBusy(false);
    }
  }

  async function copy(code: string, link: boolean) {
    const value = link ? browserInviteUrl(code) : code;
    setMessage("");
    if (await copyText(value)) {
      setMessage(t(link ? "Ссылка скопирована." : "Код скопирован."));
    } else {
      setManualCopy({ value, label: t(link ? "Ссылка для браузера" : "Код приглашения") });
    }
  }

  return (
    <section className="admin-invites personal-panel">
      <div className="admin-invite-heading">
        <h2>{t("Приглашения")}</h2>
        <button
          type="button"
          className="btn"
          disabled={!invites}
          onClick={() => {
            setExpiresAt("");
            setMaxUses("1");
            setError("");
            setCreating(true);
          }}
        >
          {t("Создать приглашение")}
        </button>
      </div>
      {loadError && (
        <ErrorBanner
          message={t("Не удалось загрузить приглашения.")}
          onRetry={() => {
            setLoadError(false);
            setAttempt((value) => value + 1);
          }}
        />
      )}
      {!invites && !loadError && <p role="status">{t("Загрузка…")}</p>}
      {invites?.length === 0 && <p className="muted">{t("Приглашений пока нет.")}</p>}
      {message && <p role="status">{message}</p>}
      {!!invites?.length && (
        <ul>
          {invites.map((invite) => {
            const expired = Date.parse(invite.expiresAt) <= Date.now();
            const exhausted = invite.uses >= invite.maxUses;
            const status = invite.revokedAt
              ? "Отозвано"
              : expired
                ? "Срок истёк"
                : exhausted
                  ? "Лимит исчерпан"
                  : "Действует";
            const code = invite.code;
            return (
              <li className="admin-invite-row" key={invite.id}>
                <div className="personal-detail">
                  <strong>{t(status)}</strong>
                  <small>
                    {t("Действует до")}{" "}
                    <time dateTime={invite.expiresAt}>{new Date(invite.expiresAt).toLocaleString(locale)}</time>
                  </small>
                  <small>
                    {t("Использовано")} {invite.uses} {t("из")} {invite.maxUses}
                  </small>
                </div>
                {!invite.revokedAt && (
                  <div className="admin-invite-actions">
                    {!expired && !exhausted && code && (
                      <>
                        <button type="button" className="btn" onClick={() => void copy(code, true)}>
                          {t("Скопировать ссылку")}
                        </button>
                        <button type="button" className="btn" onClick={() => void copy(code, false)}>
                          {t("Скопировать код")}
                        </button>
                      </>
                    )}
                    <button
                      type="button"
                      className="btn danger"
                      onClick={() => {
                        setError("");
                        setRevoking(invite);
                      }}
                    >
                      {t("Отозвать")}
                    </button>
                  </div>
                )}
              </li>
            );
          })}
        </ul>
      )}
      {creating && (
        <AdminDialog title={t("Создать приглашение")} busy={busy} onClose={() => setCreating(false)}>
          <form onSubmit={create}>
            <label>
              {t("Действует до")}
              <input
                type="datetime-local"
                required
                value={expiresAt}
                onChange={(event) => setExpiresAt(event.target.value)}
              />
            </label>
            <label>
              {t("Максимум использований")}
              <input
                type="number"
                required
                min="1"
                max={MAX_USES}
                step="1"
                value={maxUses}
                onChange={(event) => setMaxUses(event.target.value)}
              />
            </label>
            {error && <ErrorBanner message={error} />}
            <button type="submit" className="btn primary" disabled={busy}>
              {t("Создать")}
            </button>
          </form>
        </AdminDialog>
      )}
      {revoking && (
        <AdminDialog title={t("Отозвать приглашение?")} busy={busy} onClose={() => setRevoking(undefined)}>
          {error && <ErrorBanner message={error} />}
          <button type="button" className="btn danger" disabled={busy} onClick={() => void revoke(revoking)}>
            {t("Отозвать")}
          </button>
        </AdminDialog>
      )}
      {manualCopy && (
        <AdminDialog title={manualCopy.label} onClose={() => setManualCopy(undefined)}>
          <p>{t("Выделите и скопируйте текст.")}</p>
          <input
            readOnly
            aria-label={manualCopy.label}
            value={manualCopy.value}
            onFocus={(event) => event.currentTarget.select()}
          />
        </AdminDialog>
      )}
    </section>
  );
}
