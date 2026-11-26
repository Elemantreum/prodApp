import { useEffect, useState, type FormEvent } from "react";
import { Link, useNavigate } from "react-router-dom";
import { api } from "../api/client";
import { useAuth } from "../auth/AuthContext";
import { useProfile } from "../auth/Profile";
import { useSectionDirectory } from "../auth/Sections";
import { AdminDialog } from "../components/AdminUI";
import { ErrorBanner } from "../components/ErrorBanner";
import { Layout } from "../components/Layout";
import { t } from "../i18n";
import { getTelegramInitData } from "../platform/telegram";
import { getVkLaunchQuery } from "../platform/vk";
import "../styles/personal.css";

function platformName(provider: string) {
  if (provider === "browser") return t("браузер");
  return provider === "telegram" ? "Telegram" : provider === "vk" ? "VK" : provider;
}

export function AccountPage() {
  const { platform, applyTokens } = useAuth();
  const { profile, setProfile } = useProfile();
  const { choose, current, directory } = useSectionDirectory();
  const canEditProfile = directory.systemAdmin || !profile.profileLocked;
  const navigate = useNavigate();
  const [editing, setEditing] = useState(false);
  const [linking, setLinking] = useState(false);
  const [firstName, setFirstName] = useState(profile.firstName);
  const [lastName, setLastName] = useState(profile.lastName);
  const [code, setCode] = useState("");
  const [expires, setExpires] = useState(0);
  const [now, setNow] = useState(Date.now);
  const [input, setInput] = useState("");
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");
  const [message, setMessage] = useState("");
  const seconds = Math.max(0, Math.ceil((expires - now) / 1000));

  useEffect(() => {
    if (!expires) return;
    const timer = window.setInterval(() => setNow(Date.now()), 1000);
    return () => window.clearInterval(timer);
  }, [expires]);

  async function run(action: () => Promise<void>) {
    setBusy(true);
    setError("");
    setMessage("");
    try {
      await action();
    } catch (reason) {
      setError(reason instanceof Error ? reason.message : t("Не удалось сохранить"));
    } finally {
      setBusy(false);
    }
  }

  function save(event: FormEvent) {
    event.preventDefault();
    void run(async () => {
      setProfile(await api.updateProfile(firstName.trim(), lastName.trim()));
      setEditing(false);
      setMessage(t("Сохранено"));
    });
  }

  function link(event: FormEvent) {
    event.preventDefault();
    void run(async () => {
      const code = input.trim().toUpperCase();
      const data = getTelegramInitData();
      if (platform === "telegram" && !data) throw new Error(t("Telegram initData недоступен"));
      const tokens =
        platform === "telegram"
          ? await api.linkTelegram(data, code)
          : await api.linkVk(getVkLaunchQuery(), code);
      applyTokens(tokens);
      setProfile(await api.getOwnProfile());
      setInput("");
      setMessage(t("Готово, аккаунты объединены."));
    });
  }

  return (
    <Layout title={t("Профиль")} standalone={!current}>
      {error && !linking && !editing && <ErrorBanner message={error} />}
      {message && <p role="status">{message}</p>}
      <div className="grid personal-grid">
        <section className="card personal-panel">
          <h2>{[profile.firstName, profile.lastName].filter(Boolean).join(" ") || t("Без имени")}</h2>
          <small>
            {t("ID аккаунта:")} {profile.id} · {t("Вы вошли через")} {platformName(platform)}
          </small>
          {profile.identities?.map((identity) => (
            <p key={identity.provider} className="muted">
              {platformName(identity.provider)} ·{" "}
              {identity.username ? `@${identity.username.replace(/^@/, "")}` : identity.externalId}
            </p>
          ))}
          {editing && canEditProfile && (
            <AdminDialog title={t("Изменить профиль")} busy={busy} onClose={() => setEditing(false)}>
              <form className="personal-form" onSubmit={save}>
                <label>
                  {t("Имя")}
                  <input
                    autoComplete="given-name"
                    required
                    value={firstName}
                    onChange={(e) => setFirstName(e.target.value)}
                  />
                </label>
                <label>
                  {t("Фамилия")}
                  <input
                    autoComplete="family-name"
                    value={lastName}
                    onChange={(e) => setLastName(e.target.value)}
                  />
                </label>
                {error && <ErrorBanner message={error} />}
                <div className="actions">
                  <button className="btn primary" disabled={busy || !firstName.trim()}>
                    {t("Сохранить")}
                  </button>
                  <button type="button" className="btn" disabled={busy} onClick={() => setEditing(false)}>
                    {t("Отменить")}
                  </button>
                </div>
              </form>
            </AdminDialog>
          )}
          {canEditProfile ? (
            <button
              className="btn"
              onClick={() => {
                setFirstName(profile.firstName);
                setLastName(profile.lastName);
                setError("");
                setEditing(true);
              }}
            >
              {t("Изменить профиль")}
            </button>
          ) : (
            <small>{t("После вступления в секцию имя может изменить системный администратор.")}</small>
          )}
          <button
            className="btn"
            onClick={() => {
              setError("");
              setLinking(true);
            }}
          >
            {t("Привязать другую платформу")}
          </button>
        </section>
        <section className="card personal-panel">
          <h2>{t("Мои секции")}</h2>
          {directory.sections.map((section) => (
            <div className="row personal-section" key={section.id}>
              <span className="personal-avatar">
                {section.logo ? <img src={section.logo} alt="" /> : section.name.slice(0, 1)}
              </span>
              <span className="personal-detail">
                <strong>{section.name}</strong>
                <small>{section.admin ? t("Администратор секции") : t("Участник")}</small>
              </span>
              <button
                className="btn"
                onClick={() => {
                  choose(section.id);
                  navigate("/");
                }}
              >
                {t("Перейти")}
              </button>
            </div>
          ))}
          <Link className="btn" to="/invite">
            {t("Ввести код приглашения")}
          </Link>
        </section>
      </div>
      {linking && (
        <AdminDialog title={t("Привязать другую платформу")} busy={busy} onClose={() => setLinking(false)}>
          <div className="personal-panel">
            {error && <ErrorBanner message={error} />}
            <p className="muted">
              {t(
                "Один человек может входить и из Telegram, и из VK под одним аккаунтом. Получите код здесь, затем откройте это же приложение на другой платформе и введите код там, на этой же странице.",
              )}
            </p>
            {code && seconds > 0 ? (
              <>
                <strong className="personal-code">{code}</strong>
                <small>
                  {t("Действует ещё")} {seconds} {t("сек")}
                </small>
              </>
            ) : (
              <>
                {code && <small>{t("Код истёк")}</small>}
                <button
                  className="btn"
                  disabled={busy}
                  onClick={() =>
                    run(async () => {
                      const result = await api.generateLinkCode();
                      setCode(result.code);
                      setExpires(Date.parse(result.expiresAt));
                      setNow(Date.now());
                    })
                  }
                >
                  {code ? t("Получить новый") : t("Получить код")}
                </button>
              </>
            )}
          </div>
          {(platform === "telegram" || platform === "vk") && (
            <section className="personal-link-form personal-panel">
              <h2>{t("Ввести код с другой платформы")}</h2>
              <form className="personal-form" onSubmit={link}>
                <label>
                  {t("Код")}
                  <input required value={input} onChange={(e) => setInput(e.target.value)} />
                </label>
                <button className="btn primary" disabled={busy || !input.trim()}>
                  {t("Привязать")}
                </button>
              </form>
            </section>
          )}
        </AdminDialog>
      )}
    </Layout>
  );
}
