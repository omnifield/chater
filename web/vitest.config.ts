import solid from 'vite-plugin-solid';
import { defineConfig } from 'vitest/config';

// jsdom for DOM/Solid tests (shared-policy §3). Pure-logic tests run in the same
// jsdom env for simplicity — they don't touch the DOM.
export default defineConfig({
  plugins: [solid({ hot: false })],
  resolve: { dedupe: ['solid-js', 'solid-js/web'] },
  test: {
    include: ['src/**/__tests__/**/*.test.ts', 'src/**/__tests__/**/*.test.tsx'],
    environment: 'jsdom',
    globals: false,
    setupFiles: ['./src/test-setup.ts'],
  },
});
