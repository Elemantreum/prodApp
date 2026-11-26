import { t as translate } from "../i18n";
export function Spinner({ size = 24 }: { size?: number }) {
  return (
    <span
      className="spinner"
      style={{ width: size, height: size }}
      role="status"
      aria-label={translate("Загрузка")}
    />
  );
}
