export interface ThemeVars {
  "--bg": string;
  "--secondary-bg": string;
  "--text": string;
  "--hint": string;
  "--accent": string;
  "--accent-text": string;
  "--divider": string;
  "--danger": string;
}

export const LIGHT_THEME: ThemeVars = {
  "--bg": "#ffffff",
  "--secondary-bg": "#f2f2f7",
  "--text": "#0f0f10",
  "--hint": "#8b8d98",
  "--accent": "#3f8ae0",
  "--accent-text": "#ffffff",
  "--divider": "rgba(0, 0, 0, 0.08)",
  "--danger": "#e53e3e",
};

export const DARK_THEME: ThemeVars = {
  "--bg": "#18191a",
  "--secondary-bg": "#232426",
  "--text": "#f5f5f7",
  "--hint": "#8b8d98",
  "--accent": "#4d9eff",
  "--accent-text": "#ffffff",
  "--divider": "rgba(255, 255, 255, 0.09)",
  "--danger": "#ff6b6b",
};

export function applyTheme(vars: Partial<ThemeVars>): void {
  const root = document.documentElement.style;
  for (const [key, value] of Object.entries(vars)) {
    if (value) root.setProperty(key, value);
  }
}
