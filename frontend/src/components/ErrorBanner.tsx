export function ErrorBanner({ message, onRetry }: { message: string; onRetry?: () => void }) {
  return (
    <div className="error-banner">
      <span>{message}</span>
      {onRetry && (
        <button className="error-banner__retry" onClick={onRetry}>
          Повторить
        </button>
      )}
    </div>
  );
}
