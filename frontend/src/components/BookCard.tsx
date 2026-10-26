import { useState } from "react";
import { BookmarkPlus, Check, ArrowUpRight, BookOpen } from "lucide-react";
import type { Work } from "../types/api";
interface BookCardProps {
  work: Work;
  isInCollection?: boolean;
  onToggleCollection: (work: Work) => void;
  onInspectEditions: (work: Work) => void;
}
export function BookCard({
  work,
  isInCollection = false,
  onToggleCollection,
  onInspectEditions,
}: BookCardProps) {
  const [failedCover, setFailedCover] = useState<string>();
  const author =
    work.authors?.map((a) => a.name).join(", ") || "Unknown Author";
  return (
    <article className="book-card">
      <button
        className="book-cover"
        onClick={() => onInspectEditions(work)}
        aria-label={`View ${work.title} and editions`}
      >
        {work.cover_url && failedCover !== work.cover_url ? (
          <img
            src={work.cover_url}
            alt={`Cover of ${work.title}`}
            loading="lazy"
            decoding="async"
            onError={() => setFailedCover(work.cover_url)}
          />
        ) : (
          <span className="cover-fallback">
            <BookOpen size={25} />
            <strong>{work.title}</strong>
            <span>Cover Unavailable</span>
          </span>
        )}
        <span className="cover-open">
          <ArrowUpRight size={18} />
        </span>
      </button>
      <div className="book-card-copy">
        <h3>
          <button onClick={() => onInspectEditions(work)}>{work.title}</button>
        </h3>
        <p className="book-author">{author}</p>
        {work.original_year && (
          <p className="book-year">First published {work.original_year}</p>
        )}
        <button
          className={`book-save ${isInCollection ? "is-saved" : ""}`}
          aria-label={
            isInCollection ? "In your collection" : "Add to Want to Read"
          }
          onClick={() => onToggleCollection(work)}
        >
          {isInCollection ? <Check size={16} /> : <BookmarkPlus size={16} />}
          {isInCollection ? "Saved" : "Want to read"}
        </button>
      </div>
    </article>
  );
}
