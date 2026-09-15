import React from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import { BookOpen, SearchX, Sparkles } from 'lucide-react';
import { BookCard } from './BookCard';
import type { Work } from '../types/api';

interface BookGridProps {
  works: Work[];
  isLoading: boolean;
  collectionWorkIds: Set<string>;
  onToggleCollection: (work: Work) => void;
  onInspectEditions: (work: Work) => void;
  searchQuery: string;
}

export const BookGrid: React.FC<BookGridProps> = ({
  works,
  isLoading,
  collectionWorkIds,
  onToggleCollection,
  onInspectEditions,
  searchQuery,
}) => {
  // 1. Loading Skeletons
  if (isLoading) {
    return (
      <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 gap-4 sm:gap-6">
        {Array.from({ length: 10 }).map((_, i) => (
          <div
            key={i}
            className="bg-surface rounded-2xl border border-border-subtle p-3 space-y-3 animate-pulse"
          >
            <div className="aspect-[2/3] w-full bg-surface-hover rounded-xl" />
            <div className="space-y-2 pt-1">
              <div className="h-4 bg-surface-hover rounded w-3/4" />
              <div className="h-3 bg-surface-hover rounded w-1/2" />
            </div>
            <div className="h-8 bg-surface-hover rounded-lg w-full mt-2" />
          </div>
        ))}
      </div>
    );
  }

  // 2. Empty Results State
  if (searchQuery && works.length === 0) {
    return (
      <div className="text-center py-16 px-4 bg-surface/60 rounded-2xl border border-dashed border-border-subtle max-w-lg mx-auto space-y-3">
        <SearchX className="w-10 h-10 text-text-muted mx-auto" />
        <h3 className="font-serif font-semibold text-lg text-text-main">No books found</h3>
        <p className="text-xs text-text-muted leading-relaxed">
          We couldn’t find any matches for &ldquo;<span className="text-text-main font-medium">{searchQuery}</span>&rdquo;. Try searching for the author’s name, another title, or a clean 10/13-digit ISBN.
        </p>
      </div>
    );
  }

  // 3. Welcome / Blank Prompt State
  if (!searchQuery && works.length === 0) {
    return (
      <div className="text-center py-12 px-4 max-w-md mx-auto space-y-3 text-text-muted">
        <div className="w-12 h-12 rounded-2xl bg-accent-soft text-accent flex items-center justify-center mx-auto shadow-xs">
          <BookOpen className="w-6 h-6" />
        </div>
        <p className="text-sm font-medium text-text-main">Search to uncover literary works</p>
        <p className="text-xs text-text-muted">
          Type any title, author, or genre above to browse live Open Library records and inspect exact editions.
        </p>
      </div>
    );
  }

  // 4. Populated Grid with fluid spring layout reordering
  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between text-xs text-text-muted px-1">
        <span className="flex items-center gap-1.5 font-medium">
          <Sparkles className="w-3.5 h-3.5 text-amber-gold" />
          Found {works.length} {works.length === 1 ? 'work' : 'works'}
        </span>
        <span>Click any cover to inspect editions</span>
      </div>

      <motion.div
        layout
        className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 gap-4 sm:gap-6"
      >
        <AnimatePresence>
          {works.map((work) => {
            // Identify if this work is in the collection by its ID or OpenLibrary ID
            const isSaved =
              Boolean(work.id && collectionWorkIds.has(work.id)) ||
              Boolean(work.open_library_work_id && collectionWorkIds.has(work.open_library_work_id));

            return (
              <BookCard
                key={work.open_library_work_id || work.id || work.title}
                work={work}
                isInCollection={isSaved}
                onToggleCollection={onToggleCollection}
                onInspectEditions={onInspectEditions}
              />
            );
          })}
        </AnimatePresence>
      </motion.div>
    </div>
  );
};
