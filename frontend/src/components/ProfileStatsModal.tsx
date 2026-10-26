import { useCallback, useEffect, useState } from "react";
import { X, ArrowRight, Edit3 } from "lucide-react";
import { api } from "../lib/api";
import { useOptionalAuth } from "../context/AuthContext";
import { useDialog } from "../lib/useDialog";
import { RequestError } from "./RequestState";
import type { ReadingStats } from "../types/api";
interface Props {
  isOpen: boolean;
  onClose: () => void;
  userName?: string;
  embedded?: boolean;
}
const months = [
  "Jan",
  "Feb",
  "Mar",
  "Apr",
  "May",
  "Jun",
  "Jul",
  "Aug",
  "Sep",
  "Oct",
  "Nov",
  "Dec",
];
const pacingLabels = {
  AHEAD: "Ahead of Schedule",
  ON_TRACK: "On Track",
  BEHIND: "Behind Schedule",
  COMPLETED: "Goal Completed",
  NOT_SET: "No Goal Set",
};
export function ProfileStatsModal({
  isOpen,
  onClose,
  userName = "Reader",
  embedded = false,
}: Props) {
  const auth = useOptionalAuth();
  const [stats, setStats] = useState<ReadingStats | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(false);
  const [editing, setEditing] = useState(false);
  const [target, setTarget] = useState("24");
  const [saving, setSaving] = useState(false);
  const [goalError, setGoalError] = useState("");
  const dialogRef = useDialog(isOpen && !embedded, onClose);
  const fetchStats = useCallback(async () => {
    setLoading(true);
    setError(false);
    try {
      setStats(await api.getReadingStats());
    } catch {
      setError(true);
    } finally {
      setLoading(false);
    }
  }, []);
  useEffect(() => {
    if (isOpen && (!embedded || auth?.isAuthenticated)) void fetchStats();
  }, [isOpen, embedded, auth?.isAuthenticated, fetchStats]);
  if (!isOpen) return null;
  const save = async () => {
    const value = Number(target);
    if (!Number.isInteger(value) || value < 1 || value > 1000) {
      setGoalError("Please enter a valid goal between 1 and 1,000 books.");
      return;
    }
    setSaving(true);
    try {
      await api.setReadingGoal(
        stats?.current_year || new Date().getFullYear(),
        value,
      );
      setEditing(false);
      await fetchStats();
    } catch {
      setGoalError("Your goal couldn’t be saved. Try again.");
    } finally {
      setSaving(false);
    }
  };
  const year = stats?.current_year || new Date().getFullYear();
  const challenge = stats?.challenge;
  const monthly = months.map((month, i) => ({
    month,
    books:
      challenge?.monthly_progress.find((m) => m.month === i + 1)?.books || 0,
  }));
  const max = Math.max(...monthly.map((m) => m.books), 1);
  const content = (
    <>
      <div className="page-heading">
        <div>
          <h1 id="stats-heading">Reading stats</h1>
          <p>A year in books, {userName}. Make room for the next one.</p>
        </div>
        {!embedded && (
          <button
            className="icon-button"
            aria-label="Close modal"
            onClick={onClose}
          >
            <X size={20} />
          </button>
        )}
      </div>
      {embedded && !auth?.isAuthenticated ? (
        <div className="empty-state">
          <h2>Your story, by the numbers.</h2>
          <p>
            Sign in to see your reading goals, finished books, and favorite
            genres.
          </p>
          <button
            className="button primary"
            onClick={() => auth?.openAuthModal("login")}
          >
            Sign in <ArrowRight size={16} />
          </button>
        </div>
      ) : loading ? (
        <div className="stats-loading" role="status">
          <div className="skeleton" style={{ height: 150 }} />
          <p>Loading your reading stats…</p>
        </div>
      ) : error ? (
        <RequestError
          title="Your reading stats couldn’t load."
          onRetry={() => void fetchStats()}
        />
      ) : stats ? (
        <>
          <div className="stats-strip">
            {[
              ["Books in collection", stats.total_books],
              ["Finished this year", stats.books_finished_year],
              ["Currently reading", stats.currently_reading],
              [
                "Average rating",
                stats.average_rating ? stats.average_rating.toFixed(1) : "—",
              ],
            ].map(([label, value]) => (
              <div key={label}>
                <span>{label}</span>
                <strong>{value}</strong>
              </div>
            ))}
          </div>
          <section className="reading-goal">
            <div className="goal-copy">
              <h2>{year} Reading Challenge</h2>
              <p>
                {challenge?.pacing_message ||
                  "Set a goal and give your reading a little direction."}
              </p>
              <span className="pacing-label">
                {pacingLabels[challenge?.pacing_status || "NOT_SET"]}
              </span>
              <button
                className="text-action"
                aria-label="Edit reading goal"
                onClick={() => {
                  setTarget(String(challenge?.target_books || 24));
                  setGoalError("");
                  setEditing(true);
                }}
              >
                <Edit3 size={14} />{" "}
                {challenge?.target_books ? "Edit goal" : "Set a reading goal"}
              </button>
            </div>
            <div className="goal-progress">
              <div>
                <span>
                  {challenge?.books_finished || 0} of{" "}
                  {challenge?.target_books || 0} books completed
                </span>
                <strong>{Math.round(challenge?.percentage || 0)}%</strong>
              </div>
              <progress
                max={100}
                value={Math.min(challenge?.percentage || 0, 100)}
                aria-label="Reading goal progress"
              />
              <p>
                {challenge?.target_books
                  ? `${Math.max(challenge.target_books - challenge.books_finished, 0)} books to your goal`
                  : "Every book is a beginning."}
              </p>
            </div>
          </section>
          {editing && (
            <form
              className="goal-editor"
              onSubmit={(e) => {
                e.preventDefault();
                void save();
              }}
            >
              <h3>How many books this year?</h3>
              <div className="goal-presets">
                {[12, 24, 36, 52].map((n) => (
                  <button
                    type="button"
                    className="button secondary compact"
                    key={n}
                    aria-pressed={target === String(n)}
                    onClick={() => setTarget(String(n))}
                  >
                    {n} books
                  </button>
                ))}
              </div>
              <label>
                Annual target
                <input
                  type="number"
                  aria-label="Target books input"
                  value={target}
                  onChange={(e) => setTarget(e.target.value)}
                />
              </label>
              {goalError && <p role="alert">{goalError}</p>}
              <div className="empty-actions">
                <button className="button primary compact" disabled={saving}>
                  {saving ? "Saving…" : "Save Goal"}
                </button>
                <button
                  type="button"
                  className="button secondary compact"
                  onClick={() => setEditing(false)}
                >
                  Cancel
                </button>
              </div>
            </form>
          )}
          <section className="reading-chart">
            <div className="section-heading">
              <div>
                <h2>Books through the year</h2>
                <p>Finished books per month · {year}</p>
              </div>
              <span>
                {stats.books_finished_year}{" "}
                {stats.books_finished_year === 1 ? "book" : "books"}
              </span>
            </div>
            <div
              className="month-chart"
              role="img"
              aria-label={monthly
                .map((m) => `${m.month}: ${m.books} books`)
                .join(", ")}
            >
              {monthly.map((m) => (
                <div className="month-column" key={m.month}>
                  <span className="month-value">{m.books}</span>
                  <div className="month-track">
                    <div style={{ height: `${(m.books / max) * 100}%` }} />
                  </div>
                  <span>{m.month}</span>
                </div>
              ))}
            </div>
          </section>
          <div className="stats-breakdowns">
            <section>
              <h2>Subjects you return to</h2>
              {stats.top_genres.length ? (
                stats.top_genres.map((g) => (
                  <div className="ranked-row" key={g.genre}>
                    <span>{g.genre}</span>
                    <span>{g.count} books</span>
                  </div>
                ))
              ) : (
                <p className="muted">
                  Add books with subject information to see your reading mix.
                </p>
              )}
            </section>
            <section>
              <h2>Authors on your shelves</h2>
              {stats.top_authors.length ? (
                stats.top_authors.map((a) => (
                  <div className="ranked-row" key={a.author}>
                    <span>{a.author}</span>
                    <span>{a.count} books</span>
                  </div>
                ))
              ) : (
                <p className="muted">
                  Your favorite authors will appear as your collection grows.
                </p>
              )}
            </section>
          </div>
          <section className="format-summary">
            <h2>Your editions</h2>
            <div>
              {Object.entries(stats.format_distribution).map(
                ([format, count]) => (
                  <p key={format}>
                    <span>{format.replaceAll("_", " ").toLowerCase()}</span>
                    <strong>{count}</strong>
                  </p>
                ),
              )}
            </div>
            {!Object.keys(stats.format_distribution).length && (
              <p className="muted">
                Choose editions for your books to see your preferred formats.
              </p>
            )}
          </section>
        </>
      ) : null}
    </>
  );
  return embedded ? (
    <section className="stats-page">{content}</section>
  ) : (
    <div className="dialog-backdrop" onClick={onClose}>
      <div
        ref={dialogRef}
        role="dialog"
        aria-modal="true"
        aria-labelledby="stats-heading"
        tabIndex={-1}
        className="dialog-panel stats-dialog"
        onClick={(e) => e.stopPropagation()}
      >
        {content}
      </div>
    </div>
  );
}
