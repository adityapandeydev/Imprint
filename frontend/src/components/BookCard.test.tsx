import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { BookCard } from './BookCard';
import type { Work } from '../types/api';

const mockWork: Work = {
  id: 'work-1',
  title: 'Dune Messiah',
  authors: [{ id: 'author-1', name: 'Frank Herbert' }],
  original_year: 1969,
  cover_url: 'https://covers.openlibrary.org/b/id/123-M.jpg',
};

describe('BookCard component', () => {
  it('renders work title, author name, and publication year', () => {
    render(
      <BookCard
        work={mockWork}
        onToggleCollection={vi.fn()}
        onInspectEditions={vi.fn()}
      />
    );

    expect(screen.getByText('Dune Messiah')).toBeInTheDocument();
    expect(screen.getByText('Frank Herbert')).toBeInTheDocument();
    expect(screen.getByText(/First published 1969/i)).toBeInTheDocument();
  });

  it('renders literary spine fallback when cover_url is empty', () => {
    const workWithoutCover: Work = {
      ...mockWork,
      cover_url: undefined,
    };

    render(
      <BookCard
        work={workWithoutCover}
        onToggleCollection={vi.fn()}
        onInspectEditions={vi.fn()}
      />
    );

    expect(screen.getByText(/Cover Unavailable/i)).toBeInTheDocument();
  });

  it('triggers onInspectEditions when clicking the book title', () => {
    const onInspect = vi.fn();
    render(
      <BookCard
        work={mockWork}
        onToggleCollection={vi.fn()}
        onInspectEditions={onInspect}
      />
    );

    const title = screen.getByText('Dune Messiah');
    fireEvent.click(title);

    expect(onInspect).toHaveBeenCalledWith(mockWork);
  });

  it('triggers onToggleCollection when clicking the Want to Read button', () => {
    const onToggle = vi.fn();
    render(
      <BookCard
        work={mockWork}
        isInCollection={false}
        onToggleCollection={onToggle}
        onInspectEditions={vi.fn()}
      />
    );

    const addButton = screen.getByRole('button', { name: 'Add to Want to Read' });
    fireEvent.click(addButton);

    expect(onToggle).toHaveBeenCalledWith(mockWork);
  });

  it('displays Saved status when book is in collection', () => {
    render(
      <BookCard
        work={mockWork}
        isInCollection={true}
        onToggleCollection={vi.fn()}
        onInspectEditions={vi.fn()}
      />
    );

    expect(screen.getByText('Saved')).toBeInTheDocument();
    expect(screen.getByLabelText(/In your collection/i)).toBeInTheDocument();
  });
});
