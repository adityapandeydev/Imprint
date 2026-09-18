import type {
  AddWishlistPayload,
  ApiResponse,
  AuthTokens,
  Edition,
  GoodreadsImportSummary,
  LoginRequest,
  ReadingGoal,
  ReadingStats,
  ReadingStatus,
  RegisterRequest,
  TagCount,
  UpdateWishlistPayload,
  User,
  WishlistItem,
  Work,
} from '../types/api';

const API_BASE = '/api/v1';
const TOKEN_KEY = 'imprint_token';
const REFRESH_TOKEN_KEY = 'imprint_refresh_token';

export const tokenStorage = {
  get: (): string | null => {
    try {
      return localStorage.getItem(TOKEN_KEY);
    } catch {
      return null;
    }
  },
  getRefresh: (): string | null => {
    try {
      return localStorage.getItem(REFRESH_TOKEN_KEY);
    } catch {
      return null;
    }
  },
  set: (accessToken: string, refreshToken?: string): void => {
    try {
      localStorage.setItem(TOKEN_KEY, accessToken);
      if (refreshToken) {
        localStorage.setItem(REFRESH_TOKEN_KEY, refreshToken);
      }
    } catch {
      // ignore
    }
  },
  clear: (): void => {
    try {
      localStorage.removeItem(TOKEN_KEY);
      localStorage.removeItem(REFRESH_TOKEN_KEY);
    } catch {
      // ignore
    }
  },
};

function getAuthHeaders(customHeaders: Record<string, string> = {}): Record<string, string> {
  const headers: Record<string, string> = { ...customHeaders };
  const token = tokenStorage.get();
  if (token) {
    headers['Authorization'] = `Bearer ${token}`;
  }
  return headers;
}

let isRefreshing = false;
let refreshPromise: Promise<string | null> | null = null;

async function refreshAccessToken(): Promise<string | null> {
  const refreshToken = tokenStorage.getRefresh();
  if (!refreshToken) {
    tokenStorage.clear();
    return null;
  }

  if (isRefreshing && refreshPromise) {
    return refreshPromise;
  }

  isRefreshing = true;
  refreshPromise = (async () => {
    try {
      const res = await fetch(`${API_BASE}/auth/refresh`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        credentials: 'include',
        body: JSON.stringify({ refresh_token: refreshToken }),
      });

      if (!res.ok) {
        tokenStorage.clear();
        return null;
      }

      const json: ApiResponse<AuthTokens> = await res.json();
      const newTokens = json.data;
      if (newTokens?.access_token) {
        tokenStorage.set(newTokens.access_token, newTokens.refresh_token);
        return newTokens.access_token;
      }
      tokenStorage.clear();
      return null;
    } catch {
      tokenStorage.clear();
      return null;
    } finally {
      isRefreshing = false;
      refreshPromise = null;
    }
  })();

  return refreshPromise;
}

async function fetchWithAuth(url: string, options: RequestInit = {}): Promise<Response> {
  const customHeaders = (options.headers as Record<string, string>) || {};
  const headers = getAuthHeaders(customHeaders);
  const opts: RequestInit = { ...options, headers, credentials: 'include' };

  let res = await fetch(url, opts);

  // If unauthorized and we have a refresh token, attempt automatic token rotation
  if (res.status === 401 && tokenStorage.getRefresh()) {
    const newToken = await refreshAccessToken();
    if (newToken) {
      const retryHeaders = { ...headers, Authorization: `Bearer ${newToken}` };
      res = await fetch(url, { ...opts, headers: retryHeaders });
    }
  }

  return res;
}

async function handleResponse<T>(res: Response): Promise<T> {
  if (!res.ok) {
    let errorMsg = `HTTP Error ${res.status}`;
    try {
      const errPayload = await res.json();
      if (errPayload?.error?.message) {
        errorMsg = errPayload.error.message;
      }
    } catch {
      // Use fallback errorMsg
    }
    throw new Error(errorMsg);
  }
  if (res.status === 204) {
    return {} as T;
  }
  const json: ApiResponse<T> = await res.json();
  return json.data;
}

export const api = {
  // Check backend server liveness
  async checkHealth(): Promise<{ status: string; database?: string }> {
    const res = await fetch('/health/ready', { credentials: 'include' });
    if (!res.ok) {
      const liveRes = await fetch('/health', { credentials: 'include' });
      if (!liveRes.ok) throw new Error('Backend server is unreachable');
      return { status: 'live', database: 'disconnected' };
    }
    const json = await res.json();
    return json.data;
  },

  // Auth: Register new user
  async register(req: RegisterRequest): Promise<AuthTokens> {
    const res = await fetch(`${API_BASE}/auth/register`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      credentials: 'include',
      body: JSON.stringify(req),
    });
    const tokens = await handleResponse<AuthTokens>(res);
    if (tokens?.access_token) {
      tokenStorage.set(tokens.access_token, tokens.refresh_token);
    }
    return tokens;
  },

  // Auth: Log in existing user
  async login(req: LoginRequest): Promise<AuthTokens> {
    const payload = {
      login: req.login,
      email_or_username: req.email_or_username || req.login,
      password: req.password,
    };
    const res = await fetch(`${API_BASE}/auth/login`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      credentials: 'include',
      body: JSON.stringify(payload),
    });
    const tokens = await handleResponse<AuthTokens>(res);
    if (tokens?.access_token) {
      tokenStorage.set(tokens.access_token, tokens.refresh_token);
    }
    return tokens;
  },

  // Auth: Request password reset token
  async forgotPassword(email: string): Promise<{ message: string; reset_token?: string }> {
    const res = await fetch(`${API_BASE}/auth/forgot-password`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ email }),
    });
    return handleResponse<{ message: string; reset_token?: string }>(res);
  },

  // Auth: Submit new password with recovery token
  async resetPassword(token: string, newPassword: string): Promise<{ message: string }> {
    const res = await fetch(`${API_BASE}/auth/reset-password`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ token, new_password: newPassword }),
    });
    return handleResponse<{ message: string }>(res);
  },

  // Auth: Log out user
  async logout(): Promise<void> {
    try {
      await fetchWithAuth(`${API_BASE}/auth/logout`, {
        method: 'POST',
      });
    } finally {
      tokenStorage.clear();
    }
  },

  // Auth: Get current authenticated user profile
  async getMe(): Promise<User> {
    const res = await fetchWithAuth(`${API_BASE}/auth/me`);
    return handleResponse<User>(res);
  },

  // Discover books by title, author, or keyword
  async searchBooks(query: string, limit = 20, signal?: AbortSignal): Promise<Work[]> {
    if (!query.trim()) return [];
    const params = new URLSearchParams({ q: query, limit: String(limit) });
    const res = await fetchWithAuth(`${API_BASE}/books/search?${params.toString()}`, {
      signal,
    });
    const data = await handleResponse<Work[]>(res);
    return Array.isArray(data) ? data : [];
  },

  // Get full work metadata and published editions
  async getBook(id: string): Promise<{ work: Work; editions: Edition[] }> {
    const res = await fetchWithAuth(`${API_BASE}/books/${encodeURIComponent(id)}`);
    const data = await handleResponse<{ work: Work; editions: Edition[] }>(res);
    return {
      work: data?.work,
      editions: Array.isArray(data?.editions) ? data.editions : [],
    };
  },

  // Lookup edition and parent work by ISBN
  async getEditionByISBN(isbn: string): Promise<{ edition: Edition; work?: Work }> {
    const res = await fetchWithAuth(`${API_BASE}/editions/isbn/${encodeURIComponent(isbn)}`);
    return handleResponse<{ edition: Edition; work?: Work }>(res);
  },

  // Retrieve user's wishlist / collection (Protected)
  async getWishlist(status?: ReadingStatus): Promise<WishlistItem[]> {
    const params = new URLSearchParams();
    if (status) params.set('status', status);
    const url = `${API_BASE}/wishlist${params.toString() ? `?${params.toString()}` : ''}`;
    const res = await fetchWithAuth(url);
    const data = await handleResponse<WishlistItem[]>(res);
    return Array.isArray(data) ? data : [];
  },

  // Retrieve all custom tags / shelves for user
  async getTags(): Promise<TagCount[]> {
    const res = await fetchWithAuth(`${API_BASE}/wishlist/tags`);
    const data = await handleResponse<{ tags: TagCount[] }>(res);
    return Array.isArray(data?.tags) ? data.tags : [];
  },

  // Add work / edition to collection (Protected)
  async addToWishlist(payload: AddWishlistPayload): Promise<WishlistItem> {
    const res = await fetchWithAuth(`${API_BASE}/wishlist`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload),
    });
    return handleResponse<WishlistItem>(res);
  },

  // Update wishlist entry (Protected)
  async updateWishlistItem(id: string, payload: UpdateWishlistPayload): Promise<WishlistItem> {
    const res = await fetchWithAuth(`${API_BASE}/wishlist/${encodeURIComponent(id)}`, {
      method: 'PATCH',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload),
    });
    return handleResponse<WishlistItem>(res);
  },

  // Delete item from collection (Protected)
  async deleteWishlistItem(id: string): Promise<void> {
    const res = await fetchWithAuth(`${API_BASE}/wishlist/${encodeURIComponent(id)}`, {
      method: 'DELETE',
    });
    return handleResponse<void>(res);
  },

  // Upload and import Goodreads CSV export
  async importGoodreads(file: File): Promise<GoodreadsImportSummary> {
    const formData = new FormData();
    formData.append('file', file);

    const token = tokenStorage.get();
    const headers: Record<string, string> = {};
    if (token) {
      headers['Authorization'] = `Bearer ${token}`;
    }

    const res = await fetch(`${API_BASE}/wishlist/import/goodreads`, {
      method: 'POST',
      headers,
      credentials: 'include',
      body: formData,
    });
    return handleResponse<GoodreadsImportSummary>(res);
  },

  // Export full user library as CSV or JSON
  async exportCollection(format: 'csv' | 'json' = 'csv'): Promise<void> {
    const res = await fetchWithAuth(`${API_BASE}/wishlist/export?format=${format}`);
    if (!res.ok) {
      throw new Error(`Export failed: ${res.statusText}`);
    }

    const blob = await res.blob();
    const url = window.URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.style.display = 'none';
    a.href = url;
    a.download = `imprint_library_${new Date().toISOString().split('T')[0]}.${format}`;
    document.body.appendChild(a);
    a.click();
    window.URL.revokeObjectURL(url);
    document.body.removeChild(a);
  },

  // Fetch reader statistics & reading velocity
  async getReadingStats(year?: number): Promise<ReadingStats> {
    const query = year ? `?year=${year}` : '';
    const res = await fetchWithAuth(`${API_BASE}/users/stats${query}`);
    return handleResponse<ReadingStats>(res);
  },

  // Set annual reading challenge target
  async setReadingGoal(year: number, targetBooks: number): Promise<ReadingGoal> {
    const res = await fetchWithAuth(`${API_BASE}/users/goals`, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ year, target_books: targetBooks }),
    });
    return handleResponse<ReadingGoal>(res);
  },
};
