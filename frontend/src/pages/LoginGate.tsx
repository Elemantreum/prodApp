import { Layout } from "../components/Layout";
import "../styles/personal.css";
import { t } from "../i18n";
import { useState } from "react";
import type { ReactNode } from "react";
import { useAuth } from "../auth/AuthContext";
import { TelegramWebsiteLogin } from "../components/TelegramWebsiteLogin";
import { Spinner } from "../components/Spinner";
import { Button } from "../components/Button";

export function LoginGate({ children }: { children: ReactNode }) {
  const { status, error, retry, platform, loginWithToken } = useAuth();

  if (status === "authenticated") return <>{children}</>;

  return (
    <Layout publicPage title={t("Вход в ProdApp")}>
      <section className="card personal-panel personal-gate">
        {status === "loading" ? (
          <div role="status">
            <Spinner /> {t("Входим…")}
          </div>
        ) : status === "error" ? (
          <>
            <p role="alert">{error}</p>
            <button className="btn" onClick={retry}>
              {t("Повторить")}
            </button>
          </>
        ) : (
          <>
            <p className="muted">{t("Войдите с аккаунтом Telegram, чтобы открыть свои секции.")}</p>
            <TelegramWebsiteLogin />
            {import.meta.env.DEV && <DevTokenPanel platform={platform} onSubmit={loginWithToken} />}
          </>
        )}
      </section>
    </Layout>
  );
}

function DevTokenPanel({ platform, onSubmit }: { platform: string; onSubmit: (token: string) => void }) {
  const [value, setValue] = useState("");

  return (
    <details className="personal-dev">
      <summary>{t("Войти с токеном")}</summary>
      <p className="dev-panel__hint">
        {" "}
        {t("Dev-режим (")}
        {platform}
        {t("): вставь access_token, полученный вручную от")} <code>POST /api/v1/auth/telegram</code> {t("или")}{" "}
        <code>/auth/vk</code>
        {t(", чтобы посмотреть приложение без реального Telegram/VK.")}{" "}
      </p>
      <textarea
        aria-label="access_token"
        value={value}
        onChange={(e) => setValue(e.target.value)}
        rows={3}
        placeholder="eyJhbGciOi..."
      />
      <Button disabled={!value.trim()} onClick={() => onSubmit(value.trim())}>
        {" "}
        {t("Войти с токеном")}{" "}
      </Button>
    </details>
  );
}
