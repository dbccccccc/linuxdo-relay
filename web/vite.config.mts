import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';

export default defineConfig({
  plugins: [react()],
  server: {
    port: 5174,
    proxy: {
      '/auth': 'http://localhost:8080',
      '/me': 'http://localhost:8080',
      '/admin': 'http://localhost:8080',
      '/v1': 'http://localhost:8080',
    },
  },
  test: {
    environment: 'jsdom',
    setupFiles: './vitest.setup.js',
  },
});

