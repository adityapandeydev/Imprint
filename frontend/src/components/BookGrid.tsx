import { useMemo, useState } from "react";
import { BookCard } from "./BookCard";
import type { Work } from "../types/api";
interface BookGridProps {
  works: Work[];
  isLoading: boolean;
  collectionWorkIds: Set<string>;
  onToggleCollection: (work: Work) => void;
  onInspectEditions: (work: Work) => void;
  searchQuery: string;
}
export function BookGrid({
  works,
  isLoading,
  collectionWorkIds,
  onToggleCollection,
  onInspectEditions,
  searchQuery,
}: BookGridProps) {
  const [sort, setSort] = useState("relevance");
  const sorted = useMemo(
    () =>
      [...works].sort((a, b) =>
        sort === "title"
          ? a.title.localeCompare(b.title)
          : sort === "year"
            ? (b.original_year || 0) - (a.original_year || 0)
            : 0,
      ),
    [works, sort],
  );
  if (isLoading)
    return (
      <div className="book-grid" aria-busy="true" aria-label="Loading books">
        {Array.from({ length: 10 }, (_, i) => (
          <div key={i}>
            <div className="skeleton cover" />
            <div className="skeleton line" />
            <div className="skeleton line" />
          </div>
        ))}
      </div>
    );
  if (!works.length)
    return (
      <div className="empty-state">
        <h2>
          {searchQuery
            ? "A different search might help."
            : "Find your next read."}
        </h2>
        <p>
          No books matched “{searchQuery}”. Try a title, author’s name, or ISBN.
        </p>
      </div>
    );
  return (
    <div>
      <div className="results-toolbar">
        <p role="status">
          {works.length} {works.length === 1 ? "book" : "books"} for{" "}
          <strong>“{searchQuery}”</strong>
        </p>
        <label>
          Sort by{" "}
          <select
            value={sort}
            onChange={(e) => setSort(e.target.value)}
            aria-label="Sort search results"
          >
            <option value="relevance">Relevance</option>
            <option value="title">Title, A–Z</option>
            <option value="year">Publication year</option>
          </select>
        </label>
      </div>
      <div className="book-grid">
        {sorted.map((work) => (
          <BookCard
            key={
              work.id ||
              work.open_library_work_id ||
              `${work.title}-${work.authors?.[0]?.name}`
            }
            work={work}
            isInCollection={Boolean(
              (work.id && collectionWorkIds.has(work.id)) ||
              (work.open_library_work_id &&
                collectionWorkIds.has(work.open_library_work_id)),
            )}
            onToggleCollection={onToggleCollection}
            onInspectEditions={onInspectEditions}
          />
        ))}
      </div>
    </div>
  );
}
