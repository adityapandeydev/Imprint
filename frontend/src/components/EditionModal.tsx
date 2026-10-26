import { useMemo, useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { ArrowLeft, BookmarkPlus, Check, Copy, BookOpen } from "lucide-react";
import { api } from "../lib/api";
import { toast } from "../lib/toast";
import { RequestError } from "./RequestState";
import type { Work, Edition, WishlistItem } from "../types/api";
interface Props {
  work: Work | null;
  isOpen: boolean;
  onClose: () => void;
  onAddEditionToWishlist: (work: Work, edition?: Edition) => void;
  isWorkInCollection?: boolean;
  wishlistItem?: WishlistItem | null;
  onSelectWishlistEdition?: (wishlistItemId: string, edition: Edition) => void;
}
export function EditionModal({
  work,
  isOpen,
  onClose,
  onAddEditionToWishlist,
  isWorkInCollection = false,
  wishlistItem,
  onSelectWishlistEdition,
}: Props) {
  const [format, setFormat] = useState("ALL");
  const [language, setLanguage] = useState("ALL");
  const [copied, setCopied] = useState("");
  const [coverError, setCoverError] = useState("");
  const targetId = work?.id || work?.open_library_work_id;
  const { data, isLoading, isError, refetch } = useQuery({
    queryKey: ["book", targetId],
    queryFn: () => api.getBook(targetId!),
    enabled: isOpen && !!targetId,
    staleTime: 600000,
  });
  const editions = useMemo(() => {
    const byId = new Map<string, Edition>();
    for (const edition of data?.editions || []) {
      const key =
        edition.isbn13 ||
        edition.isbn10 ||
        edition.asin ||
        edition.id ||
        `${edition.title}:${edition.publisher}:${edition.format}`;
      if (!byId.has(key)) byId.set(key, edition);
    }
    return Array.from(byId.values());
  }, [data?.editions]);
  const formats = Array.from(new Set(editions.map((e) => e.format)));
  const languages = Array.from(
    new Set(editions.map((e) => e.language).filter((v): v is string => !!v)),
  );
  const shown = editions.filter(
    (e) =>
      (format === "ALL" || e.format === format) &&
      (language === "ALL" || e.language === language),
  );
  if (!isOpen || !work) return null;
  const book = data?.work || work;
  const copy = async (isbn: string) => {
    try {
      await navigator.clipboard.writeText(isbn);
      setCopied(isbn);
      setTimeout(() => setCopied(""), 2000);
    } catch {
      toast.error("ISBN couldn’t be copied");
    }
  };
  const formatName = (value: string) =>
    value === "UNKNOWN"
      ? "Format not listed"
      : value.replaceAll("_", " ").toLowerCase();
  return (
    <section className="book-details">
      <button className="back-link" onClick={onClose}>
        <ArrowLeft size={17} />
        Back to books
      </button>
      <div className="book-overview">
        <div className="detail-cover">
          {book.cover_url && coverError !== book.cover_url ? (
            <img
              src={book.cover_url}
              alt={`Cover of ${book.title}`}
              onError={() => setCoverError(book.cover_url || "")}
            />
          ) : (
            <div className="cover-fallback">
              <BookOpen size={32} />
              <strong>{book.title}</strong>
              <span>Cover unavailable</span>
            </div>
          )}
        </div>
        <div className="detail-copy">
          <h1>{book.title}</h1>
          <p className="detail-author">
            {book.authors?.map((a) => a.name).join(", ") || "Author not listed"}
          </p>
          {book.original_year && (
            <p className="detail-year">First published {book.original_year}</p>
          )}
          <button
            className="button primary"
            disabled={isWorkInCollection}
            onClick={() => onAddEditionToWishlist(book)}
          >
            {isWorkInCollection ? (
              <Check size={17} />
            ) : (
              <BookmarkPlus size={17} />
            )}{" "}
            {isWorkInCollection ? "In your collection" : "Want to read"}
          </button>
          <div className="synopsis">
            <h2>About this book</h2>
            {isLoading ? (
              <p>Loading the book’s story…</p>
            ) : (
              <p>
                {book.description ||
                  "There isn’t a synopsis in the catalog for this book yet."}
              </p>
            )}
          </div>
          {!!book.subject_tags?.length && (
            <div className="subject-tags">
              {book.subject_tags.slice(0, 6).map((t) => (
                <span key={t}>{t}</span>
              ))}
            </div>
          )}
        </div>
      </div>
      <section className="editions-section">
        <div className="section-heading">
          <div>
            <h2>Find your edition</h2>
            <p>Compare formats, publishers, languages, and ISBNs.</p>
          </div>
          <span>{editions.length} editions</span>
        </div>
        {isError ? (
          <RequestError
            title="The book’s details couldn’t load."
            onRetry={() => void refetch()}
          />
        ) : isLoading ? (
          <div
            className="skeleton"
            style={{ height: 150 }}
            aria-label="Loading editions"
          />
        ) : (
          <>
            {!!editions.length && (
              <div className="edition-filters">
                <label>
                  Format
                  <select
                    value={format}
                    onChange={(e) => setFormat(e.target.value)}
                  >
                    <option value="ALL">All formats</option>
                    {formats.map((f) => (
                      <option value={f} key={f}>
                        {formatName(f)}
                      </option>
                    ))}
                  </select>
                </label>
                <label>
                  Language
                  <select
                    value={language}
                    onChange={(e) => setLanguage(e.target.value)}
                  >
                    <option value="ALL">All languages</option>
                    {languages.map((l) => (
                      <option value={l} key={l}>
                        {languageName(l)}
                      </option>
                    ))}
                  </select>
                </label>
                <span>
                  {shown.length} matching{" "}
                  {shown.length === 1 ? "edition" : "editions"}
                </span>
              </div>
            )}
            {!shown.length ? (
              <div className="empty-state">
                <h2>
                  {editions.length
                    ? "No matching editions."
                    : "No editions listed yet."}
                </h2>
                <p>
                  {editions.length
                    ? "Try another format or language."
                    : "You can still save this book to your collection."}
                </p>
                {!!editions.length && (
                  <button
                    className="button secondary"
                    onClick={() => {
                      setFormat("ALL");
                      setLanguage("ALL");
                    }}
                  >
                    Clear filters
                  </button>
                )}
              </div>
            ) : (
              <div className="edition-list">
                {shown.map((e) => {
                  const isbn = e.isbn13 || e.isbn10;
                  const selected =
                    !!wishlistItem &&
                    (wishlistItem.edition_id === e.id ||
                      !!(
                        isbn &&
                        (wishlistItem.edition?.isbn13 === isbn ||
                          wishlistItem.edition?.isbn10 === isbn)
                      ));
                  return (
                    <article
                      className="edition-row"
                      key={e.id || isbn || `${e.publisher}-${e.title}`}
                    >
                      <div className="edition-format">
                        {formatName(e.format)}
                        <span>
                          {e.publication_year ||
                            e.publication_date ||
                            "Date not listed"}
                        </span>
                      </div>
                      <div className="edition-description">
                        <h3>{e.publisher || "Publisher not listed"}</h3>
                        <p>
                          {[
                            e.page_count ? `${e.page_count} pages` : null,
                            e.language ? languageName(e.language) : null,
                          ]
                            .filter(Boolean)
                            .join(" · ") || "Details not listed"}
                        </p>
                        {isbn ? (
                          <button
                            className="isbn-copy"
                            onClick={() => void copy(isbn)}
                            aria-label={`Copy ISBN ${isbn}`}
                          >
                            ISBN {isbn}
                            {copied === isbn ? (
                              <Check size={13} />
                            ) : (
                              <Copy size={13} />
                            )}
                          </button>
                        ) : (
                          <span className="muted">ISBN not listed</span>
                        )}
                      </div>
                      <div className="edition-action">
                        {selected ? (
                          <span className="selected-edition">
                            <Check size={16} />
                            Your edition
                          </span>
                        ) : wishlistItem ? (
                          <button
                            className="button secondary compact"
                            disabled={!e.id}
                            onClick={() =>
                              onSelectWishlistEdition?.(wishlistItem.id, e)
                            }
                          >
                            Choose edition
                          </button>
                        ) : (
                          <button
                            disabled={isWorkInCollection || !e.id}
                            className="button secondary compact"
                            onClick={() => onAddEditionToWishlist(book, e)}
                          >
                            {isWorkInCollection
                              ? "Already saved"
                              : "Save edition"}
                          </button>
                        )}
                      </div>
                    </article>
                  );
                })}
              </div>
            )}
          </>
        )}
      </section>
    </section>
  );
}
function languageName(code: string) {
  const names: Record<string, string> = {
    eng: "English",
    en: "English",
    hin: "Hindi",
    hi: "Hindi",
    fre: "French",
    fr: "French",
    spa: "Spanish",
    es: "Spanish",
    ger: "German",
    de: "German",
  };
  return names[code.toLowerCase()] || code;
}
