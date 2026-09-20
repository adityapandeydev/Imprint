import React, { createContext, useContext, useState, useEffect, useCallback } from 'react';
import { api, tokenStorage } from '../lib/api';
import type { User, LoginRequest, RegisterRequest, ProfileVisibility } from '../types/api';

interface AuthContextType {
  user: User | null;
  isAuthenticated: boolean;
  isLoading: boolean;
  login: (req: LoginRequest) => Promise<void>;
  register: (req: RegisterRequest) => Promise<void>;
  logout: () => Promise<void>;
  updateUserProfileVisibility: (visibility: ProfileVisibility) => Promise<void>;
  isAuthModalOpen: boolean;
  authModalMode: 'login' | 'register';
  openAuthModal: (mode?: 'login' | 'register') => void;
  closeAuthModal: () => void;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export const AuthProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const [user, setUser] = useState<User | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [isAuthModalOpen, setIsAuthModalOpen] = useState(false);
  const [authModalMode, setAuthModalMode] = useState<'login' | 'register'>('login');

  // Verify and restore authenticated session on initial mount
  useEffect(() => {
    let isMounted = true;
    const restoreSession = async () => {
      try {
        const currentUser = await api.getMe();
        if (isMounted && currentUser?.id) {
          setUser(currentUser);
        }
      } catch {
        if (isMounted) {
          tokenStorage.clear();
          setUser(null);
        }
      } finally {
        if (isMounted) {
          setIsLoading(false);
        }
      }
    };

    restoreSession();
    return () => {
      isMounted = false;
    };
  }, []);

  const login = useCallback(async (req: LoginRequest) => {
    const tokens = await api.login(req);
    setUser(tokens.user);
    setIsAuthModalOpen(false);
  }, []);

  const register = useCallback(async (req: RegisterRequest) => {
    const tokens = await api.register(req);
    setUser(tokens.user);
    setIsAuthModalOpen(false);
  }, []);

  const logout = useCallback(async () => {
    try {
      await api.logout();
    } finally {
      setUser(null);
    }
  }, []);

  const updateUserProfileVisibility = useCallback(async (visibility: ProfileVisibility) => {
    const res = await api.updateProfileVisibility(visibility);
    setUser((prev) => (prev ? { ...prev, profile_visibility: res.profile_visibility as ProfileVisibility } : null));
  }, []);

  const openAuthModal = useCallback((mode: 'login' | 'register' = 'login') => {
    setAuthModalMode(mode);
    setIsAuthModalOpen(true);
  }, []);

  const closeAuthModal = useCallback(() => {
    setIsAuthModalOpen(false);
  }, []);

  return (
    <AuthContext.Provider
      value={{
        user,
        isAuthenticated: Boolean(user),
        isLoading,
        login,
        register,
        logout,
        updateUserProfileVisibility,
        isAuthModalOpen,
        authModalMode,
        openAuthModal,
        closeAuthModal,
      }}
    >
      {children}
    </AuthContext.Provider>
  );
};

export function useAuth(): AuthContextType {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
}
