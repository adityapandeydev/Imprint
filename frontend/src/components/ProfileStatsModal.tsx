import React, { useEffect, useState } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import {
  BarChart3,
  Tag,
  Feather,
  BookOpen,
  Star,
  X,
  RefreshCw,
  Clock,
} from 'lucide-react';
import { api } from '../lib/api';
import type { ReadingStats } from '../types/api';

interface ProfileStatsModalProps {
  isOpen: boolean;
  onClose: () => void;
  userName?: string;
}

export const ProfileStatsModal: React.FC<ProfileStatsModalProps> = ({
  isOpen,
  onClose,
  userName = 'Reader',
}) => {
  const [stats, setStats] = useState<ReadingStats | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const fetchStats = async () => {
    try {
      setLoading(true);
      setError(null);
      const data = await api.getReadingStats();
      setStats(data);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load reading metrics');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    if (isOpen) {
      fetchStats();
    }
  }, [isOpen]);

  if (!isOpen) return null;

  return (
    <AnimatePresence>
      <div className="fixed inset-0 z-50 flex items-center justify-center p-4 sm:p-6">
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
          initial={{ opacity: 0, scale: 0.95, y: 20 }}
          animate={{ opacity: 1, scale: 1, y: 0 }}
          exit={{ opacity: 0, scale: 0.95, y: 20 }}
          transition={{ type: 'spring', damping: 25, stiffness: 300 }}
          className="relative w-full max-w-3xl max-h-[90vh] overflow-y-auto bg-surface border border-border-subtle rounded-2xl shadow-2xl p-6 sm:p-8 text-text-main z-10 custom-scrollbar"
        >
          {/* Header */}
          <div className="flex items-start justify-between border-b border-border-subtle pb-5 mb-6">
            <div>
              <div className="flex items-center gap-3">
                <span className="p-2.5 rounded-xl bg-accent-soft text-accent border border-accent/20">
                  <BarChart3 className="w-6 h-6 text-accent" />
                </span>
                <div>
                  <h2 className="text-xl sm:text-2xl font-serif font-bold tracking-tight text-text-main">
                    Literary Velocity & Analytics
                  </h2>
                  <p className="text-xs sm:text-sm text-text-muted mt-0.5">
                    Reading profile and collection metrics for{' '}
                    <span className="text-accent font-semibold">{userName}</span>
                  </p>
                </div>
              </div>
            </div>
            <button
              onClick={onClose}
              className="text-text-muted hover:text-text-main p-2 rounded-lg hover:bg-surface-hover transition-colors cursor-pointer"
              aria-label="Close modal"
            >
              <X className="w-5 h-5" />
            </button>
          </div>

          {/* Body */}
          {loading ? (
            <div className="py-16 flex flex-col items-center justify-center gap-4 text-text-muted">
              <div className="w-10 h-10 border-3 border-accent/30 border-t-accent rounded-full animate-spin" />
              <p className="text-sm font-medium">Computing reading velocity & analytics...</p>
            </div>
          ) : error ? (
            <div className="py-12 text-center">
              <p className="text-red-500 text-sm mb-4">{error}</p>
              <button
                onClick={fetchStats}
                className="inline-flex items-center gap-2 px-4 py-2 bg-surface-hover hover:opacity-90 border border-border-subtle text-text-main text-xs font-semibold rounded-lg transition-colors cursor-pointer"
              >
                <RefreshCw className="w-3.5 h-3.5" />
                <span>Retry</span>
              </button>
            </div>
          ) : stats ? (
            <div className="space-y-6">
              {/* Velocity Metric Cards */}
              <div className="grid grid-cols-2 sm:grid-cols-4 gap-3.5">
                <div className="bg-surface-hover/60 border border-border-subtle rounded-xl p-4">
                  <p className="text-xs font-medium text-text-muted">Finished in {stats.current_year}</p>
                  <p className="text-2xl sm:text-3xl font-bold text-accent mt-1">
                    {stats.books_finished_year}
                  </p>
                  <p className="text-[11px] text-text-muted mt-1">Books completed</p>
                </div>

                <div className="bg-surface-hover/60 border border-border-subtle rounded-xl p-4">
                  <p className="text-xs font-medium text-text-muted">Pages Read</p>
                  <p className="text-2xl sm:text-3xl font-bold text-emerald-600 dark:text-emerald-400 mt-1">
                    {stats.total_pages_read.toLocaleString()}
                  </p>
                  <p className="text-[11px] text-text-muted mt-1">Total volume</p>
                </div>

                <div className="bg-surface-hover/60 border border-border-subtle rounded-xl p-4">
                  <p className="text-xs font-medium text-text-muted">In Progress</p>
                  <p className="text-2xl sm:text-3xl font-bold text-sky-600 dark:text-sky-400 mt-1">
                    {stats.currently_reading}
                  </p>
                  <p className="text-[11px] text-text-muted mt-1">Active reads</p>
                </div>

                <div className="bg-surface-hover/60 border border-border-subtle rounded-xl p-4">
                  <p className="text-xs font-medium text-text-muted">Average Rating</p>
                  <div className="flex items-center gap-1.5 mt-1">
                    {stats.average_rating > 0 ? (
                      <>
                        <Star className="w-5 h-5 fill-amber-400 text-amber-400" />
                        <span className="text-2xl sm:text-3xl font-bold text-text-main">
                          {stats.average_rating.toFixed(1)}
                        </span>
                      </>
                    ) : (
                      <span className="text-2xl sm:text-3xl font-bold text-text-muted">—</span>
                    )}
                  </div>
                  <p className="text-[11px] text-text-muted mt-1">Collection review</p>
                </div>
              </div>

              {/* Genres & Authors 2-Column Section */}
              <div className="grid grid-cols-1 sm:grid-cols-2 gap-6">
                {/* Top Literary Subjects / Genres */}
                <div className="bg-surface-hover/40 border border-border-subtle rounded-xl p-5">
                  <h3 className="text-sm font-semibold text-text-main mb-3.5 flex items-center gap-2">
                    <Tag className="w-4 h-4 text-accent" />
                    <span>Top Subject Affinities</span>
                  </h3>
                  {stats.top_genres.length === 0 ? (
                    <p className="text-xs text-text-muted py-6 text-center">No genres identified yet.</p>
                  ) : (
                    <div className="space-y-3">
                      {stats.top_genres.map((g) => {
                        const maxCount = stats.top_genres[0]?.count || 1;
                        const percent = Math.round((g.count / maxCount) * 100);
                        return (
                          <div key={g.genre}>
                            <div className="flex justify-between text-xs mb-1">
                              <span className="font-medium text-text-main truncate max-w-[180px]">
                                {g.genre}
                              </span>
                              <span className="text-text-muted font-semibold">
                                {g.count} {g.count === 1 ? 'book' : 'books'}
                              </span>
                            </div>
                            <div className="w-full bg-surface rounded-full h-2 overflow-hidden border border-border-subtle">
                              <motion.div
                                initial={{ width: 0 }}
                                animate={{ width: `${percent}%` }}
                                transition={{ duration: 0.6, ease: 'easeOut' }}
                                className="bg-gradient-to-r from-accent to-accent/70 h-2 rounded-full"
                              />
                            </div>
                          </div>
                        );
                      })}
                    </div>
                  )}
                </div>

                {/* Top Authors */}
                <div className="bg-surface-hover/40 border border-border-subtle rounded-xl p-5">
                  <h3 className="text-sm font-semibold text-text-main mb-3.5 flex items-center gap-2">
                    <Feather className="w-4 h-4 text-accent" />
                    <span>Most Read Authors</span>
                  </h3>
                  {stats.top_authors.length === 0 ? (
                    <p className="text-xs text-text-muted py-6 text-center">No author affiliations tracked yet.</p>
                  ) : (
                    <div className="space-y-2.5">
                      {stats.top_authors.map((a, idx) => (
                        <div
                          key={a.author}
                          className="flex items-center justify-between p-2.5 rounded-lg bg-surface border border-border-subtle"
                        >
                          <div className="flex items-center gap-2.5">
                            <span className="w-5 h-5 flex items-center justify-center rounded-full bg-surface-hover border border-border-subtle text-[10px] font-bold text-text-muted">
                              {idx + 1}
                            </span>
                            <span className="text-xs font-medium text-text-main">{a.author}</span>
                          </div>
                          <span className="text-xs text-accent font-semibold bg-accent-soft px-2 py-0.5 rounded-full border border-accent/20">
                            {a.count} {a.count === 1 ? 'title' : 'titles'}
                          </span>
                        </div>
                      ))}
                    </div>
                  )}
                </div>
              </div>

              {/* Format Breakdown & Library Summary */}
              <div className="bg-surface-hover/30 border border-border-subtle rounded-xl p-4 flex flex-wrap items-center justify-between gap-4">
                <div className="flex items-center gap-2 text-xs text-text-muted">
                  <BookOpen className="w-4 h-4 text-accent shrink-0" />
                  <span>Total Volume:</span>
                  <span className="text-text-main font-semibold">{stats.total_books} books</span>
                  <span className="text-border-subtle">•</span>
                  <Clock className="w-3.5 h-3.5 text-accent shrink-0" />
                  <span>Queue:</span>
                  <span className="text-accent font-semibold">{stats.want_to_read} to read</span>
                </div>

                {/* Format Badges */}
                <div className="flex flex-wrap gap-2">
                  {Object.entries(stats.format_distribution).map(([fmt, count]) => {
                    if (count <= 0) return null;
                    return (
                      <span
                        key={fmt}
                        className="text-[11px] font-medium bg-surface border border-border-subtle px-2.5 py-1 rounded-lg text-text-muted"
                      >
                        {fmt.replace('_', ' ')}: <strong className="text-accent">{count}</strong>
                      </span>
                    );
                  })}
                </div>
              </div>
            </div>
          ) : null}
        </motion.div>
      </div>
    </AnimatePresence>
  );
};
