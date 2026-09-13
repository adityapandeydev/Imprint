import React, { useEffect, useState } from 'react';

type Theme = 'light' | 'dark';

export function useTheme() {
  const [theme, setTheme] = useState<Theme>(() => {
    // Check localStorage first
    const saved = localStorage.getItem('imprint-theme');
    if (saved === 'dark' || saved === 'light') {
      return saved;
    }
    // Fall back to system preference
    if (window.matchMedia('(prefers-color-scheme: dark)').matches) {
      return 'dark';
    }
    return 'light';
  });

  useEffect(() => {
    const root = document.documentElement;
    if (theme === 'dark') {
      root.classList.add('dark');
    } else {
      root.classList.remove('dark');
    }
    localStorage.setItem('imprint-theme', theme);
  }, [theme]);

  const toggleTheme = (event?: React.MouseEvent) => {
    const nextTheme: Theme = theme === 'dark' ? 'light' : 'dark';

    type ViewTransitionDoc = Document & {
      startViewTransition?: (callback: () => void) => {
        ready: Promise<void>;
      };
    };
    const doc = document as ViewTransitionDoc;

    // Fallback if View Transitions API is unsupported or user prefers reduced motion
    if (
      !doc.startViewTransition ||
      window.matchMedia('(prefers-reduced-motion: reduce)').matches
    ) {
      setTheme(nextTheme);
      return;
    }

    // Origin coordinate of click, or top-right default (e.g. keyboard toggle)
    const x = event?.clientX && event.clientX > 0 ? event.clientX : window.innerWidth - 48;
    const y = event?.clientY && event.clientY > 0 ? event.clientY : 36;
    const endRadius = Math.hypot(
      Math.max(x, window.innerWidth - x),
      Math.max(y, window.innerHeight - y)
    );

    try {
      const transition = doc.startViewTransition(() => {
        // Synchronously update DOM before snapshot
        if (nextTheme === 'dark') {
          document.documentElement.classList.add('dark');
        } else {
          document.documentElement.classList.remove('dark');
        }
        localStorage.setItem('imprint-theme', nextTheme);
        setTheme(nextTheme);
      });

      transition.ready.then(() => {
        document.documentElement.animate(
          {
            clipPath: [
              `circle(0px at ${x}px ${y}px)`,
              `circle(${endRadius}px at ${x}px ${y}px)`,
            ],
          },
          {
            duration: 450,
            easing: 'cubic-bezier(0.2, 0, 0, 1)',
            pseudoElement: '::view-transition-new(root)',
          }
        );
      });
    } catch {
      setTheme(nextTheme);
    }
  };

  return { theme, toggleTheme, isDark: theme === 'dark' };
}
