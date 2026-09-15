import tailwindcss from '@tailwindcss/vite';
import react from '@vitejs/plugin-react';
import { defineConfig } from 'vite';

// https://vite.dev/config/
export default defineConfig({
  plugins: [react(), tailwindcss()],
  server: {
    port: 5173,
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
        configure: (proxy) => {
          proxy.on('error', (err: any) => {
            // Ignore client-initiated aborts from typing in debounced search
            if (err?.code === 'ECONNRESET' || err?.message?.includes('socket hang up')) {
              return;
            }
            console.error('[vite proxy error]', err?.message || err);
          });
        },
      },
      '/health': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
    },
  },
});
