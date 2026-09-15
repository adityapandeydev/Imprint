import React, { useState, useMemo, useEffect } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import {
  BookMarked,
  ArrowUpDown,
  Compass,
  Lock,
  LogIn,
  UserPlus,
  ShieldCheck,
  Sparkles,
  BarChart3,
  Download,
  Upload,
  Tag,
  Plus,
  ChevronDown,
  Check,
} from 'lucide-react';
import { WishlistCard } from './WishlistCard';
import { ProfileStatsModal } from './ProfileStatsModal';
import { GoodreadsImportModal } from './GoodreadsImportModal';
import { useAuth } from '../context/AuthContext';
import { api } from '../lib/api';
import type { WishlistItem, ReadingStatus, TagCount } from '../types/api';

interface CollectionViewProps {
  items: WishlistItem[];
  isLoading: boolean;
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

type FilterOption = 'ALL' | ReadingStatus;
type SortOption = 'priority' | 'newest' | 'title';

const SORT_OPTIONS: { id: SortOption; label: string }[] = [
  { id: 'priority', label: 'Priority' },
  { id: 'newest', label: 'Recently Added' },
  { id: 'title', label: 'Title (A-Z)' },
];

export const CollectionView: React.FC<CollectionViewProps> = ({
  items,
  isLoading,
  onUpdateStatus,
  onUpdatePriority,
  onUpdateRating,
  onUpdateNotes,
  onUpdateTags,
  onDelete,
  onRefreshCollection,
  onGoToDiscover,
  onInspectEditions,
}) => {
  const { user, isAuthenticated, isLoading: isAuthLoading, openAuthModal } = useAuth();
  const [activeFilter, setActiveFilter] = useState<FilterOption>('ALL');
  const [selectedTag, setSelectedTag] = useState<string | null>(null);
  const [sortBy, setSortBy] = useState<SortOption>('priority');
  const [serverTags, setServerTags] = useState<TagCount[]>([]);
  const [isStatsOpen, setIsStatsOpen] = useState(false);
  const [isImportOpen, setIsImportOpen] = useState(false);
  const [isExportMenuOpen, setIsExportMenuOpen] = useState(false);
  const [isSortMenuOpen, setIsSortMenuOpen] = useState(false);
  const [isCreatingTag, setIsCreatingTag] = useState(false);
  const [newShelfName, setNewShelfName] = useState('');

  // Fetch user tags when collection changes or loads
  useEffect(() => {
    if (isAuthenticated) {
      api.getTags().then(setServerTags).catch(() => {});
    }
  }, [items, isAuthenticated]);

  // Aggregate active tags from both server and current memory state
  const allTags = useMemo(() => {
    const map = new Map<string, number>();
    // First from items
    for (const item of items) {
      for (const t of item.tags || []) {
        map.set(t, (map.get(t) || 0) + 1);
      }
    }
    // Then combine with server tags
    for (const st of serverTags) {
      if (!map.has(st.tag)) {
        map.set(st.tag, st.count);
      }
    }
    return Array.from(map.entries()).map(([tag, count]) => ({ tag, count }));
  }, [items, serverTags]);

  // Count items per filter tab (Unconditionally declared to follow React Rules of Hooks)
  const counts = useMemo(() => {
    return {
      ALL: items.length,
      WANT_TO_READ: items.filter((i) => i.status === 'WANT_TO_READ').length,
      CURRENTLY_READING: items.filter((i) => i.status === 'CURRENTLY_READING').length,
      FINISHED: items.filter((i) => i.status === 'FINISHED').length,
      ABANDONED: items.filter((i) => i.status === 'ABANDONED').length,
    };
  }, [items]);

  // Filtered and sorted items (Unconditionally declared to follow React Rules of Hooks)
  const displayItems = useMemo(() => {
    let list = items;

    // Filter by reading status tab
    if (activeFilter !== 'ALL') {
      list = list.filter((item) => item.status === activeFilter);
    }

    // Filter by custom shelf tag
    if (selectedTag) {
      list = list.filter((item) =>
        (item.tags || []).some((t) => t.toLowerCase() === selectedTag.toLowerCase())
      );
    }

    return [...list].sort((a, b) => {
      if (sortBy === 'priority') {
        const diff = (b.priority || 0) - (a.priority || 0);
        if (diff !== 0) return diff;
        const dateA = new Date(a.created_at).getTime();
        const dateB = new Date(b.created_at).getTime();
        return dateB - dateA;
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
  }, [items, activeFilter, selectedTag, sortBy]);

  const handleExport = async (format: 'csv' | 'json') => {
    setIsExportMenuOpen(false);
    try {
      await api.exportCollection(format);
    } catch (err) {
      alert(err instanceof Error ? err.message : 'Export failed');
    }
  };

  const handleCreateShelf = () => {
    const clean = newShelfName.trim();
    if (clean) {
      setSelectedTag(clean);
      setNewShelfName('');
      setIsCreatingTag(false);
    }
  };

  if (isAuthLoading) {
    return (
      <div className="grid grid-cols-1 md:grid-cols-2 gap-4 sm:gap-6 pt-4">
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
    );
  }

  if (!isAuthenticated) {
    return (
      <div className="max-w-xl mx-auto py-12 px-4 text-center space-y-6">
        <motion.div
          initial={{ opacity: 0, scale: 0.9 }}
          animate={{ opacity: 1, scale: 1 }}
          transition={{ duration: 0.3 }}
          className="w-16 h-16 mx-auto rounded-2xl bg-accent-soft border border-accent/20 flex items-center justify-center text-accent shadow-xs"
        >
          <Lock className="w-8 h-8 text-amber-gold" />
        </motion.div>
        <div className="space-y-2">
          <h2 className="font-serif text-3xl font-bold text-text-main">
            Private Reading Collection
          </h2>
          <p className="text-xs sm:text-sm text-text-muted leading-relaxed max-w-md mx-auto">
            Your personal reading shelf, custom statuses, star ratings, and private notes are protected behind authentication. Sign in to access your library.
          </p>
        </div>

        <div className="flex flex-col sm:flex-row items-center justify-center gap-3 pt-2">
          <button
            onClick={() => openAuthModal('login')}
            className="w-full sm:w-auto px-6 py-3 bg-accent text-canvas font-semibold rounded-xl text-xs sm:text-sm hover:opacity-90 transition-opacity shadow-xs flex items-center justify-center gap-2 cursor-pointer"
          >
            <LogIn className="w-4 h-4" />
            <span>Sign In to Access</span>
          </button>
          <button
            onClick={() => openAuthModal('register')}
            className="w-full sm:w-auto px-6 py-3 bg-surface hover:bg-surface-hover border border-border-subtle text-text-main font-semibold rounded-xl text-xs sm:text-sm transition-colors shadow-xs flex items-center justify-center gap-2 cursor-pointer"
          >
            <UserPlus className="w-4 h-4" />
            <span>Create Free Account</span>
          </button>
        </div>

        <div className="grid grid-cols-1 sm:grid-cols-3 gap-3 pt-6 text-left border-t border-border-subtle">
          <div className="p-3 bg-surface rounded-xl border border-border-subtle">
            <ShieldCheck className="w-4 h-4 text-accent mb-1" />
            <p className="text-xs font-semibold text-text-main">Strict Isolation</p>
            <p className="text-[11px] text-text-muted">Unique PostgreSQL collection per user</p>
          </div>
          <div className="p-3 bg-surface rounded-xl border border-border-subtle">
            <Sparkles className="w-4 h-4 text-amber-gold mb-1" />
            <p className="text-xs font-semibold text-text-main">Custom Shelves</p>
            <p className="text-[11px] text-text-muted">Track reading, finished, and priorities</p>
          </div>
          <div className="p-3 bg-surface rounded-xl border border-border-subtle">
            <Lock className="w-4 h-4 text-accent mb-1" />
            <p className="text-xs font-semibold text-text-main">Production Auth</p>
            <p className="text-[11px] text-text-muted">RFC 6819 Token Rotation + Rate Limiting</p>
          </div>
        </div>
      </div>
    );
  }

  const filterTabs: { id: FilterOption; label: string }[] = [
    { id: 'ALL', label: 'All Books' },
    { id: 'WANT_TO_READ', label: 'Want to Read' },
    { id: 'CURRENTLY_READING', label: 'Currently Reading' },
    { id: 'FINISHED', label: 'Finished' },
    { id: 'ABANDONED', label: 'Abandoned' },
  ];

  return (
    <div className="space-y-6 animate-fade-in">
      {/* Header & Main Toolbar */}
      <div className="flex flex-col md:flex-row md:items-center justify-between gap-4 border-b border-border-subtle pb-5">
        <div>
          <h1 className="font-serif text-3xl font-bold text-text-main flex items-center gap-2">
            My Collection
            <span className="text-sm font-sans font-normal text-text-muted bg-surface px-2.5 py-0.5 rounded-full border border-border-subtle">
              {items.length} {items.length === 1 ? 'book' : 'books'}
            </span>
          </h1>
          <p className="text-xs sm:text-sm text-text-muted mt-0.5">
            Organize your personal reading journey, priority list, and custom shelves.
          </p>
        </div>

        {/* Action Toolbar */}
        <div className="flex flex-wrap items-center gap-2.5 self-start md:self-auto">
          {/* Reading Stats Button */}
          <button
            onClick={() => setIsStatsOpen(true)}
            className="px-3 py-1.5 bg-surface hover:bg-surface-hover border border-border-subtle text-text-main text-xs font-semibold rounded-xl transition-colors flex items-center gap-1.5 shadow-xs cursor-pointer"
            title="View reading velocity and analytics"
          >
            <BarChart3 className="w-3.5 h-3.5 text-amber-500" />
            <span>Reading Stats</span>
          </button>

          {/* Import Goodreads Button */}
          <button
            onClick={() => setIsImportOpen(true)}
            className="px-3 py-1.5 bg-surface hover:bg-surface-hover border border-border-subtle text-text-main text-xs font-semibold rounded-xl transition-colors flex items-center gap-1.5 shadow-xs cursor-pointer"
            title="Import library from Goodreads CSV"
          >
            <Upload className="w-3.5 h-3.5 text-emerald-400" />
            <span>Import CSV</span>
          </button>

          {/* Export Dropdown */}
          <div className="relative">
            <button
              onClick={() => setIsExportMenuOpen(!isExportMenuOpen)}
              className="px-3 py-1.5 bg-surface hover:bg-surface-hover border border-border-subtle text-text-main text-xs font-semibold rounded-xl transition-colors flex items-center gap-1.5 shadow-xs cursor-pointer"
              title="Export library data"
            >
              <Download className="w-3.5 h-3.5 text-cyan-400" />
              <span>Export</span>
            </button>

            {isExportMenuOpen && (
              <>
                <div
                  className="fixed inset-0 z-30"
                  onClick={() => setIsExportMenuOpen(false)}
                />
                <div className="absolute right-0 mt-1.5 w-40 bg-surface border border-border-subtle rounded-xl shadow-xl z-40 p-1 space-y-0.5">
                  <button
                    onClick={() => handleExport('csv')}
                    className="w-full text-left px-3 py-1.5 rounded-lg text-xs font-medium text-text-main hover:bg-surface-hover flex items-center justify-between cursor-pointer"
                  >
                    <span>Download CSV</span>
                    <span className="text-[10px] text-text-muted">.csv</span>
                  </button>
                  <button
                    onClick={() => handleExport('json')}
                    className="w-full text-left px-3 py-1.5 rounded-lg text-xs font-medium text-text-main hover:bg-surface-hover flex items-center justify-between cursor-pointer"
                  >
                    <span>Download JSON</span>
                    <span className="text-[10px] text-text-muted">.json</span>
                  </button>
                </div>
              </>
            )}
          </div>
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

      {/* Shelves & Sorting Line */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3 pt-1">
        {/* Left: Custom Shelves / Tags */}
        <div className="flex flex-wrap items-center gap-2 text-xs">
          <span className="text-text-muted flex items-center gap-1 font-semibold text-[11px]">
            <Tag className="w-3 h-3 text-accent" /> Shelves:
          </span>

          {/* All Shelves Tag */}
          <button
            onClick={() => setSelectedTag(null)}
            className={`px-3 py-1 rounded-full text-xs font-semibold border transition-all cursor-pointer ${
              selectedTag === null
                ? 'bg-accent text-canvas border-accent'
                : 'bg-surface border-border-subtle text-text-muted hover:text-text-main'
            }`}
          >
            All Shelves
          </button>

          {/* Individual Tags */}
          {allTags.map((t) => {
            const isSelected = selectedTag?.toLowerCase() === t.tag.toLowerCase();
            return (
              <button
                key={t.tag}
                onClick={() => setSelectedTag(isSelected ? null : t.tag)}
                className={`px-3 py-1 rounded-full text-xs font-semibold border transition-all cursor-pointer flex items-center gap-1.5 ${
                  isSelected
                    ? 'bg-accent text-canvas border-accent'
                    : 'bg-surface border-border-subtle text-text-muted hover:text-text-main hover:border-accent/40'
                }`}
              >
                <span>#{t.tag}</span>
                <span
                  className={`text-[10px] px-1 rounded-full ${
                    isSelected
                      ? 'bg-canvas/20 text-canvas'
                      : 'bg-surface-hover text-text-muted'
                  }`}
                >
                  {t.count}
                </span>
              </button>
            );
          })}

          {/* + New Shelf Action */}
          {isCreatingTag ? (
            <div className="flex items-center gap-1 bg-surface p-0.5 rounded-full border border-border-subtle">
              <input
                type="text"
                value={newShelfName}
                onChange={(e) => setNewShelfName(e.target.value)}
                onKeyDown={(e) => {
                  if (e.key === 'Enter') handleCreateShelf();
                  if (e.key === 'Escape') setIsCreatingTag(false);
                }}
                placeholder="Shelf name..."
                className="px-2.5 py-0.5 bg-transparent text-text-main text-xs outline-none w-28"
                autoFocus
              />
              <button
                onClick={handleCreateShelf}
                className="px-2 py-0.5 rounded-full bg-accent text-canvas text-xs font-bold hover:opacity-90 cursor-pointer"
              >
                Filter
              </button>
              <button
                onClick={() => setIsCreatingTag(false)}
                className="text-text-muted hover:text-text-main px-1 text-xs cursor-pointer"
              >
                ×
              </button>
            </div>
          ) : (
            <button
              onClick={() => setIsCreatingTag(true)}
              className="inline-flex items-center gap-1 text-[11px] text-accent font-semibold px-2.5 py-1 rounded-full bg-accent-soft/40 hover:bg-accent-soft border border-accent/20 transition-colors cursor-pointer"
            >
              <Plus className="w-3 h-3" />
              <span>New Shelf</span>
            </button>
          )}
        </div>

        {/* Right: Custom Sort Dropdown */}
        <div className="relative self-end sm:self-auto shrink-0">
          <button
            onClick={() => setIsSortMenuOpen(!isSortMenuOpen)}
            className="px-3 py-1.5 bg-surface hover:bg-surface-hover border border-border-subtle text-text-main text-xs font-semibold rounded-xl transition-colors flex items-center gap-1.5 shadow-xs cursor-pointer"
            title="Sort collection"
          >
            <ArrowUpDown className="w-3.5 h-3.5 text-accent" />
            <span className="text-text-muted font-normal">Sort:</span>
            <span>{SORT_OPTIONS.find((s) => s.id === sortBy)?.label || 'Priority'}</span>
            <ChevronDown className="w-3 h-3 text-text-muted ml-0.5" />
          </button>

          {isSortMenuOpen && (
            <>
              <div
                className="fixed inset-0 z-30"
                onClick={() => setIsSortMenuOpen(false)}
              />
              <div className="absolute right-0 mt-1.5 w-44 bg-surface border border-border-subtle rounded-xl shadow-xl z-40 p-1 space-y-0.5">
                {SORT_OPTIONS.map((opt) => (
                  <button
                    key={opt.id}
                    onClick={() => {
                      setSortBy(opt.id);
                      setIsSortMenuOpen(false);
                    }}
                    className={`w-full text-left px-3 py-1.5 rounded-lg text-xs font-medium transition-colors flex items-center justify-between cursor-pointer ${
                      sortBy === opt.id
                        ? 'bg-accent-soft text-accent font-semibold'
                        : 'text-text-main hover:bg-surface-hover'
                    }`}
                  >
                    <span>{opt.label}</span>
                    {sortBy === opt.id && <Check className="w-3.5 h-3.5 text-accent" />}
                  </button>
                ))}
              </div>
            </>
          )}
        </div>
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
              {selectedTag
                ? `No books tagged with #${selectedTag}`
                : activeFilter === 'ALL'
                ? 'Your reading shelf is empty'
                : `No books in "${filterTabs.find((t) => t.id === activeFilter)?.label}"`}
            </h2>
            <p className="text-xs sm:text-sm text-text-muted max-w-md mx-auto leading-relaxed">
              {selectedTag
                ? 'Tag books in your collection with this shelf name to organize them here.'
                : activeFilter === 'ALL'
                ? 'Discover books in the catalog or import your Goodreads CSV to populate your personal collection.'
                : 'Move books between reading states using the status menu on any saved title.'}
            </p>
          </div>

          <div className="flex items-center justify-center gap-3 pt-2">
            <button
              onClick={onGoToDiscover}
              className="px-5 py-2.5 bg-accent text-canvas rounded-xl text-xs sm:text-sm font-semibold hover:opacity-90 transition-opacity shadow-xs cursor-pointer inline-flex items-center gap-2"
            >
              <Compass className="w-4 h-4" />
              <span>Discover Books</span>
            </button>
            <button
              onClick={() => setIsImportOpen(true)}
              className="px-5 py-2.5 bg-surface hover:bg-surface-hover border border-border-subtle text-text-main rounded-xl text-xs sm:text-sm font-semibold transition-colors shadow-xs cursor-pointer inline-flex items-center gap-2"
            >
              <Upload className="w-4 h-4 text-emerald-400" />
              <span>Import Goodreads</span>
            </button>
          </div>
        </div>
      ) : (
        /* Populated Shelf Grid */
        <motion.div layout className="grid grid-cols-1 md:grid-cols-2 gap-4 sm:gap-6">
          <AnimatePresence>
            {displayItems.map((item) => (
              <WishlistCard
                key={item.id}
                item={item}
                onUpdateStatus={onUpdateStatus}
                onUpdatePriority={onUpdatePriority}
                onUpdateRating={onUpdateRating}
                onUpdateNotes={onUpdateNotes}
                onUpdateTags={onUpdateTags}
                onDelete={onDelete}
                onInspectEditions={onInspectEditions}
              />
            ))}
          </AnimatePresence>
        </motion.div>
      )}

      {/* Profile & Reading Velocity Analytics Modal */}
      <ProfileStatsModal
        isOpen={isStatsOpen}
        onClose={() => setIsStatsOpen(false)}
        userName={user?.display_name || user?.username || 'Reader'}
      />

      {/* Goodreads 1-Click CSV Migration Modal */}
      <GoodreadsImportModal
        isOpen={isImportOpen}
        onClose={() => setIsImportOpen(false)}
        onSuccess={() => {
          if (onRefreshCollection) {
            onRefreshCollection();
          }
        }}
      />
    </div>
  );
};
