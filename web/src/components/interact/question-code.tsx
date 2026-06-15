import { useMemo, useState } from 'react'
import CodeMirror from '@uiw/react-codemirror'
import { EditorView } from '@codemirror/view'
import { githubLight } from '@uiw/codemirror-theme-github'
import { javascript } from '@codemirror/lang-javascript'
import { python } from '@codemirror/lang-python'
import { json } from '@codemirror/lang-json'
import { markdown } from '@codemirror/lang-markdown'
import { css } from '@codemirror/lang-css'
import { html } from '@codemirror/lang-html'
import { sql } from '@codemirror/lang-sql'
import { rust } from '@codemirror/lang-rust'
import { java } from '@codemirror/lang-java'
import { cpp } from '@codemirror/lang-cpp'
import { php } from '@codemirror/lang-php'
import { xml } from '@codemirror/lang-xml'
import { yaml } from '@codemirror/lang-yaml'
import { go } from '@codemirror/lang-go'
import { Code2 } from 'lucide-react'

import { Label } from '#/components/ui/label'

import { QuestionShell } from './question-shell'
import type { QuestionComponentProps } from './question-shell'

const EXTENSIONS: Record<string, () => any> = {
  javascript: () => javascript(),
  js: () => javascript(),
  jsx: () => javascript({ jsx: true }),
  typescript: () => javascript({ typescript: true }),
  ts: () => javascript({ typescript: true }),
  tsx: () => javascript({ jsx: true, typescript: true }),
  python: () => python(),
  py: () => python(),
  json: () => json(),
  markdown: () => markdown(),
  md: () => markdown(),
  css: () => css(),
  scss: () => css(),
  html: () => html(),
  xml: () => xml(),
  sql: () => sql(),
  rust: () => rust(),
  rs: () => rust(),
  java: () => java(),
  cpp: () => cpp(),
  c: () => cpp(),
  'c++': () => cpp(),
  php: () => php(),
  yaml: () => yaml(),
  yml: () => yaml(),
  go: () => go(),
  golang: () => go(),
  regex: () => javascript(),
  shell: () => javascript(),
  bash: () => javascript(),
  sh: () => javascript(),
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

  const extensions = useMemo(() => {
    const key = language.toLowerCase()
    const factory = EXTENSIONS[key]
    return factory ? [factory(), EditorView.lineWrapping] : [EditorView.lineWrapping]
  }, [language])

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
