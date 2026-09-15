import React, { useState } from 'react';
import { motion } from 'framer-motion';
import { Book, BookmarkPlus, Check, Layers, Calendar, User } from 'lucide-react';
import type { Work } from '../types/api';

interface BookCardProps {
  work: Work;
  isInCollection?: boolean;
  onToggleCollection: (work: Work) => void;
  onInspectEditions: (work: Work) => void;
}

export const BookCard: React.FC<BookCardProps> = ({
  work,
  isInCollection = false,
  onToggleCollection,
  onInspectEditions,
}) => {
  const [imageError, setImageError] = useState(false);

  const authorName =
    work.authors && work.authors.length > 0
      ? work.authors.map((a) => a.name).join(', ')
      : 'Unknown Author';

  return (
    <motion.article
      whileHover={{ y: -4 }}
      transition={{ duration: 0.2 }}
      className="group flex flex-col h-full bg-surface rounded-2xl border border-border-subtle hover:border-accent/40 shadow-xs hover:shadow-book transition-all overflow-hidden"
    >
      {/* Book Cover Container */}
      <div
        className="relative aspect-[2/3] w-full bg-surface-hover/70 cursor-pointer overflow-hidden flex items-center justify-center"
        onClick={() => onInspectEditions(work)}
      >
        {work.cover_url && !imageError ? (
          <img
            src={work.cover_url}
            alt={`Cover of ${work.title}`}
            onError={() => setImageError(true)}
            loading="lazy"
            className="w-full h-full object-cover group-hover:scale-105 transition-transform duration-300"
          />
        ) : (
          /* Stylized Literary Fallback Spine */
          <div className="w-full h-full p-4 flex flex-col justify-between bg-gradient-to-br from-surface to-accent-soft/30 border-b border-border-subtle text-text-main select-none">
            <div className="space-y-1">
              <span className="text-[10px] font-mono tracking-widest text-text-muted/70 uppercase">
                Cover Unavailable
              </span>
              <p className="font-serif font-bold text-sm leading-snug line-clamp-3 text-text-main">
                {work.title}
              </p>
            </div>
            <div className="flex items-center gap-1.5 text-xs text-text-muted">
              <Book className="w-4 h-4 text-accent/80" />
              <span className="truncate">{authorName}</span>
            </div>
          </div>
        )}

        {/* Floating Quick Action Button on Cover */}
        <div className="absolute top-2.5 right-2.5 z-10">
          <button
            onClick={(e) => {
              e.stopPropagation();
              onToggleCollection(work);
            }}
            aria-label={isInCollection ? 'In your collection' : 'Add to Want to Read'}
            className={`w-9 h-9 rounded-xl flex items-center justify-center transition-all shadow-sm backdrop-blur-md cursor-pointer ${
              isInCollection
                ? 'bg-emerald-600 text-white shadow-emerald-900/30'
                : 'bg-canvas/90 text-text-muted hover:text-accent hover:bg-surface border border-border-subtle'
            }`}
            title={isInCollection ? 'In Collection (click to manage)' : 'Add to Want to Read'}
          >
            {isInCollection ? <Check className="w-4 h-4" /> : <BookmarkPlus className="w-4 h-4" />}
          </button>
        </div>

        {/* Edition Preview Overlay Hint */}
        <div className="absolute inset-x-0 bottom-0 bg-gradient-to-t from-black/60 via-black/20 to-transparent p-3 pt-6 opacity-0 group-hover:opacity-100 transition-opacity flex items-center justify-center text-white text-xs font-medium gap-1.5">
          <Layers className="w-3.5 h-3.5" /> Inspect Editions
        </div>
      </div>

      {/* Book Metadata Content */}
      <div className="flex flex-col flex-1 p-4 space-y-3 justify-between">
        <div className="space-y-1.5">
          <h3
            onClick={() => onInspectEditions(work)}
            className="font-serif font-bold text-base text-text-main line-clamp-2 leading-snug group-hover:text-accent transition-colors cursor-pointer"
            title={work.title}
          >
            {work.title}
          </h3>

          <p className="text-xs text-text-muted flex items-center gap-1 line-clamp-1">
            <User className="w-3 h-3 text-accent shrink-0" />
            <span className="truncate">{authorName}</span>
          </p>

          {work.original_year && (
            <p className="text-[11px] text-text-muted flex items-center gap-1">
              <Calendar className="w-3 h-3 shrink-0" />
              <span>First published {work.original_year}</span>
            </p>
          )}
        </div>

        {/* Bottom Actions */}
        <div className="pt-2 border-t border-border-subtle/70 flex items-center justify-between gap-2">
          <button
            onClick={() => onInspectEditions(work)}
            className="text-xs text-accent hover:underline flex items-center gap-1 font-medium cursor-pointer"
          >
            <Layers className="w-3.5 h-3.5" />
            <span>Editions</span>
          </button>

          <button
            onClick={() => onToggleCollection(work)}
            className={`px-3 py-1.5 rounded-lg text-xs font-medium transition-all flex items-center gap-1.5 cursor-pointer ${
              isInCollection
                ? 'bg-emerald-500/15 text-emerald-600 dark:text-emerald-400 font-semibold border border-emerald-500/30'
                : 'bg-accent text-canvas hover:opacity-90 shadow-xs'
            }`}
          >
            {isInCollection ? (
              <>
                <Check className="w-3.5 h-3.5" />
                <span>Saved</span>
              </>
            ) : (
              <>
                <BookmarkPlus className="w-3.5 h-3.5" />
                <span>Want to Read</span>
              </>
            )}
          </button>
        </div>
      </div>
    </motion.article>
  );
};
