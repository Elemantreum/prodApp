export const LIGHT_THEME = "light";
export const DARK_THEME = "dark";
export type Theme = typeof LIGHT_THEME | typeof DARK_THEME;

let platformTheme: Theme = DARK_THEME;

export function applyTheme(next?: Theme) {
  if (next) platformTheme = next;
  const preference = localStorage.getItem("prodapp.theme");
  const theme = preference === "light" || preference === "dark" ? preference : platformTheme;
  document.documentElement.dataset.theme = theme;
}
