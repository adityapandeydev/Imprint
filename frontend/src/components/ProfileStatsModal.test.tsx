import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { ProfileStatsModal } from './ProfileStatsModal';
import { api } from '../lib/api';
import type { ReadingStats } from '../types/api';

vi.mock('../lib/api', () => ({
  api: {
    getReadingStats: vi.fn(),
    setReadingGoal: vi.fn(),
  },
}));

describe('ProfileStatsModal component', () => {
  const mockOnClose = vi.fn();

  const sampleStats: ReadingStats = {
    total_books: 15,
    books_finished_year: 8,
    total_pages_read: 2840,
    currently_reading: 2,
    want_to_read: 5,
    average_rating: 4.5,
    current_year: 2026,
    top_genres: [
      { genre: 'Science Fiction', count: 5 },
      { genre: 'Philosophy', count: 3 },
    ],
    top_authors: [
      { author: 'Frank Herbert', count: 3 },
      { author: 'Ursula K. Le Guin', count: 2 },
    ],
    format_distribution: {
      HARDCOVER: 6,
      PAPERBACK: 9,
    },
    challenge: {
      year: 2026,
      target_books: 24,
      books_finished: 8,
      percentage: 33,
      days_elapsed: 75,
      total_days: 365,
      expected_finished: 4.9,
      pacing_diff: 3.1,
      pacing_status: 'AHEAD',
      pacing_message: 'You are 3 books ahead of schedule!',
      monthly_progress: [
        { month: 1, books: 3, pages: 1020 },
        { month: 2, books: 2, pages: 640 },
        { month: 3, books: 3, pages: 1180 },
        { month: 4, books: 0, pages: 0 },
        { month: 5, books: 0, pages: 0 },
        { month: 6, books: 0, pages: 0 },
        { month: 7, books: 0, pages: 0 },
        { month: 8, books: 0, pages: 0 },
        { month: 9, books: 0, pages: 0 },
        { month: 10, books: 0, pages: 0 },
        { month: 11, books: 0, pages: 0 },
        { month: 12, books: 0, pages: 0 },
      ],
    },
  };

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders nothing when isOpen is false', () => {
    render(<ProfileStatsModal isOpen={false} onClose={mockOnClose} />);
    expect(screen.queryByText(/Literary Velocity & Analytics/i)).not.toBeInTheDocument();
  });

  it('renders loading state initially', async () => {
    vi.mocked(api.getReadingStats).mockReturnValue(new Promise(() => {})); // pending promise
    render(<ProfileStatsModal isOpen={true} onClose={mockOnClose} />);

    expect(screen.getByText(/Computing reading velocity & analytics.../i)).toBeInTheDocument();
  });

  it('renders reading stats and challenge metrics when loaded', async () => {
    vi.mocked(api.getReadingStats).mockResolvedValue(sampleStats);
    render(<ProfileStatsModal isOpen={true} onClose={mockOnClose} userName="Aditya" />);

    await waitFor(() => {
      expect(screen.getByText('2026 Reading Challenge')).toBeInTheDocument();
    });

    // Pacing badge & message
    expect(screen.getByText('Ahead of Schedule')).toBeInTheDocument();
    expect(screen.getByText('You are 3 books ahead of schedule!')).toBeInTheDocument();
    expect(screen.getByText('33%')).toBeInTheDocument();
    expect(screen.getByText('8 of 24 books completed')).toBeInTheDocument();

    // 12-Month Velocity Chart
    expect(screen.getByText(/12-Month Reading Velocity \(2026\)/i)).toBeInTheDocument();
    expect(screen.getByText('Jan')).toBeInTheDocument();
    expect(screen.getByText('Dec')).toBeInTheDocument();

    // Author & Subject
    expect(screen.getByText('Science Fiction')).toBeInTheDocument();
    expect(screen.getByText('Frank Herbert')).toBeInTheDocument();
  });

  it('allows user to open goal editor, select preset, and save new target', async () => {
    vi.mocked(api.getReadingStats).mockResolvedValue(sampleStats);
    vi.mocked(api.setReadingGoal).mockResolvedValue({
      id: 'g-1',
      user_id: 'u-1',
      year: 2026,
      target_books: 36,
      created_at: new Date().toISOString(),
      updated_at: new Date().toISOString(),
    });

    render(<ProfileStatsModal isOpen={true} onClose={mockOnClose} />);

    await waitFor(() => {
      expect(screen.getByRole('button', { name: /Edit reading goal/i })).toBeInTheDocument();
    });

    // Click edit button
    fireEvent.click(screen.getByRole('button', { name: /Edit reading goal/i }));

    // Verify presets are displayed
    expect(screen.getByText('36 books')).toBeInTheDocument();
    expect(screen.getByText('52 books')).toBeInTheDocument();

    // Click 36 books preset
    fireEvent.click(screen.getByText('36 books'));

    // Verify input value changed to 36
    const input = screen.getByLabelText(/Target books input/i) as HTMLInputElement;
    expect(input.value).toBe('36');

    // Click save
    const saveBtn = screen.getByRole('button', { name: /Save Goal/i });
    fireEvent.click(saveBtn);

    await waitFor(() => {
      expect(api.setReadingGoal).toHaveBeenCalledWith(2026, 36);
    });
  });

  it('allows user to type custom target and save', async () => {
    vi.mocked(api.getReadingStats).mockResolvedValue(sampleStats);
    vi.mocked(api.setReadingGoal).mockResolvedValue({
      id: 'g-1',
      user_id: 'u-1',
      year: 2026,
      target_books: 45,
      created_at: new Date().toISOString(),
      updated_at: new Date().toISOString(),
    });

    render(<ProfileStatsModal isOpen={true} onClose={mockOnClose} />);

    await waitFor(() => {
      expect(screen.getByRole('button', { name: /Edit reading goal/i })).toBeInTheDocument();
    });

    fireEvent.click(screen.getByRole('button', { name: /Edit reading goal/i }));

    const input = screen.getByLabelText(/Target books input/i);
    fireEvent.change(input, { target: { value: '45' } });

    const saveBtn = screen.getByRole('button', { name: /Save Goal/i });
    fireEvent.click(saveBtn);

    await waitFor(() => {
      expect(api.setReadingGoal).toHaveBeenCalledWith(2026, 45);
    });
  });

  it('displays error if goal is invalid', async () => {
    vi.mocked(api.getReadingStats).mockResolvedValue(sampleStats);
    render(<ProfileStatsModal isOpen={true} onClose={mockOnClose} />);

    await waitFor(() => {
      expect(screen.getByRole('button', { name: /Edit reading goal/i })).toBeInTheDocument();
    });

    fireEvent.click(screen.getByRole('button', { name: /Edit reading goal/i }));

    const input = screen.getByLabelText(/Target books input/i);
    fireEvent.change(input, { target: { value: '0' } });

    const saveBtn = screen.getByRole('button', { name: /Save Goal/i });
    fireEvent.click(saveBtn);

    expect(screen.getByText(/Please enter a valid goal between 1 and 1,000 books/i)).toBeInTheDocument();
    expect(api.setReadingGoal).not.toHaveBeenCalled();
  });

  it('calls onClose when close button is clicked', async () => {
    vi.mocked(api.getReadingStats).mockResolvedValue(sampleStats);
    render(<ProfileStatsModal isOpen={true} onClose={mockOnClose} />);

    await waitFor(() => {
      expect(screen.getByRole('button', { name: /Close modal/i })).toBeInTheDocument();
    });

    fireEvent.click(screen.getByRole('button', { name: /Close modal/i }));
    expect(mockOnClose).toHaveBeenCalled();
  });
});
