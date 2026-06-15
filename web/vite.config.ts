import { defineConfig } from 'vite'
import { tanstackRouter } from '@tanstack/router-plugin/vite'
import viteReact from '@vitejs/plugin-react'
import tailwindcss from '@tailwindcss/vite'

const config = defineConfig({
  resolve: { tsconfigPaths: true },
  server: {
    proxy: {
      '/api': {
        target: 'http://localhost:8800',
        changeOrigin: true,
        ws: true,
      },
      '/swagger': {
        target: 'http://localhost:8800',
        changeOrigin: true,
      },
    },
  },
  plugins: [
    // MUST come before react()
    tanstackRouter({
      target: 'react',
      autoCodeSplitting: true,
    }),
    tailwindcss(),
    viteReact(),
  ],
  build: {
    chunkSizeWarningLimit: 700,
    rolldownOptions: {
      output: {
        codeSplitting: {
          groups: [
            {
              name: 'vendor-react',
              test: /node_modules[\/](react|react-dom|scheduler)/,
              priority: 20,
            },
            {
              name: 'vendor-ui',
              test: /node_modules[\/](lucide-react|@radix-ui|class-variance-authority|tailwind-merge|clsx|@tailwindcss|sonner)/,
              priority: 15,
            },
          ],
        },
      },
    },
  },
})

export default config
