import { useCallback, useEffect, useState } from "react";
import type { FormEvent } from "react";
import { api, ApiError } from "../api/client";
import type { User } from "../api/types";
import { useAuth } from "../auth/AuthContext";
import { getTelegramInitData } from "../platform/telegram";
import { getVkLaunchQuery } from "../platform/vk";
import { Layout } from "../components/Layout";
import { Button } from "../components/Button";
import { Spinner } from "../components/Spinner";
import { ErrorBanner } from "../components/ErrorBanner";

const PLATFORM_LABEL: Record<string, string> = {
  telegram: "Telegram",
  vk: "VK",
  browser: "браузер",
};

export function AccountPage() {
  const { platform, claims, loginWithToken } = useAuth();

  const [profile, setProfile] = useState<User | null>(null);
  const [profileError, setProfileError] = useState<string | null>(null);
  const [firstName, setFirstName] = useState("");
  const [lastName, setLastName] = useState("");
  const [savingName, setSavingName] = useState(false);
  const [nameError, setNameError] = useState<string | null>(null);
  const [nameSaved, setNameSaved] = useState(false);

  const [code, setCode] = useState<string | null>(null);
  const [expiresAt, setExpiresAt] = useState<string | null>(null);
  const [generating, setGenerating] = useState(false);
  const [genError, setGenError] = useState<string | null>(null);
  const [secondsLeft, setSecondsLeft] = useState(0);

  const [inputCode, setInputCode] = useState("");
  const [linking, setLinking] = useState(false);
  const [linkError, setLinkError] = useState<string | null>(null);
  const [linkSuccess, setLinkSuccess] = useState(false);

  const loadProfile = useCallback(async () => {
    setProfileError(null);
    try {
      const p = await api.getOwnProfile();
      setProfile(p);
      setFirstName(p.firstName);
      setLastName(p.lastName);
    } catch (e) {
      setProfileError(e instanceof ApiError ? e.message : "Не удалось загрузить профиль");
    }
  }, []);

  useEffect(() => {
    loadProfile();
  }, [loadProfile]);

  useEffect(() => {
    if (!expiresAt) return;
    const tick = () => setSecondsLeft(Math.max(0, Math.round((new Date(expiresAt).getTime() - Date.now()) / 1000)));
    tick();
    const id = setInterval(tick, 1000);
    return () => clearInterval(id);
  }, [expiresAt]);

  async function handleSaveName(e: FormEvent) {
    e.preventDefault();
    setSavingName(true);
    setNameError(null);
    setNameSaved(false);
    try {
      const updated = await api.updateProfile(firstName.trim(), lastName.trim());
      setProfile(updated);
      setNameSaved(true);
    } catch (e) {
      setNameError(e instanceof ApiError ? e.message : "Не удалось сохранить");
    } finally {
      setSavingName(false);
    }
  }

  async function handleGenerate() {
    setGenerating(true);
    setGenError(null);
    try {
      const res = await api.generateLinkCode();
      setCode(res.code);
      setExpiresAt(res.expiresAt);
    } catch (e) {
      setGenError(e instanceof ApiError ? e.message : "Не удалось получить код");
    } finally {
      setGenerating(false);
    }
  }

  async function handleLink(e: FormEvent) {
    e.preventDefault();
    setLinking(true);
    setLinkError(null);
    setLinkSuccess(false);
    try {
      let token: string;
      if (platform === "telegram") {
        const initData = getTelegramInitData();
        if (!initData) throw new Error("Telegram initData недоступен");
        token = (await api.linkTelegram(initData, inputCode.trim().toUpperCase())).access_token;
      } else if (platform === "vk") {
        token = (await api.linkVk(getVkLaunchQuery(), inputCode.trim().toUpperCase())).access_token;
      } else {
        throw new Error("Привязка доступна только внутри Telegram или VK");
      }

      loginWithToken(token);
      setLinkSuccess(true);
      setInputCode("");
      loadProfile(); // подтянуть только что появившуюся вторую identity
    } catch (e) {
      setLinkError(e instanceof ApiError ? e.message : e instanceof Error ? e.message : "Не удалось привязать аккаунт");
    } finally {
      setLinking(false);
    }
  }

  // Зеркалит UserService.UpdateProfile на бэке: админ может всегда, обычный
  // юзер — только пока не одобрен.
  const canEditName = profile ? profile.isAdmin || !profile.isApproved : false;

  return (
    <Layout title="Аккаунт">
      <div className="list">
        <div className="card account-card">
          <div className="card__title">Вы вошли через {PLATFORM_LABEL[platform] ?? platform}</div>
          <div className="card__meta">ID аккаунта: {claims?.user_id}</div>

          {profileError && <ErrorBanner message={profileError} onRetry={loadProfile} />}

          {profile === null && !profileError && (
            <div className="centered">
              <Spinner size={20} />
            </div>
          )}

          {profile && (
            <div className="card__meta">
              Привязанные платформы:{" "}
              {profile.identities.length === 0
                ? "нет данных"
                : profile.identities.map((i) => PLATFORM_LABEL[i.provider] ?? i.provider).join(", ")}
            </div>
          )}
        </div>

        {profile && (
          <div className="card account-card">
            <div className="card__title">Имя и фамилия</div>
            {canEditName ? (
              <>
                <div className="card__meta">Можно менять, пока аккаунт не одобрен администратором.</div>
                <form className="form" onSubmit={handleSaveName}>
                  <label className="form__field">
                    <span>Имя</span>
                    <input type="text" value={firstName} onChange={(e) => setFirstName(e.target.value)} />
                  </label>
                  <label className="form__field">
                    <span>Фамилия</span>
                    <input type="text" value={lastName} onChange={(e) => setLastName(e.target.value)} />
                  </label>
                  <Button type="submit" loading={savingName}>
                    Сохранить
                  </Button>
                </form>
                {nameError && <ErrorBanner message={nameError} />}
                {nameSaved && <div className="success-banner">Сохранено</div>}
              </>
            ) : (
              <>
                <div className="card__title">{[profile.firstName, profile.lastName].filter(Boolean).join(" ") || "—"}</div>
                <div className="card__meta">Аккаунт одобрен — изменить имя может только администратор.</div>
              </>
            )}
          </div>
        )}

        <div className="card account-card">
          <div className="card__title">Привязать другую платформу</div>
          <div className="card__meta">
            Один человек может входить и из Telegram, и из VK под одним аккаунтом. Получите код здесь, затем
            откройте это же приложение на другой платформе и введите код там, на этой же странице.
          </div>

          {code ? (
            secondsLeft > 0 ? (
              <div className="link-code">
                <div className="link-code__value">{code}</div>
                <div className="link-code__hint">Действует ещё {secondsLeft} сек</div>
              </div>
            ) : (
              <div className="link-code">
                <div className="link-code__hint">Код истёк</div>
                <Button variant="secondary" loading={generating} onClick={handleGenerate}>
                  Получить новый
                </Button>
              </div>
            )
          ) : (
            <Button variant="secondary" loading={generating} onClick={handleGenerate}>
              Получить код
            </Button>
          )}
          {genError && <ErrorBanner message={genError} />}
        </div>

        <div className="card account-card">
          <div className="card__title">Ввести код с другой платформы</div>
          <form className="form" onSubmit={handleLink}>
            <label className="form__field">
              <span>Код</span>
              <input
                type="text"
                value={inputCode}
                onChange={(e) => setInputCode(e.target.value)}
                placeholder="XXXXXXXX"
                required
              />
            </label>
            <Button type="submit" loading={linking} disabled={!inputCode.trim()}>
              Привязать
            </Button>
          </form>
          {linkError && <ErrorBanner message={linkError} />}
          {linkSuccess && <div className="success-banner">Готово, аккаунты объединены.</div>}
        </div>
      </div>
    </Layout>
  );
}
