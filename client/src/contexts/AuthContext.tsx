// TODO: Implement authentication - See docs/AUTH_IMPLEMENTATION.md
// This file currently bypasses authentication. To implement real auth:
// 1. Remove the auto-login in useEffect
// 2. Validate JWT tokens from the backend
// 3. Store and use real tokens for API requests

import React, { createContext, useState, useContext, useEffect, useMemo } from 'react';

interface AuthContextType {
  isLoggedIn: boolean;
  token: string | null;
  loading: boolean;
  isAdmin: boolean;
  role: string | null;
  pages: string[];
  hasPageAccess: (page: string) => boolean;
  login: (token: string) => void;
  logout: () => void;
}



const AuthContext = createContext<AuthContextType>({
  isLoggedIn: false,
  token: null,
  loading: true,
  isAdmin: false,
  role: null,
  pages: ['reports'],
  hasPageAccess: () => false,
  login: () => { },
  logout: () => { },
});

export const useAuth = () => useContext(AuthContext);

export const AuthProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const [token, setToken] = useState<string | null>(null);
  const [isLoggedIn, setIsLoggedIn] = useState(false);
  const [loading, setLoading] = useState(true);
  const [role, setRole] = useState<string | null>(null);
  const [pages, setPages] = useState<string[]>(['reports']);
  const [isAdmin, setIsAdmin] = useState(false);

  const decodeJwtPayload = (jwtToken: string): Record<string, unknown> | null => {
    const parts = jwtToken.split('.');
    if (parts.length !== 3) {
      return null;
    }
    try {
      const normalized = parts[1].replace(/-/g, '+').replace(/_/g, '/');
      const padded = normalized.padEnd(Math.ceil(normalized.length / 4) * 4, '=');
      return JSON.parse(atob(padded)) as Record<string, unknown>;
    } catch {
      return null;
    }
  };

  const syncFromToken = (jwtToken: string | null) => {
    if (!jwtToken) {
      setRole(null);
      setPages(['reports']);
      setIsAdmin(false);
      return;
    }
    const payload = decodeJwtPayload(jwtToken);
    const parsedRole = typeof payload?.role === 'string' ? payload.role.toLowerCase() : 'user';
    const rawPages = Array.isArray(payload?.pages) ? payload.pages : ['reports'];
    const parsedPages = rawPages
      .filter((page): page is string => typeof page === 'string' && page.trim().length > 0)
      .map((page) => page.toLowerCase());

    setRole(parsedRole);
    setPages(parsedPages.length > 0 ? parsedPages : ['reports']);
    setIsAdmin(parsedRole === 'admin');
  };

  useEffect(() => {
    const storedToken = localStorage.getItem('token');
    if (storedToken) {
      setToken(storedToken);
      setIsLoggedIn(true);
      syncFromToken(storedToken);
    }
    setLoading(false);
  }, []);

  const login = (newToken: string) => {
    localStorage.setItem('token', newToken);
    setToken(newToken);
    setIsLoggedIn(true);
    syncFromToken(newToken);
  };

  const logout = () => {
    localStorage.removeItem('token');
    setToken(null);
    setIsLoggedIn(false);
    syncFromToken(null);
  };

  const hasPageAccess = useMemo(
    () => (page: string) => isAdmin || pages.includes(page.toLowerCase()),
    [isAdmin, pages]
  );

  const value = {
    token,
    isLoggedIn,
    loading,
    isAdmin,
    role,
    pages,
    hasPageAccess,
    login,
    logout,
  };

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
};

export default AuthContext; 