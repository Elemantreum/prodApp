import { t as translate } from "../i18n";
import { useEffect, useRef, useState } from "react";
import { api } from "../api/client";
import { useAuth } from "../auth/AuthContext";

declare global {
  interface Window {
    prodAppTelegramLogin?: (fields: Record<string, string>) => void;
  }
}

export function TelegramWebsiteLogin() {
  const { applyTokens } = useAuth();
  const container = useRef<HTMLDivElement>(null);
  const [error, setError] = useState("");
  const [attempt, setAttempt] = useState(0);
  useEffect(() => {
    let active = true;
    const root = container.current;
    setError("");
    window.prodAppTelegramLogin = async (fields) => {
      try {
        const tokens = await api.websiteLogin(fields);
        if (active) applyTokens(tokens);
      } catch {
        if (active) setError(translate("Не удалось подтвердить вход через Telegram."));
      }
    };
    api
      .websiteBot()
      .then(({ bot }) => {
        if (!active || !root) return;
        if (!bot) throw new Error(translate("Вход на сайте пока не настроен. Открой приложение в Telegram."));
        const script = document.createElement("script");
        script.src = "https://telegram.org/js/telegram-widget.js?22";
        script.async = true;
        script.dataset.telegramLogin = bot;
        script.dataset.size = "large";
        script.dataset.onauth = "prodAppTelegramLogin(user)";
        script.onerror = () => {
          if (active) setError(translate("Не удалось загрузить Telegram."));
        };
        root.appendChild(script);
      })
      .catch((e) => {
        if (active) setError(e instanceof Error ? e.message : translate("Не удалось загрузить вход."));
      });
    return () => {
      active = false;
      root?.replaceChildren();
      delete window.prodAppTelegramLogin;
    };
  }, [applyTokens, attempt]);
  return (
    <>
      <div ref={container} aria-label={translate("Вход через Telegram")} />
      {error && (
        <div role="alert">
          {error} <button onClick={() => setAttempt((n) => n + 1)}>{translate("Повторить")}</button>
        </div>
      )}
    </>
  );
}
