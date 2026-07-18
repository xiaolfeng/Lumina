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
    // 构建产物统一输出到 resources/web/dist，由根级 resources/embed.go 通过 go:embed 嵌入。
    // outDir 位于 web/ 之外，必须显式 emptyOutDir 以允许 Vite 清空目标目录。
    outDir: '../resources/web/dist',
    emptyOutDir: true,
    // mermaid 通过 rehype-mermaid 拖入完整 mermaid 库（~2.8MB），
    // 已通过 React.lazy 按需加载，仅在内容含 ```mermaid 时才下载。
    chunkSizeWarningLimit: 2900,
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
            {
              name: 'vendor-motion',
              test: /node_modules[\/](motion|motion-dom|motion-utils|framer-motion)/,
              priority: 12,
            },
            {
              name: 'vendor-mermaid',
              test: /node_modules[\/](rehype-mermaid|mermaid|cytoscape|dagre|d3|elkjs|khroma|web-worker)/,
              priority: 10,
            },
          ],
        },
      },
    },
  },
})

export default config
