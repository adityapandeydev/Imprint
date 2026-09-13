import { useState } from 'react';
import { QueryClient, QueryClientProvider, useQuery } from '@tanstack/react-query';
import { motion, AnimatePresence } from 'framer-motion';
import { Navbar } from './components/Navbar';
import { api } from './lib/api';
import { BookMarked, Search, Layers, ShieldCheck, Sparkles, ArrowRight } from 'lucide-react';

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 1000 * 60 * 5, // 5 minutes cache
      retry: 1,
      refetchOnWindowFocus: false,
    },
  },
});

function ImprintApp() {
  const [activeTab, setActiveTab] = useState<'discover' | 'collection'>('discover');

  // Check backend server status
  const { data: healthData, isError: isHealthError } = useQuery({
    queryKey: ['health'],
    queryFn: api.checkHealth,
    refetchInterval: 30000,
  });

  // Query collection count for badge
  const { data: wishlistItems = [] } = useQuery({
    queryKey: ['wishlist'],
    queryFn: () => api.getWishlist(),
  });

  const serverStatus = isHealthError
    ? 'disconnected'
    : healthData
    ? 'connected'
    : 'checking';

  return (
    <div className="min-h-screen flex flex-col bg-canvas text-text-main transition-colors duration-200">
      <Navbar
        activeTab={activeTab}
        onSelectTab={setActiveTab}
        collectionCount={wishlistItems.length}
        serverStatus={serverStatus}
      />

      <main className="flex-1 max-w-7xl w-full mx-auto px-4 sm:px-6 lg:px-8 py-8 sm:py-12">
        <AnimatePresence mode="wait">
          {activeTab === 'discover' ? (
            <motion.section
              key="discover"
              initial={{ opacity: 0, y: 12 }}
              animate={{ opacity: 1, y: 0 }}
              exit={{ opacity: 0, y: -12 }}
              transition={{ duration: 0.22, ease: 'easeOut' }}
              className="space-y-8"
            >
              {/* Hero Header */}
              <div className="text-center max-w-2xl mx-auto space-y-4 pt-4 pb-6">
                <motion.div
                  initial={{ opacity: 0, scale: 0.95 }}
                  animate={{ opacity: 1, scale: 1 }}
                  transition={{ delay: 0.05, duration: 0.3 }}
                >
                  <span className="inline-flex items-center gap-1.5 px-3.5 py-1.5 rounded-full text-xs font-semibold bg-accent-soft text-accent border border-accent/20 shadow-xs">
                    <Sparkles className="w-3.5 h-3.5 text-amber-gold" /> Reader's Sanctuary
                  </span>
                </motion.div>

                <h1 className="font-serif text-4xl sm:text-5xl lg:text-6xl font-bold tracking-tight text-text-main leading-tight">
                  Discover Books You’ll Treasure.
                </h1>
                <p className="text-text-muted text-base sm:text-lg leading-relaxed">
                  Search works, inspect exact published editions, and curate your personal reading collection.
                </p>
              </div>

              {/* Architecture Preview Banner */}
              <motion.div
                initial={{ opacity: 0, y: 16 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ delay: 0.1, duration: 0.35 }}
                className="bg-surface rounded-2xl border border-border-subtle p-6 sm:p-8 shadow-book max-w-3xl mx-auto space-y-6"
              >
                <div className="flex items-center gap-3 border-b border-border-subtle pb-4">
                  <div className="w-9 h-9 rounded-lg bg-accent-soft flex items-center justify-center text-accent">
                    <Layers className="w-5 h-5" />
                  </div>
                  <div>
                    <h2 className="font-serif font-semibold text-xl text-text-main">
                      Dual Theme & Motion Architecture
                    </h2>
                    <p className="text-xs text-text-muted">
                      React 19 + Framer Motion + Tailwind v4 + Midnight Velvet / Antiquarian Modes
                    </p>
                  </div>
                </div>

                <div className="grid sm:grid-cols-3 gap-4 text-left">
                  <motion.div
                    whileHover={{ y: -3 }}
                    transition={{ duration: 0.15 }}
                    className="p-4 rounded-xl bg-canvas border border-border-subtle hover:border-accent/30 transition-colors"
                  >
                    <h3 className="font-semibold text-sm text-text-main mb-1 flex items-center gap-1.5">
                      <Search className="w-4 h-4 text-accent" /> Book Discovery
                    </h3>
                    <p className="text-xs text-text-muted">
                      Open Library integration backed by Go REST API with local PostgreSQL caching.
                    </p>
                  </motion.div>

                  <motion.div
                    whileHover={{ y: -3 }}
                    transition={{ duration: 0.15 }}
                    className="p-4 rounded-xl bg-canvas border border-border-subtle hover:border-accent/30 transition-colors"
                  >
                    <h3 className="font-semibold text-sm text-text-main mb-1 flex items-center gap-1.5">
                      <Layers className="w-4 h-4 text-accent" /> Work vs Edition
                    </h3>
                    <p className="text-xs text-text-muted">
                      Decoupled domain model architected for retailer edition price comparisons.
                    </p>
                  </motion.div>

                  <motion.div
                    whileHover={{ y: -3 }}
                    transition={{ duration: 0.15 }}
                    className="p-4 rounded-xl bg-canvas border border-border-subtle hover:border-accent/30 transition-colors"
                  >
                    <h3 className="font-semibold text-sm text-text-main mb-1 flex items-center gap-1.5">
                      <ShieldCheck className="w-4 h-4 text-accent" /> Type-Safe Core
                    </h3>
                    <p className="text-xs text-text-muted">
                      End-to-end synchronized types across Go structs and TypeScript interfaces.
                    </p>
                  </motion.div>
                </div>

                <div className="pt-2 flex flex-col sm:flex-row items-center justify-between gap-3 text-xs text-text-muted border-t border-border-subtle/60 pt-4">
                  <span>Toggle between tabs to see the sliding pill in action.</span>
                  <button
                    onClick={() => setActiveTab('collection')}
                    className="text-accent hover:underline flex items-center gap-1 font-medium cursor-pointer"
                  >
                    View My Collection <ArrowRight className="w-3.5 h-3.5" />
                  </button>
                </div>
              </motion.div>
            </motion.section>
          ) : (
            <motion.section
              key="collection"
              initial={{ opacity: 0, y: 12 }}
              animate={{ opacity: 1, y: 0 }}
              exit={{ opacity: 0, y: -12 }}
              transition={{ duration: 0.22, ease: 'easeOut' }}
              className="space-y-6"
            >
              <div className="flex items-center justify-between border-b border-border-subtle pb-4">
                <div>
                  <h1 className="font-serif text-3xl font-bold text-text-main">My Collection</h1>
                  <p className="text-sm text-text-muted">Organize your reading wishlist, priorities, and notes.</p>
                </div>
              </div>

              <motion.div
                initial={{ opacity: 0, scale: 0.97 }}
                animate={{ opacity: 1, scale: 1 }}
                transition={{ duration: 0.25 }}
                className="text-center py-16 bg-surface rounded-2xl border border-dashed border-border-subtle p-8 max-w-xl mx-auto"
              >
                <motion.div
                  animate={{ y: [0, -5, 0] }}
                  transition={{ repeat: Infinity, duration: 3.5, ease: 'easeInOut' }}
                >
                  <BookMarked className="w-12 h-12 text-accent/70 mx-auto mb-3" />
                </motion.div>
                <h2 className="font-serif font-semibold text-lg text-text-main mb-1">Your bookshelf is waiting</h2>
                <p className="text-sm text-text-muted mb-5">
                  Discover books in the catalog and click "Want to Read" to start building your personal library.
                </p>
                <motion.button
                  whileHover={{ scale: 1.04 }}
                  whileTap={{ scale: 0.96 }}
                  onClick={() => setActiveTab('discover')}
                  className="px-5 py-2.5 bg-accent text-canvas rounded-xl text-sm font-semibold hover:opacity-95 transition-opacity shadow-xs cursor-pointer inline-flex items-center gap-2"
                >
                  <span>Go to Discover</span>
                  <ArrowRight className="w-4 h-4" />
                </motion.button>
              </motion.div>
            </motion.section>
          )}
        </AnimatePresence>
      </main>

      {/* Footer */}
      <footer className="border-t border-border-subtle py-6 text-center text-xs text-text-muted">
        Imprint © 2026 — Crafted with React 19, Framer Motion, Tailwind CSS v4, and Go
      </footer>
    </div>
  );
}

export default function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <ImprintApp />
    </QueryClientProvider>
  );
}
