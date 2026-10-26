import { RefreshCw } from "lucide-react";
export function RequestError({
  title,
  onRetry,
}: {
  title: string;
  onRetry: () => void;
}) {
  return (
    <div className="request-error" role="alert">
      <div>
        <h2>{title}</h2>
        <p>Check your connection and try again.</p>
      </div>
      <button className="button secondary" onClick={onRetry}>
        <RefreshCw size={16} />
        Try again
      </button>
    </div>
  );
}
