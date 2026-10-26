import { describe, it, expect, vi } from "vitest";
import { render, screen, fireEvent } from "@testing-library/react";
import { useState } from "react";
import { useDialog } from "./useDialog";

function DialogTest({ closed }: { closed: () => void }) {
  const [open, setOpen] = useState(false);
  const ref = useDialog(open, () => {
    closed();
    setOpen(false);
  });
  return (
    <>
      <button onClick={() => setOpen(true)}>Open</button>
      {open && (
        <div
          ref={ref}
          role="dialog"
          aria-modal="true"
          aria-label="Test dialog"
          tabIndex={-1}
        >
          <button>First</button>
          <input aria-label="Name" />
          <button>Last</button>
        </div>
      )}
    </>
  );
}
describe("Dialog keyboard behavior", () => {
  it("moves focus inside, wraps Tab, closes with Escape, and restores the trigger", () => {
    const closed = vi.fn();
    render(<DialogTest closed={closed} />);
    const trigger = screen.getByRole("button", { name: "Open" });
    trigger.focus();
    fireEvent.click(trigger);
    const first = screen.getByRole("button", { name: "First" });
    const last = screen.getByRole("button", { name: "Last" });
    expect(first).toHaveFocus();
    fireEvent.keyDown(first, { key: "Tab", shiftKey: true });
    expect(last).toHaveFocus();
    fireEvent.keyDown(last, { key: "Tab" });
    expect(first).toHaveFocus();
    fireEvent.keyDown(first, { key: "Escape" });
    expect(closed).toHaveBeenCalledOnce();
    expect(screen.queryByRole("dialog")).not.toBeInTheDocument();
    expect(trigger).toHaveFocus();
    expect(document.body.style.overflow).toBe("");
  });
});
