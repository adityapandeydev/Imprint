import React, { useState, useRef } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import {
  UploadCloud,
  FileText,
  CheckCircle2,
  AlertCircle,
  X,
  ExternalLink,
} from 'lucide-react';
import { api } from '../lib/api';
import type { GoodreadsImportSummary } from '../types/api';

interface GoodreadsImportModalProps {
  isOpen: boolean;
  onClose: () => void;
  onSuccess: () => void;
}

export const GoodreadsImportModal: React.FC<GoodreadsImportModalProps> = ({
  isOpen,
  onClose,
  onSuccess,
}) => {
  const [file, setFile] = useState<File | null>(null);
  const [uploading, setUploading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [summary, setSummary] = useState<GoodreadsImportSummary | null>(null);
  const [dragOver, setDragOver] = useState(false);
  const fileInputRef = useRef<HTMLInputElement>(null);

  if (!isOpen) return null;

  const handleFile = (selectedFile: File) => {
    setError(null);
    if (!selectedFile.name.toLowerCase().endsWith('.csv')) {
      setError('Please select a valid CSV file (e.g. goodreads_library_export.csv)');
      return;
    }
    // 5MB limit
    if (selectedFile.size > 5 * 1024 * 1024) {
      setError('File exceeds the 5MB upload size limit.');
      return;
    }
    setFile(selectedFile);
  };

  const handleDrop = (e: React.DragEvent) => {
    e.preventDefault();
    setDragOver(false);
    if (e.dataTransfer.files && e.dataTransfer.files.length > 0) {
      handleFile(e.dataTransfer.files[0]);
    }
  };

  const handleSubmit = async () => {
    if (!file) {
      setError('Please select a CSV file to import.');
      return;
    }

    try {
      setUploading(true);
      setError(null);
      const result = await api.importGoodreads(file);
      setSummary(result);
      onSuccess();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Import failed. Please verify your CSV format.');
    } finally {
      setUploading(false);
    }
  };

  const resetModal = () => {
    setFile(null);
    setSummary(null);
    setError(null);
    onClose();
  };

  return (
    <AnimatePresence>
      <div className="fixed inset-0 z-50 flex items-center justify-center p-4 sm:p-6">
        {/* Backdrop */}
        <motion.div
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          exit={{ opacity: 0 }}
          onClick={resetModal}
          className="fixed inset-0 bg-black/60 backdrop-blur-sm"
        />

        {/* Modal Window */}
        <motion.div
          initial={{ opacity: 0, scale: 0.95, y: 20 }}
          animate={{ opacity: 1, scale: 1, y: 0 }}
          exit={{ opacity: 0, scale: 0.95, y: 20 }}
          transition={{ type: 'spring', damping: 25, stiffness: 300 }}
          className="relative w-full max-w-lg bg-surface border border-border-subtle rounded-2xl shadow-2xl p-6 sm:p-7 text-text-main z-10"
        >
          {/* Header */}
          <div className="flex items-start justify-between border-b border-border-subtle pb-4 mb-5">
            <div className="flex items-center gap-3">
              <span className="p-2.5 rounded-xl bg-accent-soft text-accent border border-accent/20">
                <UploadCloud className="w-5 h-5 text-accent" />
              </span>
              <div>
                <h2 className="text-lg sm:text-xl font-serif font-bold text-text-main">
                  Migrate from Goodreads
                </h2>
                <p className="text-xs text-text-muted mt-0.5">
                  1-Click import of your reading logs, shelves, and ratings
                </p>
              </div>
            </div>
            <button
              onClick={resetModal}
              className="text-text-muted hover:text-text-main p-1.5 rounded-lg hover:bg-surface-hover transition-colors cursor-pointer"
              aria-label="Close modal"
            >
              <X className="w-5 h-5" />
            </button>
          </div>

          {/* Body */}
          {summary ? (
            /* Results Screen */
            <div className="py-4 space-y-5 text-center">
              <div className="w-14 h-14 mx-auto rounded-full bg-emerald-500/10 border border-emerald-500/20 flex items-center justify-center text-emerald-500">
                <CheckCircle2 className="w-8 h-8" />
              </div>
              <div>
                <h3 className="text-base font-semibold text-text-main">Import Complete!</h3>
                <p className="text-xs text-text-muted mt-1">
                  Your Goodreads library has been migrated into Imprint.
                </p>
              </div>

              <div className="grid grid-cols-3 gap-2.5 bg-surface-hover/50 p-3.5 rounded-xl border border-border-subtle">
                <div>
                  <p className="text-[11px] text-text-muted">Imported</p>
                  <p className="text-xl font-bold text-emerald-600 dark:text-emerald-400 mt-0.5">
                    {summary.imported_count}
                  </p>
                </div>
                <div>
                  <p className="text-[11px] text-text-muted">Skipped (Exists)</p>
                  <p className="text-xl font-bold text-text-main mt-0.5">{summary.skipped_count}</p>
                </div>
                <div>
                  <p className="text-[11px] text-text-muted">Failed</p>
                  <p className="text-xl font-bold text-amber-600 dark:text-amber-400 mt-0.5">
                    {summary.failed_count}
                  </p>
                </div>
              </div>

              <button
                onClick={resetModal}
                className="w-full py-2.5 bg-accent hover:opacity-90 text-canvas font-bold text-xs rounded-xl shadow-xs transition-all cursor-pointer"
              >
                View Collection
              </button>
            </div>
          ) : (
            /* Upload Screen */
            <div className="space-y-4">
              {/* Instructions */}
              <div className="bg-surface-hover/50 border border-border-subtle rounded-xl p-3.5 text-xs text-text-muted space-y-1.5">
                <p className="font-semibold text-text-main">How to get your Goodreads file:</p>
                <ol className="list-decimal list-inside space-y-1 text-[11px]">
                  <li>
                    Log in to{' '}
                    <a
                      href="https://www.goodreads.com/review/import"
                      target="_blank"
                      rel="noreferrer"
                      className="text-accent hover:underline inline-flex items-center gap-0.5 font-medium"
                    >
                      <span>Goodreads.com</span>
                      <ExternalLink className="w-3 h-3 inline" />
                    </a>
                  </li>
                  <li>
                    Click <strong>Export Library</strong> under Import/Export
                  </li>
                  <li>
                    Upload the exported <code className="text-accent font-mono">.csv</code> file below
                  </li>
                </ol>
              </div>

              {/* Drag-and-drop box */}
              <div
                onDragOver={(e) => {
                  e.preventDefault();
                  setDragOver(true);
                }}
                onDragLeave={() => setDragOver(false)}
                onDrop={handleDrop}
                onClick={() => fileInputRef.current?.click()}
                className={`border-2 border-dashed rounded-xl p-6 text-center cursor-pointer transition-colors ${
                  dragOver
                    ? 'border-accent bg-accent-soft/40'
                    : file
                    ? 'border-emerald-500/60 bg-emerald-500/5'
                    : 'border-border-subtle hover:border-accent/40 bg-surface-hover/20'
                }`}
              >
                <input
                  ref={fileInputRef}
                  type="file"
                  accept=".csv"
                  className="hidden"
                  onChange={(e) => e.target.files?.[0] && handleFile(e.target.files[0])}
                />

                {file ? (
                  <div className="space-y-1.5">
                    <FileText className="w-8 h-8 text-emerald-500 mx-auto" />
                    <p className="text-xs font-semibold text-text-main truncate max-w-[280px] mx-auto">
                      {file.name}
                    </p>
                    <p className="text-[10px] text-text-muted">
                      {(file.size / 1024).toFixed(1)} KB — Click or drop another to replace
                    </p>
                  </div>
                ) : (
                  <div className="space-y-2">
                    <UploadCloud className="w-8 h-8 text-accent/70 mx-auto" />
                    <p className="text-xs font-medium text-text-main">
                      Drag & drop your Goodreads CSV here, or{' '}
                      <span className="text-accent font-semibold underline underline-offset-2">
                        browse
                      </span>
                    </p>
                    <p className="text-[10px] text-text-muted">
                      Supports standard Goodreads CSV exports up to 5 MB
                    </p>
                  </div>
                )}
              </div>

              {error && (
                <div className="flex items-center gap-1.5 text-xs text-red-500 font-medium">
                  <AlertCircle className="w-3.5 h-3.5 shrink-0" />
                  <span>{error}</span>
                </div>
              )}

              {/* Actions */}
              <div className="flex gap-2.5 pt-2">
                <button
                  type="button"
                  onClick={resetModal}
                  className="flex-1 py-2.5 bg-surface hover:bg-surface-hover border border-border-subtle text-text-muted hover:text-text-main font-semibold text-xs rounded-xl transition-colors cursor-pointer"
                >
                  Cancel
                </button>
                <button
                  type="button"
                  disabled={!file || uploading}
                  onClick={handleSubmit}
                  className="flex-1 py-2.5 bg-accent hover:opacity-90 disabled:opacity-50 text-canvas font-bold text-xs rounded-xl shadow-xs transition-all flex items-center justify-center gap-2 cursor-pointer"
                >
                  {uploading ? (
                    <>
                      <div className="w-3.5 h-3.5 border-2 border-canvas border-t-transparent rounded-full animate-spin" />
                      <span>Importing Books...</span>
                    </>
                  ) : (
                    <span>Import Books</span>
                  )}
                </button>
              </div>
            </div>
          )}
        </motion.div>
      </div>
    </AnimatePresence>
  );
};
