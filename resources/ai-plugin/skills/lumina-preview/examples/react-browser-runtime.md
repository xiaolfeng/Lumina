# React 浏览器端原型

本示例不需要 npm 或构建命令。React 由浏览器从固定版本 CDN 加载，应用使用经典 `app.js` 与 `React.createElement`，适配 Preview 的 opaque-origin 沙盒。

详细边界见 [`../references/framework-runtime.md`](../references/framework-runtime.md)。

## `index.html`

```html
<!doctype html>
<html lang="zh-CN">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>React Preview</title>
  <link rel="stylesheet" href="style.css">
</head>
<body>
  <main id="root"></main>
  <script crossorigin src="https://unpkg.com/react@18/umd/react.production.min.js"></script>
  <script crossorigin src="https://unpkg.com/react-dom@18/umd/react-dom.production.min.js"></script>
  <script src="app.js"></script>
</body>
</html>
```

## `app.js`

```js
const { useState } = React

function Counter() {
  const [count, setCount] = useState(0)
  return React.createElement(
    'section',
    { className: 'card' },
    React.createElement('h1', null, 'React Preview'),
    React.createElement('p', null, `Count: ${count}`),
    React.createElement(
      'button',
      { type: 'button', onClick: () => setCount((value) => value + 1) },
      'Increment',
    ),
  )
}

ReactDOM.createRoot(document.getElementById('root')).render(
  React.createElement(Counter),
)
```

## `style.css`

```css
* { box-sizing: border-box; }
body { margin: 0; min-height: 100vh; display: grid; place-items: center; font-family: sans-serif; }
.card { width: min(28rem, calc(100vw - 2rem)); padding: 2rem; border: 1px solid #d8d4cc; }
button { padding: 0.65rem 1rem; cursor: pointer; }
```

将三个文件逐一上传后，调用 `preview_file_list` 核对入口和清单。若用户明确需要 JSX，可按框架运行约束引入浏览器端转换器，但优先保持此无转换方案。
