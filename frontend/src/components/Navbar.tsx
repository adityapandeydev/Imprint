import React from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import { BookOpen, BookmarkCheck, Compass, Sparkles, Sun, Moon } from 'lucide-react';
import { useTheme } from '../lib/theme';

interface NavbarProps {
  activeTab: 'discover' | 'collection';
  onSelectTab: (tab: 'discover' | 'collection') => void;
  collectionCount?: number;
  serverStatus?: 'connected' | 'disconnected' | 'checking';
}

export const Navbar: React.FC<NavbarProps> = ({
  activeTab,
  onSelectTab,
  collectionCount = 0,
  serverStatus = 'checking',
}) => {
  const { isDark, toggleTheme } = useTheme();

  const navTabs = [
    { id: 'discover' as const, label: 'Discover', icon: Compass },
    { id: 'collection' as const, label: 'My Collection', icon: BookmarkCheck, badge: collectionCount },
  ];

  return (
    <header className="sticky top-0 z-40 bg-canvas/90 backdrop-blur-md border-b border-border-subtle transition-colors">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="flex items-center justify-between h-16 sm:h-20">
          {/* Brand Logo */}
          <motion.div
            className="flex items-center gap-3 cursor-pointer group select-none"
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

          {/* Navigation Tabs with sliding pill indicator */}
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

          {/* Right Side Actions: Theme Toggle & Server Status */}
          <div className="flex items-center gap-3">
            {/* Dark / Light Mode Toggle with micro-spin transition */}
            <motion.button
              onClick={toggleTheme}
              aria-label="Toggle theme"
              whileHover={{ scale: 1.06 }}
              whileTap={{ scale: 0.92 }}
              transition={{ type: 'spring', stiffness: 400, damping: 20 }}
              className="w-10 h-10 rounded-xl bg-surface border border-border-subtle flex items-center justify-center text-text-muted hover:text-text-main hover:bg-surface-hover transition-colors shadow-xs relative overflow-hidden"
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

            {/* Backend Status Indicator */}
            <div className="hidden md:flex items-center gap-2 text-xs text-text-muted bg-surface/60 border border-border-subtle px-2.5 py-1.5 rounded-lg">
              <span
                className={`w-2 h-2 rounded-full transition-colors ${
                  serverStatus === 'connected'
                    ? 'bg-emerald-500 shadow-[0_0_8px_rgba(16,185,129,0.6)]'
                    : serverStatus === 'disconnected'
                    ? 'bg-amber-500'
                    : 'bg-text-muted animate-pulse'
                }`}
              />
              <span className="capitalize font-medium">
                {serverStatus === 'connected'
                  ? 'API Ready'
                  : serverStatus === 'disconnected'
                  ? 'Standalone'
                  : 'Checking'}
              </span>
            </div>
          </div>
        </div>
      </div>
    </header>
  );
};
