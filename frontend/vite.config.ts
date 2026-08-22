import { loadEnv } from 'vite'
import { defineConfig } from 'vitest/config'
import react from '@vitejs/plugin-react'
import { resolveApiBaseUrl } from './src/config/api-url'

// https://vite.dev/config/
export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), '')
  const publicApiUrl = process.env.VITE_API_URL || env.VITE_API_URL
  const proxyTarget = process.env.VITE_API_PROXY_TARGET
    || env.VITE_API_PROXY_TARGET
    || 'http://localhost:8081'

  // Fail config loading instead of shipping an insecure or malformed
  // browser-facing API origin. Same-origin /api/v1 is the default.
  resolveApiBaseUrl(publicApiUrl)

  return {
    plugins: [react()],
    server: {
      proxy: {
        '/api': {
          target: proxyTarget,
          changeOrigin: true,
        },
      },
    },
    test: {
      globals: true,
      environment: 'jsdom',
      coverage: {
        provider: 'v8',
        reporter: ['text', 'json-summary', 'lcov'],
        reportsDirectory: './coverage',
        exclude: [
          'src/**/*.test.{ts,tsx}',
          'src/**/__tests__/**',
          'src/main.tsx',
          'src/vite-env.d.ts',
        ],
      },
    },
  }
})
