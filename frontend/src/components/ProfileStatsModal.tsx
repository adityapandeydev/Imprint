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
  Target,
  Trophy,
  TrendingUp,
  Check,
  Edit3,
  Calendar,
  Globe,
  Lock,
  Eye,
  Share2,
  ExternalLink,
} from 'lucide-react';
import { api } from '../lib/api';
import { useAuth } from '../context/AuthContext';
import { ShareModal } from './ShareModal';
import { toast } from '../lib/toast';
import type { ProfileVisibility, ReadingChallenge, ReadingStats } from '../types/api';

const useSafeAuth = () => {
  try {
    return useAuth();
  } catch {
    return {
      user: null,
      isAuthenticated: false,
      isLoading: false,
      login: async () => {},
      register: async () => {},
      logout: async () => {},
      updateUserProfileVisibility: async () => {},
      isAuthModalOpen: false,
      authModalMode: 'login' as const,
      openAuthModal: () => {},
      closeAuthModal: () => {},
    };
  }
};

interface ProfileStatsModalProps {
  isOpen: boolean;
  onClose: () => void;
  userName?: string;
}

const MONTH_NAMES = [
  'Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun',
  'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec',
];

const PRESET_GOALS = [12, 24, 36, 52];

export const ProfileStatsModal: React.FC<ProfileStatsModalProps> = ({
  isOpen,
  onClose,
  userName = 'Reader',
}) => {
  const [stats, setStats] = useState<ReadingStats | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  // Reading Goal Edit State
  const [isEditingGoal, setIsEditingGoal] = useState(false);
  const [targetInput, setTargetInput] = useState<string>('');
  const [isSavingGoal, setIsSavingGoal] = useState(false);
  const [goalSaveError, setGoalSaveError] = useState<string | null>(null);
  const [hoveredMonth, setHoveredMonth] = useState<number | null>(null);

  // Privacy & Sharing State
  const auth = useSafeAuth();
  const user = auth.user;
  const [isUpdatingPrivacy, setIsUpdatingPrivacy] = useState(false);
  const [isShareModalOpen, setIsShareModalOpen] = useState(false);

  const currentVisibility = user?.profile_visibility || 'PUBLIC';

  const handlePrivacyChange = async (visibility: ProfileVisibility) => {
    if (visibility === currentVisibility) return;
    try {
      setIsUpdatingPrivacy(true);
      await auth.updateUserProfileVisibility(visibility);
      toast.success(`Library visibility set to ${visibility}`);
    } catch {
      toast.error('Failed to update privacy settings');
    } finally {
      setIsUpdatingPrivacy(false);
    }
  };

  const handleViewPublicProfile = () => {
    if (!user?.username) return;
    onClose();
    window.location.hash = `#/u/${encodeURIComponent(user.username)}`;
  };

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
      setIsEditingGoal(false);
      setGoalSaveError(null);
    }
  }, [isOpen]);

  if (!isOpen) return null;

  const handleOpenEditGoal = () => {
    setTargetInput(stats?.challenge?.target_books ? String(stats.challenge.target_books) : '24');
    setGoalSaveError(null);
    setIsEditingGoal(true);
  };

  const handleSaveGoal = async () => {
    const target = parseInt(targetInput, 10);
    if (isNaN(target) || target < 1 || target > 1000) {
      setGoalSaveError('Please enter a valid goal between 1 and 1,000 books.');
      return;
    }

    try {
      setIsSavingGoal(true);
      setGoalSaveError(null);
      const year = stats?.current_year || new Date().getFullYear();
      await api.setReadingGoal(year, target);
      setIsEditingGoal(false);
      await fetchStats();
    } catch (err) {
      setGoalSaveError(err instanceof Error ? err.message : 'Failed to save reading goal');
    } finally {
      setIsSavingGoal(false);
    }
  };

  const getPacingConfig = (status?: ReadingChallenge['pacing_status']) => {
    switch (status) {
      case 'COMPLETED':
        return {
          label: 'Goal Completed',
          badgeClass: 'bg-emerald-500/10 text-emerald-600 dark:text-emerald-400 border-emerald-500/20',
          Icon: Trophy,
        };
      case 'AHEAD':
        return {
          label: 'Ahead of Schedule',
          badgeClass: 'bg-emerald-500/10 text-emerald-600 dark:text-emerald-400 border-emerald-500/20',
          Icon: TrendingUp,
        };
      case 'ON_TRACK':
        return {
          label: 'On Track',
          badgeClass: 'bg-sky-500/10 text-sky-600 dark:text-sky-400 border-sky-500/20',
          Icon: Check,
        };
      case 'BEHIND':
        return {
          label: 'Behind Schedule',
          badgeClass: 'bg-amber-500/10 text-amber-600 dark:text-amber-400 border-amber-500/20',
          Icon: Clock,
        };
      case 'NOT_SET':
      default:
        return {
          label: 'No Goal Set',
          badgeClass: 'bg-surface text-text-muted border-border-subtle',
          Icon: Target,
        };
    }
  };

  const challenge = stats?.challenge;
  const targetBooks = challenge?.target_books || 0;
  const booksFinished = challenge?.books_finished || 0;
  const percentage = challenge?.percentage || 0;
  const radius = 44;
  const circumference = 2 * Math.PI * radius;
  const clampedPercent = Math.min(Math.max(percentage, 0), 100);
  const strokeDashoffset = circumference - (clampedPercent / 100) * circumference;

  const pacingConfig = getPacingConfig(challenge?.pacing_status);
  const PacingIcon = pacingConfig.Icon;

  const monthlyData = challenge?.monthly_progress || [];
  const maxMonthlyBooks = Math.max(...monthlyData.map((m) => m.books), 1);
  const currentMonthIdx = new Date().getMonth(); // 0-11
  const isCurrentYear = !stats || stats.current_year === new Date().getFullYear();

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
              {/* Annual Reading Challenge Section */}
              <div className="bg-surface-hover/50 border border-border-subtle rounded-2xl p-5 sm:p-6 transition-all">
                <div className="flex items-center justify-between border-b border-border-subtle/80 pb-4 mb-4">
                  <div className="flex items-center gap-2.5">
                    <div className="p-2 rounded-lg bg-accent/10 text-accent border border-accent/20">
                      <Trophy className="w-4 h-4 text-accent" />
                    </div>
                    <div>
                      <h3 className="text-sm sm:text-base font-semibold text-text-main">
                        {stats.current_year} Reading Challenge
                      </h3>
                      <p className="text-[11px] text-text-muted">
                        Track your yearly pace and reading milestones
                      </p>
                    </div>
                  </div>

                  {!isEditingGoal && (
                    <button
                      onClick={handleOpenEditGoal}
                      className="inline-flex items-center gap-1.5 px-3 py-1.5 text-xs font-medium rounded-lg bg-surface border border-border-subtle hover:bg-surface-hover text-text-main transition-colors cursor-pointer"
                      aria-label="Edit reading goal"
                    >
                      <Edit3 className="w-3.5 h-3.5 text-accent" />
                      <span>{targetBooks > 0 ? 'Edit Goal' : 'Set Goal'}</span>
                    </button>
                  )}
                </div>

                {isEditingGoal ? (
                  /* Goal Editor Drawer */
                  <div className="bg-surface border border-border-subtle rounded-xl p-4 sm:p-5 space-y-4">
                    <div className="flex items-center justify-between">
                      <span className="text-xs font-semibold uppercase tracking-wider text-text-muted">
                        Set your {stats.current_year} reading target
                      </span>
                      <button
                        onClick={() => {
                          setIsEditingGoal(false);
                          setGoalSaveError(null);
                        }}
                        className="text-xs text-text-muted hover:text-text-main transition-colors cursor-pointer"
                      >
                        Cancel
                      </button>
                    </div>

                    {/* Presets */}
                    <div>
                      <p className="text-xs text-text-muted mb-2">Quick presets:</p>
                      <div className="grid grid-cols-4 gap-2">
                        {PRESET_GOALS.map((preset) => (
                          <button
                            key={preset}
                            type="button"
                            onClick={() => setTargetInput(String(preset))}
                            className={`py-2 px-1 text-xs font-semibold rounded-lg border transition-all cursor-pointer ${
                              targetInput === String(preset)
                                ? 'bg-accent text-white border-accent shadow-sm'
                                : 'bg-surface-hover/80 text-text-main border-border-subtle hover:border-accent/40'
                            }`}
                          >
                            {preset} books
                          </button>
                        ))}
                      </div>
                    </div>

                    {/* Custom Input */}
                    <div>
                      <label
                        htmlFor="reading-goal-input"
                        className="block text-xs text-text-muted mb-1.5 font-medium"
                      >
                        Or custom target (books):
                      </label>
                      <div className="flex items-center gap-2">
                        <input
                          id="reading-goal-input"
                          type="number"
                          min="1"
                          max="1000"
                          value={targetInput}
                          onChange={(e) => setTargetInput(e.target.value)}
                          placeholder="e.g. 30"
                          aria-label="Target books input"
                          className="flex-1 bg-surface-hover/60 border border-border-subtle rounded-lg px-3 py-2 text-sm text-text-main focus:outline-none focus:border-accent"
                        />
                        <button
                          onClick={handleSaveGoal}
                          disabled={isSavingGoal}
                          className="px-4 py-2 bg-accent text-white text-xs font-semibold rounded-lg hover:opacity-90 disabled:opacity-50 transition-opacity cursor-pointer inline-flex items-center gap-1.5"
                        >
                          {isSavingGoal ? (
                            <>
                              <RefreshCw className="w-3.5 h-3.5 animate-spin" />
                              <span>Saving...</span>
                            </>
                          ) : (
                            <>
                              <Check className="w-3.5 h-3.5" />
                              <span>Save Goal</span>
                            </>
                          )}
                        </button>
                      </div>
                      {goalSaveError && (
                        <p className="text-xs text-red-500 mt-2 font-medium">{goalSaveError}</p>
                      )}
                    </div>
                  </div>
                ) : (
                  /* Challenge Overview with Circular Ring and Pacing */
                  <div className="flex flex-col sm:flex-row items-center sm:items-stretch gap-6">
                    {/* Circular Progress Ring */}
                    <div className="relative flex items-center justify-center shrink-0 w-28 h-28">
                      <svg className="w-28 h-28 transform -rotate-90" viewBox="0 0 100 100">
                        {/* Background track */}
                        <circle
                          cx="50"
                          cy="50"
                          r={radius}
                          stroke="currentColor"
                          strokeWidth="8"
                          className="text-surface-hover fill-none"
                        />
                        {/* Progress ring */}
                        <motion.circle
                          cx="50"
                          cy="50"
                          r={radius}
                          stroke="currentColor"
                          strokeWidth="8"
                          strokeLinecap="round"
                          className="text-accent fill-none"
                          initial={{ strokeDashoffset: circumference }}
                          animate={{ strokeDashoffset }}
                          transition={{ duration: 0.8, ease: 'easeOut' }}
                          style={{
                            strokeDasharray: circumference,
                          }}
                        />
                      </svg>
                      {/* Inner Ring Text */}
                      <div className="absolute inset-0 flex flex-col items-center justify-center text-center">
                        <span className="text-xl font-bold font-serif text-text-main tracking-tight">
                          {percentage}%
                        </span>
                        <span className="text-[10px] uppercase font-semibold text-text-muted tracking-wider">
                          Goal
                        </span>
                      </div>
                    </div>

                    {/* Challenge Stats & Pacing Badge */}
                    <div className="flex-1 flex flex-col justify-center text-center sm:text-left space-y-2.5">
                      <div className="flex flex-wrap items-center justify-center sm:justify-start gap-2">
                        <span
                          className={`inline-flex items-center gap-1.5 px-2.5 py-0.5 rounded-full text-xs font-semibold border ${pacingConfig.badgeClass}`}
                        >
                          <PacingIcon className="w-3.5 h-3.5" />
                          <span>{pacingConfig.label}</span>
                        </span>

                        {targetBooks > 0 && (
                          <span className="text-xs text-text-muted font-medium">
                            {booksFinished} of {targetBooks} books completed
                          </span>
                        )}
                      </div>

                      <p className="text-xs sm:text-sm text-text-main font-medium leading-relaxed">
                        {challenge?.pacing_message ||
                          (targetBooks > 0
                            ? `You have completed ${booksFinished} books toward your ${targetBooks} book goal.`
                            : 'Set an annual reading goal to track your pacing and celebrate reading milestones!')}
                      </p>

                      {targetBooks > 0 && (
                        <div className="flex flex-wrap items-center justify-center sm:justify-start gap-x-4 gap-y-1 text-xs text-text-muted pt-0.5">
                          <span>
                            Remaining:{' '}
                            <strong className="text-text-main">
                              {Math.max(0, targetBooks - booksFinished)}
                            </strong>{' '}
                            books
                          </span>
                          <span>•</span>
                          <span>
                            Expected pace:{' '}
                            <strong className="text-text-main">
                              {(challenge?.expected_finished || 0).toFixed(1)}
                            </strong>{' '}
                            books by day {challenge?.days_elapsed || 1}
                          </span>
                        </div>
                      )}
                    </div>
                  </div>
                )}
              </div>

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

              {/* 12-Month Reading Velocity Bar Chart */}
              <div className="bg-surface-hover/40 border border-border-subtle rounded-xl p-5">
                <div className="flex items-center justify-between mb-4">
                  <div className="flex items-center gap-2">
                    <Calendar className="w-4 h-4 text-accent" />
                    <h3 className="text-sm font-semibold text-text-main">
                      12-Month Reading Velocity ({stats.current_year})
                    </h3>
                  </div>
                  <span className="text-xs text-text-muted font-medium">
                    {stats.books_finished_year} {stats.books_finished_year === 1 ? 'book' : 'books'} completed
                  </span>
                </div>

                {/* Bars Container */}
                <div className="pt-6 pb-2">
                  <div className="flex items-end justify-between gap-1 sm:gap-2 h-28 px-1">
                    {monthlyData.map((m) => {
                      const isCurrent = isCurrentYear && m.month === currentMonthIdx + 1;
                      const heightPercent =
                        m.books > 0
                          ? Math.max(Math.round((m.books / maxMonthlyBooks) * 100), 16)
                          : 4;
                      const isHovered = hoveredMonth === m.month;

                      return (
                        <div
                          key={m.month}
                          className="flex-1 flex flex-col items-center group relative cursor-pointer"
                          onMouseEnter={() => setHoveredMonth(m.month)}
                          onMouseLeave={() => setHoveredMonth(null)}
                        >
                          {/* Hover Tooltip */}
                          {isHovered && (
                            <div className="absolute -top-12 z-20 bg-surface border border-border-subtle shadow-xl px-2.5 py-1 rounded-md text-[11px] whitespace-nowrap pointer-events-none transform -translate-x-1/2 left-1/2">
                              <span className="font-semibold text-text-main">
                                {MONTH_NAMES[m.month - 1]}
                              </span>
                              :{' '}
                              <span className="text-accent font-bold">
                                {m.books} {m.books === 1 ? 'book' : 'books'}
                              </span>
                              <span className="text-text-muted">
                                {' '}
                                ({m.pages.toLocaleString()} pages)
                              </span>
                            </div>
                          )}

                          {/* Book count above bar */}
                          <span
                            className={`text-[10px] font-semibold mb-1 h-3 transition-opacity ${
                              m.books > 0 ? 'text-accent opacity-100' : 'opacity-0'
                            }`}
                          >
                            {m.books > 0 ? m.books : ''}
                          </span>

                          {/* Bar Graphic */}
                          <div className="w-full max-w-[28px] h-20 bg-surface rounded-t-sm flex flex-col justify-end overflow-hidden border-b border-border-subtle">
                            <motion.div
                              initial={{ height: 0 }}
                              animate={{ height: `${heightPercent}%` }}
                              transition={{ duration: 0.5, ease: 'easeOut' }}
                              className={`w-full rounded-t-sm transition-colors ${
                                m.books > 0
                                  ? isCurrent
                                    ? 'bg-accent'
                                    : 'bg-accent/80 hover:bg-accent'
                                  : 'bg-border-subtle/40'
                              }`}
                            />
                          </div>

                          {/* Month Label */}
                          <span
                            className={`text-[10px] mt-1.5 font-medium transition-colors ${
                              isCurrent
                                ? 'text-accent font-bold underline underline-offset-2'
                                : 'text-text-muted group-hover:text-text-main'
                            }`}
                          >
                            {MONTH_NAMES[m.month - 1]}
                          </span>
                        </div>
                      );
                    })}
                  </div>
                </div>

                <div className="flex items-center justify-between text-[11px] text-text-muted pt-3 border-t border-border-subtle/60 mt-2">
                  <span>Pacing breakdown across 12 calendar months</span>
                  <span>Hover over any month to view page volume</span>
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

              {/* Privacy & Public Sharing Controls */}
              {user && (
                <div className="bg-surface-hover/40 border border-border-subtle rounded-2xl p-5 sm:p-6 space-y-4">
                  <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
                    <div className="flex items-center gap-2">
                      <Globe className="w-4 h-4 text-accent" />
                      <h3 className="text-sm font-semibold text-text-main">
                        Library Visibility & Social Sharing
                      </h3>
                    </div>
                    <div className="flex items-center gap-2">
                      {currentVisibility !== 'PRIVATE' && (
                        <>
                          <button
                            onClick={() => setIsShareModalOpen(true)}
                            className="inline-flex items-center gap-1.5 px-3 py-1.5 bg-accent-soft border border-accent/20 text-accent text-xs font-semibold rounded-lg hover:opacity-90 transition-opacity cursor-pointer"
                          >
                            <Share2 className="w-3.5 h-3.5" />
                            <span>Share Library</span>
                          </button>
                          <button
                            onClick={handleViewPublicProfile}
                            className="inline-flex items-center gap-1.5 px-3 py-1.5 bg-surface border border-border-subtle hover:bg-surface-hover text-text-main text-xs font-semibold rounded-lg transition-colors cursor-pointer"
                          >
                            <ExternalLink className="w-3.5 h-3.5 text-accent" />
                            <span>View Public Profile</span>
                          </button>
                        </>
                      )}
                    </div>
                  </div>

                  {/* Visibility Choice Cards */}
                  <div className="grid grid-cols-1 sm:grid-cols-3 gap-3">
                    {/* PUBLIC */}
                    <button
                      type="button"
                      disabled={isUpdatingPrivacy}
                      onClick={() => handlePrivacyChange('PUBLIC')}
                      className={`p-3.5 rounded-xl border text-left transition-all cursor-pointer ${
                        currentVisibility === 'PUBLIC'
                          ? 'bg-surface border-accent shadow-sm ring-1 ring-accent/30'
                          : 'bg-surface/60 border-border-subtle hover:border-accent/40'
                      }`}
                    >
                      <div className="flex items-center justify-between mb-1.5">
                        <span className="text-xs font-bold text-text-main flex items-center gap-1.5">
                          <Globe className="w-3.5 h-3.5 text-emerald-500" />
                          <span>Public</span>
                        </span>
                        {currentVisibility === 'PUBLIC' && (
                          <Check className="w-3.5 h-3.5 text-accent" />
                        )}
                      </div>
                      <p className="text-[11px] text-text-muted leading-relaxed">
                        Anyone can discover and view your collection, challenge progress, and shelves.
                      </p>
                    </button>

                    {/* UNLISTED */}
                    <button
                      type="button"
                      disabled={isUpdatingPrivacy}
                      onClick={() => handlePrivacyChange('UNLISTED')}
                      className={`p-3.5 rounded-xl border text-left transition-all cursor-pointer ${
                        currentVisibility === 'UNLISTED'
                          ? 'bg-surface border-accent shadow-sm ring-1 ring-accent/30'
                          : 'bg-surface/60 border-border-subtle hover:border-accent/40'
                      }`}
                    >
                      <div className="flex items-center justify-between mb-1.5">
                        <span className="text-xs font-bold text-text-main flex items-center gap-1.5">
                          <Eye className="w-3.5 h-3.5 text-amber-500" />
                          <span>Unlisted</span>
                        </span>
                        {currentVisibility === 'UNLISTED' && (
                          <Check className="w-3.5 h-3.5 text-accent" />
                        )}
                      </div>
                      <p className="text-[11px] text-text-muted leading-relaxed">
                        Only people with your direct link can view your shelves. Hidden from public discovery.
                      </p>
                    </button>

                    {/* PRIVATE */}
                    <button
                      type="button"
                      disabled={isUpdatingPrivacy}
                      onClick={() => handlePrivacyChange('PRIVATE')}
                      className={`p-3.5 rounded-xl border text-left transition-all cursor-pointer ${
                        currentVisibility === 'PRIVATE'
                          ? 'bg-surface border-accent shadow-sm ring-1 ring-accent/30'
                          : 'bg-surface/60 border-border-subtle hover:border-accent/40'
                      }`}
                    >
                      <div className="flex items-center justify-between mb-1.5">
                        <span className="text-xs font-bold text-text-main flex items-center gap-1.5">
                          <Lock className="w-3.5 h-3.5 text-rose-500" />
                          <span>Private</span>
                        </span>
                        {currentVisibility === 'PRIVATE' && (
                          <Check className="w-3.5 h-3.5 text-accent" />
                        )}
                      </div>
                      <p className="text-[11px] text-text-muted leading-relaxed">
                        Only you can view your collection and stats. Your public profile link is disabled.
                      </p>
                    </button>
                  </div>
                </div>
              )}
            </div>
          ) : null}

          {/* Share Modal */}
          {user && (
            <ShareModal
              isOpen={isShareModalOpen}
              onClose={() => setIsShareModalOpen(false)}
              username={user.username}
              displayName={user.display_name}
            />
          )}
        </motion.div>
      </div>
    </AnimatePresence>
  );
};
