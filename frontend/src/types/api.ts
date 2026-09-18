export type ReadingStatus =
  | 'WANT_TO_READ'
  | 'CURRENTLY_READING'
  | 'FINISHED'
  | 'ABANDONED';

export type BookFormat =
  | 'HARDCOVER'
  | 'PAPERBACK'
  | 'MASS_MARKET'
  | 'EBOOK'
  | 'AUDIOBOOK'
  | 'UNKNOWN';

export interface Author {
  id?: string;
  name: string;
  bio?: string;
  open_library_id?: string;
}

export interface Work {
  id?: string;
  title: string;
  subtitle?: string;
  original_year?: number;
  description?: string;
  cover_url?: string;
  open_library_work_id?: string;
  subject_tags?: string[];
  authors?: Author[];
  created_at?: string;
  updated_at?: string;
}

export interface Edition {
  id?: string;
  work_id?: string;
  title: string;
  publisher?: string;
  publication_date?: string;
  publication_year?: number;
  page_count?: number;
  language?: string;
  format: BookFormat;
  isbn10?: string;
  isbn13?: string;
  asin?: string;
  cover_url?: string;
  description?: string;
  open_library_edition_id?: string;
  created_at?: string;
  updated_at?: string;
}

export interface WishlistItem {
  id: string;
  user_id: string;
  work_id: string;
  work?: Work;
  edition_id?: string;
  edition?: Edition;
  status: ReadingStatus;
  priority: number; // 1 to 5
  rating?: number;  // 1 to 5
  tags: string[];
  notes?: string;
  started_at?: string;
  finished_at?: string;
  created_at: string;
  updated_at: string;
}

export interface AddWishlistPayload {
  work_id: string;
  edition_id?: string;
  status?: ReadingStatus;
  priority?: number;
  rating?: number;
  tags?: string[];
  notes?: string;
  title?: string;
  author?: string;
  cover_url?: string;
  original_year?: number;
}

export interface UpdateWishlistPayload {
  edition_id?: string;
  status?: ReadingStatus;
  priority?: number;
  rating?: number;
  tags?: string[];
  notes?: string;
}

export interface TagCount {
  tag: string;
  count: number;
}

export interface GenreCount {
  genre: string;
  count: number;
}

export interface AuthorCount {
  author: string;
  count: number;
}

export interface MonthlyReadingProgress {
  month: number; // 1 = Jan .. 12 = Dec
  books: number;
  pages: number;
}

export interface ReadingGoal {
  id: string;
  user_id: string;
  year: number;
  target_books: number;
  created_at: string;
  updated_at: string;
}

export interface ReadingChallenge {
  year: number;
  target_books: number;
  books_finished: number;
  percentage: number;
  days_elapsed: number;
  total_days: number;
  expected_finished: number;
  pacing_diff: number;
  pacing_status: 'AHEAD' | 'ON_TRACK' | 'BEHIND' | 'COMPLETED' | 'NOT_SET';
  pacing_message: string;
  monthly_progress: MonthlyReadingProgress[];
}

export interface ReadingStats {
  total_books: number;
  books_finished_year: number;
  total_pages_read: number;
  currently_reading: number;
  want_to_read: number;
  average_rating: number;
  top_genres: GenreCount[];
  top_authors: AuthorCount[];
  format_distribution: Record<string, number>;
  current_year: number;
  challenge?: ReadingChallenge;
}

export interface GoodreadsImportSummary {
  total_rows: number;
  imported_count: number;
  skipped_count: number;
  failed_count: number;
}

export interface ApiResponse<T> {
  data: T;
}

export interface ApiError {
  error: {
    code: string;
    message: string;
    request_id?: string;
  };
}

export interface User {
  id: string;
  email: string;
  username: string;
  display_name: string;
  created_at: string;
  updated_at: string;
}

export interface AuthTokens {
  access_token: string;
  refresh_token?: string;
  expires_at: string;
  user: User;
}

export interface RegisterRequest {
  email: string;
  username: string;
  password: string;
  display_name?: string;
}

export interface LoginRequest {
  login: string;
  email_or_username?: string;
  password: string;
}
