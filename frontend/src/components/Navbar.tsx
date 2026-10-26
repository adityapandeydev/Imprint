import {
  BookOpen,
  Compass,
  Library,
  BarChart3,
  Settings,
  Sun,
  Moon,
  LogOut,
} from "lucide-react";
import { useTheme } from "../lib/theme";
import { useAuth } from "../context/AuthContext";
export type AppTab = "discover" | "collection" | "stats" | "settings";
interface NavbarProps {
  activeTab: AppTab;
  onSelectTab: (tab: AppTab) => void;
  collectionCount?: number;
}
export function Navbar({
  activeTab,
  onSelectTab,
  collectionCount = 0,
}: NavbarProps) {
  const { isDark, toggleTheme } = useTheme();
  const { user, isAuthenticated, logout, openAuthModal } = useAuth();
  const tabs = [
    { id: "discover" as const, label: "Discover", icon: Compass },
    { id: "collection" as const, label: "My Collection", icon: Library },
    { id: "stats" as const, label: "Reading stats", icon: BarChart3 },
    { id: "settings" as const, label: "Settings", icon: Settings },
  ];
  return (
    <header className="site-header">
      <div className="header-inner">
        <button
          className="wordmark"
          onClick={() => onSelectTab("discover")}
          aria-label="Imprint home"
        >
          <BookOpen size={25} strokeWidth={1.6} />
          <span>
            imprint<span className="wordmark-dot">.</span>
          </span>
        </button>
        <nav className="main-nav" aria-label="Main navigation">
          {tabs.map(({ id, label, icon: Icon }) => (
            <button
              key={id}
              aria-current={activeTab === id ? "page" : undefined}
              onClick={() => onSelectTab(id)}
            >
              <Icon size={17} />
              <span>{label}</span>
              {id === "collection" &&
                isAuthenticated &&
                collectionCount > 0 && (
                  <span className="nav-count">{collectionCount}</span>
                )}
            </button>
          ))}
        </nav>
        <div className="header-actions">
          <button
            className="icon-button"
            onClick={toggleTheme}
            aria-label="Toggle theme"
            title={isDark ? "Switch to light mode" : "Switch to dark mode"}
          >
            {isDark ? <Sun size={18} /> : <Moon size={18} />}
          </button>
          {isAuthenticated && user ? (
            <>
              <button
                className="account-button"
                onClick={() => onSelectTab("settings")}
                aria-label="Account settings"
              >
                <span className="avatar">
                  {(user.display_name || user.username).charAt(0)}
                </span>
                <span className="account-name">
                  {user.display_name || user.username}
                </span>
              </button>
              <button
                className="icon-button"
                onClick={() => void logout()}
                aria-label="Sign out"
              >
                <LogOut size={17} />
              </button>
            </>
          ) : (
            <button
              className="button primary compact"
              onClick={() => openAuthModal("login")}
            >
              Sign In
            </button>
          )}
        </div>
      </div>
    </header>
  );
}
