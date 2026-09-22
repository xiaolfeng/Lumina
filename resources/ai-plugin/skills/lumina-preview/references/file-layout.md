# Preview 文件布局约束

Preview 沙盒用 `iframe sandbox="allow-scripts"`（不加 `allow-same-origin`）渲染。文件必须能在同层互相引用。

## 文件名

- 扁平单层：`index.html`、`style.css`、`app.js`、`data.json`
- 禁止 `/`、`\`、`..`，禁止 `assets/style.css` 这种子目录
- 长度 1–255

## 相对引用

HTML 里可用同层相对路径加载经典 CSS/JS：

```html
<link rel="stylesheet" href="style.css">
<script src="app.js"></script>
```

不要写 `/style.css`、`assets/app.js` 或绝对 URL 指到本会话里还不存在的文件。沙盒没有 `allow-same-origin`，模块脚本从同层 `.js` 文件加载不属于可靠支持范围；React/Vue、CDN 和 ESM 的具体方式见 [`framework-runtime.md`](./framework-runtime.md)。

## 大小与编码

- `content` 是完整文件文本，不要包 Markdown 围栏
- UTF-8 编码后单文件不超过 **256 KiB**
- 允许空文件，但会话必须最终有一个 HTML 入口

## 支持的文本类型

HTML、CSS、JavaScript / MJS、JSON、SVG、纯文本。React/Vue 可通过浏览器构建或预先构建的静态产物运行，但 Lumina 不执行 npm 安装或构建命令。二进制图片请改用 SVG 或 Q&A `image` 题型，不要塞进 Preview。

## 清单核对

全部上传后调用 `preview_file_list`：

- 没有文件 → 继续 `preview_file_upload`
- 没有 HTML 入口 → 先上传 HTML
- `workflow` 区段下的 `state` 标签为 `ready_for_review` 且依赖齐全 → 才能交付 URL 或挂 Q&A
