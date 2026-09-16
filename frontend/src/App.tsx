import { useState, useMemo, useEffect, useCallback } from 'react';
import {
  QueryClient,
  QueryClientProvider,
  useQuery,
  useMutation,
  useQueryClient,
} from '@tanstack/react-query';
import { motion, AnimatePresence } from 'framer-motion';
import { Navbar } from './components/Navbar';
import { SearchBar } from './components/SearchBar';
import { BookGrid } from './components/BookGrid';
import { EditionModal } from './components/EditionModal';
import { CollectionView } from './components/CollectionView';
import { AuthModal } from './components/AuthModal';
import { ToastContainer } from './components/Toast';
import { AuthProvider, useAuth } from './context/AuthContext';
import { api } from './lib/api';
import { toast } from './lib/toast';
import { Sparkles, Layers, Search, ShieldCheck } from 'lucide-react';
import type { Work, Edition, ReadingStatus, WishlistItem } from './types/api';

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 1000 * 60 * 5, // 5 minutes cache
      retry: 1,
      refetchOnWindowFocus: false,
    },
  },
});

function getInitialTab(): 'discover' | 'collection' {
  if (typeof window === 'undefined') return 'discover';
  const hash = window.location.hash.toLowerCase();
  if (hash === '#collection') return 'collection';
  if (hash === '#discover') return 'discover';
  try {
    const saved = localStorage.getItem('imprint_active_tab');
    if (saved === 'collection' || saved === 'discover') return saved;
  } catch {
    // Storage access might fail in restricted iframe / sandbox
  }
  return 'discover';
}

function ImprintApp() {
  const { user, isAuthenticated, openAuthModal } = useAuth();
  const [activeTab, setActiveTab] = useState<'discover' | 'collection'>(getInitialTab);
  const [searchQuery, setSearchQuery] = useState('');
  const [selectedWork, setSelectedWork] = useState<Work | null>(null);
  const [activeWishlistItem, setActiveWishlistItem] = useState<WishlistItem | null>(null);
  const [isEditionModalOpen, setIsEditionModalOpen] = useState(false);

  const queryClientInstance = useQueryClient();

  const switchTab = useCallback((tab: 'discover' | 'collection') => {
    setActiveTab(tab);
    try {
      localStorage.setItem('imprint_active_tab', tab);
    } catch {
      // Ignore
    }
    const targetHash = tab === 'collection' ? '#collection' : '#discover';
    if (window.location.hash.toLowerCase() !== targetHash) {
      window.history.replaceState(null, '', targetHash);
    }
  }, []);

  // Listen for browser Back/Forward and manual URL hash modifications
  useEffect(() => {
    const handleLocationChange = () => {
      const hash = window.location.hash.toLowerCase();
      if (hash === '#collection') {
        setActiveTab('collection');
        try {
          localStorage.setItem('imprint_active_tab', 'collection');
        } catch {}
      } else if (hash === '#discover' || hash === '' || hash === '#') {
        setActiveTab('discover');
        try {
          localStorage.setItem('imprint_active_tab', 'discover');
        } catch {}
      }
    };

    window.addEventListener('hashchange', handleLocationChange);
    window.addEventListener('popstate', handleLocationChange);

    // Keep URL in sync with restored active tab if no hash was set
    const currentHash = window.location.hash.toLowerCase();
    if (!currentHash) {
      const initialHash = activeTab === 'collection' ? '#collection' : '#discover';
      window.history.replaceState(null, '', initialHash);
    }

    return () => {
      window.removeEventListener('hashchange', handleLocationChange);
      window.removeEventListener('popstate', handleLocationChange);
    };
  }, [activeTab]);

  // 1. Query user wishlist collection (Strictly behind authentication)
  const { data: rawWishlist = [], isLoading: isWishlistLoading } = useQuery({
    queryKey: ['wishlist', user?.id],
    queryFn: () => api.getWishlist(),
    enabled: isAuthenticated,
  });
  const wishlistItems = isAuthenticated && Array.isArray(rawWishlist) ? rawWishlist : [];

  // Fast lookup set of collection works
  const collectionWorkIds = useMemo(() => {
    const set = new Set<string>();
    wishlistItems.forEach((item) => {
      if (item.work_id) set.add(item.work_id);
      if (item.work?.id) set.add(item.work.id);
      if (item.work?.open_library_work_id) set.add(item.work.open_library_work_id);
    });
    return set;
  }, [wishlistItems]);

  // 3. Search books query with AbortSignal cancellation
  const {
    data: rawSearchResults = [],
    isLoading: isSearchLoading,
    isFetching: isSearchFetching,
  } = useQuery({
    queryKey: ['books', searchQuery],
    queryFn: ({ signal }) => api.searchBooks(searchQuery, 20, signal),
    enabled: Boolean(searchQuery.trim()),
    staleTime: 1000 * 60 * 10,
  });
  const searchResults = Array.isArray(rawSearchResults) ? rawSearchResults : [];

  // 4. Wishlist Mutations (Optimistic UI: 0ms visual feedback)
  const addMutation = useMutation({
    mutationFn: (variables: { work: Work; edition?: Edition }) => {
      const workId = variables.work.id || variables.work.open_library_work_id || '';
      const authorName =
        variables.work.authors && variables.work.authors.length > 0
          ? variables.work.authors[0].name
          : undefined;
      return api.addToWishlist({
        work_id: workId,
        edition_id: variables.edition?.id,
        title: variables.work.title,
        author: authorName,
        cover_url: variables.work.cover_url,
        original_year: variables.work.original_year,
        status: 'WANT_TO_READ',
        priority: 3,
      });
    },
    onMutate: async (variables) => {
      // Cancel outgoing refetches so they don't overwrite optimistic update
      await queryClientInstance.cancelQueries({ queryKey: ['wishlist', user?.id] });

      // Snapshot previous collection
      const previousWishlist =
        queryClientInstance.getQueryData<WishlistItem[]>(['wishlist', user?.id]) || [];

      // Create optimistic item with immediate presence in collectionWorkIds
      const tempId = `optimistic-${Date.now()}`;
      const optimisticItem: WishlistItem = {
        id: tempId,
        user_id: user?.id || 'guest',
        work_id: variables.work.id || variables.work.open_library_work_id || tempId,
        edition_id: variables.edition?.id,
        status: 'WANT_TO_READ',
        priority: 3,
        tags: [],
        work: variables.work,
        edition: variables.edition,
        created_at: new Date().toISOString(),
        updated_at: new Date().toISOString(),
      };

      // Instantly inject into cache: button turns green/saved in 0ms
      queryClientInstance.setQueryData<WishlistItem[]>(['wishlist', user?.id], (old = []) => [
        optimisticItem,
        ...old,
      ]);

      return { previousWishlist, tempId };
    },
    onSuccess: (newItem, variables, context) => {
      const resolvedItem: WishlistItem = {
        ...newItem,
        work: newItem.work || variables.work,
        edition: newItem.edition || variables.edition,
      };
      // Swap out optimistic temp item with persistent server item
      queryClientInstance.setQueryData<WishlistItem[]>(['wishlist', user?.id], (old = []) => [
        resolvedItem,
        ...old.filter((it) => it.id !== context?.tempId && it.id !== resolvedItem.id),
      ]);
      queryClientInstance.invalidateQueries({ queryKey: ['wishlist', user?.id] });
      toast.success('Added to Want to Read', variables.work.title);
    },
    onError: (err: Error, variables, context) => {
      // Revert optimistic change on network failure
      if (context?.previousWishlist) {
        queryClientInstance.setQueryData(['wishlist', user?.id], context.previousWishlist);
      }
      if (err.message.includes('already in your collection') || err.message.includes('duplicate')) {
        toast.info('Already in your collection', variables.work.title);
      } else {
        toast.error('Could not save book', err.message);
      }
    },
  });

  const updateMutation = useMutation({
    mutationFn: (variables: {
      id: string;
      status?: ReadingStatus;
      priority?: number;
      rating?: number;
      notes?: string;
      edition_id?: string;
      tags?: string[];
    }) => {
      return api.updateWishlistItem(variables.id, variables);
    },
    onMutate: async (variables) => {
      // 1. Cancel in-flight wishlist queries so background fetches don't overwrite optimistic updates
      await queryClientInstance.cancelQueries({ queryKey: ['wishlist', user?.id] });

      // 2. Snapshot current state for rollback on error
      const previousWishlist =
        queryClientInstance.getQueryData<WishlistItem[]>(['wishlist', user?.id]) || [];

      // 3. Apply optimistic update in 0ms directly to cache
      queryClientInstance.setQueryData<WishlistItem[]>(['wishlist', user?.id], (old = []) =>
        old.map((item) => {
          if (item.id !== variables.id) return item;
          return {
            ...item,
            ...(variables.priority !== undefined ? { priority: variables.priority } : {}),
            ...(variables.status !== undefined ? { status: variables.status } : {}),
            ...(variables.rating !== undefined ? { rating: variables.rating } : {}),
            ...(variables.notes !== undefined ? { notes: variables.notes } : {}),
            ...(variables.edition_id !== undefined ? { edition_id: variables.edition_id } : {}),
            ...(variables.tags !== undefined ? { tags: variables.tags } : {}),
            updated_at: new Date().toISOString(),
          };
        })
      );

      return { previousWishlist };
    },
    onSuccess: (updated) => {
      // Reconcile cache with confirmed server payload (preserving existing work/edition objects)
      queryClientInstance.setQueryData<WishlistItem[]>(['wishlist', user?.id], (old = []) =>
        old.map((it) => {
          if (it.id !== updated.id) return it;
          return {
            ...it,
            ...updated,
            work: updated.work || it.work,
            edition: updated.edition || it.edition,
          };
        })
      );
      // NOTE: Do not call invalidateQueries here. A remote refetch over the network would wipe
      // optimistic state and disrupt smooth layout reordering animations.
    },
    onError: (err: Error, _variables, context) => {
      if (context?.previousWishlist) {
        queryClientInstance.setQueryData(['wishlist', user?.id], context.previousWishlist);
      }
      toast.error('Failed to update item', err.message);
    },
  });

  const deleteMutation = useMutation({
    mutationFn: (id: string) => api.deleteWishlistItem(id),
    onMutate: async (deletedId) => {
      await queryClientInstance.cancelQueries({ queryKey: ['wishlist', user?.id] });
      const previousWishlist =
        queryClientInstance.getQueryData<WishlistItem[]>(['wishlist', user?.id]) || [];

      // Optimistically remove from collection in 0ms
      queryClientInstance.setQueryData<WishlistItem[]>(['wishlist', user?.id], (old = []) =>
        old.filter((it) => it.id !== deletedId)
      );

      return { previousWishlist };
    },
    onSuccess: () => {
      toast.info('Removed from collection');
    },
    onError: (err: Error, _deletedId, context) => {
      if (context?.previousWishlist) {
        queryClientInstance.setQueryData(['wishlist', user?.id], context.previousWishlist);
      }
      toast.error('Failed to remove item', err.message);
    },
  });

  // Handlers
  const handleToggleCollection = (work: Work) => {
    if (!isAuthenticated) {
      toast.info('Sign in required', 'Please sign in to save books to your collection');
      openAuthModal('login');
      return;
    }

    const isSaved =
      Boolean(work.id && collectionWorkIds.has(work.id)) ||
      Boolean(work.open_library_work_id && collectionWorkIds.has(work.open_library_work_id));

    if (isSaved) {
      switchTab('collection');
      toast.info('Opening book in collection', work.title);
    } else {
      addMutation.mutate({ work });
    }
  };

  const handleInspectEditions = (work: Work) => {
    setActiveWishlistItem(null);
    setSelectedWork(work);
    setIsEditionModalOpen(true);
  };

  const handleInspectWishlistEditions = (item: WishlistItem) => {
    setActiveWishlistItem(item);
    if (item.work) {
      setSelectedWork(item.work);
    } else {
      setSelectedWork({
        id: item.work_id,
        open_library_work_id: item.work_id.startsWith('OL') ? item.work_id : undefined,
        title: 'Book Details',
        authors: [],
      });
    }
    setIsEditionModalOpen(true);
  };

  const handleSelectWishlistEdition = (wishlistItemId: string, edition: Edition) => {
    updateMutation.mutate({
      id: wishlistItemId,
      edition_id: edition.id,
    });
    setActiveWishlistItem((prev) =>
      prev && prev.id === wishlistItemId ? { ...prev, edition_id: edition.id, edition } : prev
    );
    toast.success(
      'Edition updated',
      `${edition.format || 'Edition'} (${edition.publisher || 'Catalog'})`
    );
  };

  return (
    <div className="min-h-screen flex flex-col bg-canvas text-text-main transition-colors duration-200">
      <Navbar
        activeTab={activeTab}
        onSelectTab={switchTab}
        collectionCount={wishlistItems.length}
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
              <div className="text-center max-w-2xl mx-auto space-y-4 pt-2 pb-4">
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

              {/* Interactive Search Bar */}
              <SearchBar
                initialQuery={searchQuery}
                onSearch={setSearchQuery}
                isLoading={isSearchLoading || isSearchFetching}
              />

              {/* Book Results Grid */}
              <BookGrid
                works={searchResults}
                isLoading={isSearchLoading || (isSearchFetching && searchResults.length === 0)}
                collectionWorkIds={collectionWorkIds}
                onToggleCollection={handleToggleCollection}
                onInspectEditions={handleInspectEditions}
                searchQuery={searchQuery}
              />

              {/* Architecture Features Hint (When no search has been performed) */}
              {!searchQuery && searchResults.length === 0 && (
                <motion.div
                  initial={{ opacity: 0, y: 16 }}
                  animate={{ opacity: 1, y: 0 }}
                  transition={{ delay: 0.15, duration: 0.35 }}
                  className="bg-surface rounded-2xl border border-border-subtle p-6 sm:p-8 shadow-book max-w-3xl mx-auto space-y-6 mt-8"
                >
                  <div className="flex items-center gap-3 border-b border-border-subtle pb-4">
                    <div className="w-9 h-9 rounded-lg bg-accent-soft flex items-center justify-center text-accent">
                      <Layers className="w-5 h-5" />
                    </div>
                    <div>
                      <h2 className="font-serif font-semibold text-xl text-text-main">
                        Live Catalog & Edition Architecture
                      </h2>
                      <p className="text-xs text-text-muted">
                        Connected to Go REST API with Open Library & Google Books hybrid providers and Neon PostgreSQL caching
                      </p>
                    </div>
                  </div>

                  <div className="grid sm:grid-cols-3 gap-4 text-left">
                    <div className="p-4 rounded-xl bg-canvas border border-border-subtle">
                      <h3 className="font-semibold text-sm text-text-main mb-1 flex items-center gap-1.5">
                        <Search className="w-4 h-4 text-accent" /> Live Search
                      </h3>
                      <p className="text-xs text-text-muted">
                        Type any title above to query live bibliographic records.
                      </p>
                    </div>

                    <div className="p-4 rounded-xl bg-canvas border border-border-subtle">
                      <h3 className="font-semibold text-sm text-text-main mb-1 flex items-center gap-1.5">
                        <Layers className="w-4 h-4 text-accent" /> Work vs Edition
                      </h3>
                      <p className="text-xs text-text-muted">
                        Inspect exact ISBN-10, ISBN-13, page counts, and publishers.
                      </p>
                    </div>

                    <div className="p-4 rounded-xl bg-canvas border border-border-subtle">
                      <h3 className="font-semibold text-sm text-text-main mb-1 flex items-center gap-1.5">
                        <ShieldCheck className="w-4 h-4 text-accent" /> Type-Safe Core
                      </h3>
                      <p className="text-xs text-text-muted">
                        Zero data mismatch between Go domain models and TypeScript interfaces.
                      </p>
                    </div>
                  </div>
                </motion.div>
              )}
            </motion.section>
          ) : (
            <motion.section
              key="collection"
              initial={{ opacity: 0, y: 12 }}
              animate={{ opacity: 1, y: 0 }}
              exit={{ opacity: 0, y: -12 }}
              transition={{ duration: 0.22, ease: 'easeOut' }}
            >
              <CollectionView
                items={wishlistItems}
                isLoading={isWishlistLoading}
                onUpdateStatus={(id, status) => updateMutation.mutate({ id, status })}
                onUpdatePriority={(id, priority) => updateMutation.mutate({ id, priority })}
                onUpdateRating={(id, rating) => updateMutation.mutate({ id, rating })}
                onUpdateNotes={(id, notes) => updateMutation.mutate({ id, notes })}
                onUpdateTags={(id, tags) => updateMutation.mutate({ id, tags })}
                onDelete={(id) => deleteMutation.mutate(id)}
                onRefreshCollection={() =>
                  queryClientInstance.invalidateQueries({ queryKey: ['wishlist', user?.id] })
                }
                onGoToDiscover={() => switchTab('discover')}
                onInspectEditions={handleInspectWishlistEditions}
              />
            </motion.section>
          )}
        </AnimatePresence>
      </main>

      {/* Edition Inspection Modal */}
      <EditionModal
        work={selectedWork}
        isOpen={isEditionModalOpen}
        onClose={() => {
          setIsEditionModalOpen(false);
          setActiveWishlistItem(null);
        }}
        wishlistItem={activeWishlistItem}
        onSelectWishlistEdition={handleSelectWishlistEdition}
        onAddEditionToWishlist={(work, edition) => {
          if (!isAuthenticated) {
            toast.info('Sign in required', 'Please sign in to save books to your collection');
            openAuthModal('login');
            return;
          }
          addMutation.mutate({ work, edition });
        }}
        isWorkInCollection={
          Boolean(selectedWork?.id && collectionWorkIds.has(selectedWork.id)) ||
          Boolean(
            selectedWork?.open_library_work_id &&
              collectionWorkIds.has(selectedWork.open_library_work_id)
          )
        }
      />

      {/* Authentication Dialog Modal */}
      <AuthModal />

      {/* Global Floating Toast Container */}
      <ToastContainer />

      {/* Footer */}
      <footer className="border-t border-border-subtle py-6 text-center text-xs text-text-muted">
        Imprint © 2026
      </footer>
    </div>
  );
}

export default function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <AuthProvider>
        <ImprintApp />
      </AuthProvider>
    </QueryClientProvider>
  );
}
