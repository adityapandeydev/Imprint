import type {
  ApiResponse,
  Edition,
  ReadingStatus,
  WishlistItem,
  Work,
} from '../types/api';

const API_BASE = '/api/v1';

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
    const res = await fetch('/health/ready');
    if (!res.ok) {
      const liveRes = await fetch('/health');
      if (!liveRes.ok) throw new Error('Backend server is unreachable');
      return { status: 'live', database: 'disconnected' };
    }
    const json = await res.json();
    return json.data;
  },

  // Discover books by title, author, or keyword
  async searchBooks(query: string, limit = 20): Promise<Work[]> {
    if (!query.trim()) return [];
    const params = new URLSearchParams({ q: query, limit: String(limit) });
    const res = await fetch(`${API_BASE}/books/search?${params.toString()}`);
    return handleResponse<Work[]>(res);
  },

  // Get full work metadata and published editions
  async getBook(id: string): Promise<{ work: Work; editions: Edition[] }> {
    const res = await fetch(`${API_BASE}/books/${encodeURIComponent(id)}`);
    return handleResponse<{ work: Work; editions: Edition[] }>(res);
  },

  // Lookup edition and parent work by ISBN
  async getEditionByISBN(isbn: string): Promise<{ edition: Edition; work?: Work }> {
    const res = await fetch(`${API_BASE}/editions/isbn/${encodeURIComponent(isbn)}`);
    return handleResponse<{ edition: Edition; work?: Work }>(res);
  },

  // Retrieve user's wishlist / collection
  async getWishlist(status?: ReadingStatus): Promise<WishlistItem[]> {
    const params = new URLSearchParams();
    if (status) params.set('status', status);
    const url = `${API_BASE}/wishlist${params.toString() ? `?${params.toString()}` : ''}`;
    const res = await fetch(url);
    return handleResponse<WishlistItem[]>(res);
  },

  // Add work / edition to collection
  async addToWishlist(payload: {
    work_id: string;
    edition_id?: string;
    status?: ReadingStatus;
    priority?: number;
    notes?: string;
  }): Promise<WishlistItem> {
    const res = await fetch(`${API_BASE}/wishlist`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload),
    });
    return handleResponse<WishlistItem>(res);
  },

  // Update wishlist entry
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
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload),
    });
    return handleResponse<WishlistItem>(res);
  },

  // Delete item from collection
  async deleteWishlistItem(id: string): Promise<void> {
    const res = await fetch(`${API_BASE}/wishlist/${encodeURIComponent(id)}`, {
      method: 'DELETE',
    });
    return handleResponse<void>(res);
  },
};
