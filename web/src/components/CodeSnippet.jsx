import { useState } from 'react'

export function CodeSnippet({ code, language = 'cpp' }) {
  const [copied, setCopied] = useState(false)

  function copy() {
    navigator.clipboard.writeText(code)
    setCopied(true)
    setTimeout(() => setCopied(false), 2000)
  }

  return (
    <div className="relative rounded-xl overflow-hidden border border-zinc-800">
      <div className="flex items-center justify-between bg-[#161616] px-4 py-2 border-b border-zinc-800">
        <span className="text-xs text-zinc-500">{language}</span>
        <button
          onClick={copy}
          className="text-xs text-zinc-400 hover:text-white transition-colors"
        >
          {copied ? '✓ copied' : 'copy'}
        </button>
      </div>
      <pre className="bg-[#0d0d0d] p-4 overflow-x-auto text-sm mono text-zinc-300 leading-relaxed">
        <code>{code}</code>
      </pre>
    </div>
  )
}
