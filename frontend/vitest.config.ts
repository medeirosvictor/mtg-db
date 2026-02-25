import { defineConfig } from 'vitest/config';
import { svelte } from '@sveltejs/vite-plugin-svelte';
import path from 'path';

export default defineConfig({
  plugins: [svelte({ hot: !process.env.VITEST })],
  test: {
    environment: 'jsdom',
    globals: true,
    include: ['src/**/*.test.ts', 'src/**/*.test.js'],
    alias: {
      '@wails': path.resolve(__dirname, 'wailsjs'),
    },
  },
  resolve: {
    alias: {
      '@wails': path.resolve(__dirname, 'wailsjs'),
      '../../wailsjs': path.resolve(__dirname, 'wailsjs'),
    },
  },
});
