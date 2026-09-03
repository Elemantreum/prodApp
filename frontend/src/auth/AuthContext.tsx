import { createContext, useCallback, useContext, useEffect, useMemo, useState } from "react";
import type { ReactNode } from "react";
import { api, setAuthToken, setUnauthorizedHandler } from "../api/client";
import { detectPlatform } from "../platform/detect";
import type { Platform } from "../platform/detect";
import { applyTheme, DARK_THEME, LIGHT_THEME } from "../platform/theme";
import { getTelegramWebApp } from "../platform/telegram";
import { getVkLaunchQuery, initVk, subscribeVkTheme } from "../platform/vk";
import { decodeJwt, isExpired } from "./jwt";
import type { JwtClaims } from "./jwt";

const STORAGE_KEY = "prodapp.token";

type AuthStatus = "loading" | "authenticated" | "error" | "unsupported";

interface AuthContextValue {
  status: AuthStatus;
  platform: Platform;
  token: string | null;
  claims: JwtClaims | null;
  error: string | null;
  retry: () => void;
  /** Dev-only escape hatch for testing outside Telegram/VK — see LoginDevPanel. */
  loginWithToken: (token: string) => void;
}

const AuthContext = createContext<AuthContextValue | null>(null);

async function performLogin(platform: Platform): Promise<string> {
  if (platform === "telegram") {
    const tg = getTelegramWebApp();
    if (!tg?.initData) throw new Error("Telegram initData недоступен");
    const { access_token } = await api.authTelegram(tg.initData);
    return access_token;
  }
  if (platform === "vk") {
    const { access_token } = await api.authVk(getVkLaunchQuery());
    return access_token;
  }
  throw new Error("Открой это приложение внутри Telegram или VK");
}

export function AuthProvider({ children }: { children: ReactNode }) {
  const [platform] = useState<Platform>(() => detectPlatform());
  const [status, setStatus] = useState<AuthStatus>("loading");
  const [token, setToken] = useState<string | null>(null);
  const [claims, setClaims] = useState<JwtClaims | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [attempt, setAttempt] = useState(0);

  const applyToken = useCallback((next: string) => {
    const parsed = decodeJwt(next);
    if (!parsed) throw new Error("Сервер вернул некорректный токен");
    localStorage.setItem(STORAGE_KEY, next);
    setAuthToken(next);
    setToken(next);
    setClaims(parsed);
    setStatus("authenticated");
    setError(null);
  }, []);

  // Platform SDK bootstrap + theme wiring. Runs once.
  useEffect(() => {
    if (platform === "telegram") {
      const tg = getTelegramWebApp();
      tg?.ready();
      tg?.expand();
      const apply = () => {
        const tp = tg?.themeParams ?? {};
        applyTheme({
          "--bg": tp.bg_color,
          "--secondary-bg": tp.secondary_bg_color,
          "--text": tp.text_color,
          "--hint": tp.hint_color,
          "--accent": tp.button_color ?? tp.link_color,
          "--accent-text": tp.button_text_color,
        });
        if (tp.bg_color) {
          tg?.setHeaderColor(tp.bg_color);
          tg?.setBackgroundColor(tp.bg_color);
        }
      };
      apply();
      tg?.onEvent("themeChanged", apply);
      return () => tg?.offEvent("themeChanged", apply);
    }

    if (platform === "vk") {
      initVk().catch(() => {
        /* VKWebAppInit failing shouldn't block the app from rendering */
      });
      applyTheme(LIGHT_THEME);
      return subscribeVkTheme((scheme) => applyTheme(scheme === "dark" ? DARK_THEME : LIGHT_THEME));
    }

    applyTheme(LIGHT_THEME);
    return undefined;
  }, [platform]);

  // Auth bootstrap: reuse a cached, still-valid token; otherwise log in fresh.
  useEffect(() => {
    let cancelled = false;
    setStatus("loading");
    setError(null);

    async function run() {
      const cached = localStorage.getItem(STORAGE_KEY);
      if (cached) {
        const parsed = decodeJwt(cached);
        if (parsed && !isExpired(parsed)) {
          setAuthToken(cached);
          if (!cancelled) {
            setToken(cached);
            setClaims(parsed);
            setStatus("authenticated");
          }
          return;
        }
      }

      if (platform === "browser") {
        if (!cancelled) setStatus("unsupported");
        return;
      }

      try {
        const fresh = await performLogin(platform);
        if (!cancelled) applyToken(fresh);
      } catch (e) {
        if (!cancelled) {
          setStatus("error");
          setError(e instanceof Error ? e.message : "Не удалось войти");
        }
      }
    }

    run();
    return () => {
      cancelled = true;
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [platform, attempt]);

  // Let the API client silently mint a fresh token when one expires mid-session.
  useEffect(() => {
    setUnauthorizedHandler(async () => {
      if (platform === "browser") return null;
      try {
        const fresh = await performLogin(platform);
        applyToken(fresh);
        return fresh;
      } catch {
        localStorage.removeItem(STORAGE_KEY);
        setStatus("error");
        setError("Сессия истекла, перезапусти приложение");
        return null;
      }
    });
    return () => setUnauthorizedHandler(null);
  }, [platform, applyToken]);

  const retry = useCallback(() => setAttempt((n) => n + 1), []);

  const loginWithToken = useCallback(
    (raw: string) => {
      try {
        applyToken(raw);
      } catch (e) {
        setStatus("error");
        setError(e instanceof Error ? e.message : "Некорректный токен");
      }
    },
    [applyToken],
  );

  const value = useMemo<AuthContextValue>(
    () => ({ status, platform, token, claims, error, retry, loginWithToken }),
    [status, platform, token, claims, error, retry, loginWithToken],
  );

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuth(): AuthContextValue {
  const ctx = useContext(AuthContext);
  if (!ctx) throw new Error("useAuth must be used within AuthProvider");
  return ctx;
}
