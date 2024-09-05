import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [react()],
  resolve: {
    alias: {
      "@apiclient": "/src/api-client",
      "@components": "/src/components",
      src: "/src",
    },
  },
  server: {
    proxy: {
      '/instaman/': {
        changeOrigin: true,
        target: 'http://127.0.0.1:10000',
      },
    },
  },
})
