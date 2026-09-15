import React, { useState, useEffect } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import {
  X,
  Mail,
  Lock,
  User as UserIcon,
  Eye,
  EyeOff,
  Sparkles,
  LogIn,
  UserPlus,
  AlertCircle,
  Loader2,
} from 'lucide-react';
import { useAuth } from '../context/AuthContext';
import { toast } from '../lib/toast';

export const AuthModal: React.FC = () => {
  const {
    isAuthModalOpen,
    authModalMode,
    openAuthModal,
    closeAuthModal,
    login,
    register,
  } = useAuth();

  const [identifier, setIdentifier] = useState('');
  const [password, setPassword] = useState('');
  const [username, setUsername] = useState('');
  const [displayName, setDisplayName] = useState('');
  const [showPassword, setShowPassword] = useState(false);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [errorMsg, setErrorMsg] = useState<string | null>(null);

  // Reset state when modal opens or mode changes
  useEffect(() => {
    if (isAuthModalOpen) {
      setErrorMsg(null);
      setPassword('');
      setShowPassword(false);
    }
  }, [isAuthModalOpen, authModalMode]);

  // Handle ESC key to dismiss
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === 'Escape' && isAuthModalOpen) {
        closeAuthModal();
      }
    };
    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [isAuthModalOpen, closeAuthModal]);

  if (!isAuthModalOpen) return null;

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setErrorMsg(null);
    setIsSubmitting(true);

    try {
      if (authModalMode === 'login') {
        if (!identifier.trim()) {
          throw new Error('Please enter your email or username');
        }
        if (!password) {
          throw new Error('Please enter your password');
        }
        await login({ login: identifier.trim(), password });
        toast.success('Welcome back to Imprint');
      } else {
        if (!identifier.trim() || !identifier.includes('@')) {
          throw new Error('Please enter a valid email address');
        }
        if (!username.trim() || username.length < 3) {
          throw new Error('Username must be at least 3 alphanumeric characters');
        }
        if (password.length < 8) {
          throw new Error('Password must be at least 8 characters');
        }
        await register({
          email: identifier.trim(),
          username: username.trim().toLowerCase(),
          password,
          display_name: displayName.trim() || username.trim(),
        });
        toast.success('Account created successfully', 'Your collection is now ready');
      }
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : 'Authentication failed';
      setErrorMsg(msg);
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <AnimatePresence>
      <div className="fixed inset-0 z-50 flex items-center justify-center p-4">
        {/* Backdrop */}
        <motion.div
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          exit={{ opacity: 0 }}
          transition={{ duration: 0.2 }}
          onClick={closeAuthModal}
          className="fixed inset-0 bg-canvas/80 backdrop-blur-md"
        />

        {/* Modal Dialog */}
        <motion.div
          initial={{ opacity: 0, scale: 0.96, y: 12 }}
          animate={{ opacity: 1, scale: 1, y: 0 }}
          exit={{ opacity: 0, scale: 0.96, y: 12 }}
          transition={{ type: 'spring', stiffness: 450, damping: 32 }}
          className="relative w-full max-w-md bg-surface border border-border-subtle rounded-2xl p-6 sm:p-8 shadow-book z-10 space-y-6 overflow-hidden"
          onClick={(e) => e.stopPropagation()}
        >
          {/* Close button */}
          <button
            onClick={closeAuthModal}
            className="absolute top-4 right-4 p-2 text-text-muted hover:text-text-main rounded-xl hover:bg-surface-hover transition-colors cursor-pointer"
            aria-label="Close dialog"
          >
            <X className="w-5 h-5" />
          </button>

          {/* Header */}
          <div className="text-center space-y-2">
            <div className="w-12 h-12 mx-auto rounded-2xl bg-accent-soft border border-accent/20 flex items-center justify-center text-accent shadow-xs">
              <Sparkles className="w-6 h-6 text-amber-gold" />
            </div>
            <h2 className="font-serif text-2xl sm:text-3xl font-bold text-text-main tracking-tight">
              {authModalMode === 'login' ? 'Welcome Back' : 'Create an Account'}
            </h2>
            <p className="text-xs sm:text-sm text-text-muted max-w-xs mx-auto">
              {authModalMode === 'login'
                ? 'Sign in to access your reading collection, private notes, and saved editions.'
                : 'Join Imprint to curate your own personal bibliographic sanctuary.'}
            </p>
          </div>

          {/* Mode Switcher Tabs */}
          <div className="grid grid-cols-2 p-1 bg-surface-hover/60 border border-border-subtle rounded-xl relative">
            <button
              type="button"
              onClick={() => openAuthModal('login')}
              className={`py-2 text-xs sm:text-sm font-semibold rounded-lg transition-colors relative z-10 cursor-pointer ${
                authModalMode === 'login' ? 'text-accent' : 'text-text-muted hover:text-text-main'
              }`}
            >
              {authModalMode === 'login' && (
                <motion.div
                  layoutId="authTabIndicator"
                  className="absolute inset-0 bg-surface border border-border-subtle shadow-xs rounded-lg -z-10"
                  transition={{ type: 'spring', stiffness: 450, damping: 35 }}
                />
              )}
              <span className="flex items-center justify-center gap-1.5">
                <LogIn className="w-3.5 h-3.5" /> Sign In
              </span>
            </button>

            <button
              type="button"
              onClick={() => openAuthModal('register')}
              className={`py-2 text-xs sm:text-sm font-semibold rounded-lg transition-colors relative z-10 cursor-pointer ${
                authModalMode === 'register' ? 'text-accent' : 'text-text-muted hover:text-text-main'
              }`}
            >
              {authModalMode === 'register' && (
                <motion.div
                  layoutId="authTabIndicator"
                  className="absolute inset-0 bg-surface border border-border-subtle shadow-xs rounded-lg -z-10"
                  transition={{ type: 'spring', stiffness: 450, damping: 35 }}
                />
              )}
              <span className="flex items-center justify-center gap-1.5">
                <UserPlus className="w-3.5 h-3.5" /> Create Account
              </span>
            </button>
          </div>

          {/* Error Banner */}
          {errorMsg && (
            <motion.div
              initial={{ opacity: 0, y: -8 }}
              animate={{ opacity: 1, y: 0 }}
              className="flex items-start gap-2.5 p-3 rounded-xl bg-red-500/10 border border-red-500/25 text-red-600 dark:text-red-400 text-xs sm:text-sm leading-snug"
            >
              <AlertCircle className="w-4 h-4 shrink-0 mt-0.5" />
              <span>{errorMsg}</span>
            </motion.div>
          )}

          {/* Form */}
          <form onSubmit={handleSubmit} className="space-y-4">
            {authModalMode === 'register' && (
              <>
                <div>
                  <label className="block text-xs font-semibold text-text-muted mb-1 uppercase tracking-wider">
                    Display Name
                  </label>
                  <div className="relative">
                    <UserIcon className="absolute left-3.5 top-1/2 -translate-y-1/2 w-4 h-4 text-text-muted" />
                    <input
                      type="text"
                      placeholder="e.g. Mary Shelley"
                      value={displayName}
                      onChange={(e) => setDisplayName(e.target.value)}
                      className="w-full bg-canvas border border-border-subtle rounded-xl pl-10 pr-4 py-2.5 text-sm text-text-main placeholder:text-text-muted/60 focus:outline-none focus:border-accent transition-colors"
                    />
                  </div>
                </div>

                <div>
                  <label className="block text-xs font-semibold text-text-muted mb-1 uppercase tracking-wider">
                    Username
                  </label>
                  <div className="relative">
                    <span className="absolute left-3.5 top-1/2 -translate-y-1/2 text-sm text-text-muted font-mono">
                      @
                    </span>
                    <input
                      type="text"
                      required
                      placeholder="username (letters, numbers, underscores)"
                      value={username}
                      onChange={(e) => setUsername(e.target.value)}
                      className="w-full bg-canvas border border-border-subtle rounded-xl pl-9 pr-4 py-2.5 text-sm text-text-main placeholder:text-text-muted/60 focus:outline-none focus:border-accent font-mono transition-colors"
                    />
                  </div>
                </div>
              </>
            )}

            <div>
              <label className="block text-xs font-semibold text-text-muted mb-1 uppercase tracking-wider">
                {authModalMode === 'login' ? 'Email or Username' : 'Email Address'}
              </label>
              <div className="relative">
                <Mail className="absolute left-3.5 top-1/2 -translate-y-1/2 w-4 h-4 text-text-muted" />
                <input
                  type={authModalMode === 'login' ? 'text' : 'email'}
                  required
                  placeholder={
                    authModalMode === 'login'
                      ? 'reader@domain.com or username'
                      : 'reader@domain.com'
                  }
                  value={identifier}
                  onChange={(e) => setIdentifier(e.target.value)}
                  className="w-full bg-canvas border border-border-subtle rounded-xl pl-10 pr-4 py-2.5 text-sm text-text-main placeholder:text-text-muted/60 focus:outline-none focus:border-accent transition-colors"
                />
              </div>
            </div>

            <div>
              <label className="block text-xs font-semibold text-text-muted mb-1 uppercase tracking-wider">
                Password
              </label>
              <div className="relative">
                <Lock className="absolute left-3.5 top-1/2 -translate-y-1/2 w-4 h-4 text-text-muted" />
                <input
                  type={showPassword ? 'text' : 'password'}
                  required
                  placeholder={
                    authModalMode === 'register'
                      ? 'Minimum 8 characters'
                      : 'Enter your password'
                  }
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                  className="w-full bg-canvas border border-border-subtle rounded-xl pl-10 pr-11 py-2.5 text-sm text-text-main placeholder:text-text-muted/60 focus:outline-none focus:border-accent transition-colors"
                />
                <button
                  type="button"
                  onClick={() => setShowPassword(!showPassword)}
                  className="absolute right-3 top-1/2 -translate-y-1/2 p-1 text-text-muted hover:text-text-main rounded-md transition-colors cursor-pointer"
                  aria-label={showPassword ? 'Hide password' : 'Show password'}
                >
                  {showPassword ? (
                    <EyeOff className="w-4 h-4" />
                  ) : (
                    <Eye className="w-4 h-4" />
                  )}
                </button>
              </div>
            </div>

            <button
              type="submit"
              disabled={isSubmitting}
              className="w-full mt-2 py-3 px-4 bg-accent text-canvas font-semibold rounded-xl text-sm hover:opacity-90 transition-opacity shadow-xs flex items-center justify-center gap-2 cursor-pointer disabled:opacity-50"
            >
              {isSubmitting ? (
                <>
                  <Loader2 className="w-4 h-4 animate-spin" />
                  <span>{authModalMode === 'login' ? 'Signing In...' : 'Creating Account...'}</span>
                </>
              ) : (
                <span>{authModalMode === 'login' ? 'Sign In to Imprint' : 'Create My Account'}</span>
              )}
            </button>
          </form>

          {/* Footer toggle */}
          <div className="text-center text-xs text-text-muted pt-2 border-t border-border-subtle">
            {authModalMode === 'login' ? (
              <p>
                New reader?{' '}
                <button
                  type="button"
                  onClick={() => openAuthModal('register')}
                  className="font-semibold text-accent hover:underline cursor-pointer"
                >
                  Create an account
                </button>
              </p>
            ) : (
              <p>
                Already have an account?{' '}
                <button
                  type="button"
                  onClick={() => openAuthModal('login')}
                  className="font-semibold text-accent hover:underline cursor-pointer"
                >
                  Sign in
                </button>
              </p>
            )}
          </div>
        </motion.div>
      </div>
    </AnimatePresence>
  );
};
