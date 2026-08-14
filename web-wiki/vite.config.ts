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
    modulePreload: {
      // mermaid 应随 Markdown 的 ```mermaid 代码块按需加载，禁止在入口首屏 preload
      resolveDependencies: (_filename, deps) =>
        deps.filter((dep) => !dep.includes('vendor-mermaid')),
    },
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
              name: 'vendor-orama',
              test: /node_modules[/]@orama/,
              priority: 18,
            },
            {
              // 完整捕获 mermaid 生态（含 rehype-mermaid/cytoscape/dagre/d3 等），
              // priority 高于 vendor-markdown，避免被 rehype- 规则抢走导致拆分碎片化
              name: 'vendor-mermaid',
              test: /node_modules[/](rehype-mermaid|mermaid|cytoscape|dagre|d3|elkjs|khroma|web-worker)/,
              priority: 16,
            },
            {
              name: 'vendor-markdown',
              test: /node_modules[/](react-markdown|remark-|rehype-(?!mermaid)|unified|micromark|highlight\.js)/,
              priority: 12,
            },
          ],
        },
      },
    },
  },
})

export default config
