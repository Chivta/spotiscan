import { defineConfig } from 'vite'
import { svelte } from '@sveltejs/vite-plugin-svelte'

export default defineConfig({
  base: '/static/',
  plugins: [svelte()],
  server: {
    proxy: {
      '/playlist': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
    },
  },
});