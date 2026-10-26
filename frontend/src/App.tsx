import {
  useState,
  useMemo,
  useEffect,
  useCallback,
  lazy,
  Suspense,
} from "react";
import {
  QueryClient,
  QueryClientProvider,
  useQuery,
  useMutation,
  useQueryClient,
} from "@tanstack/react-query";
import { MotionConfig } from "framer-motion";
import { Navbar, type AppTab } from "./components/Navbar";
import { Discovery } from "./components/Discovery";
import { RequestError } from "./components/RequestState";
const ProfileStatsModal = lazy(() =>
  import("./components/ProfileStatsModal").then((module) => ({
    default: module.ProfileStatsModal,
  })),
);
const SettingsView = lazy(() =>
  import("./components/SettingsView").then((module) => ({
    default: module.SettingsView,
  })),
);
import { SearchBar } from "./components/SearchBar";
import { BookGrid } from "./components/BookGrid";
const EditionModal = lazy(() =>
  import("./components/EditionModal").then((module) => ({
    default: module.EditionModal,
  })),
);
import { CollectionView } from "./components/CollectionView";
import { AuthModal } from "./components/AuthModal";
import { ToastContainer } from "./components/Toast";
import { PublicProfileView } from "./components/PublicProfileView";
import { AuthProvider, useAuth } from "./context/AuthContext";
import { api } from "./lib/api";
import { toast } from "./lib/toast";

import type { Work, Edition, ReadingStatus, WishlistItem } from "./types/api";

const EMPTY_WISHLIST: WishlistItem[] = [];

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 1000 * 60 * 5, // 5 minutes cache
      retry: 1,
      refetchOnWindowFocus: false,
    },
  },
});

interface PublicRoute {
  username: string;
  shelf?: string;
}

function decodeRoute(value: string) {
  try {
    return decodeURIComponent(value);
  } catch {
    return value;
  }
}

function parsePublicRoute(): PublicRoute | null {
  if (typeof window === "undefined") return null;

  const pathname = window.location.pathname;
  if (pathname.startsWith("/u/")) {
    const parts = pathname.slice(3).split("/");
    const username = decodeRoute(parts[0] || "");
    if (username) {
      const shelfIdx = parts.indexOf("shelf");
      const shelf =
        shelfIdx !== -1 && parts[shelfIdx + 1]
          ? decodeRoute(parts[shelfIdx + 1])
          : undefined;
      return { username, shelf };
    }
  }

  const hash = window.location.hash;
  if (hash.startsWith("#/u/")) {
    const parts = hash.slice(4).split("/");
    const username = decodeRoute(parts[0] || "");
    if (username) {
      const shelfIdx = parts.indexOf("shelf");
      const shelf =
        shelfIdx !== -1 && parts[shelfIdx + 1]
          ? decodeRoute(parts[shelfIdx + 1])
          : undefined;
      return { username, shelf };
    }
  }

  return null;
}

function getInitialTab(): AppTab {
  if (typeof window === "undefined") return "discover";
  const hash = window.location.hash.toLowerCase();
  if (hash === "#collection") return "collection";
  if (hash.startsWith("#discover")) return "discover";
  if (hash === "#stats") return "stats";
  if (hash === "#settings") return "settings";
  try {
    const saved = localStorage.getItem("imprint_active_tab");
    if (
      saved === "collection" ||
      saved === "discover" ||
      saved === "stats" ||
      saved === "settings"
    )
      return saved;
  } catch {
    // Storage access might fail in restricted iframe / sandbox
  }
  return "discover";
}

function ImprintApp() {
  const { user, isAuthenticated, openAuthModal } = useAuth();
  const [activeTab, setActiveTab] = useState<AppTab>(getInitialTab);
  const [publicRoute, setPublicRoute] = useState<PublicRoute | null>(
    parsePublicRoute,
  );
  const [searchQuery, setSearchQuery] = useState(
    () =>
      new URLSearchParams(window.location.hash.split("?")[1]).get("q") || "",
  );
  const [searchLimit, setSearchLimit] = useState(20);
  const searchBooks = useCallback((query: string) => {
    setSearchQuery(query);
    setSearchLimit(20);
    const params = new URLSearchParams();
    if (query) params.set("q", query);
    window.history.replaceState(
      null,
      "",
      "/#discover" + (query ? "?" + params.toString() : ""),
    );
  }, []);
  const [selectedWork, setSelectedWork] = useState<Work | null>(null);
  const [activeWishlistItem, setActiveWishlistItem] =
    useState<WishlistItem | null>(null);
  const [isEditionModalOpen, setIsEditionModalOpen] = useState(false);

  const queryClientInstance = useQueryClient();

  const switchTab = useCallback(
    (tab: AppTab) => {
      setIsEditionModalOpen(false);
      setActiveTab(tab);
      setPublicRoute(null);
      try {
        localStorage.setItem("imprint_active_tab", tab);
      } catch {
        // Ignore
      }
      const targetHash =
        "#" +
        tab +
        (tab === "discover" && searchQuery
          ? "?" + new URLSearchParams({ q: searchQuery }).toString()
          : "");
      if (window.location.hash.toLowerCase() !== targetHash) {
        window.history.pushState(null, "", targetHash);
      }
      window.scrollTo({ top: 0 });
    },
    [searchQuery],
  );

  // Listen for browser Back/Forward and manual URL hash modifications
  useEffect(() => {
    const handleLocationChange = () => {
      const pub = parsePublicRoute();
      setPublicRoute(pub);

      if (!pub) {
        if (window.location.hash.startsWith("#book/")) {
          const id = decodeRoute(window.location.hash.slice(6));
          setSelectedWork({ id, title: "Book details" });
          setActiveWishlistItem(null);
          setIsEditionModalOpen(true);
          return;
        }
        setIsEditionModalOpen(false);
        const hash = window.location.hash.toLowerCase();
        if (hash === "#stats" || hash === "#settings") {
          setActiveTab(hash.slice(1) as AppTab);
        } else if (hash === "#collection") {
          setActiveTab("collection");
          try {
            localStorage.setItem("imprint_active_tab", "collection");
          } catch {}
        } else if (
          hash.startsWith("#discover") ||
          hash === "" ||
          hash === "#"
        ) {
          setActiveTab("discover");
          setSearchQuery(
            new URLSearchParams(window.location.hash.split("?")[1]).get("q") ||
              "",
          );
          try {
            localStorage.setItem("imprint_active_tab", "discover");
          } catch {}
        }
      }
    };

    window.addEventListener("hashchange", handleLocationChange);
    window.addEventListener("popstate", handleLocationChange);

    // Initial check
    const pub = parsePublicRoute();
    setPublicRoute(pub);
    if (!pub) {
      const currentHash = window.location.hash.toLowerCase();
      if (currentHash.startsWith("#book/")) handleLocationChange();
      if (!currentHash) {
        const initialHash = "#" + activeTab;
        window.history.replaceState(null, "", initialHash);
      }
    }

    return () => {
      window.removeEventListener("hashchange", handleLocationChange);
      window.removeEventListener("popstate", handleLocationChange);
    };
  }, [activeTab]);

  // 1. Query user wishlist collection (Strictly behind authentication)
  const {
    data: rawWishlist = EMPTY_WISHLIST,
    isLoading: isWishlistLoading,
    isError: isWishlistError,
  } = useQuery({
    queryKey: ["wishlist", user?.id],
    queryFn: () => api.getWishlist(),
    enabled: isAuthenticated,
  });
  const wishlistItems =
    isAuthenticated && Array.isArray(rawWishlist)
      ? rawWishlist
      : EMPTY_WISHLIST;

  // Fast lookup set of collection works
  const collectionWorkIds = useMemo(() => {
    const set = new Set<string>();
    wishlistItems.forEach((item) => {
      if (item.work_id) set.add(item.work_id);
      if (item.work?.id) set.add(item.work.id);
      if (item.work?.open_library_work_id)
        set.add(item.work.open_library_work_id);
    });
    return set;
  }, [wishlistItems]);

  // 3. Search books query with AbortSignal cancellation
  const {
    data: rawSearchResults = [],
    isLoading: isSearchLoading,
    isFetching: isSearchFetching,
    isError: isSearchError,
    refetch: refetchSearch,
  } = useQuery({
    queryKey: ["books", searchQuery, searchLimit],
    queryFn: ({ signal }) => api.searchBooks(searchQuery, searchLimit, signal),
    enabled: Boolean(searchQuery.trim()),
    staleTime: 1000 * 60 * 10,
  });
  const searchResults = Array.isArray(rawSearchResults) ? rawSearchResults : [];

  // 4. Wishlist Mutations (Optimistic UI: 0ms visual feedback)
  const addMutation = useMutation({
    mutationFn: (variables: { work: Work; edition?: Edition }) => {
      const workId =
        variables.work.id || variables.work.open_library_work_id || "";
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
        status: "WANT_TO_READ",
        priority: 3,
      });
    },
    onMutate: async (variables) => {
      // Cancel outgoing refetches so they don't overwrite optimistic update
      await queryClientInstance.cancelQueries({
        queryKey: ["wishlist", user?.id],
      });

      // Snapshot previous collection
      const previousWishlist =
        queryClientInstance.getQueryData<WishlistItem[]>([
          "wishlist",
          user?.id,
        ]) || [];

      // Create optimistic item with immediate presence in collectionWorkIds
      const tempId = `optimistic-${Date.now()}`;
      const optimisticItem: WishlistItem = {
        id: tempId,
        user_id: user?.id || "guest",
        work_id:
          variables.work.id || variables.work.open_library_work_id || tempId,
        edition_id: variables.edition?.id,
        status: "WANT_TO_READ",
        priority: 3,
        tags: [],
        work: variables.work,
        edition: variables.edition,
        created_at: new Date().toISOString(),
        updated_at: new Date().toISOString(),
      };

      // Instantly inject into cache: button turns green/saved in 0ms
      queryClientInstance.setQueryData<WishlistItem[]>(
        ["wishlist", user?.id],
        (old = []) => [optimisticItem, ...old],
      );

      return { previousWishlist, tempId };
    },
    onSuccess: (newItem, variables, context) => {
      const resolvedItem: WishlistItem = {
        ...newItem,
        work: newItem.work || variables.work,
        edition: newItem.edition || variables.edition,
      };
      // Swap out optimistic temp item with persistent server item
      queryClientInstance.setQueryData<WishlistItem[]>(
        ["wishlist", user?.id],
        (old = []) => [
          resolvedItem,
          ...old.filter(
            (it) => it.id !== context?.tempId && it.id !== resolvedItem.id,
          ),
        ],
      );
      queryClientInstance.invalidateQueries({
        queryKey: ["wishlist", user?.id],
      });
      toast.success("Added to Want to Read", variables.work.title);
    },
    onError: (err: Error, variables, context) => {
      // Revert optimistic change on network failure
      if (context?.previousWishlist) {
        queryClientInstance.setQueryData(
          ["wishlist", user?.id],
          context.previousWishlist,
        );
      }
      if (
        err.message.includes("already in your collection") ||
        err.message.includes("duplicate")
      ) {
        toast.info("Already in your collection", variables.work.title);
      } else {
        toast.error("Could not save book", err.message);
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
      await queryClientInstance.cancelQueries({
        queryKey: ["wishlist", user?.id],
      });

      // 2. Snapshot current state for rollback on error
      const previousWishlist =
        queryClientInstance.getQueryData<WishlistItem[]>([
          "wishlist",
          user?.id,
        ]) || [];

      // 3. Apply optimistic update in 0ms directly to cache
      queryClientInstance.setQueryData<WishlistItem[]>(
        ["wishlist", user?.id],
        (old = []) =>
          old.map((item) => {
            if (item.id !== variables.id) return item;
            return {
              ...item,
              ...(variables.priority !== undefined
                ? { priority: variables.priority }
                : {}),
              ...(variables.status !== undefined
                ? { status: variables.status }
                : {}),
              ...(variables.rating !== undefined
                ? { rating: variables.rating }
                : {}),
              ...(variables.notes !== undefined
                ? { notes: variables.notes }
                : {}),
              ...(variables.edition_id !== undefined
                ? { edition_id: variables.edition_id }
                : {}),
              ...(variables.tags !== undefined ? { tags: variables.tags } : {}),
              updated_at: new Date().toISOString(),
            };
          }),
      );

      return { previousWishlist };
    },
    onSuccess: (updated) => {
      // Reconcile cache with confirmed server payload (preserving existing work/edition objects)
      queryClientInstance.setQueryData<WishlistItem[]>(
        ["wishlist", user?.id],
        (old = []) =>
          old.map((it) => {
            if (it.id !== updated.id) return it;
            return {
              ...it,
              ...updated,
              work: updated.work || it.work,
              edition: updated.edition || it.edition,
            };
          }),
      );
      // NOTE: Do not call invalidateQueries here. A remote refetch over the network would wipe
      // optimistic state and disrupt smooth layout reordering animations.
    },
    onError: (err: Error, _variables, context) => {
      if (context?.previousWishlist) {
        queryClientInstance.setQueryData(
          ["wishlist", user?.id],
          context.previousWishlist,
        );
      }
      toast.error("Failed to update item", err.message);
    },
  });

  const deleteMutation = useMutation({
    mutationFn: (id: string) => api.deleteWishlistItem(id),
    onMutate: async (deletedId) => {
      await queryClientInstance.cancelQueries({
        queryKey: ["wishlist", user?.id],
      });
      const previousWishlist =
        queryClientInstance.getQueryData<WishlistItem[]>([
          "wishlist",
          user?.id,
        ]) || [];

      // Optimistically remove from collection in 0ms
      queryClientInstance.setQueryData<WishlistItem[]>(
        ["wishlist", user?.id],
        (old = []) => old.filter((it) => it.id !== deletedId),
      );

      return { previousWishlist };
    },
    onSuccess: () => {
      toast.info("Removed from collection");
    },
    onError: (err: Error, _deletedId, context) => {
      if (context?.previousWishlist) {
        queryClientInstance.setQueryData(
          ["wishlist", user?.id],
          context.previousWishlist,
        );
      }
      toast.error("Failed to remove item", err.message);
    },
  });

  // Handlers
  const handleToggleCollection = (work: Work) => {
    if (!isAuthenticated) {
      toast.info(
        "Sign in required",
        "Please sign in to save books to your collection",
      );
      openAuthModal("login");
      return;
    }

    const isSaved =
      Boolean(work.id && collectionWorkIds.has(work.id)) ||
      Boolean(
        work.open_library_work_id &&
        collectionWorkIds.has(work.open_library_work_id),
      );

    if (isSaved) {
      switchTab("collection");
      toast.info("Opening book in collection", work.title);
    } else {
      addMutation.mutate({ work });
    }
  };

  const handleInspectEditions = (work: Work) => {
    window.history.pushState(
      null,
      "",
      "#book/" + encodeURIComponent(work.id || work.open_library_work_id || ""),
    );
    window.scrollTo({ top: 0 });
    setActiveWishlistItem(null);
    setSelectedWork(work);
    setIsEditionModalOpen(true);
  };

  const handleInspectWishlistEditions = (item: WishlistItem) => {
    window.history.pushState(
      null,
      "",
      "#book/" + encodeURIComponent(item.work_id),
    );
    window.scrollTo({ top: 0 });
    setActiveWishlistItem(item);
    if (item.work) {
      setSelectedWork(item.work);
    } else {
      setSelectedWork({
        id: item.work_id,
        open_library_work_id: item.work_id.startsWith("OL")
          ? item.work_id
          : undefined,
        title: "Book Details",
        authors: [],
      });
    }
    setIsEditionModalOpen(true);
  };

  const handleSelectWishlistEdition = (
    wishlistItemId: string,
    edition: Edition,
  ) => {
    updateMutation.mutate(
      { id: wishlistItemId, edition_id: edition.id },
      {
        onSuccess: () => {
          setActiveWishlistItem((previous) =>
            previous?.id === wishlistItemId
              ? { ...previous, edition_id: edition.id, edition }
              : previous,
          );
          toast.success("Edition updated");
        },
      },
    );
  };

  if (publicRoute) {
    return (
      <div
        key={`public:${publicRoute.username}:${isEditionModalOpen ? selectedWork?.id : "profile"}`}
        className="page-transition min-h-screen bg-bg text-text-main transition-colors duration-300"
      >
        {!isEditionModalOpen && (
          <PublicProfileView
            username={publicRoute.username}
            initialShelf={publicRoute.shelf}
            onNavigateHome={() => {
              setPublicRoute(null);
              window.history.pushState(null, "", "/#discover");
              setActiveTab("discover");
            }}
            onInspectWork={(work) => {
              setSelectedWork(work);
              setIsEditionModalOpen(true);
            }}
          />
        )}

        {/* Book detail page for public viewers */}
        <div className="app-main">
          <Suspense
            fallback={
              <div
                className="skeleton"
                style={{ height: 240 }}
                aria-label="Loading book"
              />
            }
          >
            <EditionModal
              work={selectedWork}
              isOpen={isEditionModalOpen}
              onClose={() => {
                setIsEditionModalOpen(false);
                setActiveWishlistItem(null);
              }}
              wishlistItem={null}
              onSelectWishlistEdition={() => {}}
              onAddEditionToWishlist={(work, edition) => {
                if (!isAuthenticated) {
                  toast.info(
                    "Sign in required",
                    "Please sign in to save books to your collection",
                  );
                  openAuthModal("login");
                  return;
                }
                addMutation.mutate({ work, edition });
              }}
              isWorkInCollection={
                Boolean(
                  selectedWork?.id && collectionWorkIds.has(selectedWork.id),
                ) ||
                Boolean(
                  selectedWork?.open_library_work_id &&
                  collectionWorkIds.has(selectedWork.open_library_work_id),
                )
              }
            />
          </Suspense>
        </div>

        <AuthModal />
        <ToastContainer />
      </div>
    );
  }

  return (
    <div className="min-h-screen flex flex-col bg-canvas text-text-main transition-colors duration-200">
      <a
        className="skip-link"
        href="#main-content"
        onClick={(event) => {
          event.preventDefault();
          document.getElementById("main-content")?.focus();
        }}
      >
        Skip to content
      </a>
      <Navbar
        activeTab={activeTab}
        onSelectTab={switchTab}
        collectionCount={wishlistItems.length}
      />

      <main id="main-content" tabIndex={-1} className="app-main">
        <Suspense
          fallback={
            <div
              className="skeleton"
              style={{ height: 240 }}
              aria-label="Loading page"
            />
          }
        >
          <div
            key={
              isEditionModalOpen
                ? `book:${selectedWork?.id || selectedWork?.open_library_work_id}`
                : activeTab
            }
            className="page-transition"
          >
            {isEditionModalOpen ? (
              <EditionModal
                work={selectedWork}
                isOpen={isEditionModalOpen}
                onClose={() => switchTab(activeTab)}
                wishlistItem={
                  activeWishlistItem
                    ? wishlistItems.find(
                        (item) => item.id === activeWishlistItem.id,
                      ) || activeWishlistItem
                    : wishlistItems.find(
                        (item) =>
                          item.work_id === selectedWork?.id ||
                          (selectedWork?.open_library_work_id &&
                            item.work?.open_library_work_id ===
                              selectedWork.open_library_work_id),
                      ) || null
                }
                onSelectWishlistEdition={handleSelectWishlistEdition}
                onAddEditionToWishlist={(work, edition) => {
                  if (!isAuthenticated) {
                    openAuthModal("login");
                    return;
                  }
                  addMutation.mutate({ work, edition });
                }}
                isWorkInCollection={
                  Boolean(
                    selectedWork?.id && collectionWorkIds.has(selectedWork.id),
                  ) ||
                  Boolean(
                    selectedWork?.open_library_work_id &&
                    collectionWorkIds.has(selectedWork.open_library_work_id),
                  )
                }
              />
            ) : activeTab === "discover" ? (
              <section>
                <div className="page-heading">
                  <div>
                    <h1>
                      {searchQuery
                        ? "Find your next read."
                        : "Your next chapter."}
                    </h1>
                    <p>
                      {searchQuery
                        ? "A familiar favorite or a new discovery. Make it yours."
                        : "Good books. Beautiful editions. A collection that’s yours."}
                    </p>
                  </div>
                  {!searchQuery && (
                    <p className="reading-note">
                      For the books you find.
                      <br />
                      And the ones that find you.
                    </p>
                  )}
                </div>
                <SearchBar
                  initialQuery={searchQuery}
                  onSearch={searchBooks}
                  isLoading={isSearchLoading || isSearchFetching}
                />
                {!searchQuery ? (
                  <Discovery
                    onSearch={searchBooks}
                    onCollection={() => switchTab("collection")}
                  />
                ) : isSearchError ? (
                  <RequestError
                    title="The search couldn’t load."
                    onRetry={() => void refetchSearch()}
                  />
                ) : (
                  <>
                    <BookGrid
                      works={searchResults}
                      isLoading={
                        isSearchLoading ||
                        (isSearchFetching && !searchResults.length)
                      }
                      collectionWorkIds={collectionWorkIds}
                      onToggleCollection={handleToggleCollection}
                      onInspectEditions={handleInspectEditions}
                      searchQuery={searchQuery}
                    />
                    {searchResults.length >= searchLimit &&
                      searchLimit < 50 && (
                        <div className="load-more">
                          <button
                            className="button secondary"
                            disabled={isSearchFetching}
                            onClick={() => setSearchLimit(50)}
                          >
                            {isSearchFetching ? "Loading…" : "Show more books"}
                          </button>
                        </div>
                      )}
                    {searchResults.length >= 50 && (
                      <p className="search-limit-note">
                        Showing up to 50 matches. Refine your search to find a
                        specific book.
                      </p>
                    )}
                  </>
                )}
              </section>
            ) : activeTab === "collection" ? (
              <CollectionView
                items={wishlistItems}
                isLoading={isWishlistLoading}
                error={isWishlistError}
                onUpdateStatus={(id, status) =>
                  updateMutation.mutate({ id, status })
                }
                onUpdatePriority={(id, priority) =>
                  updateMutation.mutate({ id, priority })
                }
                onUpdateRating={(id, rating) =>
                  updateMutation.mutate({ id, rating })
                }
                onUpdateNotes={(id, notes) =>
                  updateMutation.mutate({ id, notes })
                }
                onUpdateTags={(id, tags) => updateMutation.mutate({ id, tags })}
                onDelete={(id) => deleteMutation.mutate(id)}
                onRefreshCollection={() =>
                  void queryClientInstance.invalidateQueries({
                    queryKey: ["wishlist", user?.id],
                  })
                }
                onGoToDiscover={() => switchTab("discover")}
                onInspectEditions={handleInspectWishlistEditions}
              />
            ) : activeTab === "stats" ? (
              <ProfileStatsModal
                isOpen
                onClose={() => switchTab("collection")}
                userName={user?.display_name || user?.username || "Reader"}
                embedded
              />
            ) : (
              <SettingsView
                onRefreshCollection={() =>
                  void queryClientInstance.invalidateQueries({
                    queryKey: ["wishlist", user?.id],
                  })
                }
              />
            )}
          </div>
        </Suspense>
      </main>

      {/* Authentication Dialog Modal */}
      <AuthModal />

      {/* Global Floating Toast Container */}
      <ToastContainer />

      {/* Footer */}
      <footer className="site-footer">
        <span>Imprint © {new Date().getFullYear()}</span>
        <span>A little space for a good book.</span>
      </footer>
    </div>
  );
}

export default function App() {
  return (
    <MotionConfig reducedMotion="user">
      <QueryClientProvider client={queryClient}>
        <AuthProvider>
          <ImprintApp />
        </AuthProvider>
      </QueryClientProvider>
    </MotionConfig>
  );
}
