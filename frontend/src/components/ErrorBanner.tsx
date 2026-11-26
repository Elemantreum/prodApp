import { t as translate } from "../i18n";
export function ErrorBanner({ message, onRetry }: { message: string; onRetry?: () => void }) {
  return (
    <div className="error-banner" role="alert">
      <span>{message}</span>
      {onRetry && (
        <button className="error-banner__retry" onClick={onRetry}>
          {" "}
          {translate("Повторить")}{" "}
        </button>
      )}
    </div>
  );
}
