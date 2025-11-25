import { useCallback, useEffect, useState } from 'react';
import axios from 'axios';

const READY_STATE = { mode: 'ready' };

export function useSetupStatus() {
  const [status, setStatus] = useState(READY_STATE);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  const refresh = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const res = await axios.get('/setup/status');
      setStatus(res.data);
    } catch (err) {
      if (err.response?.status === 404) {
        setStatus(READY_STATE);
      } else {
        setError(err.response?.data?.error || err.message);
      }
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    refresh();
  }, [refresh]);

  return { status, loading, error, refresh };
}
