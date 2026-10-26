import { useEffect, useRef } from "react";

// Nested dialogs suspend the underlying focus boundary.
const dialogStack: symbol[] = [];
export function useDialog(open: boolean, onClose: () => void) {
  const ref = useRef<HTMLDivElement>(null);
  const close = useRef(onClose);
  useEffect(() => {
    close.current = onClose;
  }, [onClose]);
  useEffect(() => {
    if (!open || !ref.current) return;
    const element = ref.current;
    const previous = document.activeElement as HTMLElement | null;
    const id = Symbol("dialog");
    const originalOverflow = document.body.style.overflow;
    dialogStack.push(id);
    document.body.style.overflow = "hidden";
    const selectors =
      'button:not([disabled]), a[href], input:not([disabled]), select:not([disabled]), textarea:not([disabled]), [tabindex="0"]';
    const focusables = () =>
      Array.from(element.querySelectorAll<HTMLElement>(selectors)).filter(
        (node) => !node.closest("[hidden]"),
      );
    (focusables()[0] || element).focus();
    const keydown = (event: KeyboardEvent) => {
      if (dialogStack.at(-1) !== id) return;
      if (event.key === "Escape") {
        event.stopImmediatePropagation();
        close.current();
      }
      if (event.key === "Tab") {
        const targets = focusables();
        const first = targets[0];
        const last = targets.at(-1);
        if (!first) {
          event.preventDefault();
          element.focus();
          return;
        }
        if (
          event.shiftKey &&
          (document.activeElement === first ||
            document.activeElement === element)
        ) {
          event.preventDefault();
          last?.focus();
        } else if (
          !event.shiftKey &&
          (document.activeElement === last ||
            !element.contains(document.activeElement))
        ) {
          event.preventDefault();
          first.focus();
        }
      }
    };
    const focusin = (event: FocusEvent) => {
      if (dialogStack.at(-1) === id && !element.contains(event.target as Node))
        (focusables()[0] || element).focus();
    };
    document.addEventListener("keydown", keydown, true);
    document.addEventListener("focusin", focusin);
    return () => {
      dialogStack.splice(dialogStack.indexOf(id), 1);
      document.removeEventListener("keydown", keydown, true);
      document.removeEventListener("focusin", focusin);
      document.body.style.overflow = originalOverflow;
      if (previous?.isConnected) previous.focus();
    };
  }, [open]);
  return ref;
}
