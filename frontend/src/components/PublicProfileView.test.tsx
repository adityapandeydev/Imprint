import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { PublicProfileView } from './PublicProfileView';
import { api } from '../lib/api';
import type { PublicProfile, PublicCollectionResponse } from '../types/api';

vi.mock('../lib/api', () => ({
  api: {
    getPublicProfile: vi.fn(),
    getPublicCollection: vi.fn(),
    getPublicOGCardUrl: vi.fn((u) => `/api/v1/public/users/${u}/og.svg`),
  },
}));

describe('PublicProfileView component', () => {
  const mockNavigateHome = vi.fn();
  const mockInspectWork = vi.fn();

  const mockProfile: PublicProfile = {
    username: 'aditya',
    display_name: 'Aditya Pandey',
    profile_visibility: 'PUBLIC',
    member_since: '2026-01-01T00:00:00Z',
    shelves: [
      { tag: 'favorites', count: 4 },
      { tag: 'sci-fi', count: 6 },
    ],
    stats: {
      total_books: 12,
      books_finished_year: 7,
      total_pages_read: 2450,
      currently_reading: 2,
      want_to_read: 3,
      average_rating: 4.6,
      top_genres: [{ genre: 'Sci-Fi', count: 5 }],
      top_authors: [{ author: 'Frank Herbert', count: 3 }],
      format_distribution: { HARDCOVER: 5, PAPERBACK: 7 },
      current_year: 2026,
      challenge: {
        year: 2026,
        target_books: 20,
        books_finished: 7,
        percentage: 35,
        days_elapsed: 80,
        total_days: 365,
        expected_finished: 4.4,
        pacing_diff: 2.6,
        pacing_status: 'AHEAD',
        pacing_message: 'You are 2.6 books ahead of schedule!',
        monthly_progress: [],
      },
    },
  };

  const mockCollection: PublicCollectionResponse = {
    profile: mockProfile,
    items: [
      {
        id: 'item-1',
        user_id: 'u-aditya',
        work_id: 'w-dune',
        status: 'FINISHED',
        priority: 5,
        rating: 5,
        tags: ['favorites', 'sci-fi'],
        work: {
          id: 'w-dune',
          title: 'Dune',
          cover_url: 'https://covers.openlibrary.org/b/id/123-M.jpg',
          authors: [{ name: 'Frank Herbert' }],
        },
        created_at: '2026-01-10T00:00:00Z',
        updated_at: '2026-01-20T00:00:00Z',
      },
      {
        id: 'item-2',
        user_id: 'u-aditya',
        work_id: 'w-neuromancer',
        status: 'CURRENTLY_READING',
        priority: 4,
        tags: ['sci-fi'],
        work: {
          id: 'w-neuromancer',
          title: 'Neuromancer',
          authors: [{ name: 'William Gibson' }],
        },
        created_at: '2026-02-01T00:00:00Z',
        updated_at: '2026-02-15T00:00:00Z',
      },
    ],
    count: 2,
  };

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders loading state initially', async () => {
    vi.mocked(api.getPublicProfile).mockReturnValue(new Promise(() => {}));
    vi.mocked(api.getPublicCollection).mockReturnValue(new Promise(() => {}));

    render(
      <PublicProfileView
        username="aditya"
        onNavigateHome={mockNavigateHome}
        onInspectWork={mockInspectWork}
      />
    );

    expect(screen.getByText(/Opening @aditya's library\.\.\./i)).toBeInTheDocument();
  });

  it('renders reader public profile details, challenge ring, shelves, and books', async () => {
    vi.mocked(api.getPublicProfile).mockResolvedValue(mockProfile);
    vi.mocked(api.getPublicCollection).mockResolvedValue(mockCollection);

    render(
      <PublicProfileView
        username="aditya"
        onNavigateHome={mockNavigateHome}
        onInspectWork={mockInspectWork}
      />
    );

    await waitFor(() => {
      expect(screen.getByText('Aditya Pandey')).toBeInTheDocument();
    });

    expect(screen.getByText(/@aditya/)).toBeInTheDocument();
    expect(screen.getByText('35%')).toBeInTheDocument();
    expect(screen.getByText(/7 of 20 read/i)).toBeInTheDocument();

    // Curated shelves
    expect(screen.getByRole('button', { name: /#favorites/i })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /#sci-fi/i })).toBeInTheDocument();

    // Book cards
    expect(screen.getByRole('heading', { name: 'Dune' })).toBeInTheDocument();
    expect(screen.getByRole('heading', { name: 'Neuromancer' })).toBeInTheDocument();
  });

  it('renders private profile barrier when profile is private (403)', async () => {
    vi.mocked(api.getPublicProfile).mockRejectedValue(new Error('this reader profile is private'));

    render(
      <PublicProfileView
        username="aditya"
        onNavigateHome={mockNavigateHome}
        onInspectWork={mockInspectWork}
      />
    );

    await waitFor(() => {
      expect(screen.getByText(/This Library is Private/i)).toBeInTheDocument();
    });

    expect(
      screen.getByText(/@aditya has chosen to keep their reading collection.*private/i)
    ).toBeInTheDocument();

    // Return to Imprint button
    const returnBtn = screen.getByRole('button', { name: /Return to Imprint/i });
    fireEvent.click(returnBtn);
    expect(mockNavigateHome).toHaveBeenCalled();
  });

  it('allows clicking a shelf tag to filter collection', async () => {
    vi.mocked(api.getPublicProfile).mockResolvedValue(mockProfile);
    vi.mocked(api.getPublicCollection).mockResolvedValue(mockCollection);

    render(
      <PublicProfileView
        username="aditya"
        onNavigateHome={mockNavigateHome}
        onInspectWork={mockInspectWork}
      />
    );

    await waitFor(() => {
      expect(screen.getByRole('button', { name: /#favorites/i })).toBeInTheDocument();
    });

    fireEvent.click(screen.getByRole('button', { name: /#favorites/i }));

    await waitFor(() => {
      expect(api.getPublicCollection).toHaveBeenCalledWith('aditya', {
        status: undefined,
        tag: 'favorites',
      });
    });
  });

  it('opens ShareModal when Share Collection is clicked', async () => {
    vi.mocked(api.getPublicProfile).mockResolvedValue(mockProfile);
    vi.mocked(api.getPublicCollection).mockResolvedValue(mockCollection);

    render(
      <PublicProfileView
        username="aditya"
        onNavigateHome={mockNavigateHome}
        onInspectWork={mockInspectWork}
      />
    );

    await waitFor(() => {
      expect(screen.getByRole('button', { name: /^Share Collection$/i })).toBeInTheDocument();
    });

    fireEvent.click(screen.getByRole('button', { name: /^Share Collection$/i }));

    await waitFor(() => {
      expect(screen.getByText(/Social Card Preview/i)).toBeInTheDocument();
    });
  });
});
