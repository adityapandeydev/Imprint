import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';
import { api, tokenStorage } from './api';

describe('api client & tokenStorage', () => {
  beforeEach(() => {
    localStorage.clear();
    vi.restoreAllMocks();
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  describe('tokenStorage', () => {
    it('stores, retrieves, and clears tokens', () => {
      expect(tokenStorage.get()).toBeNull();
      expect(tokenStorage.getRefresh()).toBeNull();

      tokenStorage.set('access-123', 'refresh-456');
      expect(tokenStorage.get()).toBe('access-123');
      expect(tokenStorage.getRefresh()).toBe('refresh-456');

      tokenStorage.clear();
      expect(tokenStorage.get()).toBeNull();
      expect(tokenStorage.getRefresh()).toBeNull();
    });
  });

  describe('api.searchBooks', () => {
    it('constructs correct query URL and returns works', async () => {
      const mockWorks = [
        {
          id: 'work-1',
          title: 'Dune',
          author: 'Frank Herbert',
          authors: ['Frank Herbert'],
          original_year: 1965,
        },
      ];

      const fetchSpy = vi.spyOn(globalThis, 'fetch').mockResolvedValueOnce({
        ok: true,
        status: 200,
        json: async () => ({ data: mockWorks }),
      } as Response);

      const results = await api.searchBooks('Dune', 15);

      expect(fetchSpy).toHaveBeenCalledWith(
        '/api/v1/books/search?q=Dune&limit=15',
        expect.objectContaining({
          headers: expect.any(Object),
        })
      );
      expect(results).toEqual(mockWorks);
    });

    it('throws error when server responds with 500 error payload', async () => {
      vi.spyOn(globalThis, 'fetch').mockResolvedValueOnce({
        ok: false,
        status: 500,
        json: async () => ({
          error: {
            code: 'INTERNAL_ERROR',
            message: 'database connection failed',
          },
        }),
      } as Response);

      await expect(api.searchBooks('Dune')).rejects.toThrow(
        'database connection failed'
      );
    });
  });

  describe('api.getBook', () => {
    it('fetches work by ID and returns work with editions', async () => {
      const mockWork = {
        id: 'work-123',
        title: 'Project Hail Mary',
        author: 'Andy Weir',
      };
      const mockEditions = [{ id: 'ed-1', isbn13: '9780593135204' }];

      vi.spyOn(globalThis, 'fetch').mockResolvedValueOnce({
        ok: true,
        status: 200,
        json: async () => ({ data: { work: mockWork, editions: mockEditions } }),
      } as Response);

      const result = await api.getBook('work-123');
      expect(result).toEqual({ work: mockWork, editions: mockEditions });
    });
  });
});
