/**
 * SandboxFrame — 沙盒化 HTML 即时渲染器
 *
 * 基于浏览器原生 <iframe sandbox="allow-scripts" srcDoc> 渲染不可信的
 * HTML / CSS / JS 三件套。与旧 ShadowHtml（Shadow DOM + DOMPurify 白名单）的
 * 核心差异：
 *
 * - 完整支持 <script> / <style> / <form> —— JS 与交互 demo 可真实运行，
 *   不再被标签黑名单阉割
 * - 安全边界从「白名单清洗」升级为「浏览器 opaque origin 隔离」：
 *   sandbox="allow-scripts" 刻意不配 allow-same-origin，让脚本运行在一个
 *   无法访问父页面 cookie / localStorage / DOM / API 的隔离 origin 中
 *   （一旦 allow-scripts 与 allow-same-origin 并用，iframe 可自行移除 sandbox，
 *   隔离即失效，故二者永不共存）
 * - 主题变量通过读取根元素计算样式注入 iframe 的 :root，暗黑模式随 .dark 类切换
 * - 内容高度经 postMessage 回传，iframe 自适应内容高度（无内部滚动条）
 */

import { useEffect, useMemo, useRef, useState } from 'react';

export interface SandboxFrameProps {
	/** 原始 HTML 片段（body 内容，可含 <style>/<script>） */
	content: string;
	/** 额外注入 <head> 的 CSS 字符串（可选） */
	css?: string;
	/** 作用于 iframe 元素的 className（布局，不作用于内容排版） */
	className?: string;
	/** iframe 无障碍标题 */
	title?: string;
}

/** 高度回传消息类型 —— 与 iframe 内注入脚本约定 */
const HEIGHT_MSG_TYPE = 'lumina:frame-height';

/**
 * 注入 iframe 的主题 CSS 变量白名单（微明主题语义前缀）。
 * 仅收集主题命名空间变量，避免把父页面 `:root` 上其它可能携带敏感值的
 * `--*` 自定义属性（如动态注入的 token / 密钥）一并交给沙箱内不可信内容脚本读取。
 */
const THEME_VAR_PREFIXES = [
	'--sea-',
	'--lagoon',
	'--palm',
	'--sand',
	'--foam',
	'--line',
];

/** 收集根元素上主题命名空间的 CSS 自定义属性（仅主题变量） */
function collectThemeVariables(): string {
	const styles = getComputedStyle(document.documentElement);
	return Array.from({ length: styles.length }, (_, i) => styles.item(i))
		.filter((name) =>
			THEME_VAR_PREFIXES.some((prefix) => name.startsWith(prefix)),
		)
		.map((name) => `${name}: ${styles.getPropertyValue(name).trim()};`)
		.join('\n');
}

/**
 * 基础排版样式 —— 与旧 ShadowHtml 的 BASE_STYLE 对齐，仅作用于 iframe 内文档。
 * 作为「可读性底线」兜底：Agent 推送的纯文档 HTML 也能排版美观、主题一致；
 * 用户自带 <style>（位于 body，靠后声明）可自然覆盖本层。
 * 注：iframe 无法继承宿主的 Manrope @font-face，故回退系统 UI 字体栈。
 */
const BASE_STYLE = `
	html {
		font-family: ui-sans-serif, system-ui, -apple-system, 'Segoe UI', sans-serif;
		color: var(--sea-ink, #33271c);
	}
	body { margin: 0; line-height: 1.6; background: transparent; }
	* { box-sizing: border-box; }
	h1, h2, h3, h4, h5, h6 { margin: 0 0 0.5em; font-weight: 600; line-height: 1.3; color: var(--sea-ink, #33271c); }
	h1 { font-size: 1.75em; }
	h2 { font-size: 1.3em; }
	h3 { font-size: 1.1em; }
	p { margin: 0 0 0.75em; color: var(--sea-ink-soft, #6b5f52); }
	a { color: var(--lagoon-deep, #7a4e1a); text-decoration: underline; }
	strong { color: var(--sea-ink, #33271c); font-weight: 600; }
	ul, ol { margin: 0 0 0.75em; padding-left: 1.5em; color: var(--sea-ink-soft, #6b5f52); }
	li { margin-bottom: 0.25em; }
	blockquote {
		margin: 0 0 0.75em;
		padding: 0.5em 1em;
		border-left: 3px solid var(--lagoon, #c9883a);
		background: color-mix(in srgb, var(--lagoon, #c9883a) 5%, transparent);
		border-radius: 0 6px 6px 0;
		color: var(--sea-ink-soft, #6b5f52);
	}
	table { width: 100%; border-collapse: collapse; margin-bottom: 0.75em; font-size: 0.875em; }
	th {
		background: var(--foam, #fdf8f0);
		padding: 0.5em 0.75em;
		text-align: left;
		font-weight: 600;
		color: var(--sea-ink, #33271c);
		border: 1px solid var(--line, rgba(51,39,28,0.11));
	}
	td {
		padding: 0.5em 0.75em;
		border-top: 1px solid var(--line, rgba(51,39,28,0.11));
		color: var(--sea-ink-soft, #6b5f52);
	}
	code {
		padding: 0.1em 0.35em;
		border-radius: 4px;
		font-size: 0.85em;
		font-family: ui-monospace, 'SF Mono', Menlo, monospace;
		background: color-mix(in srgb, var(--lagoon, #c9883a) 10%, transparent);
		color: var(--lagoon-deep, #7a4e1a);
	}
	pre {
		margin: 0 0 0.75em;
		padding: 0.75em;
		border-radius: 8px;
		border: 1px solid var(--line, rgba(51,39,28,0.11));
		background: var(--foam, #fdf8f0);
		font-size: 0.8125em;
		font-family: ui-monospace, 'SF Mono', Menlo, monospace;
		overflow-x: auto;
		color: var(--sea-ink, #33271c);
	}
	pre code { background: transparent; padding: 0; color: inherit; font-size: inherit; }
	hr { border: none; border-top: 1px solid var(--line, rgba(51,39,28,0.11)); margin: 1em 0; }
	img { max-width: 100%; height: auto; border-radius: 8px; }
`;

/**
 * 构建 iframe 完整 HTML 文档。
 * base 样式置于 <head>，用户 content 置于 <body>（靠后声明，其 <style> 自然覆盖 base）。
 */
function buildDoc(content: string, css: string, themeVars: string): string {
	return `<!DOCTYPE html>
<html lang="zh-CN">
<head>
<meta charset="utf-8" />
<meta name="viewport" content="width=device-width, initial-scale=1" />
<style>
:root {
${themeVars}
}
${BASE_STYLE}
${css}
</style>
</head>
<body>
${content}
<script>
(function () {
	var last = -1;
	function postHeight() {
		var h = Math.max(
			document.documentElement.scrollHeight,
			document.body ? document.body.scrollHeight : 0
		);
		if (h !== last) {
			last = h;
			parent.postMessage({ type: '${HEIGHT_MSG_TYPE}', height: h }, '*');
		}
	}
	window.addEventListener('load', postHeight);
	document.addEventListener('DOMContentLoaded', postHeight);
	if (window.ResizeObserver) {
		new ResizeObserver(postHeight).observe(document.documentElement);
	}
	// 兜底轮询：捕获图片/字体懒加载与 JS 动态渲染导致的高度变化
	setInterval(postHeight, 500);
})();
</script>
</body>
</html>`;
}

export function SandboxFrame({ content, css, className, title }: SandboxFrameProps) {
	const frameRef = useRef<HTMLIFrameElement>(null);
	const [mounted, setMounted] = useState(false);
	const [height, setHeight] = useState(0);
	const [themeVars, setThemeVars] = useState<string>(() =>
		typeof document !== 'undefined' ? collectThemeVariables() : ''
	);

	// 首次挂载后再渲染 iframe，规避 SSR 下 srcDoc 序列化与 document 访问问题
	useEffect(() => {
		setMounted(true);
	}, []);

	// 主题切换（.dark 类增删）时重新收集变量 → srcDoc 重算 → iframe 重载
	useEffect(() => {
		if (typeof MutationObserver === 'undefined') return;
		const observer = new MutationObserver(() => setThemeVars(collectThemeVariables()));
		observer.observe(document.documentElement, { attributes: true, attributeFilter: ['class'] });
		return () => observer.disconnect();
	}, []);

	// 监听 iframe 高度回传 —— 用 e.source 校验来源（sandbox 下 origin 恒为 "null"，不可靠）
	useEffect(() => {
		const onMessage = (e: MessageEvent) => {
			const frame = frameRef.current;
			if (!frame || e.source !== frame.contentWindow) return;
			const data = e.data;
			if (data && data.type === HEIGHT_MSG_TYPE && typeof data.height === 'number') {
				setHeight(data.height);
			}
		};
		window.addEventListener('message', onMessage);
		return () => window.removeEventListener('message', onMessage);
	}, []);

	const doc = useMemo(
		() => buildDoc(content, css ?? '', themeVars),
		[content, css, themeVars]
	);

	// SSR / 首次渲染占位，与旧 ShadowHtml 行为一致（内容在客户端挂载后呈现）
	if (!mounted) {
		return <div className={className} />;
	}

	return (
		<iframe
			ref={frameRef}
			title={title ?? '即时渲染预览'}
			sandbox="allow-scripts"
			srcDoc={doc}
			style={{ height }}
			className={`block w-full border-0 ${className ?? ''}`}
		/>
	);
}
