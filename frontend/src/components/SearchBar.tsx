import { useEffect, useState } from "react";
import { Search, X, Loader2 } from "lucide-react";
interface SearchBarProps {
  initialQuery?: string;
  onSearch: (query: string) => void;
  isLoading?: boolean;
}
const suggestions = [
  "Dune",
  "The Hobbit",
  "Frank Herbert",
  "Project Hail Mary",
  "Neuromancer",
  "1984",
];
export function SearchBar({
  initialQuery = "",
  onSearch,
  isLoading = false,
}: SearchBarProps) {
  const [input, setInput] = useState(initialQuery);
  const [previousQuery, setPreviousQuery] = useState(initialQuery);
  if (previousQuery !== initialQuery) {
    setPreviousQuery(initialQuery);
    setInput(initialQuery);
  }
  useEffect(() => {
    const timer = setTimeout(() => {
      if (input.trim() !== initialQuery) onSearch(input.trim());
    }, 350);
    return () => clearTimeout(timer);
  }, [input, initialQuery, onSearch]);
  useEffect(() => {
    if (initialQuery.trim().length < 3) return;
    try {
      const recent = readRecentSearches();
      localStorage.setItem(
        "imprint_recent_searches",
        JSON.stringify(
          [
            initialQuery,
            ...recent.filter((term) => term !== initialQuery),
          ].slice(0, 4),
        ),
      );
    } catch {
      /* Search remains available when storage is restricted. */
    }
  }, [initialQuery]);
  const recentSearches = readRecentSearches();
  const isbn = /^(?:\d{9}[\dXx]|\d{13})$/.test(input.replace(/[-\s]/g, ""));
  return (
    <div className="search-area">
      <form
        className="catalog-search"
        role="search"
        onSubmit={(e) => {
          e.preventDefault();
          onSearch(input.trim());
        }}
      >
        {isLoading ? (
          <Loader2 className="animate-spin" size={20} />
        ) : (
          <Search size={20} />
        )}
        <input
          aria-label="Search books"
          type="search"
          value={input}
          onChange={(e) => setInput(e.target.value)}
          placeholder="Search by title, author, series, or ISBN..."
        />
        {isbn && <span className="isbn-label">ISBN</span>}
        {input && (
          <button
            type="button"
            className="icon-button"
            aria-label="Clear search"
            onClick={() => {
              setInput("");
              onSearch("");
            }}
          >
            <X size={18} />
          </button>
        )}
        <button className="search-submit" type="submit">
          Search
        </button>
      </form>
      {!initialQuery && recentSearches.length > 0 && (
        <div className="recent-searches">
          <span>Recent searches</span>
          {recentSearches.map((term) => (
            <button
              key={term}
              onClick={() => {
                setInput(term);
                onSearch(term);
              }}
            >
              {term}
            </button>
          ))}
        </div>
      )}
      <div className="search-suggestions">
        <span>Try a favorite</span>
        {suggestions.map((term) => (
          <button
            key={term}
            onClick={() => {
              setInput(term);
              onSearch(term);
            }}
          >
            {term}
          </button>
        ))}
      </div>
    </div>
  );
}

function readRecentSearches(): string[] {
  try {
    const value: unknown = JSON.parse(
      localStorage.getItem("imprint_recent_searches") || "[]",
    );
    return Array.isArray(value)
      ? value
          .filter((term: unknown): term is string => typeof term === "string")
          .slice(0, 4)
      : [];
  } catch {
    return [];
  }
}
