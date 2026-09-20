# Vue 浏览器端原型

本示例不需要 npm 或构建命令。Vue 由浏览器从固定版本 CDN 加载，应用使用带模板编译器的全局浏览器构建和经典 `app.js`，适配 Preview 的 opaque-origin 沙盒。

详细边界见 [`../references/framework-runtime.md`](../references/framework-runtime.md)。

## `index.html`

```html
<!doctype html>
<html lang="zh-CN">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>Vue Preview</title>
  <link rel="stylesheet" href="style.css">
</head>
<body>
  <main id="app"></main>
  <script src="https://unpkg.com/vue@3/dist/vue.global.prod.js"></script>
  <script src="app.js"></script>
</body>
</html>
```

## `app.js`

```js
Vue.createApp({
  data: () => ({ count: 0 }),
  template: `
    <section class="card">
      <h1>Vue Preview</h1>
      <p>Count: {{ count }}</p>
      <button type="button" @click="count++">Increment</button>
    </section>
  `,
}).mount('#app')
```

## `style.css`

```css
* { box-sizing: border-box; }
body { margin: 0; min-height: 100vh; display: grid; place-items: center; font-family: sans-serif; }
.card { width: min(28rem, calc(100vw - 2rem)); padding: 2rem; border: 1px solid #d8d4cc; }
button { padding: 0.65rem 1rem; cursor: pointer; }
```

将三个文件逐一上传后，调用 `preview_file_list` 核对入口和清单。标准 `.vue` 单文件组件不能直接交给 Preview 编译；需要这类源码时，先在真实项目中构建，再上传浏览器产物。
