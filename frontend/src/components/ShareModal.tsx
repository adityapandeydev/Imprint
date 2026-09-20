import React, { useState } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import {
  X,
  Copy,
  Check,
  Share2,
  ExternalLink,
  MessageCircle,
} from 'lucide-react';
import { api } from '../lib/api';
import { toast } from '../lib/toast';

const XIcon: React.FC<{ className?: string }> = ({ className }) => (
  <svg className={className} viewBox="0 0 24 24" fill="currentColor">
    <path d="M18.244 2.25h3.308l-7.227 8.26 8.502 11.24H16.17l-5.214-6.817L4.99 21.75H1.68l7.73-8.835L1.254 2.25H8.08l4.713 6.231zm-1.161 17.52h1.833L7.084 4.126H5.117z" />
  </svg>
);

interface ShareModalProps {
  isOpen: boolean;
  onClose: () => void;
  username: string;
  displayName: string;
  shelf?: string;
}

export const ShareModal: React.FC<ShareModalProps> = ({
  isOpen,
  onClose,
  username,
  displayName,
  shelf,
}) => {
  const [copied, setCopied] = useState(false);

  if (!isOpen) return null;

  const origin = typeof window !== 'undefined' ? window.location.origin : '';
  const sharePath = shelf
    ? `/u/${encodeURIComponent(username)}/shelf/${encodeURIComponent(shelf)}`
    : `/u/${encodeURIComponent(username)}`;
  const shareUrl = `${origin}${sharePath}`;
  const ogCardUrl = api.getPublicOGCardUrl(username);

  const shareTitle = shelf
    ? `Explore ${displayName}'s "${shelf}" shelf on Imprint`
    : `Explore ${displayName}'s curated library on Imprint`;
  const shareText = `Check out my reading collection and literary velocity on Imprint: ${shareUrl}`;

  const handleCopy = async () => {
    try {
      if (navigator.clipboard) {
        await navigator.clipboard.writeText(shareUrl);
      } else {
        const textarea = document.createElement('textarea');
        textarea.value = shareUrl;
        document.body.appendChild(textarea);
        textarea.select();
        document.execCommand('copy');
        document.body.removeChild(textarea);
      }
      setCopied(true);
      toast.success('Share link copied to clipboard!');
      setTimeout(() => setCopied(false), 2500);
    } catch {
      toast.error('Failed to copy link');
    }
  };

  const handleTwitterShare = () => {
    const tweetUrl = `https://twitter.com/intent/tweet?text=${encodeURIComponent(
      shareTitle
    )}&url=${encodeURIComponent(shareUrl)}`;
    window.open(tweetUrl, '_blank', 'noopener,noreferrer');
  };

  const handleWhatsAppShare = () => {
    const waUrl = `https://api.whatsapp.com/send?text=${encodeURIComponent(
      `${shareTitle}: ${shareUrl}`
    )}`;
    window.open(waUrl, '_blank', 'noopener,noreferrer');
  };

  const handleNativeShare = async () => {
    if (navigator.share) {
      try {
        await navigator.share({
          title: shareTitle,
          text: shareText,
          url: shareUrl,
        });
      } catch {
        // User cancelled or share failed
      }
    } else {
      handleCopy();
    }
  };

  return (
    <AnimatePresence>
      <div className="fixed inset-0 z-50 flex items-center justify-center p-4 sm:p-6">
        {/* Backdrop */}
        <motion.div
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          exit={{ opacity: 0 }}
          onClick={onClose}
          className="fixed inset-0 bg-black/60 backdrop-blur-sm"
        />

        {/* Modal Window */}
        <motion.div
          initial={{ opacity: 0, scale: 0.95, y: 15 }}
          animate={{ opacity: 1, scale: 1, y: 0 }}
          exit={{ opacity: 0, scale: 0.95, y: 15 }}
          transition={{ type: 'spring', damping: 25, stiffness: 300 }}
          className="relative w-full max-w-lg bg-surface border border-border-subtle rounded-2xl shadow-2xl p-6 text-text-main z-10 custom-scrollbar"
        >
          {/* Header */}
          <div className="flex items-center justify-between border-b border-border-subtle pb-4 mb-5">
            <div className="flex items-center gap-2.5">
              <span className="p-2 rounded-xl bg-accent-soft text-accent border border-accent/20">
                <Share2 className="w-5 h-5" />
              </span>
              <div>
                <h2 className="text-lg font-serif font-bold text-text-main">
                  {shelf ? `Share "${shelf}" Shelf` : 'Share Reader Library'}
                </h2>
                <p className="text-xs text-text-muted">
                  Share your public collection, challenge pacing, and shelves
                </p>
              </div>
            </div>
            <button
              onClick={onClose}
              className="text-text-muted hover:text-text-main p-1.5 rounded-lg hover:bg-surface-hover transition-colors cursor-pointer"
              aria-label="Close dialog"
            >
              <X className="w-5 h-5" />
            </button>
          </div>

          {/* OpenGraph Preview Card */}
          <div className="mb-5">
            <p className="text-xs font-semibold uppercase tracking-wider text-text-muted mb-2">
              Social Card Preview (Twitter, WhatsApp, iMessage)
            </p>
            <div className="rounded-xl overflow-hidden border border-border-subtle bg-black/30 shadow-md aspect-[1200/630] relative group">
              <img
                src={ogCardUrl}
                alt="OpenGraph Social Card"
                className="w-full h-full object-cover transition-transform duration-300 group-hover:scale-[1.01]"
                loading="lazy"
              />
              <div className="absolute inset-0 ring-1 ring-inset ring-white/10 rounded-xl pointer-events-none" />
            </div>
          </div>

          {/* Share Link Copy Field */}
          <div className="mb-5">
            <label className="block text-xs text-text-muted mb-1.5 font-medium">
              Public Share Link:
            </label>
            <div className="flex items-center gap-2">
              <input
                type="text"
                readOnly
                value={shareUrl}
                className="flex-1 bg-surface-hover/60 border border-border-subtle rounded-lg px-3 py-2 text-xs font-mono text-text-main truncate focus:outline-none select-all"
              />
              <button
                onClick={handleCopy}
                className="inline-flex items-center gap-1.5 px-3.5 py-2 bg-accent text-white text-xs font-semibold rounded-lg hover:opacity-90 transition-opacity cursor-pointer shrink-0"
              >
                {copied ? (
                  <>
                    <Check className="w-3.5 h-3.5" />
                    <span>Copied</span>
                  </>
                ) : (
                  <>
                    <Copy className="w-3.5 h-3.5" />
                    <span>Copy</span>
                  </>
                )}
              </button>
            </div>
          </div>

          {/* Social Quick Share Actions */}
          <div className="grid grid-cols-2 sm:grid-cols-3 gap-2.5 pt-2 border-t border-border-subtle">
            <button
              onClick={handleTwitterShare}
              className="flex items-center justify-center gap-2 px-3 py-2 rounded-lg bg-surface-hover/60 border border-border-subtle hover:border-accent/40 text-xs font-medium text-text-main transition-colors cursor-pointer"
            >
              <XIcon className="w-3.5 h-3.5" />
              <span>Share on X</span>
            </button>

            <button
              onClick={handleWhatsAppShare}
              className="flex items-center justify-center gap-2 px-3 py-2 rounded-lg bg-surface-hover/60 border border-border-subtle hover:border-accent/40 text-xs font-medium text-text-main transition-colors cursor-pointer"
            >
              <MessageCircle className="w-3.5 h-3.5 text-[#25D366]" />
              <span>WhatsApp</span>
            </button>

            {typeof navigator !== 'undefined' && typeof navigator.share === 'function' && (
              <button
                onClick={handleNativeShare}
                className="col-span-2 sm:col-span-1 flex items-center justify-center gap-2 px-3 py-2 rounded-lg bg-accent-soft border border-accent/20 text-accent hover:opacity-90 text-xs font-medium transition-colors cursor-pointer"
              >
                <ExternalLink className="w-3.5 h-3.5" />
                <span>Device Share</span>
              </button>
            )}
          </div>
        </motion.div>
      </div>
    </AnimatePresence>
  );
};
