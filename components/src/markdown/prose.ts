/**
 * Prose class 常量 —— 收敛 Markdown 排版样式。
 *
 * 三个语义层级：
 * - proseQuestion：Q&A 问题主文案（紧凑 14px）
 * - proseHint：次级/补充文案（更紧凑 12px）
 * - proseArticle：Wiki 全文阅读型排版（标准 16px，含 h1-h6 完整层级）
 *
 * proseArticle 采用 GitHub-style 标题层级方案：
 * - h1：横线分隔 + 衬线标题字体（页面唯一主标题）
 * - h2：semibold 无横线（章节）
 * - h3-h6：渐进缩放，h4 等同正文字号靠粗体区分，h6 降级 muted 色
 */

const baseCode =
	'[&_code]:rounded [&_code]:bg-lagoon/10 [&_code]:px-1.5 [&_code]:py-0.5 [&_code]:text-xs [&_code]:font-mono [&_code]:text-lagoon-deep'
const basePre =
	'[&_pre]:rounded-lg [&_pre]:border [&_pre]:border-line [&_pre]:bg-white [&_pre]:p-3 [&_pre]:text-xs [&_pre]:leading-relaxed [&_pre]:font-mono [&_pre]:text-gray-800 [&_pre_code]:bg-transparent [&_pre_code]:p-0 [&_pre_code]:text-inherit'
const baseLink = '[&_a]:text-lagoon-deep [&_a]:underline'

/** 问题主文案 —— 正文 14px，墨褐 */
export const proseQuestion = `prose prose-sm max-w-none ${baseCode} ${basePre} ${baseLink} [&_p]:mb-1 [&_p]:text-sm [&_p]:leading-relaxed [&_p]:text-sea-ink`

/** 次级/补充文案 —— 正文 12px，柔褐（description、提示语） */
export const proseHint = `prose prose-sm max-w-none ${baseCode} ${basePre} ${baseLink} [&_p]:mt-1 [&_p]:mb-2 [&_p]:text-xs [&_p]:leading-relaxed [&_p]:text-sea-ink-soft`

/**
 * Wiki 全文阅读型排版（prose base = 16px）。
 *
 * 标题层级参考 GitHub markdown-body：
 * - 全部 heading 用 font-semibold(600) + line-height 1.25
 * - h1 底部横线分隔，h2-h6 无横线
 * - h1/h2 用衬线 display-title，h3-h6 用无衬线
 * - h4 等同正文字号（16px），纯靠粗体区分
 * - h6 降级为 muted 色，表示层级末端
 * - 代码块 14px，与正文匹配
 */
export const proseArticle = `prose prose-slate max-w-none
${baseLink}
[&_h1]:display-title [&_h1]:text-2xl [&_h1]:font-bold [&_h1]:text-sea-ink [&_h1]:mt-0 [&_h1]:mb-4 [&_h1]:pb-2 [&_h1]:scroll-mt-20 [&_h1]:border-b [&_h1]:border-line
[&_h2]:display-title [&_h2]:text-xl [&_h2]:font-semibold [&_h2]:text-sea-ink [&_h2]:mt-8 [&_h2]:mb-3 [&_h2]:scroll-mt-20
[&_h3]:text-lg [&_h3]:font-semibold [&_h3]:text-sea-ink [&_h3]:mt-6 [&_h3]:mb-2 [&_h3]:scroll-mt-20
[&_h4]:text-base [&_h4]:font-semibold [&_h4]:text-sea-ink [&_h4]:mt-4 [&_h4]:mb-2 [&_h4]:scroll-mt-20
[&_h5]:text-sm [&_h5]:font-semibold [&_h5]:text-sea-ink [&_h5]:mt-4 [&_h5]:mb-2 [&_h5]:scroll-mt-20
[&_h6]:text-sm [&_h6]:font-semibold [&_h6]:text-sea-ink-soft [&_h6]:mt-4 [&_h6]:mb-2 [&_h6]:scroll-mt-20
[&_p]:text-base [&_p]:leading-relaxed [&_p]:text-sea-ink
[&_strong]:text-sea-ink [&_strong]:font-semibold
[&_blockquote]:border-l-lagoon [&_blockquote]:bg-lagoon/5 [&_blockquote]:rounded-r-lg [&_blockquote]:px-4 [&_blockquote]:py-2 [&_blockquote]:text-base [&_blockquote]:text-sea-ink-soft
[&_code]:rounded [&_code]:bg-lagoon/10 [&_code]:px-1.5 [&_code]:py-0.5 [&_code]:text-sm [&_code]:font-mono [&_code]:text-lagoon-deep
[&_pre]:rounded-lg [&_pre]:border [&_pre]:border-line [&_pre]:bg-white [&_pre]:p-4 [&_pre]:text-sm [&_pre]:leading-relaxed [&_pre]:font-mono [&_pre]:text-gray-800 [&_pre_code]:bg-transparent [&_pre_code]:p-0 [&_pre_code]:text-inherit
[&_table]:w-full [&_table]:text-sm
[&_th]:bg-foam [&_th]:px-3 [&_th]:py-2 [&_th]:text-left [&_th]:text-sea-ink [&_th]:font-semibold
[&_td]:px-3 [&_td]:py-2 [&_td]:text-sea-ink-soft [&_td]:border-t [&_td]:border-line
[&_hr]:border-line
[&_ul]:text-base [&_ul]:text-sea-ink [&_ol]:text-base [&_ol]:text-sea-ink [&_li]:leading-relaxed [&_li]:text-sea-ink
[&_em]:text-sea-ink-soft`
