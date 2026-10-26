import { useState } from "react";
import { Upload, Download, ExternalLink, Copy, Check } from "lucide-react";
import { useAuth } from "../context/AuthContext";
import { GoodreadsImportModal } from "./GoodreadsImportModal";
import { api } from "../lib/api";
import { toast } from "../lib/toast";
import type { ProfileVisibility } from "../types/api";
export function SettingsView({
  onRefreshCollection,
}: {
  onRefreshCollection: () => void;
}) {
  const { user, isAuthenticated, openAuthModal, updateUserProfileVisibility } =
    useAuth();
  const [importing, setImporting] = useState(false);
  const [busy, setBusy] = useState(false);
  const [exporting, setExporting] = useState(false);
  const [copied, setCopied] = useState(false);
  const privacy = async (value: ProfileVisibility) => {
    setBusy(true);
    try {
      await updateUserProfileVisibility(value);
      toast.success("Visibility updated");
    } catch {
      toast.error("Visibility couldn’t be saved");
    } finally {
      setBusy(false);
    }
  };
  const exportData = async (format: "csv" | "json") => {
    setExporting(true);
    try {
      await api.exportCollection(format);
    } catch {
      toast.error("Export failed", "Please try again.");
    } finally {
      setExporting(false);
    }
  };
  const copy = async () => {
    try {
      await navigator.clipboard.writeText(
        `${window.location.origin}/#/u/${encodeURIComponent(user!.username)}`,
      );
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    } catch {
      toast.error("The link couldn’t be copied");
    }
  };
  return (
    <section className="settings-page">
      <div className="page-heading">
        <div>
          <h1>Your space</h1>
          <p>Account, sharing, and the books you bring with you.</p>
        </div>
      </div>
      {!isAuthenticated || !user ? (
        <div className="empty-state">
          <h2>Make yourself at home.</h2>
          <p>Sign in to manage your account and collection.</p>
          <button
            className="button primary"
            onClick={() => openAuthModal("login")}
          >
            Sign in
          </button>
        </div>
      ) : (
        <>
          <section className="settings-section">
            <div>
              <h2>Reader profile</h2>
              <p>Your account on Imprint.</p>
            </div>
            <div className="account-profile">
              <span className="avatar">
                {(user.display_name || user.username).charAt(0)}
              </span>
              <div>
                <strong>{user.display_name || user.username}</strong>
                <p>
                  @{user.username} · {user.email}
                </p>
              </div>
            </div>
          </section>
          <section className="settings-section">
            <div>
              <h2>Collection visibility</h2>
              <p>Choose who can view your shelves and reading statistics.</p>
            </div>
            <fieldset className="privacy-options" disabled={busy}>
              <legend className="sr-only">Collection visibility</legend>
              {(
                [
                  {
                    id: "PUBLIC",
                    label: "Public",
                    description:
                      "Anyone with your profile link can view your collection.",
                  },
                  {
                    id: "UNLISTED",
                    label: "Unlisted",
                    description:
                      "Accessible by direct link. This is not a private library.",
                  },
                  {
                    id: "PRIVATE",
                    label: "Private",
                    description:
                      "Keep your collection and reading stats visible only to you.",
                  },
                ] as const
              ).map((option) => (
                <label key={option.id}>
                  <input
                    type="radio"
                    name="visibility"
                    value={option.id}
                    checked={user.profile_visibility === option.id}
                    onChange={() => void privacy(option.id)}
                  />
                  <span>
                    <strong>{option.label}</strong>
                    <span>{option.description}</span>
                  </span>
                </label>
              ))}
            </fieldset>
          </section>
          {user.profile_visibility !== "PRIVATE" && (
            <section className="settings-section">
              <div>
                <h2>Share your shelves</h2>
                <p>A link to your reader profile.</p>
              </div>
              <div className="empty-actions">
                <a
                  className="button secondary"
                  href={`#/u/${encodeURIComponent(user.username)}`}
                >
                  <ExternalLink size={16} />
                  View profile
                </a>
                <button
                  className="button secondary"
                  onClick={() => void copy()}
                >
                  {copied ? <Check size={16} /> : <Copy size={16} />}{" "}
                  {copied ? "Copied" : "Copy link"}
                </button>
              </div>
            </section>
          )}
          <section className="settings-section">
            <div>
              <h2>Bring your books</h2>
              <p>
                Import a Goodreads CSV, or take a copy of your collection with
                you.
              </p>
            </div>
            <div className="empty-actions">
              <button
                className="button secondary"
                onClick={() => setImporting(true)}
              >
                <Upload size={16} />
                Import Goodreads
              </button>
              <button
                disabled={exporting}
                className="button secondary"
                onClick={() => void exportData("csv")}
              >
                <Download size={16} />
                Export CSV
              </button>
              <button
                disabled={exporting}
                className="button secondary"
                onClick={() => void exportData("json")}
              >
                Export JSON
              </button>
            </div>
          </section>
          <GoodreadsImportModal
            isOpen={importing}
            onClose={() => setImporting(false)}
            onSuccess={onRefreshCollection}
          />
        </>
      )}
    </section>
  );
}
