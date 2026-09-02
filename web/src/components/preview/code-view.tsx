import { useEffect, useMemo, useState } from 'react'
import type { Extension } from '@codemirror/state'
import CodeMirror from '@uiw/react-codemirror'
import { EditorView } from '@codemirror/view'
import { githubLight } from '@uiw/codemirror-theme-github'

import { previewLanguageFromFilename } from '#/lib/preview-file'

const LANG_LOADERS: Record<string, (() => Promise<Extension>) | undefined> = {
  js: () => import('@codemirror/lang-javascript').then((m) => m.javascript()),
  jsx: () =>
    import('@codemirror/lang-javascript').then((m) =>
      m.javascript({ jsx: true }),
    ),
  ts: () =>
    import('@codemirror/lang-javascript').then((m) =>
      m.javascript({ typescript: true }),
    ),
  tsx: () =>
    import('@codemirror/lang-javascript').then((m) =>
      m.javascript({ jsx: true, typescript: true }),
    ),
  css: () => import('@codemirror/lang-css').then((m) => m.css()),
  json: () => import('@codemirror/lang-json').then((m) => m.json()),
  html: () => import('@codemirror/lang-html').then((m) => m.html()),
  xml: () => import('@codemirror/lang-xml').then((m) => m.xml()),
  yaml: () => import('@codemirror/lang-yaml').then((m) => m.yaml()),
  markdown: () => import('@codemirror/lang-markdown').then((m) => m.markdown()),
  go: () => import('@codemirror/lang-go').then((m) => m.go()),
  python: () => import('@codemirror/lang-python').then((m) => m.python()),
  py: () => import('@codemirror/lang-python').then((m) => m.python()),
  rust: () => import('@codemirror/lang-rust').then((m) => m.rust()),
  rs: () => import('@codemirror/lang-rust').then((m) => m.rust()),
  java: () => import('@codemirror/lang-java').then((m) => m.java()),
  sql: () => import('@codemirror/lang-sql').then((m) => m.sql()),
}

export function PreviewCodeView({
  filename,
  source,
}: {
  filename: string
  source: string
}) {
  const lang = previewLanguageFromFilename(filename)
  const [langExtension, setLangExtension] = useState<Extension | null>(null)

  useEffect(() => {
    let cancelled = false
    const loader = LANG_LOADERS[lang]
    if (!loader) {
      setLangExtension(null)
      return
    }
    void loader().then((ext) => {
      if (!cancelled) setLangExtension(ext)
    })
    return () => {
      cancelled = true
    }
  }, [lang])

  const extensions = useMemo(() => {
    const base = [EditorView.lineWrapping, EditorView.editable.of(false)]
    if (langExtension) base.unshift(langExtension)
    return base
  }, [langExtension])

  return (
    <div className="min-h-0 flex-1 overflow-hidden">
      <CodeMirror
        value={source}
        height="100%"
        theme={githubLight}
        extensions={extensions}
        editable={false}
        basicSetup={{
          lineNumbers: true,
          highlightActiveLine: false,
          highlightActiveLineGutter: false,
          foldGutter: true,
          autocompletion: false,
          tabSize: 2,
        }}
        className="h-full text-sm"
      />
    </div>
  )
}
