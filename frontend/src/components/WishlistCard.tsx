import React, { useState } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import {
  BookOpen,
  Trash2,
  Star,
  Flame,
  ChevronDown,
  MessageSquare,
  Check,
  Layers,
} from 'lucide-react';
import type { WishlistItem, ReadingStatus } from '../types/api';

interface WishlistCardProps {
  item: WishlistItem;
  onUpdateStatus: (id: string, status: ReadingStatus) => void;
  onUpdatePriority: (id: string, priority: number) => void;
  onUpdateRating: (id: string, rating: number) => void;
  onUpdateNotes: (id: string, notes: string) => void;
  onDelete: (id: string) => void;
  onInspectEditions?: (item: WishlistItem) => void;
}

const STATUS_LABELS: Record<ReadingStatus, { label: string; color: string }> = {
  WANT_TO_READ: { label: 'Want to Read', color: 'text-amber-500 bg-amber-500/10 border-amber-500/20' },
  CURRENTLY_READING: { label: 'Reading', color: 'text-sky-500 bg-sky-500/10 border-sky-500/20' },
  FINISHED: { label: 'Finished', color: 'text-emerald-500 bg-emerald-500/10 border-emerald-500/20' },
  ABANDONED: { label: 'Abandoned', color: 'text-zinc-400 bg-zinc-400/10 border-zinc-400/20' },
};

export const WishlistCard: React.FC<WishlistCardProps> = ({
  item,
  onUpdateStatus,
  onUpdatePriority,
  onUpdateRating,
  onUpdateNotes,
  onDelete,
  onInspectEditions,
}) => {
  const [isNotesOpen, setIsNotesOpen] = useState(false);
  const [noteText, setNoteText] = useState(item.notes || '');
  const [isStatusMenuOpen, setIsStatusMenuOpen] = useState(false);
  const [confirmDelete, setConfirmDelete] = useState(false);

  const work = item.work;
  const edition = item.edition;
  const authorName =
    work?.authors && work.authors.length > 0
      ? work.authors.map((a) => a.name).join(', ')
      : 'Unknown Author';

  const handleSaveNotes = () => {
    onUpdateNotes(item.id, noteText);
    setIsNotesOpen(false);
  };

  return (
    <motion.article
      layout
      initial={{ opacity: 0, y: 15 }}
      animate={{ opacity: 1, y: 0 }}
      exit={{ opacity: 0, scale: 0.95 }}
      transition={{ duration: 0.2 }}
      className="bg-surface rounded-2xl border border-border-subtle hover:border-accent/30 p-4 sm:p-5 shadow-xs hover:shadow-book transition-all space-y-4"
    >
      <div className="flex gap-4 sm:gap-5 items-start">
        {/* Cover Thumbnail */}
        <div
          onClick={() => onInspectEditions?.(item)}
          className="w-16 sm:w-20 aspect-[2/3] rounded-xl overflow-hidden bg-surface-hover shrink-0 border border-border-subtle shadow-xs cursor-pointer group relative"
        >
          {work?.cover_url ? (
            <img
              src={work.cover_url}
              alt={work.title}
              className="w-full h-full object-cover group-hover:scale-105 transition-transform"
            />
          ) : (
            <div className="w-full h-full p-2 flex flex-col justify-between bg-accent-soft/40 text-text-main text-[9px]">
              <span className="font-bold line-clamp-2">{work?.title || 'Book'}</span>
              <BookOpen className="w-3.5 h-3.5 text-accent" />
            </div>
          )}
        </div>

        {/* Info Column */}
        <div className="flex-1 min-w-0 space-y-2">
          <div className="flex items-start justify-between gap-2">
            <div>
              <h3
                onClick={() => onInspectEditions?.(item)}
                className="font-serif font-bold text-base sm:text-lg text-text-main line-clamp-1 leading-snug hover:text-accent cursor-pointer transition-colors"
                title={work?.title}
              >
                {work?.title || 'Untitled Book'}
              </h3>
              <p className="text-xs text-text-muted line-clamp-1">{authorName}</p>
            </div>

            {/* Status Dropdown */}
            <div className="relative">
              <button
                onClick={() => setIsStatusMenuOpen(!isStatusMenuOpen)}
                className={`px-2.5 py-1 rounded-lg text-xs font-semibold border transition-all flex items-center gap-1.5 cursor-pointer ${
                  STATUS_LABELS[item.status].color
                }`}
              >
                <span>{STATUS_LABELS[item.status].label}</span>
                <ChevronDown className="w-3 h-3" />
              </button>

              {isStatusMenuOpen && (
                <>
                  <div
                    className="fixed inset-0 z-30"
                    onClick={() => setIsStatusMenuOpen(false)}
                  />
                  <div className="absolute right-0 mt-1.5 w-44 bg-surface border border-border-subtle rounded-xl shadow-xl z-40 p-1 space-y-0.5">
                    {(
                      [
                        'WANT_TO_READ',
                        'CURRENTLY_READING',
                        'FINISHED',
                        'ABANDONED',
                      ] as ReadingStatus[]
                    ).map((st) => (
                      <button
                        key={st}
                        onClick={() => {
                          onUpdateStatus(item.id, st);
                          setIsStatusMenuOpen(false);
                        }}
                        className={`w-full text-left px-3 py-1.5 rounded-lg text-xs font-medium flex items-center justify-between cursor-pointer ${
                          item.status === st
                            ? 'bg-accent-soft text-accent font-semibold'
                            : 'text-text-main hover:bg-surface-hover'
                        }`}
                      >
                        <span>{STATUS_LABELS[st].label}</span>
                        {item.status === st && <Check className="w-3 h-3 text-accent" />}
                      </button>
                    ))}
                  </div>
                </>
              )}
            </div>
          </div>

          {/* Edition / Format Tag & Selection Button */}
          {edition ? (
            <div className="flex items-center gap-2 text-[11px] text-text-muted flex-wrap">
              <span className="inline-flex items-center gap-1.5 bg-canvas px-2.5 py-1 rounded-lg border border-border-subtle font-medium">
                <Layers className="w-3.5 h-3.5 text-accent" />
                <span className="uppercase font-bold text-text-main text-[10px] bg-accent-soft text-accent px-1.5 py-0.5 rounded">
                  {edition.format || 'EDITION'}
                </span>
                {edition.publisher && <span className="text-text-main">{edition.publisher}</span>}
                {edition.publication_year && <span>({edition.publication_year})</span>}
              </span>
              <button
                onClick={() => onInspectEditions?.(item)}
                className="text-accent hover:underline text-xs font-semibold cursor-pointer"
              >
                Change Edition
              </button>
            </div>
          ) : (
            <div className="flex items-center gap-2 pt-0.5">
              <button
                onClick={() => onInspectEditions?.(item)}
                className="inline-flex items-center gap-1.5 text-xs text-accent bg-accent-soft hover:bg-accent-soft/80 border border-accent/20 px-3 py-1 rounded-lg font-medium transition-colors cursor-pointer"
              >
                <Layers className="w-3.5 h-3.5" />
                <span>+ Select Edition (Paperback, Hardcover, etc.)</span>
              </button>
            </div>
          )}

          {/* Priority & Star Ratings Controls */}
          <div className="flex flex-wrap items-center gap-4 pt-1 text-xs">
            {/* Priority Selector (1-5) */}
            <div className="flex items-center gap-1.5 text-text-muted">
              <span className="text-[11px] flex items-center gap-1">
                <Flame className="w-3 h-3 text-amber-500" /> Priority:
              </span>
              <div className="flex items-center gap-0.5">
                {[1, 2, 3, 4, 5].map((p) => (
                  <button
                    key={p}
                    onClick={() => onUpdatePriority(item.id, p)}
                    className={`w-4 h-4 rounded-full flex items-center justify-center text-[10px] font-bold transition-transform cursor-pointer ${
                      p <= item.priority
                        ? 'bg-amber-500 text-canvas scale-105'
                        : 'bg-surface-hover text-text-muted hover:bg-amber-500/30'
                    }`}
                    title={`Priority ${p} of 5`}
                  >
                    {p}
                  </button>
                ))}
              </div>
            </div>

            {/* Rating Stars (shown when finished) */}
            {item.status === 'FINISHED' && (
              <div className="flex items-center gap-1 text-text-muted">
                <span className="text-[11px]">Rating:</span>
                <div className="flex items-center gap-0.5">
                  {[1, 2, 3, 4, 5].map((r) => (
                    <button
                      key={r}
                      onClick={() => onUpdateRating(item.id, r)}
                      className="cursor-pointer transition-transform hover:scale-110"
                      title={`${r} Stars`}
                    >
                      <Star
                        className={`w-3.5 h-3.5 ${
                          r <= (item.rating || 0)
                            ? 'text-amber-gold fill-amber-gold'
                            : 'text-border-subtle'
                        }`}
                      />
                    </button>
                  ))}
                </div>
              </div>
            )}
          </div>
        </div>
      </div>

      {/* Card Footer: Notes Button & Remove Action */}
      <div className="pt-2 border-t border-border-subtle/70 flex items-center justify-between gap-2 text-xs">
        <button
          onClick={() => setIsNotesOpen(!isNotesOpen)}
          className={`px-2.5 py-1 rounded-lg border transition-colors flex items-center gap-1.5 cursor-pointer ${
            item.notes
              ? 'bg-accent-soft text-accent border-accent/30 font-medium'
              : 'bg-surface border-border-subtle text-text-muted hover:text-text-main'
          }`}
        >
          <MessageSquare className="w-3.5 h-3.5" />
          <span>{item.notes ? 'View Notes' : 'Add Notes'}</span>
        </button>

        {confirmDelete ? (
          <div className="flex items-center gap-2">
            <span className="text-text-muted text-[11px]">Remove book?</span>
            <button
              onClick={() => onDelete(item.id)}
              className="px-2 py-0.5 rounded bg-rose-500 text-white font-semibold text-[11px] hover:bg-rose-600 transition-colors cursor-pointer"
            >
              Yes, Remove
            </button>
            <button
              onClick={() => setConfirmDelete(false)}
              className="text-text-muted text-[11px] hover:underline cursor-pointer"
            >
              Cancel
            </button>
          </div>
        ) : (
          <button
            onClick={() => setConfirmDelete(true)}
            className="p-1.5 text-text-muted hover:text-rose-500 hover:bg-rose-500/10 rounded-lg transition-colors cursor-pointer"
            title="Remove from collection"
            aria-label="Remove book from collection"
          >
            <Trash2 className="w-3.5 h-3.5" />
          </button>
        )}
      </div>

      {/* Expandable Notes Drawer */}
      <AnimatePresence>
        {isNotesOpen && (
          <motion.div
            initial={{ opacity: 0, height: 0 }}
            animate={{ opacity: 1, height: 'auto' }}
            exit={{ opacity: 0, height: 0 }}
            className="overflow-hidden space-y-2 pt-2 border-t border-border-subtle"
          >
            <textarea
              value={noteText}
              onChange={(e) => setNoteText(e.target.value)}
              placeholder="Record your thoughts, quotes, or reading goals..."
              rows={3}
              className="w-full p-3 bg-canvas text-text-main placeholder:text-text-muted rounded-xl border border-border-subtle focus:border-accent text-xs outline-none resize-none transition-colors"
            />
            <div className="flex items-center justify-end gap-2">
              <button
                onClick={() => setIsNotesOpen(false)}
                className="px-3 py-1 rounded-lg text-xs text-text-muted hover:text-text-main cursor-pointer"
              >
                Close
              </button>
              <button
                onClick={handleSaveNotes}
                className="px-3 py-1 rounded-lg bg-accent text-canvas text-xs font-semibold hover:opacity-90 transition-opacity cursor-pointer flex items-center gap-1"
              >
                <Check className="w-3.5 h-3.5" /> Save Note
              </button>
            </div>
          </motion.div>
        )}
      </AnimatePresence>
    </motion.article>
  );
};
