import { useMemo, useState } from "react";
import {
  Grid2X2,
  List,
  Search,
  Plus,
  X,
  ArrowRight,
  Upload,
  Download,
  SlidersHorizontal,
} from "lucide-react";
import { WishlistCard, STATUS_LABELS } from "./WishlistCard";
import { GoodreadsImportModal } from "./GoodreadsImportModal";
import { useAuth } from "../context/AuthContext";
import { api } from "../lib/api";
import { toast } from "../lib/toast";
import { RequestError } from "./RequestState";
import type { WishlistItem, ReadingStatus } from "../types/api";
interface CollectionViewProps {
  items: WishlistItem[];
  isLoading: boolean;
  error?: boolean;
  onUpdateStatus: (id: string, status: ReadingStatus) => void;
  onUpdatePriority: (id: string, priority: number) => void;
  onUpdateRating: (id: string, rating: number) => void;
  onUpdateNotes: (id: string, notes: string) => void;
  onUpdateTags?: (id: string, tags: string[]) => void;
  onDelete: (id: string) => void;
  onRefreshCollection?: () => void;
  onGoToDiscover: () => void;
  onInspectEditions?: (item: WishlistItem) => void;
}
export function CollectionView({
  items,
  isLoading,
  error,
  onRefreshCollection,
  onGoToDiscover,
  ...actions
}: CollectionViewProps) {
  const {
    isAuthenticated,
    isLoading: isAuthLoading,
    openAuthModal,
    user,
  } = useAuth();
  const [filter, setFilter] = useState<ReadingStatus | "ALL">("ALL");
  const [shelf, setShelf] = useState<string | null>(null);
  const [query, setQuery] = useState("");
  const [sort, setSort] = useState("newest");
  const [view, setView] = useState("grid");
  const [mobileFilters, setMobileFilters] = useState(false);
  const [importing, setImporting] = useState(false);
  const [newShelf, setNewShelf] = useState("");
  const [creating, setCreating] = useState(false);
  const shelfKey = `imprint_shelves_${user?.id || "guest"}`;
  const [localShelves, setLocalShelves] = useState<{
    owner: string;
    names: string[];
  }>(() => ({ owner: shelfKey, names: readShelves(shelfKey) }));
  if (localShelves.owner !== shelfKey)
    setLocalShelves({ owner: shelfKey, names: readShelves(shelfKey) });
  const tags = useMemo(
    () =>
      Array.from(
        new Set([...localShelves.names, ...items.flatMap((i) => i.tags)]),
      ).sort((a, b) => a.localeCompare(b)),
    [items, localShelves],
  );
  const shown = useMemo(
    () =>
      items
        .filter(
          (i) =>
            (filter === "ALL" || i.status === filter) &&
            (!shelf ||
              i.tags.some((t) => t.toLowerCase() === shelf.toLowerCase())) &&
            `${i.work?.title} ${i.work?.authors?.map((a) => a.name).join(" ")}`
              .toLowerCase()
              .includes(query.toLowerCase()),
        )
        .sort((a, b) =>
          sort === "title"
            ? (a.work?.title || "").localeCompare(b.work?.title || "")
            : sort === "priority"
              ? b.priority - a.priority
              : Date.parse(b.created_at) - Date.parse(a.created_at),
        ),
    [items, filter, shelf, query, sort],
  );
  const exportLibrary = async () => {
    try {
      await api.exportCollection("csv");
    } catch {
      toast.error(
        "Export failed",
        "Try again when your connection is restored.",
      );
    }
  };
  const createShelf = () => {
    const name = newShelf.trim();
    if (!name) return;
    const names = Array.from(new Set([...localShelves.names, name]));
    setLocalShelves({ owner: shelfKey, names });
    try {
      localStorage.setItem(shelfKey, JSON.stringify(names));
    } catch {
      toast.info("Shelf is available for this session");
    }
    setShelf(name);
    setNewShelf("");
    setCreating(false);
  };
  if (isAuthLoading)
    return (
      <div
        className="skeleton"
        style={{ height: 250 }}
        aria-label="Loading collection"
      />
    );
  if (!isAuthenticated)
    return (
      <section className="collection-welcome">
        <div>
          <h1>
            Your reading life,
            <br />
            all in one place.
          </h1>
          <p>
            Build a collection that’s yours. Save books, find the right edition,
            and keep track of what comes next.
          </p>
          <div className="empty-actions">
            <button
              className="button primary"
              onClick={() => openAuthModal("register")}
            >
              Create your collection <ArrowRight size={17} />
            </button>
            <button
              className="button secondary"
              onClick={() => openAuthModal("login")}
            >
              Sign in
            </button>
          </div>
        </div>
        <div className="reading-shelves-preview">
          <h2>A shelf for every chapter</h2>
          {["Want to read", "Currently reading", "Finished", "Put aside"].map(
            (label, i) => (
              <div key={label}>
                <span>{label}</span>
                <span className="preview-spines" aria-hidden="true">
                  {Array.from({ length: 4 - i }, (_, j) => (
                    <i key={j} />
                  ))}
                </span>
              </div>
            ),
          )}
          <p>Already have a Goodreads library? Bring your CSV when you join.</p>
        </div>
      </section>
    );
  return (
    <section className="collection-page">
      <div className="page-heading">
        <div>
          <h1>My collection</h1>
          <p>
            {items.length} {items.length === 1 ? "book" : "books"}. A little
            more you with every read.
          </p>
        </div>
        <div className="empty-actions">
          <button
            className="button secondary compact"
            onClick={() => setImporting(true)}
          >
            <Upload size={15} />
            Import
          </button>
          <button
            className="button secondary compact"
            onClick={() => void exportLibrary()}
          >
            <Download size={15} />
            Export
          </button>
        </div>
      </div>
      {error ? (
        <RequestError
          title="Your collection couldn’t load."
          onRetry={() => onRefreshCollection?.()}
        />
      ) : (
        <div className="collection-layout">
          <aside
            className={`collection-sidebar ${mobileFilters ? "mobile-open" : ""}`}
            aria-label="Collection filters"
          >
            <div className="sidebar-heading">
              <h2>Your shelves</h2>
              <button
                className="icon-button mobile-only"
                aria-label="Close filters"
                onClick={() => setMobileFilters(false)}
              >
                <X size={18} />
              </button>
            </div>
            <button
              className={`shelf-filter ${filter === "ALL" ? "selected" : ""}`}
              onClick={() => setFilter("ALL")}
              aria-pressed={filter === "ALL"}
            >
              <span>All books</span>
              <span>{items.length}</span>
            </button>
            {Object.entries(STATUS_LABELS).map(([status, label]) => (
              <button
                className={`shelf-filter ${filter === status ? "selected" : ""}`}
                key={status}
                aria-pressed={filter === status}
                onClick={() => setFilter(status as ReadingStatus)}
              >
                <span>{label}</span>
                <span>{items.filter((i) => i.status === status).length}</span>
              </button>
            ))}
            <div className="sidebar-heading custom-shelf-heading">
              <h2>Custom shelves</h2>
              <button
                className="icon-button"
                aria-label="Create shelf"
                onClick={() => setCreating(!creating)}
              >
                <Plus size={16} />
              </button>
            </div>
            <button
              className={`shelf-filter ${!shelf ? "selected" : ""}`}
              aria-pressed={!shelf}
              onClick={() => setShelf(null)}
            >
              All shelves
            </button>
            {tags.map((t) => (
              <button
                className={`shelf-filter ${shelf === t ? "selected" : ""}`}
                key={t}
                aria-pressed={shelf === t}
                onClick={() => setShelf(t)}
              >
                <span>{t}</span>
                <span>{items.filter((i) => i.tags.includes(t)).length}</span>
              </button>
            ))}
            {creating && (
              <form
                className="new-shelf-form"
                onSubmit={(e) => {
                  e.preventDefault();
                  createShelf();
                }}
              >
                <input
                  autoFocus
                  aria-label="New shelf name"
                  value={newShelf}
                  onChange={(e) => setNewShelf(e.target.value)}
                  placeholder="Shelf name"
                  maxLength={60}
                />
                <button
                  className="button primary compact"
                  disabled={!newShelf.trim()}
                >
                  Create
                </button>
                <p>
                  Empty shelves stay on this device. Add a book to sync its
                  shelf tag.
                </p>
              </form>
            )}
          </aside>
          <div className="collection-content">
            <div className="collection-toolbar">
              <label className="collection-search">
                <Search size={17} />
                <input
                  type="search"
                  aria-label="Search your collection"
                  placeholder="Find a book in your collection…"
                  value={query}
                  onChange={(e) => setQuery(e.target.value)}
                />
              </label>
              <button
                className="icon-button mobile-only"
                aria-label="Show shelf filters"
                onClick={() => setMobileFilters(!mobileFilters)}
              >
                <SlidersHorizontal size={18} />
              </button>
              <select
                aria-label="Sort collection"
                value={sort}
                onChange={(e) => setSort(e.target.value)}
              >
                <option value="newest">Recently added</option>
                <option value="title">Title, A–Z</option>
                <option value="priority">Priority</option>
              </select>
              <div className="view-switch">
                <button
                  className="icon-button"
                  aria-label="Grid view"
                  aria-pressed={view === "grid"}
                  onClick={() => setView("grid")}
                >
                  <Grid2X2 size={16} />
                </button>
                <button
                  className="icon-button"
                  aria-label="List view"
                  aria-pressed={view === "list"}
                  onClick={() => setView("list")}
                >
                  <List size={17} />
                </button>
              </div>
            </div>
            <div className="collection-scope">
              <span>
                {shelf ||
                  (filter === "ALL" ? "All books" : STATUS_LABELS[filter])}
              </span>
              <span>
                {shown.length} {shown.length === 1 ? "book" : "books"}
              </span>
            </div>
            {isLoading ? (
              <div className="collection-cards">
                {[0, 1, 2, 3].map((i) => (
                  <div className="skeleton" style={{ height: 180 }} key={i} />
                ))}
              </div>
            ) : !shown.length ? (
              <div className="empty-state">
                <h2>
                  {items.length
                    ? "No books on this shelf yet."
                    : "Your next chapter starts here."}
                </h2>
                <p>
                  {items.length
                    ? "Try another filter, or open a book’s Manage menu to add it to this shelf."
                    : "Find something you love, or bring over your Goodreads collection."}
                </p>
                <div className="empty-actions">
                  <button className="button primary" onClick={onGoToDiscover}>
                    Discover books <ArrowRight size={16} />
                  </button>
                  {items.length > 0 && (
                    <button
                      className="button secondary"
                      onClick={() => {
                        setFilter("ALL");
                        setShelf(null);
                        setQuery("");
                      }}
                    >
                      Clear filters
                    </button>
                  )}
                </div>
              </div>
            ) : (
              <div
                className={`collection-cards ${view === "list" ? "list-view" : ""}`}
              >
                {shown.map((item) => (
                  <WishlistCard key={item.id} item={item} {...actions} />
                ))}
              </div>
            )}
          </div>
        </div>
      )}
      <GoodreadsImportModal
        isOpen={importing}
        onClose={() => setImporting(false)}
        onSuccess={() => onRefreshCollection?.()}
      />
    </section>
  );
}
function readShelves(key: string): string[] {
  try {
    const value = JSON.parse(localStorage.getItem(key) || "[]");
    return Array.isArray(value)
      ? value.filter((s: unknown): s is string => typeof s === "string")
      : [];
  } catch {
    return [];
  }
}
