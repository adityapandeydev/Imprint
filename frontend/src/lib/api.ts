import type {
  ApiResponse,
  AuthTokens,
  Edition,
  LoginRequest,
  ReadingStatus,
  RegisterRequest,
  User,
  WishlistItem,
  Work,
} from '../types/api';

const API_BASE = '/api/v1';
const TOKEN_KEY = 'imprint_token';

export const tokenStorage = {
  get: (): string | null => {
    try {
      return localStorage.getItem(TOKEN_KEY);
    } catch {
      return null;
    }
  },
  set: (token: string): void => {
    try {
      localStorage.setItem(TOKEN_KEY, token);
    } catch {
      // ignore
    }
  },
  clear: (): void => {
    try {
      localStorage.removeItem(TOKEN_KEY);
    } catch {
      // ignore
    }
  },
};

function getAuthHeaders(customHeaders: Record<string, string> = {}): HeadersInit {
  const headers: Record<string, string> = { ...customHeaders };
  const token = tokenStorage.get();
  if (token) {
    headers['Authorization'] = `Bearer ${token}`;
  }
  return headers;
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
      tokenStorage.set(tokens.access_token);
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
      tokenStorage.set(tokens.access_token);
    }
    return tokens;
  },

  // Auth: Log out user
  async logout(): Promise<void> {
    try {
      await fetch(`${API_BASE}/auth/logout`, {
        method: 'POST',
        headers: getAuthHeaders(),
        credentials: 'include',
      });
    } finally {
      tokenStorage.clear();
    }
  },

  // Auth: Get current authenticated user profile
  async getMe(): Promise<User> {
    const res = await fetch(`${API_BASE}/auth/me`, {
      headers: getAuthHeaders(),
      credentials: 'include',
    });
    return handleResponse<User>(res);
  },

  // Discover books by title, author, or keyword
  async searchBooks(query: string, limit = 20, signal?: AbortSignal): Promise<Work[]> {
    if (!query.trim()) return [];
    const params = new URLSearchParams({ q: query, limit: String(limit) });
    const res = await fetch(`${API_BASE}/books/search?${params.toString()}`, {
      headers: getAuthHeaders(),
      credentials: 'include',
      signal,
    });
    const data = await handleResponse<Work[]>(res);
    return Array.isArray(data) ? data : [];
  },

  // Get full work metadata and published editions
  async getBook(id: string): Promise<{ work: Work; editions: Edition[] }> {
    const res = await fetch(`${API_BASE}/books/${encodeURIComponent(id)}`, {
      headers: getAuthHeaders(),
      credentials: 'include',
    });
    const data = await handleResponse<{ work: Work; editions: Edition[] }>(res);
    return {
      work: data?.work,
      editions: Array.isArray(data?.editions) ? data.editions : [],
    };
  },

  // Lookup edition and parent work by ISBN
  async getEditionByISBN(isbn: string): Promise<{ edition: Edition; work?: Work }> {
    const res = await fetch(`${API_BASE}/editions/isbn/${encodeURIComponent(isbn)}`, {
      headers: getAuthHeaders(),
      credentials: 'include',
    });
    return handleResponse<{ edition: Edition; work?: Work }>(res);
  },

  // Retrieve user's wishlist / collection (Protected)
  async getWishlist(status?: ReadingStatus): Promise<WishlistItem[]> {
    const params = new URLSearchParams();
    if (status) params.set('status', status);
    const url = `${API_BASE}/wishlist${params.toString() ? `?${params.toString()}` : ''}`;
    const res = await fetch(url, {
      headers: getAuthHeaders(),
      credentials: 'include',
    });
    const data = await handleResponse<WishlistItem[]>(res);
    return Array.isArray(data) ? data : [];
  },

  // Add work / edition to collection (Protected)
  async addToWishlist(payload: {
    work_id: string;
    edition_id?: string;
    status?: ReadingStatus;
    priority?: number;
    notes?: string;
    title?: string;
    author?: string;
    cover_url?: string;
    original_year?: number;
  }): Promise<WishlistItem> {
    const res = await fetch(`${API_BASE}/wishlist`, {
      method: 'POST',
      headers: getAuthHeaders({ 'Content-Type': 'application/json' }),
      credentials: 'include',
      body: JSON.stringify(payload),
    });
    return handleResponse<WishlistItem>(res);
  },

  // Update wishlist entry (Protected)
  async updateWishlistItem(
    id: string,
    payload: {
      edition_id?: string;
      status?: ReadingStatus;
      priority?: number;
      rating?: number;
      notes?: string;
    }
  ): Promise<WishlistItem> {
    const res = await fetch(`${API_BASE}/wishlist/${encodeURIComponent(id)}`, {
      method: 'PATCH',
      headers: getAuthHeaders({ 'Content-Type': 'application/json' }),
      credentials: 'include',
      body: JSON.stringify(payload),
    });
    return handleResponse<WishlistItem>(res);
  },

  // Delete item from collection (Protected)
  async deleteWishlistItem(id: string): Promise<void> {
    const res = await fetch(`${API_BASE}/wishlist/${encodeURIComponent(id)}`, {
      method: 'DELETE',
      headers: getAuthHeaders(),
      credentials: 'include',
    });
    return handleResponse<void>(res);
  },
};
