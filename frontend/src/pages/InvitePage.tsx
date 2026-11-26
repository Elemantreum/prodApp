import { useEffect, useState, type FormEvent } from "react";
import { useLocation, useNavigate } from "react-router-dom";
import { api, ApiError } from "../api/client";
import type { InvitedSection } from "../api/types";
import { useProfile } from "../auth/Profile";
import { useSectionDirectory } from "../auth/Sections";
import { ErrorBanner } from "../components/ErrorBanner";
import { Layout } from "../components/Layout";
import { t } from "../i18n";
import "../styles/personal.css";

function inviteError(reason: unknown, fallback: string) {
  if (reason instanceof ApiError && [400, 404, 409, 410, 422].includes(reason.status)) {
    return t("Приглашение недействительно или больше не доступно.");
  }
  return t(fallback);
}

export function InvitePage() {
  const { hash, search, pathname } = useLocation();
  const navigate = useNavigate();
  const fragmentCode = new URLSearchParams(hash.slice(1)).get("invite")?.trim();
  const code = fragmentCode || new URLSearchParams(search).get("invite")?.trim() || "";

  useEffect(() => {
    const query = new URLSearchParams(search);
    if (!query.has("invite")) return;
    query.delete("invite");
    navigate(
      { pathname, search: query.toString(), hash: code ? `#${new URLSearchParams({ invite: code })}` : hash },
      { replace: true },
    );
  }, [code, hash, navigate, pathname, search]);

  return (
    <InviteForm
      key={code}
      code={code}
      onCodeChange={(invite) =>
        navigate({ pathname, search, hash: `#${new URLSearchParams({ invite })}` }, { replace: true })
      }
    />
  );
}

function InviteForm({ code, onCodeChange }: { code: string; onCodeChange: (code: string) => void }) {
  const { current, reload } = useSectionDirectory();
  const { setProfile } = useProfile();
  const navigate = useNavigate();
  const [input, setInput] = useState(code);
  const [attempt, setAttempt] = useState(0);
  const [preview, setPreview] = useState<{ section?: InvitedSection; error?: string } | null>(null);
  const [joined, setJoined] = useState<InvitedSection>();
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");
  const sameCode = code !== "" && input.trim() === code;
  const checking = sameCode && !preview;
  const section = joined ?? (sameCode ? preview?.section : undefined);

  useEffect(() => {
    if (!code) return;
    let active = true;
    api
      .previewSectionInvite(code)
      .then((section) => {
        if (active) setPreview({ section });
      })
      .catch((reason) => {
        if (active) setPreview({ error: inviteError(reason, "Не удалось проверить приглашение.") });
      });
    return () => {
      active = false;
    };
  }, [code, attempt]);

  function check(event: FormEvent) {
    event.preventDefault();
    const next = input.trim();
    if (!next || checking || busy) return;
    setError("");
    if (next === code) {
      setPreview(null);
      setAttempt((value) => value + 1);
    } else {
      onCodeChange(next);
    }
  }

  async function join() {
    if (!section || busy) return;
    setError("");
    setBusy(true);
    let result = joined;
    try {
      if (!result) {
        result = await api.redeemSectionInvite(code);
        setJoined(result);
      }
      await Promise.all([reload(result.id), api.getOwnProfile().then(setProfile)]);
      navigate("/", { replace: true });
    } catch (reason) {
      setError(
        result
          ? t("Вы вступили в секцию. Не удалось обновить данные.")
          : inviteError(reason, "Не удалось вступить в секцию."),
      );
    } finally {
      setBusy(false);
    }
  }

  return (
    <Layout title={t("Приглашение в секцию")} standalone={!current}>
      <section className="card personal-panel personal-gate">
        {!current && !code && <p>{t("Введите код, который вам прислал администратор секции.")}</p>}
        {!joined && (
          <form className="personal-form" onSubmit={check}>
            <label>
              {t("Код приглашения")}
              <input
                required
                autoCapitalize="none"
                autoComplete="off"
                spellCheck={false}
                disabled={busy}
                value={input}
                onChange={(event) => {
                  setInput(event.target.value);
                  setError("");
                }}
              />
            </label>
            <button className="btn" disabled={!input.trim() || busy || checking}>
              {t("Проверить код")}
            </button>
          </form>
        )}
        {checking && <p role="status">{t("Проверяем приглашение…")}</p>}
        {sameCode && preview?.error && <ErrorBanner message={preview.error} />}
        {error && <ErrorBanner message={error} />}
        {section && (
          <>
            <div className="personal-section">
              <span className="personal-avatar" aria-hidden="true">
                {section.logo ? (
                  <img src={section.logo} alt="" referrerPolicy="no-referrer" />
                ) : (
                  section.name.slice(0, 1)
                )}
              </span>
              <strong className="personal-detail">{section.name}</strong>
            </div>
            <button type="button" className="btn primary" disabled={busy} onClick={join}>
              {t(joined ? "Повторить" : "Вступить в секцию")}
            </button>
          </>
        )}
      </section>
    </Layout>
  );
}
