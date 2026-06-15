/**
 * Prose class 常量 —— 收敛 interact 模块内 30+ 处重复的 Markdown 排版样式。
 *
 * 三个语义层级：
 * - proseQuestion：问题主文案（标准字号，墨褐）
 * - proseHint：次级/补充文案（更小字号，柔褐）
 * - proseArticle：详情面板全文（含 h1/h2/blockquote/table 全套，阅读型）
 *
 * 通用基底（三者共享）：行内代码琥珀底、pre 代码块烛晕底、链接琥珀深色。
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

/** 详情面板全文 —— 阅读型排版，含标题层级、引用、表格 */
export const proseArticle = `prose prose-sm prose-slate max-w-none ${baseLink}
[&_h1]:display-title [&_h1]:text-2xl [&_h1]:font-bold [&_h1]:text-sea-ink [&_h1]:mb-4 [&_h1]:mt-0
[&_h2]:display-title [&_h2]:text-lg [&_h2]:font-bold [&_h2]:text-sea-ink [&_h2]:mt-6 [&_h2]:mb-2
[&_p]:text-sm [&_p]:leading-relaxed [&_p]:text-sea-ink-soft
[&_strong]:text-sea-ink [&_strong]:font-semibold
[&_blockquote]:border-l-lagoon [&_blockquote]:bg-lagoon/5 [&_blockquote]:rounded-r-lg [&_blockquote]:px-4 [&_blockquote]:py-2 [&_blockquote]:text-sm [&_blockquote]:text-sea-ink-soft
[&_table]:w-full [&_table]:text-sm
[&_th]:bg-foam [&_th]:px-3 [&_th]:py-2 [&_th]:text-left [&_th]:text-sea-ink [&_th]:font-semibold
[&_td]:px-3 [&_td]:py-2 [&_td]:text-sea-ink-soft [&_td]:border-t [&_td]:border-line
[&_hr]:border-line
[&_ul]:text-sm [&_ul]:text-sea-ink-soft [&_ol]:text-sm [&_ol]:text-sea-ink-soft [&_li]:leading-relaxed
[&_em]:text-sea-ink-soft`
