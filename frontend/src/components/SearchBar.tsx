import React, { useEffect, useState } from 'react';
import { Search, X, Loader2, BookKey, Sparkles } from 'lucide-react';

interface SearchBarProps {
  initialQuery?: string;
  onSearch: (query: string) => void;
  isLoading?: boolean;
}

const QUICK_SUGGESTIONS = [
  'Dune',
  'The Hobbit',
  'Frank Herbert',
  'Project Hail Mary',
  'Neuromancer',
  '1984',
];

export const SearchBar: React.FC<SearchBarProps> = ({
  initialQuery = '',
  onSearch,
  isLoading = false,
}) => {
  const [inputVal, setInputVal] = useState(initialQuery);

  // 350ms debounce effect
  useEffect(() => {
    const timer = setTimeout(() => {
      onSearch(inputVal.trim());
    }, 350);

    return () => clearTimeout(timer);
  }, [inputVal, onSearch]);

  const handleClear = () => {
    setInputVal('');
    onSearch('');
  };

  const handleChipClick = (term: string) => {
    setInputVal(term);
    onSearch(term);
  };

  // Detect if query resembles an ISBN
  const cleanNum = inputVal.replace(/[-\s]/g, '');
  const isLikelyISBN =
    (cleanNum.length === 10 || cleanNum.length === 13) && /^[0-9]+[0-9Xx]?$/.test(cleanNum);

  return (
    <div className="w-full max-w-2xl mx-auto space-y-3">
      <div className="relative flex items-center">
        <div className="absolute left-4 pointer-events-none text-text-muted flex items-center">
          {isLoading ? (
            <Loader2 className="w-5 h-5 animate-spin text-accent" />
          ) : (
            <Search className="w-5 h-5 text-accent" />
          )}
        </div>

        <input
          type="text"
          value={inputVal}
          onChange={(e) => setInputVal(e.target.value)}
          placeholder="Search by title, author, series, or ISBN..."
          className="w-full pl-12 pr-24 py-3.5 bg-surface text-text-main placeholder:text-text-muted rounded-2xl border border-border-subtle focus:border-accent focus:ring-2 focus:ring-accent/20 outline-none transition-all shadow-book text-base font-normal"
        />

        <div className="absolute right-3 flex items-center gap-1.5">
          {isLikelyISBN && (
            <span className="hidden sm:inline-flex items-center gap-1 px-2 py-0.5 rounded-md bg-accent-soft text-accent text-xs font-medium border border-accent/20">
              <BookKey className="w-3 h-3" /> ISBN
            </span>
          )}

          {inputVal && (
            <button
              onClick={handleClear}
              className="p-1 text-text-muted hover:text-text-main rounded-lg hover:bg-surface-hover transition-colors cursor-pointer"
              aria-label="Clear search"
            >
              <X className="w-4 h-4" />
            </button>
          )}
        </div>
      </div>

      {/* Quick Discovery Chips */}
      <div className="flex items-center gap-1.5 flex-wrap justify-center text-xs text-text-muted pt-1">
        <span className="inline-flex items-center gap-1 font-medium mr-1 text-text-muted/80">
          <Sparkles className="w-3 h-3 text-amber-gold" /> Popular:
        </span>
        {QUICK_SUGGESTIONS.map((term) => (
          <button
            key={term}
            onClick={() => handleChipClick(term)}
            className={`px-2.5 py-1 rounded-lg border transition-all cursor-pointer ${
              inputVal.toLowerCase() === term.toLowerCase()
                ? 'bg-accent text-canvas border-accent font-semibold shadow-xs'
                : 'bg-surface/80 border-border-subtle text-text-main hover:border-accent/40 hover:bg-surface-hover'
            }`}
          >
            {term}
          </button>
        ))}
      </div>
    </div>
  );
};
