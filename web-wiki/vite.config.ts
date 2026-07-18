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
    // 构建产物统一输出到 resources/web-wiki/dist，由根级 resources/embed.go 通过 go:embed 嵌入。
    // outDir 位于 web-wiki/ 之外，必须显式 emptyOutDir 以允许 Vite 清空目标目录。
    outDir: '../resources/web-wiki/dist',
    emptyOutDir: true,
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
