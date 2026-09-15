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
  notes?: string;
  started_at?: string;
  finished_at?: string;
  created_at: string;
  updated_at: string;
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

