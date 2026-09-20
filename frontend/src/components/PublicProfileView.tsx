import React, { useEffect, useState } from 'react';
import { motion } from 'framer-motion';
import {
  BookOpen,
  Share2,
  ShieldAlert,
  ArrowLeft,
  Trophy,
  Star,
  Tag,
} from 'lucide-react';
import { api } from '../lib/api';
import type {
  PublicProfile,
  WishlistItem,
  ReadingStatus,
  Work,
} from '../types/api';
import { ShareModal } from './ShareModal';

interface PublicProfileViewProps {
  username: string;
  initialShelf?: string;
  onNavigateHome: () => void;
  onInspectWork?: (work: Work) => void;
}

export const PublicProfileView: React.FC<PublicProfileViewProps> = ({
  username,
  initialShelf,
  onNavigateHome,
  onInspectWork,
}) => {
  const [profile, setProfile] = useState<PublicProfile | null>(null);
  const [items, setItems] = useState<WishlistItem[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<{ status?: number; message: string } | null>(null);

  // Filters
  const [statusFilter, setStatusFilter] = useState<ReadingStatus | 'ALL'>('ALL');
  const [activeShelf, setActiveShelf] = useState<string | null>(initialShelf || null);
  const [isShareModalOpen, setIsShareModalOpen] = useState(false);

  useEffect(() => {
    let isMounted = true;

    const loadPublicData = async () => {
      try {
        setLoading(true);
        setError(null);

        // 1. Fetch Profile
        const prof = await api.getPublicProfile(username);
        if (!isMounted) return;
        setProfile(prof);

        // 2. Fetch Collection
        const collectionResp = await api.getPublicCollection(username, {
          status: statusFilter === 'ALL' ? undefined : statusFilter,
          tag: activeShelf || undefined,
        });
        if (!isMounted) return;
        setItems(collectionResp.items || []);
      } catch (err: any) {
        if (!isMounted) return;
        const msg = err instanceof Error ? err.message : 'Failed to load reader profile';
        const isPrivate =
          msg.toLowerCase().includes('private') ||
          msg.toLowerCase().includes('forbidden') ||
          err?.status === 403;
        setError({
          status: isPrivate ? 403 : err?.status || 404,
          message: isPrivate
            ? "This reader's library is private"
            : 'Reader profile not found',
        });
      } finally {
        if (isMounted) setLoading(false);
      }
    };

    loadPublicData();

    return () => {
      isMounted = false;
    };
  }, [username, statusFilter, activeShelf]);

  // Handle shelf tag toggle
  const handleToggleShelf = (shelfTag: string) => {
    const nextShelf = activeShelf === shelfTag ? null : shelfTag;
    setActiveShelf(nextShelf);

    // Update browser URL without reload
    const cleanUsername = encodeURIComponent(username);
    const targetUrl = nextShelf
      ? `#/u/${cleanUsername}/shelf/${encodeURIComponent(nextShelf)}`
      : `#/u/${cleanUsername}`;
    if (window.location.hash !== targetUrl) {
      window.history.replaceState(null, '', targetUrl);
    }
  };

  // Circular ring math for challenge
  const challenge = profile?.stats?.challenge;
  const radius = 34;
  const circumference = 2 * Math.PI * radius;
  const clampedPercent = Math.min(Math.max(challenge?.percentage || 0, 0), 100);
  const strokeDashoffset = circumference - (clampedPercent / 100) * circumference;

  // Private profile Barrier Screen
  if (error?.status === 403) {
    return (
      <div className="min-h-screen bg-bg text-text-main flex flex-col items-center justify-center p-6 text-center">
        <motion.div
          initial={{ opacity: 0, scale: 0.95 }}
          animate={{ opacity: 1, scale: 1 }}
          className="max-w-md bg-surface border border-border-subtle rounded-3xl p-8 shadow-2xl space-y-5"
        >
          <div className="w-16 h-16 rounded-2xl bg-amber-500/10 text-amber-500 border border-amber-500/20 flex items-center justify-center mx-auto">
            <ShieldAlert className="w-8 h-8" />
          </div>
          <div>
            <h1 className="text-2xl font-serif font-bold text-text-main">
              This Library is Private
            </h1>
            <p className="text-xs sm:text-sm text-text-muted mt-2 leading-relaxed">
              @{username} has chosen to keep their reading collection, shelves, and reading analytics private.
            </p>
          </div>
          <div className="pt-2">
            <button
              onClick={onNavigateHome}
              className="inline-flex items-center gap-2 px-5 py-2.5 bg-accent text-white text-xs font-semibold rounded-xl hover:opacity-90 transition-opacity cursor-pointer shadow-md"
            >
              <ArrowLeft className="w-4 h-4" />
              <span>Return to Imprint</span>
            </button>
          </div>
        </motion.div>
      </div>
    );
  }

  // Not Found Screen
  if (error && error.status !== 403) {
    return (
      <div className="min-h-screen bg-bg text-text-main flex flex-col items-center justify-center p-6 text-center">
        <div className="max-w-md bg-surface border border-border-subtle rounded-3xl p-8 shadow-2xl space-y-4">
          <p className="text-text-muted text-sm">{error.message}</p>
          <button
            onClick={onNavigateHome}
            className="inline-flex items-center gap-2 px-5 py-2.5 bg-accent text-white text-xs font-semibold rounded-xl hover:opacity-90 transition-opacity cursor-pointer"
          >
            <ArrowLeft className="w-4 h-4" />
            <span>Discover Books</span>
          </button>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-bg text-text-main transition-colors duration-300">
      {/* Top Floating Navbar */}
      <header className="sticky top-0 z-40 bg-bg/80 backdrop-blur-md border-b border-border-subtle">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 h-16 flex items-center justify-between">
          <button
            onClick={onNavigateHome}
            className="inline-flex items-center gap-2 text-xs sm:text-sm font-medium text-text-muted hover:text-text-main transition-colors cursor-pointer group"
          >
            <ArrowLeft className="w-4 h-4 transition-transform group-hover:-translate-x-1 text-accent" />
            <span>Back to Discovery</span>
          </button>

          <div className="flex items-center gap-2">
            <span className="text-xs font-serif font-bold text-accent tracking-wider uppercase">
              Imprint Library
            </span>
          </div>

          <button
            onClick={() => setIsShareModalOpen(true)}
            className="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-lg bg-surface border border-border-subtle hover:bg-surface-hover text-text-main text-xs font-semibold transition-colors cursor-pointer"
          >
            <Share2 className="w-3.5 h-3.5 text-accent" />
            <span>Share Library</span>
          </button>
        </div>
      </header>

      {/* Main Content Container */}
      <main className="max-w-7xl mx-auto px-4 sm:px-6 py-8 space-y-8">
        {loading && !profile ? (
          <div className="py-24 flex flex-col items-center justify-center gap-4 text-text-muted">
            <div className="w-10 h-10 border-3 border-accent/30 border-t-accent rounded-full animate-spin" />
            <p className="text-sm font-medium">Opening @{username}'s library...</p>
          </div>
        ) : profile ? (
          <>
            {/* Reader Header Card */}
            <motion.div
              initial={{ opacity: 0, y: 15 }}
              animate={{ opacity: 1, y: 0 }}
              className="bg-surface border border-border-subtle rounded-3xl p-6 sm:p-8 shadow-sm relative overflow-hidden"
            >
              {/* Background watermark */}
              <div className="absolute right-0 bottom-0 translate-x-8 translate-y-8 opacity-5 pointer-events-none">
                <BookOpen className="w-72 h-72 text-accent" />
              </div>

              <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-6 relative z-10">
                <div className="flex items-center gap-5">
                  {/* Initials Avatar */}
                  <div className="w-16 h-16 sm:w-20 sm:h-20 rounded-2xl bg-gradient-to-br from-accent to-accent/60 text-white font-serif font-bold text-2xl sm:text-3xl flex items-center justify-center shadow-lg shrink-0">
                    {profile.display_name.charAt(0).toUpperCase()}
                  </div>

                  <div>
                    <div className="flex flex-wrap items-center gap-2.5">
                      <h1 className="text-2xl sm:text-3xl font-serif font-bold text-text-main tracking-tight">
                        {profile.display_name}
                      </h1>
                      {profile.profile_visibility === 'UNLISTED' && (
                        <span className="text-[10px] font-semibold uppercase tracking-wider px-2 py-0.5 rounded-full bg-surface-hover text-text-muted border border-border-subtle">
                          Unlisted
                        </span>
                      )}
                    </div>
                    <p className="text-xs sm:text-sm text-text-muted mt-0.5">
                      @{profile.username}
                      {profile.member_since && (
                        <>
                          {' '}
                          • Reader since{' '}
                          {new Date(profile.member_since).getFullYear()}
                        </>
                      )}
                    </p>
                  </div>
                </div>

                <div className="flex items-center gap-2.5 w-full sm:w-auto">
                  <button
                    onClick={() => setIsShareModalOpen(true)}
                    className="flex-1 sm:flex-initial inline-flex items-center justify-center gap-2 px-4 py-2.5 bg-accent text-white text-xs font-semibold rounded-xl hover:opacity-90 transition-opacity shadow-sm cursor-pointer"
                  >
                    <Share2 className="w-3.5 h-3.5" />
                    <span>Share Collection</span>
                  </button>
                </div>
              </div>

              {/* Challenge & Quick Stats Row */}
              {profile.stats && (
                <div className="grid grid-cols-2 sm:grid-cols-4 gap-3.5 mt-6 pt-6 border-t border-border-subtle/70">
                  {/* Annual Challenge mini ring */}
                  <div className="col-span-2 sm:col-span-1 bg-surface-hover/40 border border-border-subtle rounded-2xl p-4 flex items-center gap-3.5">
                    {challenge && challenge.target_books > 0 ? (
                      <>
                        <div className="relative w-16 h-16 shrink-0 flex items-center justify-center">
                          <svg className="w-16 h-16 transform -rotate-90" viewBox="0 0 80 80">
                            <circle
                              cx="40"
                              cy="40"
                              r={radius}
                              stroke="currentColor"
                              strokeWidth="6"
                              className="text-surface-hover fill-none"
                            />
                            <motion.circle
                              cx="40"
                              cy="40"
                              r={radius}
                              stroke="currentColor"
                              strokeWidth="6"
                              strokeLinecap="round"
                              className="text-accent fill-none"
                              initial={{ strokeDashoffset: circumference }}
                              animate={{ strokeDashoffset }}
                              transition={{ duration: 0.8, ease: 'easeOut' }}
                              style={{ strokeDasharray: circumference }}
                            />
                          </svg>
                          <span className="absolute text-xs font-bold text-text-main font-serif">
                            {challenge.percentage}%
                          </span>
                        </div>
                        <div>
                          <p className="text-xs font-semibold text-text-main">
                            {challenge.year} Goal
                          </p>
                          <p className="text-[11px] text-text-muted">
                            {challenge.books_finished} of {challenge.target_books} read
                          </p>
                        </div>
                      </>
                    ) : (
                      <div className="flex items-center gap-2.5">
                        <Trophy className="w-5 h-5 text-accent" />
                        <div>
                          <p className="text-xs font-semibold text-text-main">Reading Goal</p>
                          <p className="text-[11px] text-text-muted">Not set</p>
                        </div>
                      </div>
                    )}
                  </div>

                  <div className="bg-surface-hover/40 border border-border-subtle rounded-2xl p-4">
                    <p className="text-[11px] font-medium text-text-muted">Collection</p>
                    <p className="text-xl sm:text-2xl font-bold text-accent mt-0.5">
                      {profile.stats.total_books}
                    </p>
                    <p className="text-[10px] text-text-muted mt-0.5">Total books curated</p>
                  </div>

                  <div className="bg-surface-hover/40 border border-border-subtle rounded-2xl p-4">
                    <p className="text-[11px] font-medium text-text-muted">Pages Read</p>
                    <p className="text-xl sm:text-2xl font-bold text-emerald-600 dark:text-emerald-400 mt-0.5">
                      {profile.stats.total_pages_read.toLocaleString()}
                    </p>
                    <p className="text-[10px] text-text-muted mt-0.5">Literary volume</p>
                  </div>

                  <div className="bg-surface-hover/40 border border-border-subtle rounded-2xl p-4">
                    <p className="text-[11px] font-medium text-text-muted">Currently Reading</p>
                    <p className="text-xl sm:text-2xl font-bold text-sky-600 dark:text-sky-400 mt-0.5">
                      {profile.stats.currently_reading}
                    </p>
                    <p className="text-[10px] text-text-muted mt-0.5">Active books</p>
                  </div>
                </div>
              )}
            </motion.div>

            {/* Custom Shelves (Tags) Bar */}
            {profile.shelves.length > 0 && (
              <div className="space-y-2">
                <div className="flex items-center justify-between">
                  <span className="text-xs font-semibold uppercase tracking-wider text-text-muted flex items-center gap-1.5">
                    <Tag className="w-3.5 h-3.5 text-accent" />
                    <span>Curated Shelves</span>
                  </span>
                  {activeShelf && (
                    <button
                      onClick={() => handleToggleShelf(activeShelf)}
                      className="text-xs text-accent hover:underline cursor-pointer"
                    >
                      Clear shelf filter
                    </button>
                  )}
                </div>

                <div className="flex flex-wrap gap-2">
                  {profile.shelves.map((s) => {
                    const isSelected = activeShelf === s.tag;
                    return (
                      <button
                        key={s.tag}
                        onClick={() => handleToggleShelf(s.tag)}
                        className={`inline-flex items-center gap-1.5 px-3 py-1.5 rounded-xl text-xs font-medium border transition-all cursor-pointer ${
                          isSelected
                            ? 'bg-accent text-white border-accent shadow-sm'
                            : 'bg-surface hover:bg-surface-hover text-text-main border-border-subtle'
                        }`}
                      >
                        <span>#{s.tag}</span>
                        <span
                          className={`text-[10px] px-1.5 py-0.2 rounded-full ${
                            isSelected ? 'bg-white/20 text-white' : 'bg-surface-hover text-text-muted'
                          }`}
                        >
                          {s.count}
                        </span>
                      </button>
                    );
                  })}
                </div>
              </div>
            )}

            {/* Status Filter Tabs */}
            <div className="flex items-center justify-between border-b border-border-subtle pb-3">
              <div className="flex flex-wrap gap-1.5">
                {(['ALL', 'CURRENTLY_READING', 'FINISHED', 'WANT_TO_READ'] as const).map(
                  (tab) => {
                    const isSelected = statusFilter === tab;
                    const label =
                      tab === 'ALL'
                        ? 'All Books'
                        : tab === 'CURRENTLY_READING'
                        ? 'Reading'
                        : tab === 'FINISHED'
                        ? 'Finished'
                        : 'Queue';

                    return (
                      <button
                        key={tab}
                        onClick={() => setStatusFilter(tab)}
                        className={`px-3 py-1.5 text-xs font-medium rounded-lg transition-colors cursor-pointer ${
                          isSelected
                            ? 'bg-surface-hover text-text-main border border-border-subtle shadow-xs'
                            : 'text-text-muted hover:text-text-main'
                        }`}
                      >
                        {label}
                      </button>
                    );
                  }
                )}
              </div>

              <span className="text-xs text-text-muted">
                {items.length} {items.length === 1 ? 'book' : 'books'}
              </span>
            </div>

            {/* Books Grid */}
            {items.length === 0 ? (
              <div className="py-20 text-center bg-surface border border-border-subtle rounded-3xl p-8 space-y-3">
                <BookOpen className="w-10 h-10 text-accent/40 mx-auto" />
                <h3 className="text-base font-serif font-bold text-text-main">
                  No books in this view
                </h3>
                <p className="text-xs text-text-muted max-w-sm mx-auto">
                  {activeShelf
                    ? `No books tagged with #${activeShelf} matching the current filter.`
                    : 'This reader has not added any books to this shelf yet.'}
                </p>
              </div>
            ) : (
              <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6 gap-4 sm:gap-6">
                {items.map((item) => {
                  const work = item.work;
                  if (!work) return null;

                  return (
                    <motion.div
                      key={item.id}
                      initial={{ opacity: 0, y: 10 }}
                      animate={{ opacity: 1, y: 0 }}
                      className="group flex flex-col bg-surface border border-border-subtle rounded-2xl overflow-hidden shadow-xs hover:shadow-lg transition-all duration-300"
                    >
                      {/* Cover Area */}
                      <div
                        onClick={() => onInspectWork && onInspectWork(work)}
                        className="relative aspect-[2/3] bg-surface-hover/80 overflow-hidden cursor-pointer"
                      >
                        {work.cover_url ? (
                          <img
                            src={work.cover_url}
                            alt={work.title}
                            className="w-full h-full object-cover transition-transform duration-500 group-hover:scale-105"
                            loading="lazy"
                          />
                        ) : (
                          <div className="w-full h-full flex flex-col items-center justify-center p-3 text-center bg-gradient-to-br from-surface to-surface-hover">
                            <BookOpen className="w-8 h-8 text-accent/40 mb-2" />
                            <p className="text-xs font-serif font-bold text-text-main line-clamp-2">
                              {work.title}
                            </p>
                          </div>
                        )}

                        {/* Status Badge */}
                        <div className="absolute top-2 left-2">
                          <span
                            className={`text-[9px] uppercase font-bold px-2 py-0.5 rounded-full shadow-md backdrop-blur-md ${
                              item.status === 'FINISHED'
                                ? 'bg-emerald-600/90 text-white'
                                : item.status === 'CURRENTLY_READING'
                                ? 'bg-sky-600/90 text-white'
                                : 'bg-black/60 text-white'
                            }`}
                          >
                            {item.status === 'CURRENTLY_READING'
                              ? 'Reading'
                              : item.status === 'FINISHED'
                              ? 'Finished'
                              : 'Want to Read'}
                          </span>
                        </div>

                        {/* Rating overlay if present */}
                        {item.rating && item.rating > 0 && (
                          <div className="absolute top-2 right-2 bg-black/70 backdrop-blur-md rounded-full px-2 py-0.5 flex items-center gap-1">
                            <Star className="w-3 h-3 fill-amber-400 text-amber-400" />
                            <span className="text-[10px] font-bold text-white">
                              {item.rating}
                            </span>
                          </div>
                        )}
                      </div>

                      {/* Info Area */}
                      <div className="p-3 flex-1 flex flex-col justify-between">
                        <div>
                          <h4
                            onClick={() => onInspectWork && onInspectWork(work)}
                            className="text-xs sm:text-sm font-serif font-bold text-text-main line-clamp-1 hover:text-accent transition-colors cursor-pointer"
                            title={work.title}
                          >
                            {work.title}
                          </h4>
                          <p className="text-[11px] text-text-muted truncate mt-0.5">
                            {work.authors && work.authors.length > 0
                              ? work.authors.map((a) => a.name).join(', ')
                              : 'Unknown Author'}
                          </p>
                        </div>

                        {/* Tags / Shelves on Book */}
                        {item.tags && item.tags.length > 0 && (
                          <div className="flex flex-wrap gap-1 mt-2.5 pt-2 border-t border-border-subtle/60">
                            {item.tags.slice(0, 2).map((t) => (
                              <span
                                key={t}
                                onClick={() => handleToggleShelf(t)}
                                className="text-[9px] font-medium bg-surface-hover hover:bg-accent hover:text-white px-1.5 py-0.5 rounded text-text-muted transition-colors cursor-pointer truncate max-w-[85px]"
                              >
                                #{t}
                              </span>
                            ))}
                            {item.tags.length > 2 && (
                              <span className="text-[9px] text-text-muted">
                                +{item.tags.length - 2}
                              </span>
                            )}
                          </div>
                        )}
                      </div>
                    </motion.div>
                  );
                })}
              </div>
            )}
          </>
        ) : null}
      </main>

      {/* Share Modal */}
      {profile && (
        <ShareModal
          isOpen={isShareModalOpen}
          onClose={() => setIsShareModalOpen(false)}
          username={profile.username}
          displayName={profile.display_name}
          shelf={activeShelf || undefined}
        />
      )}
    </div>
  );
};
