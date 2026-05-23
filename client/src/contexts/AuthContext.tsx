import React, { createContext, useState, useContext, useEffect } from 'react';

interface AuthContextType {
  isLoggedIn: boolean;
  token: string | null;
  loading: boolean;
  isAdmin: boolean;
  role: 'admin' | 'user' | null;
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
  pages: [],
  hasPageAccess: () => false,
  login: () => { },
  logout: () => { },
});

export const useAuth = () => useContext(AuthContext);

export const AuthProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const [token, setToken] = useState<string | null>(null);
  const [isLoggedIn, setIsLoggedIn] = useState(false);
  const [loading, setLoading] = useState(true);
  const [role, setRole] = useState<'admin' | 'user' | null>(null);
  const [pages, setPages] = useState<string[]>([]);
  const [isAdmin, setIsAdmin] = useState(false);

  const decodeJwtPayload = (jwtToken: string): Record<string, unknown> | null => {
    try {
      const payloadBase64 = jwtToken.split('.')[1];
      if (!payloadBase64) return null;
      const payloadJson = atob(payloadBase64.replace(/-/g, '+').replace(/_/g, '/'));
      return JSON.parse(payloadJson);
    } catch {
      return null;
    }
  };

  const syncFromToken = (jwtToken: string) => {
    const payload = decodeJwtPayload(jwtToken);
    if (!payload) {
      throw new Error('Invalid token');
    }
    const tokenRole = payload.role === 'admin' ? 'admin' : 'user';
    const tokenPages = Array.isArray(payload.pages)
      ? payload.pages.filter((p): p is string => typeof p === 'string')
      : ['reports'];

    setRole(tokenRole);
    setPages(tokenRole === 'admin' ? ['photos', 'devices', 'statistics', 'reports', 'users'] : tokenPages);
    setIsAdmin(tokenRole === 'admin');
  };

  useEffect(() => {
    const storedToken = localStorage.getItem('token');
    if (storedToken) {
      try {
        syncFromToken(storedToken);
        setToken(storedToken);
        setIsLoggedIn(true);
      } catch {
        localStorage.removeItem('token');
      }
    }
    setLoading(false);
  }, []);

  const login = (newToken: string) => {
    syncFromToken(newToken);
    localStorage.setItem('token', newToken);
    setToken(newToken);
    setIsLoggedIn(true);
  };

  const logout = () => {
    localStorage.removeItem('token');
    setToken(null);
    setIsLoggedIn(false);
    setRole(null);
    setPages([]);
    setIsAdmin(false);
  };

  const hasPageAccess = (page: string) => {
    if (!isLoggedIn) return false;
    if (isAdmin) return true;
    return pages.includes(page);
  };

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