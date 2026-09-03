// Minimal hand-rolled typing for the bits of the Telegram Mini Apps JS SDK
// (window.Telegram.WebApp, injected by telegram-web-app.js) that this app
// actually uses. Not a full SDK — see index.html for the script tag.

export interface TelegramThemeParams {
  bg_color?: string;
  text_color?: string;
  hint_color?: string;
  link_color?: string;
  button_color?: string;
  button_text_color?: string;
  secondary_bg_color?: string;
  destructive_text_color?: string;
}

interface TelegramMainButton {
  text: string;
  isVisible: boolean;
  isActive: boolean;
  isProgressVisible: boolean;
  setText(text: string): void;
  onClick(cb: () => void): void;
  offClick(cb: () => void): void;
  show(): void;
  hide(): void;
  enable(): void;
  disable(): void;
  showProgress(leaveActive?: boolean): void;
  hideProgress(): void;
}

interface TelegramBackButton {
  isVisible: boolean;
  onClick(cb: () => void): void;
  offClick(cb: () => void): void;
  show(): void;
  hide(): void;
}

interface TelegramHapticFeedback {
  impactOccurred(style: "light" | "medium" | "heavy" | "rigid" | "soft"): void;
  notificationOccurred(type: "error" | "success" | "warning"): void;
  selectionChanged(): void;
}

export interface TelegramWebApp {
  initData: string;
  colorScheme: "light" | "dark";
  themeParams: TelegramThemeParams;
  MainButton: TelegramMainButton;
  BackButton: TelegramBackButton;
  HapticFeedback: TelegramHapticFeedback;
  ready(): void;
  expand(): void;
  close(): void;
  setHeaderColor(color: string): void;
  setBackgroundColor(color: string): void;
  onEvent(event: "themeChanged" | "viewportChanged", cb: () => void): void;
  offEvent(event: "themeChanged" | "viewportChanged", cb: () => void): void;
}

declare global {
  interface Window {
    Telegram?: { WebApp?: TelegramWebApp };
  }
}

export function getTelegramWebApp(): TelegramWebApp | null {
  return window.Telegram?.WebApp ?? null;
}

export function isTelegram(): boolean {
  return !!getTelegramWebApp()?.initData;
}

/**
 * Deliberately a LIVE read on every call, unlike VK's getVkLaunchQuery()
 * (which caches at module load). Tried caching this the same way once —
 * wrong move: Telegram's client keeps window.Telegram.WebApp.initData fresh
 * over the WebView's lifetime (e.g. when the app regains focus after being
 * backgrounded), so a one-time snapshot goes stale while the session is
 * still open. That stale snapshot is exactly what made "Повторить" after a
 * 401 keep failing with a real "unauthorized" from the backend — the retry
 * resent the frozen old initData instead of the currently-valid one — while
 * a full relaunch worked, because relaunching re-runs this module and grabs
 * a fresh value. VK's query-string caching stays as-is; that one's forced by
 * React Router wiping location.search on navigation, a different problem.
 */
export function getTelegramInitData(): string {
  return window.Telegram?.WebApp?.initData ?? "";
}
