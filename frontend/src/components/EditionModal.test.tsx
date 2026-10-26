import { describe, it, expect, vi } from "vitest";
import { render, screen, fireEvent } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { EditionModal } from "./EditionModal";
import { api } from "../lib/api";

vi.mock("../lib/api", () => ({ api: { getBook: vi.fn() } }));
describe("Edition detail browsing", () => {
  it("filters published editions by format and language and saves the chosen edition", async () => {
    const paper = {
      id: "paper",
      title: "Dune",
      format: "PAPERBACK" as const,
      publisher: "Ace",
      language: "eng",
      isbn13: "9780441172719",
    };
    const hard = {
      id: "hard",
      title: "Dune",
      format: "HARDCOVER" as const,
      publisher: "French publisher",
      language: "fre",
    };
    vi.mocked(api.getBook).mockResolvedValue({
      work: { id: "dune", title: "Dune" },
      editions: [paper, hard],
    });
    const save = vi.fn();
    render(
      <QueryClientProvider
        client={
          new QueryClient({ defaultOptions: { queries: { retry: false } } })
        }
      >
        <EditionModal
          isOpen
          work={{ id: "dune", title: "Dune" }}
          onClose={vi.fn()}
          onAddEditionToWishlist={save}
        />
      </QueryClientProvider>,
    );
    await screen.findByText("Ace");
    fireEvent.change(screen.getByLabelText("Format"), {
      target: { value: "PAPERBACK" },
    });
    expect(screen.queryByText("French publisher")).not.toBeInTheDocument();
    fireEvent.change(screen.getByLabelText("Language"), {
      target: { value: "eng" },
    });
    fireEvent.click(screen.getByRole("button", { name: "Save edition" }));
    expect(save).toHaveBeenCalledWith({ id: "dune", title: "Dune" }, paper);
  });
});
