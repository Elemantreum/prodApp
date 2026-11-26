import { useEffect, useState, type FormEvent, type ReactNode } from "react";
import { api } from "../api/client";
import type { User } from "../api/types";
import { useAuth } from "../auth/AuthContext";
import { ProfileContext } from "../auth/Profile";
import { ErrorBanner } from "../components/ErrorBanner";
import { Layout } from "../components/Layout";
import { Spinner } from "../components/Spinner";
import { t } from "../i18n";
import "../styles/personal.css";

export function RegistrationGate({ children }: { children: ReactNode }) {
  const { claims } = useAuth();
  return <Registration key={claims?.user_id}>{children}</Registration>;
}

function Registration({ children }: { children: ReactNode }) {
  const [profile, setProfile] = useState<User | null>(null);
  const [firstName, setFirstName] = useState("");
  const [lastName, setLastName] = useState("");
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState("");
  const [attempt, setAttempt] = useState(0);

  useEffect(() => {
    let active = true;
    setError("");
    api
      .getOwnProfile()
      .then((user) => {
        if (!active) return;
        setProfile(user);
        setFirstName(user.firstName);
        setLastName(user.lastName);
      })
      .catch((reason) => {
        if (active) setError(reason instanceof Error ? reason.message : t("Не удалось проверить профиль"));
      });
    return () => {
      active = false;
    };
  }, [attempt]);

  async function submit(event: FormEvent) {
    event.preventDefault();
    setSaving(true);
    setError("");
    try {
      setProfile(await api.updateProfile(firstName.trim(), lastName.trim()));
    } catch (reason) {
      setError(reason instanceof Error ? reason.message : t("Не удалось сохранить"));
    } finally {
      setSaving(false);
    }
  }
  if (profile?.isRegistered)
    return <ProfileContext.Provider value={{ profile, setProfile }}>{children}</ProfileContext.Provider>;
  return (
    <Layout publicPage title={t("Регистрация")}>
      <section className="card personal-panel personal-gate">
        {error && (
          <ErrorBanner message={error} onRetry={profile ? undefined : () => setAttempt((value) => value + 1)} />
        )}
        {!profile && !error && <Spinner />}
        {profile && (
          <form className="personal-form" onSubmit={submit}>
            <h2>{t("Как к вам обращаться?")}</h2>
            <label>
              {t("Имя")}
              <input
                autoComplete="given-name"
                value={firstName}
                required
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
            <button type="submit" className="btn primary" disabled={saving || !firstName.trim()}>
              {t("Завершить регистрацию")}
            </button>
          </form>
        )}
      </section>
    </Layout>
  );
}
