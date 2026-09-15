import React, { useState } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import {
  BookOpen,
  BookmarkCheck,
  Compass,
  Sparkles,
  Sun,
  Moon,
  LogIn,
  LogOut,
  ChevronDown,
} from 'lucide-react';
import { useTheme } from '../lib/theme';
import { useAuth } from '../context/AuthContext';

interface NavbarProps {
  activeTab: 'discover' | 'collection';
  onSelectTab: (tab: 'discover' | 'collection') => void;
  collectionCount?: number;
}

export const Navbar: React.FC<NavbarProps> = ({
  activeTab,
  onSelectTab,
  collectionCount = 0,
}) => {
  const { isDark, toggleTheme } = useTheme();
  const { user, isAuthenticated, logout, openAuthModal } = useAuth();
  const [isUserMenuOpen, setIsUserMenuOpen] = useState(false);

  const navTabs = [
    { id: 'discover' as const, label: 'Discover', icon: Compass },
    {
      id: 'collection' as const,
      label: 'My Collection',
      icon: BookmarkCheck,
      badge: isAuthenticated ? collectionCount : undefined,
    },
  ];

  return (
    <header className="sticky top-0 z-40 bg-canvas/90 backdrop-blur-md border-b border-border-subtle transition-colors">
      <div className="w-full px-4 sm:px-6 lg:px-8 xl:px-10">
        <div className="relative flex items-center justify-between h-16 sm:h-20">
          {/* Brand Logo */}
          <motion.div
            className="flex items-center gap-3 cursor-pointer group select-none z-10"
            onClick={() => onSelectTab('discover')}
            whileHover={{ scale: 1.02 }}
            whileTap={{ scale: 0.98 }}
            transition={{ type: 'spring', stiffness: 400, damping: 25 }}
          >
            <div className="w-10 h-10 rounded-xl bg-surface border border-border-subtle flex items-center justify-center text-accent shadow-xs group-hover:border-accent/40 transition-colors">
              <BookOpen className="w-5 h-5 transition-transform group-hover:-rotate-3" />
            </div>
            <div>
              <span className="font-serif font-bold text-2xl tracking-tight text-text-main flex items-center gap-1.5">
                Imprint
                <Sparkles className="w-3.5 h-3.5 text-amber-gold" />
              </span>
              <p className="text-[11px] uppercase tracking-wider text-text-muted font-medium -mt-1 hidden sm:block">
                Catalog & Collection
              </p>
            </div>
          </motion.div>

          {/* Navigation Tabs with sliding pill indicator (Dead-Center Alignment) */}
          <div className="absolute left-1/2 -translate-x-1/2 flex items-center justify-center pointer-events-auto">
            <nav className="relative flex items-center gap-1 sm:gap-1.5 bg-surface/80 p-1 rounded-xl border border-border-subtle shadow-xs">
              {navTabs.map((tab) => {
                const Icon = tab.icon;
                const isActive = activeTab === tab.id;

                return (
                  <button
                    key={tab.id}
                    onClick={() => onSelectTab(tab.id)}
                    className={`relative flex items-center justify-center gap-2 px-3.5 sm:px-4 py-2 rounded-lg text-sm font-medium transition-colors z-10 select-none cursor-pointer ${
                      isActive ? 'text-accent font-semibold' : 'text-text-muted hover:text-text-main'
                    }`}
                  >
                    {/* Sliding active pill */}
                    {isActive && (
                      <motion.div
                        layoutId="activeTabPill"
                        className="absolute inset-0 bg-accent-soft border border-accent/25 rounded-lg shadow-xs -z-10 pointer-events-none"
                        transition={{
                          type: 'spring',
                          stiffness: 450,
                          damping: 35,
                        }}
                      />
                    )}

                    <Icon className={`w-4 h-4 shrink-0 transition-transform ${isActive ? 'scale-105 text-accent' : ''}`} />
                    <span className="leading-none">{tab.label}</span>

                    {tab.badge !== undefined && tab.badge > 0 && (
                      <motion.span
                        key={tab.badge}
                        initial={{ scale: 0.7, opacity: 0 }}
                        animate={{ scale: 1, opacity: 1 }}
                        className="ml-0.5 px-1.5 py-0.5 text-xs font-semibold bg-accent text-canvas rounded-full leading-none"
                      >
                        {tab.badge}
                      </motion.span>
                    )}
                  </button>
                );
              })}
            </nav>
          </div>

          {/* Right Side Actions: Theme Toggle & User Auth */}
          <div className="flex items-center gap-3 z-10">
            {/* Dark / Light Mode Toggle with micro-spin transition */}
            <motion.button
              onClick={toggleTheme}
              aria-label="Toggle theme"
              whileHover={{ scale: 1.06 }}
              whileTap={{ scale: 0.92 }}
              transition={{ type: 'spring', stiffness: 400, damping: 20 }}
              className="w-10 h-10 rounded-xl bg-surface border border-border-subtle flex items-center justify-center text-text-muted hover:text-text-main hover:bg-surface-hover transition-colors shadow-xs relative overflow-hidden cursor-pointer"
              title={isDark ? 'Switch to light mode' : 'Switch to dark mode'}
            >
              <AnimatePresence mode="wait" initial={false}>
                {isDark ? (
                  <motion.div
                    key="sun"
                    initial={{ rotate: -90, opacity: 0, scale: 0.6 }}
                    animate={{ rotate: 0, opacity: 1, scale: 1 }}
                    exit={{ rotate: 90, opacity: 0, scale: 0.6 }}
                    transition={{ duration: 0.2 }}
                  >
                    <Sun className="w-4 h-4 text-amber-gold" />
                  </motion.div>
                ) : (
                  <motion.div
                    key="moon"
                    initial={{ rotate: 90, opacity: 0, scale: 0.6 }}
                    animate={{ rotate: 0, opacity: 1, scale: 1 }}
                    exit={{ rotate: -90, opacity: 0, scale: 0.6 }}
                    transition={{ duration: 0.2 }}
                  >
                    <Moon className="w-4 h-4 text-accent" />
                  </motion.div>
                )}
              </AnimatePresence>
            </motion.button>

            {/* Auth Action: Sign In Button or User Dropdown */}
            {isAuthenticated && user ? (
              <div className="relative">
                <button
                  onClick={() => setIsUserMenuOpen(!isUserMenuOpen)}
                  className="flex items-center gap-2 px-2.5 sm:px-3 py-1.5 rounded-xl bg-surface hover:bg-surface-hover border border-border-subtle text-text-main text-xs sm:text-sm font-medium transition-colors shadow-xs cursor-pointer"
                >
                  <div className="w-6 h-6 rounded-lg bg-accent text-canvas font-bold flex items-center justify-center text-xs uppercase">
                    {user.username.charAt(0)}
                  </div>
                  <span className="hidden sm:inline font-medium max-w-[100px] truncate">
                    {user.display_name || user.username}
                  </span>
                  <ChevronDown className="w-3.5 h-3.5 text-text-muted transition-transform" />
                </button>

                {/* Dropdown Menu */}
                <AnimatePresence>
                  {isUserMenuOpen && (
                    <>
                      <div
                        className="fixed inset-0 z-20"
                        onClick={() => setIsUserMenuOpen(false)}
                      />
                      <motion.div
                        initial={{ opacity: 0, scale: 0.95, y: -4 }}
                        animate={{ opacity: 1, scale: 1, y: 0 }}
                        exit={{ opacity: 0, scale: 0.95, y: -4 }}
                        transition={{ duration: 0.15 }}
                        className="absolute right-0 mt-2 w-56 bg-surface border border-border-subtle rounded-xl p-2 shadow-book z-30 space-y-1"
                      >
                        <div className="px-3 py-2 border-b border-border-subtle">
                          <p className="font-semibold text-xs text-text-main truncate">
                            {user.display_name || user.username}
                          </p>
                          <p className="text-[11px] text-text-muted font-mono truncate">
                            @{user.username}
                          </p>
                          <p className="text-[10px] text-text-muted/70 truncate mt-0.5">
                            {user.email}
                          </p>
                        </div>

                        <button
                          onClick={() => {
                            setIsUserMenuOpen(false);
                            logout();
                          }}
                          className="w-full flex items-center gap-2 px-3 py-2 text-xs text-red-600 dark:text-red-400 hover:bg-red-500/10 rounded-lg transition-colors cursor-pointer text-left"
                        >
                          <LogOut className="w-3.5 h-3.5" />
                          <span>Sign Out</span>
                        </button>
                      </motion.div>
                    </>
                  )}
                </AnimatePresence>
              </div>
            ) : (
              <motion.button
                whileHover={{ scale: 1.02 }}
                whileTap={{ scale: 0.98 }}
                onClick={() => openAuthModal('login')}
                className="flex items-center gap-1.5 px-3.5 py-2 rounded-xl bg-accent text-canvas text-xs sm:text-sm font-semibold hover:opacity-90 transition-opacity shadow-xs cursor-pointer"
              >
                <LogIn className="w-3.5 h-3.5" />
                <span>Sign In</span>
              </motion.button>
            )}
          </div>
        </div>
      </div>
    </header>
  );
};
