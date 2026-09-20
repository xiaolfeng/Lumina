# Preview 浏览器端框架运行约束

Preview / Pages 托管的是浏览器可直接执行的静态文本文件，不提供 Node.js、包管理器、开发服务器、打包器、SSR 或服务端模板运行时。

## 可用方式

### React

优先使用 UMD 浏览器构建，把 JSX-free 应用代码放在经典 `app.js` 中：

```html
<script crossorigin src="https://unpkg.com/react@18/umd/react.production.min.js"></script>
<script crossorigin src="https://unpkg.com/react-dom@18/umd/react-dom.production.min.js"></script>
<script src="app.js"></script>
```

```js
const root = ReactDOM.createRoot(document.getElementById('root'))
root.render(React.createElement('h1', null, 'React Preview'))
```

需要 JSX 时，可以在入口 HTML 中额外引入浏览器端转换器（例如 Babel Standalone），但会增加下载与启动开销；用于轻量原型即可，不要把它描述为生产构建。

### Vue

优先使用带模板编译器的全局浏览器构建，应用代码仍放在经典 `app.js` 中：

```html
<script src="https://unpkg.com/vue@3/dist/vue.global.prod.js"></script>
<script src="app.js"></script>
```

```js
Vue.createApp({
  data: () => ({ count: 0 }),
  template: '<button @click="count++">{{ count }}</button>',
}).mount('#app')
```

### ESM 与 Import Map

入口 HTML 可内联 `<script type="module">`，并从允许跨源模块加载的 CDN 导入 React、Vue 或其它 ESM 依赖。iframe 没有 `allow-same-origin`，因此文档的 origin 为 opaque origin；不要让模块脚本从 Preview / Pages 的同层 `.js` 文件加载，也不要假设任意 CDN 都允许这种跨源请求。

如果需要本地多模块源码、JSX/TSX、Vue SFC、npm 依赖解析或构建插件，先在真实项目里完成构建，再把浏览器可直接执行的扁平静态产物上传；产物仍须满足文件名与大小限制。

## CDN 与离线边界

- Lumina 不会替原型安装依赖；框架代码由浏览器从明确的 HTTPS CDN 地址获取。
- 固定框架和插件版本，避免 `latest` 导致 Preview 与 Pages 随上游变化。
- 外网不可用、CDN 被 CSP / 网络策略拦截或 CDN 未提供 CORS 时，外部依赖无法加载。
- 需要离线或内网运行时，改用可访问的内网静态源，或上传已经构建且满足大小限制的浏览器产物。

## “实时”含义

Preview 文件上传、覆写、行级编辑或删除后，WebSocket 会通知已打开的工作台，工作台通过新的文件 URL 重新加载当前 iframe。它是自动刷新，不是 React Fast Refresh、Vue HMR，也不保证保留组件状态。

Pages 是不可变快照，没有实时同步。Preview 中核对通过后可晋升；继续修改已发布页面时，先 Fork 为新 Preview 会话，再晋升新版本。外部 CDN 依赖不会被复制进快照，因此 CDN 版本和可用性同样影响已发布 Pages。
