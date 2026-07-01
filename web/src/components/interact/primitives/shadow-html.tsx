/**
 * ShadowHtml — 沙盒化 HTML 渲染器
 *
 * 使用 Shadow DOM 将外部 HTML 的 CSS 完全隔离在影子边界内，
 * 不会泄漏到 Lumina 全局样式；同时 DOMPurify 清洗 XSS 向量。
 *
 * 关键特性：
 * - Shadow DOM open 模式，CSS 隔离但可调试
 * - CSS 自定义属性（var(--sea-ink) 等）天然穿透 Shadow 边界 → 主题色继承
 * - inheritable 属性（font-family、color、line-height）从宿主继承
 * - DOMPurify 清洗 <script>、onerror、onclick 等 XSS 向量
 * - 内容变化时自动更新 shadowRoot.innerHTML
 */

import { useEffect, useRef } from 'react';
import DOMPurify from 'dompurify';
import type { Config as PurifyConfig } from 'dompurify';

export interface ShadowHtmlProps {
	/** 原始 HTML 字符串（未清洗，组件内部会过 DOMPurify） */
	content: string;
	/** 额外注入影子根的 CSS 字符串（可选） */
	css?: string;
	/** 容器 className（作用于宿主元素，不进入 Shadow DOM） */
	className?: string;
}

/** 影子根内的基础排版样式 —— 薄层，仅保证可读性 */
const BASE_STYLE = `
:host {
	display: block;
	/* 继承宿主的字体栈和颜色 */
	font-family: inherit;
	color: inherit;
	line-height: inherit;
}
/* 基础排版重置 —— 仅作用于影子根内部 */
* { box-sizing: border-box; }
h1, h2, h3, h4, h5, h6 { margin: 0 0 0.5em; font-weight: 600; line-height: 1.3; }
h1 { font-size: 1.75em; color: var(--sea-ink, #2a2420); }
h2 { font-size: 1.3em; color: var(--sea-ink, #2a2420); }
h3 { font-size: 1.1em; color: var(--sea-ink, #2a2420); }
p { margin: 0 0 0.75em; color: var(--sea-ink-soft, #5c534a); }
a { color: var(--lagoon-deep, #7a4e1a); text-decoration: underline; }
strong { color: var(--sea-ink, #2a2420); font-weight: 600; }
ul, ol { margin: 0 0 0.75em; padding-left: 1.5em; color: var(--sea-ink-soft, #5c534a); }
li { margin-bottom: 0.25em; }
blockquote {
	margin: 0 0 0.75em;
	padding: 0.5em 1em;
	border-left: 3px solid var(--lagoon, #c9883a);
	background: color-mix(in srgb, var(--lagoon, #c9883a) 5%, transparent);
	border-radius: 0 6px 6px 0;
	color: var(--sea-ink-soft, #5c534a);
}
table {
	width: 100%;
	border-collapse: collapse;
	margin-bottom: 0.75em;
	font-size: 0.875em;
}
th {
	background: var(--foam, #f8f4ee);
	padding: 0.5em 0.75em;
	text-align: left;
	font-weight: 600;
	color: var(--sea-ink, #2a2420);
	border: 1px solid var(--line, rgba(42,36,32,0.11));
}
td {
	padding: 0.5em 0.75em;
	border-top: 1px solid var(--line, rgba(42,36,32,0.11));
	color: var(--sea-ink-soft, #5c534a);
}
code {
	padding: 0.1em 0.35em;
	border-radius: 4px;
	font-size: 0.85em;
	font-family: ui-monospace, 'SF Mono', monospace;
	background: color-mix(in srgb, var(--lagoon, #c9883a) 10%, transparent);
	color: var(--lagoon-deep, #7a4e1a);
}
pre {
	margin: 0 0 0.75em;
	padding: 0.75em;
	border-radius: 8px;
	border: 1px solid var(--line, rgba(42,36,32,0.11));
	background: #fff;
	font-size: 0.8125em;
	font-family: ui-monospace, 'SF Mono', monospace;
	overflow-x: auto;
	color: #1e293b;
}
pre code {
	background: transparent;
	padding: 0;
	color: inherit;
	font-size: inherit;
}
hr {
	border: none;
	border-top: 1px solid var(--line, rgba(42,36,32,0.11));
	margin: 1em 0;
}
img { max-width: 100%; height: auto; border-radius: 8px; }
`;

/** DOMPurify 配置 —— 允许常见排版标签，剥离脚本和事件处理器 */
const PURIFY_CONFIG: PurifyConfig = {
	ALLOWED_TAGS: [
		'h1', 'h2', 'h3', 'h4', 'h5', 'h6',
		'p', 'br', 'hr',
		'strong', 'b', 'em', 'i', 'u', 's', 'del', 'ins', 'mark', 'sub', 'sup', 'small',
		'ul', 'ol', 'li', 'dl', 'dt', 'dd',
		'blockquote', 'q', 'cite',
		'code', 'pre', 'kbd', 'samp', 'var',
		'a', 'img',
		'table', 'thead', 'tbody', 'tfoot', 'tr', 'th', 'td', 'caption', 'colgroup', 'col',
		'div', 'span', 'p',
		'figure', 'figcaption',
		'abbr', 'address', 'time',
		'details', 'summary',
	],
	ALLOWED_ATTR: [
		'href', 'src', 'alt', 'title', 'width', 'height',
		'class', 'id', 'style',
		'target', 'rel',
		'colspan', 'rowspan',
		'datetime',
		'open',
	],
	ALLOW_DATA_ATTR: false,
	FORBID_TAGS: ['script', 'iframe', 'object', 'embed', 'form', 'input', 'textarea', 'button', 'select', 'style', 'link', 'meta'],
	FORBID_ATTR: ['onerror', 'onclick', 'onload', 'onmouseover', 'onfocus', 'onblur', 'onchange', 'onsubmit'],
};

export function ShadowHtml({ content, css, className }: ShadowHtmlProps) {
	const hostRef = useRef<HTMLDivElement>(null);
	const shadowRef = useRef<ShadowRoot | null>(null);

	// 初始化 Shadow Root（仅一次）
	useEffect(() => {
		if (!hostRef.current || shadowRef.current) return;
		const shadow = hostRef.current.attachShadow({ mode: 'open' });
		const styleEl = document.createElement('style');
		styleEl.textContent = BASE_STYLE + (css ?? '');
		shadow.appendChild(styleEl);
		const contentEl = document.createElement('div');
		shadow.appendChild(contentEl);
		shadowRef.current = shadow;
	}, [css]);

	// 内容变化时更新（DOMPurify 清洗后写入）
	useEffect(() => {
		if (!shadowRef.current) return;
		const contentEl = shadowRef.current.lastElementChild as HTMLDivElement | null;
		if (!contentEl) return;
		contentEl.innerHTML = DOMPurify.sanitize(content, PURIFY_CONFIG);
	}, [content]);

	return <div ref={hostRef} className={className} />;
}
