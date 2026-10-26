import { describe, it, expect, vi } from "vitest";
import { render, screen, fireEvent } from "@testing-library/react";
import { CollectionView } from "./CollectionView";
import type { WishlistItem } from "../types/api";

vi.mock("../context/AuthContext", () => ({
  useAuth: () => ({
    user: { id: "reader-1" },
    isAuthenticated: true,
    isLoading: false,
  }),
}));
const items: WishlistItem[] = [
  {
    id: "1",
    user_id: "reader-1",
    work_id: "dune",
    status: "FINISHED",
    priority: 3,
    tags: ["Favorites"],
    work: { title: "Dune", authors: [{ name: "Frank Herbert" }] },
    created_at: "2026-01-01",
    updated_at: "2026-01-01",
  },
  {
    id: "2",
    user_id: "reader-1",
    work_id: "hobbit",
    status: "CURRENTLY_READING",
    priority: 4,
    tags: [],
    work: { title: "The Hobbit", authors: [{ name: "J. R. R. Tolkien" }] },
    created_at: "2026-02-01",
    updated_at: "2026-02-01",
  },
];
const props = {
  items,
  isLoading: false,
  onUpdateStatus: vi.fn(),
  onUpdatePriority: vi.fn(),
  onUpdateRating: vi.fn(),
  onUpdateNotes: vi.fn(),
  onDelete: vi.fn(),
  onGoToDiscover: vi.fn(),
};

describe("Collection browsing", () => {
  it("combines status, shelf, and author search without altering the collection", () => {
    render(<CollectionView {...props} />);
    fireEvent.click(screen.getByRole("button", { name: /^Finished\s*1$/ }));
    expect(screen.getByRole("heading", { name: "Dune" })).toBeInTheDocument();
    expect(
      screen.queryByRole("heading", { name: "The Hobbit" }),
    ).not.toBeInTheDocument();
    fireEvent.change(
      screen.getByRole("searchbox", { name: "Search your collection" }),
      { target: { value: "Tolkien" } },
    );
    expect(screen.getByText("No books on this shelf yet.")).toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: "Clear filters" }));
    expect(screen.getAllByRole("heading", { level: 3 })).toHaveLength(2);
  });

  it("persists an empty custom shelf for the signed-in reader", () => {
    const view = render(<CollectionView {...props} />);
    fireEvent.click(screen.getByRole("button", { name: "Create shelf" }));
    fireEvent.change(screen.getByRole("textbox", { name: "New shelf name" }), {
      target: { value: "Weekend reads" },
    });
    fireEvent.click(screen.getByRole("button", { name: /^Create$/ }));
    expect(
      JSON.parse(localStorage.getItem("imprint_shelves_reader-1")!),
    ).toEqual(["Weekend reads"]);
    view.unmount();
    render(<CollectionView {...props} />);
    expect(
      screen.getByRole("button", { name: /^Weekend reads\s*0$/ }),
    ).toBeInTheDocument();
  });

  it("offers retry instead of an empty shelf when the request fails", () => {
    const retry = vi.fn();
    render(<CollectionView {...props} error onRefreshCollection={retry} />);
    expect(screen.getByRole("alert")).toHaveTextContent(
      "Your collection couldn’t load.",
    );
    fireEvent.click(screen.getByRole("button", { name: "Try again" }));
    expect(retry).toHaveBeenCalledOnce();
  });
});
