import React, { useState, useMemo } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import {
  BookMarked,
  ArrowUpDown,
  Compass,
} from 'lucide-react';
import { WishlistCard } from './WishlistCard';
import type { WishlistItem, ReadingStatus } from '../types/api';

interface CollectionViewProps {
  items: WishlistItem[];
  isLoading: boolean;
  onUpdateStatus: (id: string, status: ReadingStatus) => void;
  onUpdatePriority: (id: string, priority: number) => void;
  onUpdateRating: (id: string, rating: number) => void;
  onUpdateNotes: (id: string, notes: string) => void;
  onDelete: (id: string) => void;
  onGoToDiscover: () => void;
  onInspectEditions?: (item: WishlistItem) => void;
}

type FilterOption = 'ALL' | ReadingStatus;
type SortOption = 'priority' | 'newest' | 'title';

export const CollectionView: React.FC<CollectionViewProps> = ({
  items,
  isLoading,
  onUpdateStatus,
  onUpdatePriority,
  onUpdateRating,
  onUpdateNotes,
  onDelete,
  onGoToDiscover,
  onInspectEditions,
}) => {
  const [activeFilter, setActiveFilter] = useState<FilterOption>('ALL');
  const [sortBy, setSortBy] = useState<SortOption>('priority');

  // Count items per filter tab
  const counts = useMemo(() => {
    return {
      ALL: items.length,
      WANT_TO_READ: items.filter((i) => i.status === 'WANT_TO_READ').length,
      CURRENTLY_READING: items.filter((i) => i.status === 'CURRENTLY_READING').length,
      FINISHED: items.filter((i) => i.status === 'FINISHED').length,
      ABANDONED: items.filter((i) => i.status === 'ABANDONED').length,
    };
  }, [items]);

  // Filtered and sorted items
  const displayItems = useMemo(() => {
    let list = items;
    if (activeFilter !== 'ALL') {
      list = list.filter((item) => item.status === activeFilter);
    }

    return [...list].sort((a, b) => {
      if (sortBy === 'priority') {
        return (b.priority || 0) - (a.priority || 0);
      }
      if (sortBy === 'title') {
        const titleA = a.work?.title || '';
        const titleB = b.work?.title || '';
        return titleA.localeCompare(titleB);
      }
      // 'newest'
      const dateA = new Date(a.created_at).getTime();
      const dateB = new Date(b.created_at).getTime();
      return dateB - dateA;
    });
  }, [items, activeFilter, sortBy]);

  const filterTabs: { id: FilterOption; label: string }[] = [
    { id: 'ALL', label: 'All Books' },
    { id: 'WANT_TO_READ', label: 'Want to Read' },
    { id: 'CURRENTLY_READING', label: 'Currently Reading' },
    { id: 'FINISHED', label: 'Finished' },
    { id: 'ABANDONED', label: 'Abandoned' },
  ];

  return (
    <div className="space-y-6 animate-fade-in">
      {/* Header & Controls */}
      <div className="flex flex-col md:flex-row md:items-center justify-between gap-4 border-b border-border-subtle pb-5">
        <div>
          <h1 className="font-serif text-3xl font-bold text-text-main flex items-center gap-2">
            My Collection
            <span className="text-sm font-sans font-normal text-text-muted bg-surface px-2.5 py-0.5 rounded-full border border-border-subtle">
              {items.length} {items.length === 1 ? 'book' : 'books'}
            </span>
          </h1>
          <p className="text-xs sm:text-sm text-text-muted mt-0.5">
            Organize your personal reading journey, priority list, and private notes.
          </p>
        </div>

        {/* Sort Dropdown */}
        <div className="flex items-center gap-2 text-xs self-start md:self-auto">
          <span className="text-text-muted flex items-center gap-1 font-medium">
            <ArrowUpDown className="w-3.5 h-3.5 text-accent" /> Sort:
          </span>
          <select
            value={sortBy}
            onChange={(e) => setSortBy(e.target.value as SortOption)}
            className="bg-surface text-text-main border border-border-subtle rounded-xl px-3 py-1.5 outline-none focus:border-accent font-medium cursor-pointer shadow-xs"
          >
            <option value="priority">Highest Priority</option>
            <option value="newest">Recently Added</option>
            <option value="title">Title (A-Z)</option>
          </select>
        </div>
      </div>

      {/* Filter Tabs with sliding pill */}
      <div className="flex items-center gap-1 bg-surface/80 p-1.5 rounded-2xl border border-border-subtle shadow-xs overflow-x-auto">
        {filterTabs.map((tab) => {
          const isActive = activeFilter === tab.id;
          const count = counts[tab.id];

          return (
            <button
              key={tab.id}
              onClick={() => setActiveFilter(tab.id)}
              className={`relative flex items-center gap-2 px-3.5 sm:px-4 py-2 rounded-xl text-xs sm:text-sm font-medium transition-colors whitespace-nowrap z-10 select-none cursor-pointer ${
                isActive ? 'text-accent font-semibold' : 'text-text-muted hover:text-text-main'
              }`}
            >
              {isActive && (
                <motion.div
                  layoutId="collectionFilterPill"
                  className="absolute inset-0 bg-accent-soft border border-accent/20 rounded-xl shadow-xs -z-10 pointer-events-none"
                  transition={{ type: 'spring', stiffness: 450, damping: 35 }}
                />
              )}

              <span>{tab.label}</span>
              {count > 0 && (
                <span
                  className={`px-1.5 py-0.2 rounded-full text-[11px] font-bold ${
                    isActive
                      ? 'bg-accent text-canvas'
                      : 'bg-surface-hover text-text-muted'
                  }`}
                >
                  {count}
                </span>
              )}
            </button>
          );
        })}
      </div>

      {/* Loading Skeletons */}
      {isLoading ? (
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4 sm:gap-6">
          {Array.from({ length: 4 }).map((_, i) => (
            <div
              key={i}
              className="bg-surface rounded-2xl border border-border-subtle p-5 animate-pulse flex gap-4"
            >
              <div className="w-16 sm:w-20 aspect-[2/3] bg-surface-hover rounded-xl shrink-0" />
              <div className="flex-1 space-y-3 pt-1">
                <div className="h-4 bg-surface-hover rounded w-3/4" />
                <div className="h-3 bg-surface-hover rounded w-1/2" />
                <div className="h-6 bg-surface-hover rounded-lg w-1/3 mt-3" />
              </div>
            </div>
          ))}
        </div>
      ) : displayItems.length === 0 ? (
        <div className="text-center py-16 bg-surface rounded-2xl border border-dashed border-border-subtle p-8 max-w-xl mx-auto space-y-4">
          <BookMarked className="w-12 h-12 text-accent/60 mx-auto" />
          <div className="space-y-1">
            <h2 className="font-serif font-semibold text-lg text-text-main">
              {activeFilter === 'ALL'
                ? 'Your reading shelf is empty'
                : `No books in "${filterTabs.find((t) => t.id === activeFilter)?.label}"`}
            </h2>
            <p className="text-xs sm:text-sm text-text-muted max-w-md mx-auto leading-relaxed">
              {activeFilter === 'ALL'
                ? 'Discover books in the catalog and click "Want to Read" to start curating your library.'
                : 'Move books between reading states using the status menu on any saved title.'}
            </p>
          </div>

          <button
            onClick={onGoToDiscover}
            className="px-5 py-2.5 bg-accent text-canvas rounded-xl text-xs sm:text-sm font-semibold hover:opacity-90 transition-opacity shadow-xs cursor-pointer inline-flex items-center gap-2"
          >
            <Compass className="w-4 h-4" />
            <span>Discover Books</span>
          </button>
        </div>
      ) : (
        /* Populated Shelf Grid */
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4 sm:gap-6">
          <AnimatePresence mode="popLayout">
            {displayItems.map((item) => (
              <WishlistCard
                key={item.id}
                item={item}
                onUpdateStatus={onUpdateStatus}
                onUpdatePriority={onUpdatePriority}
                onUpdateRating={onUpdateRating}
                onUpdateNotes={onUpdateNotes}
                onDelete={onDelete}
                onInspectEditions={onInspectEditions}
              />
            ))}
          </AnimatePresence>
        </div>
      )}
    </div>
  );
};
