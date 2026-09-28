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

export function getTelegramInitData(): string {
  return window.Telegram?.WebApp?.initData ?? "";
}
