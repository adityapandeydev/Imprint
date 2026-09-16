import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { render, screen, fireEvent, act } from '@testing-library/react';
import { SearchBar } from './SearchBar';

describe('SearchBar component', () => {
  beforeEach(() => {
    vi.useFakeTimers();
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  it('renders search input with placeholder and discovery chips', () => {
    render(<SearchBar onSearch={vi.fn()} />);

    expect(
      screen.getByPlaceholderText(/Search by title, author, series, or ISBN.../i)
    ).toBeInTheDocument();
    expect(screen.getByText('Dune')).toBeInTheDocument();
    expect(screen.getByText('Project Hail Mary')).toBeInTheDocument();
  });

  it('debounces user input before triggering onSearch', () => {
    const onSearch = vi.fn();
    render(<SearchBar onSearch={onSearch} />);

    const input = screen.getByPlaceholderText(/Search by title, author, series, or ISBN.../i);

    fireEvent.change(input, { target: { value: 'Neuromancer' } });

    // Immediately after typing, debounced call should not have fired yet
    expect(onSearch).not.toHaveBeenCalledWith('Neuromancer');

    // Fast-forward debounce timer (350ms)
    act(() => {
      vi.advanceTimersByTime(350);
    });

    expect(onSearch).toHaveBeenCalledWith('Neuromancer');
  });

  it('clears input and triggers onSearch with empty string when clear button is clicked', () => {
    const onSearch = vi.fn();
    render(<SearchBar initialQuery="Foundation" onSearch={onSearch} />);

    const clearButton = screen.getByRole('button', { name: /clear search/i });
    expect(clearButton).toBeInTheDocument();

    fireEvent.click(clearButton);

    const input = screen.getByPlaceholderText(/Search by title, author, series, or ISBN.../i) as HTMLInputElement;
    expect(input.value).toBe('');
    expect(onSearch).toHaveBeenCalledWith('');
  });

  it('triggers immediate onSearch when a suggestion chip is clicked', () => {
    const onSearch = vi.fn();
    render(<SearchBar onSearch={onSearch} />);

    const chip = screen.getByRole('button', { name: 'The Hobbit' });
    fireEvent.click(chip);

    expect(onSearch).toHaveBeenCalledWith('The Hobbit');
    const input = screen.getByPlaceholderText(/Search by title, author, series, or ISBN.../i) as HTMLInputElement;
    expect(input.value).toBe('The Hobbit');
  });

  it('detects ISBN-like input and displays ISBN badge', () => {
    render(<SearchBar onSearch={vi.fn()} />);

    const input = screen.getByPlaceholderText(/Search by title, author, series, or ISBN.../i);
    fireEvent.change(input, { target: { value: '9780593135204' } });

    expect(screen.getByText('ISBN')).toBeInTheDocument();
  });
});
