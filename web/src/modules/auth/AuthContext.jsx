import React, { createContext, useContext, useEffect, useState, useCallback } from 'react';
import axios from 'axios';

const AuthContext = createContext(null);

const STORAGE_KEY = 'linuxdo-relay-auth';

export function AuthProvider({ children }) {
  const [token, setToken] = useState(null);
  const [user, setUser] = useState(null);

  useEffect(() => {
    const raw = localStorage.getItem(STORAGE_KEY);
    if (raw) {
      try {
        const parsed = JSON.parse(raw);
        setToken(parsed.token || null);
        setUser(parsed.user || null);
      } catch {
        // ignore
      }
    }
  }, []);

  const saveAuth = (nextToken, nextUser) => {
    setToken(nextToken);
    setUser(nextUser);
    if (nextToken && nextUser) {
      localStorage.setItem(
        STORAGE_KEY,
        JSON.stringify({ token: nextToken, user: nextUser }),
      );
    } else {
      localStorage.removeItem(STORAGE_KEY);
    }
  };

  const logout = () => {
    saveAuth(null, null);
  };

  const reloadUser = useCallback(async () => {
    if (!token) return null;
    try {
      const res = await axios.get('/me', {
        headers: { Authorization: `Bearer ${token}` },
      });
      const profile = res.data;
      setUser(profile);
      localStorage.setItem(STORAGE_KEY, JSON.stringify({ token, user: profile }));
      return profile;
    } catch (err) {
      console.error('reload user failed', err);
      return null;
    }
  }, [token]);

  useEffect(() => {
    if (!token) return;
    reloadUser();
  }, [token, reloadUser]);

  const isAdmin = user?.role === 'admin';

  return (
    <AuthContext.Provider value={{ token, user, isAdmin, saveAuth, logout, reloadUser }}>
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth() {
  const ctx = useContext(AuthContext);
  if (!ctx) throw new Error('useAuth must be used within AuthProvider');
  return ctx;
}

