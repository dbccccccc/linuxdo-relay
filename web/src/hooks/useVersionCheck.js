import { useEffect } from 'react';
import { useLocation } from 'react-router-dom';

const VERSION_STORAGE_KEY = 'linuxdo-relay-version';

/**
 * Hook to check for app version updates on every route change.
 * If a new version is detected, shows an alert and forces a page reload.
 */
export function useVersionCheck() {
  const location = useLocation();

  useEffect(() => {
    const checkVersion = async () => {
      try {
        // Add cache-busting query parameter to avoid cached responses
        const res = await fetch(`/healthz?_t=${Date.now()}`);
        if (!res.ok) return;

        const data = await res.json();
        const serverVersion = data.version;

        if (!serverVersion) return;

        const cachedVersion = localStorage.getItem(VERSION_STORAGE_KEY);

        if (cachedVersion && cachedVersion !== serverVersion) {
          // Version mismatch detected
          localStorage.setItem(VERSION_STORAGE_KEY, serverVersion);
          alert(`发现新版本 ${serverVersion}，页面将自动刷新以应用更新。`);
          // Force reload, bypassing cache
          window.location.reload(true);
        } else if (!cachedVersion) {
          // First visit, just save the version
          localStorage.setItem(VERSION_STORAGE_KEY, serverVersion);
        }
      } catch (err) {
        // Silently ignore errors (network issues, etc.)
        console.debug('Version check failed:', err);
      }
    };

    checkVersion();
  }, [location.pathname]); // Re-check on every route change
}
