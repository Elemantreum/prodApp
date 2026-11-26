import { locale } from "../i18n";

export function LanguageSelect() {
  return (
    <label>
      {locale === "en" ? "Language" : "Язык"}{" "}
      <select
        aria-label={locale === "en" ? "Language" : "Язык"}
        value={locale}
        onChange={(event) => {
          localStorage.setItem("prodapp.language", event.target.value);
          window.location.reload();
        }}
      >
        <option value="ru">Русский</option>
        <option value="en">English</option>
      </select>
    </label>
  );
}
