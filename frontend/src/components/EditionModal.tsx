import React, { useEffect, useState, useMemo } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import {
  X,
  Calendar,
  Layers,
  Copy,
  Check,
  BookmarkPlus,
  Loader2,
  FileText,
  Globe,
  Tag,
} from 'lucide-react';
import { useQuery } from '@tanstack/react-query';
import { api } from '../lib/api';
import { toast } from '../lib/toast';
import type { Work, Edition, WishlistItem } from '../types/api';

interface EditionModalProps {
  work: Work | null;
  isOpen: boolean;
  onClose: () => void;
  onAddEditionToWishlist: (work: Work, edition?: Edition) => void;
  isWorkInCollection?: boolean;
  wishlistItem?: WishlistItem | null;
  onSelectWishlistEdition?: (wishlistItemId: string, edition: Edition) => void;
}

export const EditionModal: React.FC<EditionModalProps> = ({
  work,
  isOpen,
  onClose,
  onAddEditionToWishlist,
  isWorkInCollection = false,
  wishlistItem,
  onSelectWishlistEdition,
}) => {
  const [copiedISBN, setCopiedISBN] = useState<string | null>(null);

  // Close on Escape key
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === 'Escape') onClose();
    };
    if (isOpen) {
      window.addEventListener('keydown', handleKeyDown);
      document.body.style.overflow = 'hidden';
    }
    return () => {
      window.removeEventListener('keydown', handleKeyDown);
      document.body.style.overflow = '';
    };
  }, [isOpen, onClose]);

  // Query full work and editions
  const targetId = work?.id || work?.open_library_work_id;
  const { data, isLoading } = useQuery({
    queryKey: ['book', targetId],
    queryFn: () => (targetId ? api.getBook(targetId) : Promise.reject('No ID')),
    enabled: isOpen && Boolean(targetId),
    staleTime: 1000 * 60 * 10,
  });

  // Deduplicate and merge editions by ISBN/identifier to eliminate duplicate cards
  const editions = useMemo(() => {
    const raw = data?.editions || [];
    if (raw.length <= 1) return raw;

    const seen = new Map<string, number>();
    const unique: Edition[] = [];

    for (const ed of raw) {
      const key =
        ed.isbn13?.trim()
          ? `isbn13:${ed.isbn13.trim()}`
          : ed.isbn10?.trim()
          ? `isbn10:${ed.isbn10.trim()}`
          : ed.asin?.trim()
          ? `asin:${ed.asin.trim()}`
          : ed.open_library_edition_id?.trim()
          ? `ol:${ed.open_library_edition_id.trim()}`
          : ed.id
          ? `id:${ed.id}`
          : `title:${(ed.title || '').toLowerCase()}|${(ed.publisher || '').toLowerCase()}`;

      if (seen.has(key)) {
        const existingIdx = seen.get(key)!;
        const existing = unique[existingIdx];
        // Merge attributes to retain highest quality metadata
        unique[existingIdx] = {
          ...existing,
          format:
            existing.format === 'UNKNOWN' && ed.format && ed.format !== 'UNKNOWN'
              ? ed.format
              : existing.format,
          page_count: existing.page_count ?? ed.page_count,
          publisher: existing.publisher || ed.publisher,
          publication_year: existing.publication_year ?? ed.publication_year,
          publication_date: existing.publication_date || ed.publication_date,
          cover_url: existing.cover_url || ed.cover_url,
          language: existing.language || ed.language,
          isbn10: existing.isbn10 || ed.isbn10,
          isbn13: existing.isbn13 || ed.isbn13,
          asin: existing.asin || ed.asin,
        };
      } else {
        seen.set(key, unique.length);
        unique.push({ ...ed });
      }
    }

    return unique;
  }, [data?.editions]);

  const fullWork = data?.work || work;

  const handleCopy = (text: string, label: string) => {
    navigator.clipboard.writeText(text);
    setCopiedISBN(text);
    toast.info(`Copied ${label}`, text);
    setTimeout(() => setCopiedISBN(null), 2000);
  };

  const authorName =
    fullWork?.authors && fullWork.authors.length > 0
      ? fullWork.authors.map((a) => a.name).join(', ')
      : 'Unknown Author';

  if (!isOpen || !work) return null;

  return (
    <AnimatePresence>
      <div className="fixed inset-0 z-50 flex items-center justify-center p-3 sm:p-6 overflow-y-auto">
        {/* Backdrop */}
        <motion.div
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          exit={{ opacity: 0 }}
          onClick={onClose}
          className="fixed inset-0 bg-black/60 backdrop-blur-sm"
        />

        {/* Modal Window */}
        <motion.div
          initial={{ opacity: 0, scale: 0.96, y: 16 }}
          animate={{ opacity: 1, scale: 1, y: 0 }}
          exit={{ opacity: 0, scale: 0.96, y: 16 }}
          transition={{ type: 'spring', stiffness: 450, damping: 32 }}
          className="relative w-full max-w-3xl bg-surface border border-border-subtle rounded-3xl shadow-2xl overflow-hidden z-10 my-auto flex flex-col max-h-[90vh]"
        >
          {/* Header */}
          <div className="flex items-start justify-between p-6 border-b border-border-subtle bg-surface/80 backdrop-blur-md sticky top-0 z-20">
            <div className="space-y-1 pr-6">
              <span className="inline-flex items-center gap-1.5 px-2.5 py-0.5 rounded-md text-xs font-semibold bg-accent-soft text-accent border border-accent/20">
                <Layers className="w-3.5 h-3.5" />
                {wishlistItem ? 'Choose Edition for Your Collection' : 'Work & Published Editions'}
              </span>
              <h2 className="font-serif font-bold text-2xl text-text-main leading-snug">
                {fullWork?.title}
              </h2>
              <p className="text-xs text-text-muted">By {authorName}</p>
            </div>

            <button
              onClick={onClose}
              className="w-9 h-9 rounded-xl border border-border-subtle hover:bg-surface-hover flex items-center justify-center text-text-muted hover:text-text-main transition-colors shrink-0 cursor-pointer"
              aria-label="Close modal"
            >
              <X className="w-5 h-5" />
            </button>
          </div>

          {/* Scrollable Content Body */}
          <div className="p-6 overflow-y-auto space-y-6">
            {/* Top Work Overview */}
            <div className="flex flex-col sm:flex-row gap-6 items-start">
              {fullWork?.cover_url && (
                <div className="w-28 sm:w-36 aspect-[2/3] rounded-xl overflow-hidden border border-border-subtle shadow-md shrink-0 bg-surface-hover">
                  <img
                    src={fullWork.cover_url}
                    alt={fullWork.title}
                    className="w-full h-full object-cover"
                  />
                </div>
              )}

              <div className="space-y-3 flex-1">
                <div className="flex flex-wrap items-center gap-3 text-xs text-text-muted">
                  {fullWork?.original_year && (
                    <span className="flex items-center gap-1">
                      <Calendar className="w-3.5 h-3.5 text-accent" /> First published {fullWork.original_year}
                    </span>
                  )}
                  {fullWork?.open_library_work_id && (
                    <span className="font-mono text-[11px] bg-surface-hover px-2 py-0.5 rounded border border-border-subtle">
                      OL ID: {fullWork.open_library_work_id}
                    </span>
                  )}
                </div>

                {/* Synopsis */}
                {fullWork?.description ? (
                  <p className="text-xs sm:text-sm text-text-main/90 leading-relaxed max-h-36 overflow-y-auto pr-1">
                    {fullWork.description}
                  </p>
                ) : (
                  <p className="text-xs text-text-muted italic">
                    No synopsis provided by the catalog for this work.
                  </p>
                )}

                {/* Subject Tags */}
                {fullWork?.subject_tags && fullWork.subject_tags.length > 0 && (
                  <div className="flex flex-wrap items-center gap-1.5 pt-1">
                    <Tag className="w-3 h-3 text-accent" />
                    {fullWork.subject_tags.slice(0, 5).map((tag) => (
                      <span
                        key={tag}
                        className="px-2 py-0.5 text-[11px] rounded-md bg-canvas text-text-muted border border-border-subtle"
                      >
                        {tag}
                      </span>
                    ))}
                  </div>
                )}

                {/* Quick Add Whole Work to Collection */}
                <div className="pt-2">
                  <button
                    onClick={() => onAddEditionToWishlist(fullWork!)}
                    className={`px-4 py-2 rounded-xl text-xs font-semibold transition-all flex items-center gap-2 cursor-pointer ${
                      isWorkInCollection
                        ? 'bg-emerald-600 text-white shadow-xs'
                        : 'bg-accent text-canvas hover:opacity-90 shadow-xs'
                    }`}
                  >
                    {isWorkInCollection ? (
                      <>
                        <Check className="w-4 h-4" /> In Your Collection
                      </>
                    ) : (
                      <>
                        <BookmarkPlus className="w-4 h-4" /> Want to Read
                      </>
                    )}
                  </button>
                </div>
              </div>
            </div>

            {/* Published Editions Section */}
            <div className="space-y-3 pt-2">
              <div className="flex items-center justify-between border-b border-border-subtle pb-2">
                <h3 className="font-serif font-bold text-lg text-text-main flex items-center gap-2">
                  <Layers className="w-4 h-4 text-accent" />
                  Published Editions
                  <span className="text-xs font-sans font-normal text-text-muted">
                    ({editions.length})
                  </span>
                </h3>
                <span className="text-[11px] text-text-muted">
                  Future-ready for retailer price comparison
                </span>
              </div>

              {isLoading ? (
                <div className="text-center py-8 space-y-2 text-text-muted">
                  <Loader2 className="w-6 h-6 animate-spin text-accent mx-auto" />
                  <p className="text-xs">Fetching editions from catalog...</p>
                </div>
              ) : editions.length === 0 ? (
                <div className="text-center py-8 bg-surface-hover/40 rounded-xl border border-border-subtle p-4 text-text-muted text-xs">
                  No individual ISBN editions cataloged yet for this work.
                </div>
              ) : (
                <div className="grid sm:grid-cols-2 gap-3 max-h-72 overflow-y-auto pr-1">
                  {editions.map((edition) => (
                    <div
                      key={edition.id || edition.isbn13 || edition.isbn10 || edition.title}
                      className="p-3.5 rounded-xl bg-canvas border border-border-subtle hover:border-accent/40 transition-colors space-y-2.5 flex flex-col justify-between"
                    >
                      <div className="space-y-1">
                        <div className="flex items-center justify-between gap-2">
                          <span className="text-[10px] uppercase font-bold tracking-wider px-2 py-0.5 rounded bg-accent-soft text-accent border border-accent/20">
                            {edition.format || 'PAPERBACK'}
                          </span>

                          {edition.publication_year && (
                            <span className="text-xs text-text-muted font-medium">
                              {edition.publication_year}
                            </span>
                          )}
                        </div>

                        <p className="font-semibold text-xs text-text-main line-clamp-1">
                          {edition.publisher || 'Unknown Publisher'}
                        </p>

                        <div className="flex items-center gap-3 text-[11px] text-text-muted">
                          {edition.page_count && (
                            <span className="flex items-center gap-1">
                              <FileText className="w-3 h-3" /> {edition.page_count} pages
                            </span>
                          )}
                          {edition.language && (
                            <span className="flex items-center gap-1">
                              <Globe className="w-3 h-3" /> {edition.language.toUpperCase()}
                            </span>
                          )}
                        </div>
                      </div>

                      {/* Identifiers & Action */}
                      <div className="pt-2 border-t border-border-subtle/60 flex items-center justify-between gap-2 text-xs">
                        <div className="flex items-center gap-1 font-mono text-[11px] text-text-muted">
                          {edition.isbn13 ? (
                            <button
                              onClick={() => handleCopy(edition.isbn13!, 'ISBN-13')}
                              className="hover:text-accent flex items-center gap-1 cursor-pointer bg-surface px-1.5 py-0.5 rounded border border-border-subtle"
                              title="Click to copy ISBN-13"
                            >
                              <span>ISBN: {edition.isbn13}</span>
                              {copiedISBN === edition.isbn13 ? (
                                <Check className="w-3 h-3 text-emerald-500" />
                              ) : (
                                <Copy className="w-3 h-3" />
                              )}
                            </button>
                          ) : edition.isbn10 ? (
                            <button
                              onClick={() => handleCopy(edition.isbn10!, 'ISBN-10')}
                              className="hover:text-accent flex items-center gap-1 cursor-pointer bg-surface px-1.5 py-0.5 rounded border border-border-subtle"
                              title="Click to copy ISBN-10"
                            >
                              <span>ISBN: {edition.isbn10}</span>
                              {copiedISBN === edition.isbn10 ? (
                                <Check className="w-3 h-3 text-emerald-500" />
                              ) : (
                                <Copy className="w-3 h-3" />
                              )}
                            </button>
                          ) : (
                            <span className="text-text-muted/60">No ISBN</span>
                          )}
                        </div>

                        {wishlistItem ? (
                          wishlistItem.edition_id === edition.id ||
                          (wishlistItem.edition?.isbn13 &&
                            edition.isbn13 &&
                            wishlistItem.edition.isbn13 === edition.isbn13) ? (
                            <span className="inline-flex items-center gap-1 text-emerald-500 font-semibold text-xs bg-emerald-500/10 border border-emerald-500/20 px-2.5 py-1 rounded-lg">
                              <Check className="w-3.5 h-3.5" /> Current Edition
                            </span>
                          ) : (
                            <button
                              onClick={() => onSelectWishlistEdition?.(wishlistItem.id, edition)}
                              className="px-3 py-1 rounded-lg bg-accent text-canvas hover:opacity-90 font-medium text-xs transition-all flex items-center gap-1 cursor-pointer"
                            >
                              <Check className="w-3.5 h-3.5" /> Set as My Edition
                            </button>
                          )
                        ) : (
                          <button
                            onClick={() => onAddEditionToWishlist(fullWork!, edition)}
                            className="px-2.5 py-1 rounded-lg bg-surface text-accent hover:bg-accent hover:text-canvas border border-accent/30 text-[11px] font-medium transition-colors cursor-pointer"
                          >
                            Save Edition
                          </button>
                        )}
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </div>
          </div>
        </motion.div>
      </div>
    </AnimatePresence>
  );
};
