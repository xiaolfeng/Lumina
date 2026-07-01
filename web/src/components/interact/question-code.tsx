import { useEffect, useMemo, useState } from 'react'
import CodeMirror from '@uiw/react-codemirror'
import { EditorView } from '@codemirror/view'
import { githubLight } from '@uiw/codemirror-theme-github'
import { Code2 } from 'lucide-react'

import { Label } from '#/components/ui/label'

import { QuestionShell } from './question-shell'
import type { QuestionComponentProps } from './question-shell'

/**
 * 语言包懒加载映射。
 *
 * CodeMirror 的 15 个语言包合计 ~1MB，
 * 全量静态导入会把 question-code chunk 撑到 1046KB。
 * 改为按需动态 import：仅当用户实际使用某种语言时才下载对应包。
 */
const LANG_LOADERS: Record<string, (() => Promise<any>) | undefined> = {
  javascript: () => import('@codemirror/lang-javascript').then((m) => m.javascript()),
  js: () => import('@codemirror/lang-javascript').then((m) => m.javascript()),
  jsx: () => import('@codemirror/lang-javascript').then((m) => m.javascript({ jsx: true })),
  typescript: () => import('@codemirror/lang-javascript').then((m) => m.javascript({ typescript: true })),
  ts: () => import('@codemirror/lang-javascript').then((m) => m.javascript({ typescript: true })),
  tsx: () => import('@codemirror/lang-javascript').then((m) => m.javascript({ jsx: true, typescript: true })),
  python: () => import('@codemirror/lang-python').then((m) => m.python()),
  py: () => import('@codemirror/lang-python').then((m) => m.python()),
  json: () => import('@codemirror/lang-json').then((m) => m.json()),
  markdown: () => import('@codemirror/lang-markdown').then((m) => m.markdown()),
  md: () => import('@codemirror/lang-markdown').then((m) => m.markdown()),
  css: () => import('@codemirror/lang-css').then((m) => m.css()),
  scss: () => import('@codemirror/lang-css').then((m) => m.css()),
  html: () => import('@codemirror/lang-html').then((m) => m.html()),
  xml: () => import('@codemirror/lang-xml').then((m) => m.xml()),
  sql: () => import('@codemirror/lang-sql').then((m) => m.sql()),
  rust: () => import('@codemirror/lang-rust').then((m) => m.rust()),
  rs: () => import('@codemirror/lang-rust').then((m) => m.rust()),
  java: () => import('@codemirror/lang-java').then((m) => m.java()),
  cpp: () => import('@codemirror/lang-cpp').then((m) => m.cpp()),
  c: () => import('@codemirror/lang-cpp').then((m) => m.cpp()),
  'c++': () => import('@codemirror/lang-cpp').then((m) => m.cpp()),
  php: () => import('@codemirror/lang-php').then((m) => m.php()),
  yaml: () => import('@codemirror/lang-yaml').then((m) => m.yaml()),
  yml: () => import('@codemirror/lang-yaml').then((m) => m.yaml()),
  go: () => import('@codemirror/lang-go').then((m) => m.go()),
  golang: () => import('@codemirror/lang-go').then((m) => m.go()),
  regex: () => import('@codemirror/lang-javascript').then((m) => m.javascript()),
  shell: () => import('@codemirror/lang-javascript').then((m) => m.javascript()),
  bash: () => import('@codemirror/lang-javascript').then((m) => m.javascript()),
  sh: () => import('@codemirror/lang-javascript').then((m) => m.javascript()),
}

export function QuestionCode({
  question,
  onSubmit,
  onSkip,
  onRequestSupplement,
  isSupplementLoading = false,
}: QuestionComponentProps) {
  const language = (question.config?.language as string | undefined) ?? ''
  const placeholder =
    (question.config?.placeholder as string | undefined) || '输入代码...'

  const [code, setCode] = useState('')
  const [langExtension, setLangExtension] = useState<any | null>(null)

  const langKey = language.toLowerCase()

  useEffect(() => {
    let cancelled = false
    const loader = LANG_LOADERS[langKey]
    if (loader) {
      loader().then((ext) => {
        if (!cancelled) setLangExtension(ext)
      })
    } else {
      setLangExtension(null)
    }
    return () => {
      cancelled = true
    }
  }, [langKey])

  const extensions = useMemo(() => {
    const base = [EditorView.lineWrapping]
    if (langExtension) base.unshift(langExtension)
    return base
  }, [langExtension])

  const handleSubmit = () => {
    if (!code.trim()) return
    const answer: { code: string; language?: string } = { code }
    if (language) answer.language = language
    onSubmit(answer)
  }

  return (
    <QuestionShell
      question={question}
      isSupplementLoading={isSupplementLoading}
      onSkip={onSkip}
      onRequestSupplement={onRequestSupplement}
      submitDisabled={!code.trim()}
      onSubmit={handleSubmit}
    >
      {language && (
        <div className="flex items-center gap-2">
          <Code2 className="size-3.5 text-lagoon-deep" />
          <Label className="inline-flex items-center rounded-md bg-lagoon/10 px-2 py-0.5 font-mono text-xs font-semibold text-lagoon-deep">
            {language}
          </Label>
        </div>
      )}

      <div className="overflow-hidden rounded-lg border border-line">
        <CodeMirror
          value={code}
          height="260px"
          theme={githubLight}
          extensions={extensions}
          editable={!isSupplementLoading}
          placeholder={placeholder}
          onChange={(val) => setCode(val)}
          basicSetup={{
            lineNumbers: true,
            highlightActiveLineGutter: true,
            highlightSpecialChars: true,
            foldGutter: true,
            drawSelection: true,
            dropCursor: true,
            allowMultipleSelections: true,
            indentOnInput: true,
            bracketMatching: true,
            closeBrackets: true,
            autocompletion: true,
            rectangularSelection: true,
            crosshairCursor: true,
            highlightActiveLine: true,
            highlightSelectionMatches: true,
            tabSize: 2,
          }}
        />
      </div>
    </QuestionShell>
  )
}
