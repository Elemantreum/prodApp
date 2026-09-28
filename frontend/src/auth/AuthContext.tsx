import { createContext, useCallback, useContext, useEffect, useMemo, useState } from "react";
import type { ReactNode } from "react";
import { api, setAuthToken, setUnauthorizedHandler } from "../api/client";
import { detectPlatform } from "../platform/detect";
import type { Platform } from "../platform/detect";
import { applyTheme, DARK_THEME, LIGHT_THEME } from "../platform/theme";
import { getTelegramInitData, getTelegramWebApp } from "../platform/telegram";
import { getVkLaunchQuery, initVk, subscribeVkTheme } from "../platform/vk";
import { decodeJwt, isExpired } from "./jwt";
import type { JwtClaims } from "./jwt";
import type { AuthResponse } from "../api/types";

const STORAGE_KEY = "prodapp.token";
const REFRESH_STORAGE_KEY = "prodapp.refreshToken";

type AuthStatus = "loading" | "authenticated" | "error" | "unsupported";

interface AuthContextValue {
  status: AuthStatus;
  platform: Platform;
  token: string | null;
  claims: JwtClaims | null;
  error: string | null;
  retry: () => void;
  logout: () => void;
  loginWithToken: (token: string) => void;
}

const AuthContext = createContext<AuthContextValue | null>(null);

async function performLogin(platform: Platform): Promise<AuthResponse> {
  if (platform === "telegram") {
    const initData = getTelegramInitData();
    if (!initData) throw new Error("Telegram initData недоступен");
    return api.authTelegram(initData);
  }
  if (platform === "vk") {
    return api.authVk(getVkLaunchQuery());
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

  const applyTokens = useCallback((next: AuthResponse) => {
    const parsed = decodeJwt(next.access_token);
    if (!parsed) throw new Error("Сервер вернул некорректный токен");
    localStorage.setItem(STORAGE_KEY, next.access_token);
    localStorage.setItem(REFRESH_STORAGE_KEY, next.refresh_token);
    setAuthToken(next.access_token);
    setToken(next.access_token);
    setClaims(parsed);
    setStatus("authenticated");
    setError(null);
  }, []);

  const tryRefresh = useCallback(async (): Promise<AuthResponse | null> => {
    const stored = localStorage.getItem(REFRESH_STORAGE_KEY);
    if (!stored) return null;
    try {
      return await api.refreshToken(stored);
    } catch {
      localStorage.removeItem(REFRESH_STORAGE_KEY);
      return null;
    }
  }, []);

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
      initVk().catch(() => {});
      applyTheme(LIGHT_THEME);
      return subscribeVkTheme((scheme) => applyTheme(scheme === "dark" ? DARK_THEME : LIGHT_THEME));
    }

    applyTheme(LIGHT_THEME);
    return undefined;
  }, [platform]);

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

      const refreshed = await tryRefresh();
      if (refreshed) {
        if (!cancelled) applyTokens(refreshed);
        return;
      }

      if (platform === "browser") {
        if (!cancelled) setStatus("unsupported");
        return;
      }

      try {
        const fresh = await performLogin(platform);
        if (!cancelled) applyTokens(fresh);
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

  useEffect(() => {
    setUnauthorizedHandler(async () => {
      const refreshed = await tryRefresh();
      if (refreshed) {
        applyTokens(refreshed);
        return refreshed.access_token;
      }

      if (platform === "browser") return null;
      try {
        const fresh = await performLogin(platform);
        applyTokens(fresh);
        return fresh.access_token;
      } catch {
        localStorage.removeItem(STORAGE_KEY);
        localStorage.removeItem(REFRESH_STORAGE_KEY);
        setStatus("error");
        setError("Сессия истекла, перезапусти приложение");
        return null;
      }
    });
    return () => setUnauthorizedHandler(null);
  }, [platform, applyTokens, tryRefresh]);

  const retry = useCallback(() => setAttempt((n) => n + 1), []);

  const loginWithToken = useCallback((raw: string) => {
    const parsed = decodeJwt(raw);
    if (!parsed) {
      setStatus("error");
      setError("Некорректный токен");
      return;
    }
    localStorage.setItem(STORAGE_KEY, raw);
    localStorage.removeItem(REFRESH_STORAGE_KEY);
    setAuthToken(raw);
    setToken(raw);
    setClaims(parsed);
    setStatus("authenticated");
    setError(null);
  }, []);

  const logout = useCallback(() => {
    const stored = localStorage.getItem(REFRESH_STORAGE_KEY);
    localStorage.removeItem(STORAGE_KEY);
    localStorage.removeItem(REFRESH_STORAGE_KEY);
    setAuthToken(null);
    setToken(null);
    setClaims(null);
    if (stored) api.logout(stored).catch(() => {});
    setAttempt((n) => n + 1);
  }, []);

  const value = useMemo<AuthContextValue>(
    () => ({ status, platform, token, claims, error, retry, logout, loginWithToken }),
    [status, platform, token, claims, error, retry, logout, loginWithToken],
  );

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuth(): AuthContextValue {
  const ctx = useContext(AuthContext);
  if (!ctx) throw new Error("useAuth must be used within AuthProvider");
  return ctx;
}
