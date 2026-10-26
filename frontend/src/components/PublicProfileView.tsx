import { useEffect, useState } from "react";
import { ArrowLeft, Share2, ArrowUpRight } from "lucide-react";
import { api } from "../lib/api";
import { ShareModal } from "./ShareModal";
import { STATUS_LABELS } from "./WishlistCard";
import { RequestError } from "./RequestState";
import type {
  PublicProfile,
  WishlistItem,
  ReadingStatus,
  Work,
} from "../types/api";
interface Props {
  username: string;
  initialShelf?: string;
  onNavigateHome: () => void;
  onInspectWork?: (work: Work) => void;
}
export function PublicProfileView({
  username,
  initialShelf,
  onNavigateHome,
  onInspectWork,
}: Props) {
  const [profile, setProfile] = useState<PublicProfile | null>(null);
  const [items, setItems] = useState<WishlistItem[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<"private" | "missing" | null>(null);
  const [status, setStatus] = useState("ALL");
  const [shelf, setShelf] = useState(initialShelf || "");
  const [sharing, setSharing] = useState(false);
  const [retry, setRetry] = useState(0);
  useEffect(() => {
    let current = true;
    setLoading(true);
    setError(null);
    api
      .getPublicCollection(username, {
        status: status === "ALL" ? undefined : status,
        tag: shelf || undefined,
      })
      .then((response) => {
        if (current) {
          setProfile(response.profile);
          setItems(response.items || []);
        }
      })
      .catch((err: unknown) => {
        if (current)
          setError(
            err instanceof Error && /private|forbidden/i.test(err.message)
              ? "private"
              : "missing",
          );
      })
      .finally(() => {
        if (current) setLoading(false);
      });
    return () => {
      current = false;
    };
  }, [username, status, shelf, retry]);
  // The existing public endpoint includes the profile, avoiding a second waterfall request.
  useEffect(() => {
    setShelf(initialShelf || "");
  }, [initialShelf]);
  const toggleShelf = (tag: string) => {
    const next = shelf === tag ? "" : tag;
    setShelf(next);
    window.history.replaceState(
      null,
      "",
      `#/u/${encodeURIComponent(username)}${next ? `/shelf/${encodeURIComponent(next)}` : ""}`,
    );
  };
  const share = () => setSharing(true);
  const challenge = profile?.stats?.challenge;
  return (
    <div className="public-shell">
      <header className="public-header">
        <button className="back-link" onClick={onNavigateHome}>
          <ArrowLeft size={17} />
          Return to Imprint
        </button>
        <span className="wordmark">imprint.</span>
      </header>
      <main className="app-main">
        {loading ? (
          <div className="stats-loading" role="status">
            <div className="skeleton" style={{ height: 180 }} />
            <p>Opening @{username}'s library...</p>
          </div>
        ) : error === "private" ? (
          <div className="empty-state">
            <h1>This Library is Private</h1>
            <p>
              @{username} has chosen to keep their reading collection, shelves,
              and reading analytics private.
            </p>
            <button className="button primary" onClick={onNavigateHome}>
              Discover books
            </button>
          </div>
        ) : error ? (
          <RequestError
            title="Reader profile couldn’t load."
            onRetry={() => setRetry(retry + 1)}
          />
        ) : profile ? (
          <>
            <section className="public-profile-heading">
              <div className="public-reader">
                <span className="reader-initial">
                  {(profile.display_name || profile.username).charAt(0)}
                </span>
                <div>
                  <h1>{profile.display_name || profile.username}</h1>
                  <p>@{profile.username} · A reader’s collection</p>
                </div>
              </div>
              <button className="button secondary" onClick={share}>
                <Share2 size={16} />
                Share Collection
              </button>
            </section>
            {profile.stats && (
              <div className="stats-strip">
                <div>
                  <span>On the shelves</span>
                  <strong>{profile.stats.total_books}</strong>
                </div>
                <div>
                  <span>Finished this year</span>
                  <strong>{profile.stats.books_finished_year}</strong>
                </div>
                <div>
                  <span>Reading now</span>
                  <strong>{profile.stats.currently_reading}</strong>
                </div>
                <div>
                  <span>
                    {challenge?.target_books
                      ? `${challenge.books_finished} of ${challenge.target_books} read`
                      : "Reading goal"}
                  </span>
                  <strong>
                    {challenge?.target_books
                      ? `${Math.round(challenge.percentage)}%`
                      : "Not set"}
                  </strong>
                </div>
              </div>
            )}
            <div className="public-filters">
              <div className="public-shelves">
                <button
                  className={!shelf ? "active" : ""}
                  onClick={() => setShelf("")}
                >
                  All shelves
                </button>
                {profile.shelves.map((t) => (
                  <button
                    className={shelf === t.tag ? "active" : ""}
                    key={t.tag}
                    onClick={() => toggleShelf(t.tag)}
                  >
                    #{t.tag} <span>{t.count}</span>
                  </button>
                ))}
              </div>
              <label>
                <span className="sr-only">Filter reading status</span>
                <select
                  value={status}
                  onChange={(e) => setStatus(e.target.value)}
                >
                  <option value="ALL">All reading states</option>
                  {Object.entries(STATUS_LABELS).map(([id, label]) => (
                    <option key={id} value={id}>
                      {label}
                    </option>
                  ))}
                </select>
              </label>
            </div>
            {!items.length ? (
              <div className="empty-state">
                <h2>A little room on this shelf.</h2>
                <p>No books match these filters yet.</p>
                <button
                  className="button secondary"
                  onClick={() => {
                    setStatus("ALL");
                    setShelf("");
                  }}
                >
                  View all books
                </button>
              </div>
            ) : (
              <div className="book-grid">
                {items.map((item) => (
                  <article className="book-card" key={item.id}>
                    <button
                      className="book-cover"
                      onClick={() => item.work && onInspectWork?.(item.work)}
                      aria-label={`View ${item.work?.title || "book"} editions`}
                    >
                      {item.work?.cover_url ? (
                        <img
                          src={item.work.cover_url}
                          alt={`Cover of ${item.work.title}`}
                          loading="lazy"
                          onError={(e) => {
                            e.currentTarget.style.visibility = "hidden";
                          }}
                        />
                      ) : (
                        <span className="cover-fallback">
                          {item.work?.title || "Book"}
                        </span>
                      )}
                      <span className="cover-open">
                        <ArrowUpRight size={18} />
                      </span>
                    </button>
                    <div className="book-card-copy">
                      <h2 className="public-book-title">
                        <button
                          onClick={() =>
                            item.work && onInspectWork?.(item.work)
                          }
                        >
                          {item.work?.title || "Untitled book"}
                        </button>
                      </h2>
                      <p className="book-author">
                        {item.work?.authors?.map((a) => a.name).join(", ") ||
                          "Author not listed"}
                      </p>
                      <p className="public-book-status">
                        {STATUS_LABELS[item.status as ReadingStatus]}
                        {item.rating ? ` · ${item.rating}/5` : ""}
                      </p>
                    </div>
                  </article>
                ))}
              </div>
            )}
            <ShareModal
              isOpen={sharing}
              onClose={() => setSharing(false)}
              username={profile.username}
              displayName={profile.display_name}
              shelf={shelf || undefined}
            />
          </>
        ) : null}
      </main>
    </div>
  );
}
