import { defineConfig } from 'vite'
import { tanstackRouter } from '@tanstack/router-plugin/vite'
import viteReact from '@vitejs/plugin-react'
import tailwindcss from '@tailwindcss/vite'

const config = defineConfig({
  base: '/wiki/',
  resolve: { tsconfigPaths: true },
  server: {
    port: 3001,
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
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
              test: /node_modules[/](react|react-dom|scheduler)/,
              priority: 20,
            },
            {
              name: 'vendor-mermaid',
              test: /node_modules[/](mermaid)/,
              priority: 15,
            },
            {
              name: 'vendor-markdown',
              test: /node_modules[/](react-markdown|remark-|rehype-|unified|micromark|highlight\.js)/,
              priority: 12,
            },
          ],
        },
      },
    },
  },
})

export default config
