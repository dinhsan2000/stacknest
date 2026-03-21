import { useEffect, useRef, useState } from 'react'
import { Terminal as XTerm } from '@xterm/xterm'
import { FitAddon } from '@xterm/addon-fit'
import { WebLinksAddon } from '@xterm/addon-web-links'
import '@xterm/xterm/css/xterm.css'
import {
  TerminalStart,
  TerminalWrite,
  TerminalResize,
  TerminalClose,
  GetWWWPath,
} from '../../wailsjs/go/main/App'
import { EventsOn, EventsOff } from '../../wailsjs/runtime/runtime'
import { useI18n } from '../i18n'
import { FolderOpen, X, Play, Terminal as TerminalIcon } from 'lucide-react'

export default function Terminal() {
  const { t } = useI18n()
  const containerRef = useRef<HTMLDivElement>(null)
  const xtermRef    = useRef<XTerm | null>(null)
  const fitAddon    = useRef<FitAddon | null>(null)
  const [status, setStatus]   = useState<'idle' | 'running' | 'exited'>('idle')
  const [error, setError]     = useState('')
  const [wwwPath, setWWWPath] = useState('')

  // Lấy www path để dùng làm cwd mặc định
  useEffect(() => {
    GetWWWPath().then(setWWWPath)
  }, [])

  // Khởi tạo xterm.js
  useEffect(() => {
    if (!containerRef.current) return

    const term = new XTerm({
      theme: {
        background:  '#0a0f1a',
        foreground:  '#e2e8f0',
        cursor:      '#60a5fa',
        cursorAccent:'#0a0f1a',
        black:       '#1e2535',
        brightBlack: '#4a5568',
        blue:        '#60a5fa',
        brightBlue:  '#93c5fd',
        green:       '#34d399',
        brightGreen: '#6ee7b7',
        red:         '#f87171',
        brightRed:   '#fca5a5',
        yellow:      '#fbbf24',
        brightYellow:'#fde68a',
        cyan:        '#22d3ee',
        brightCyan:  '#67e8f9',
        white:       '#e2e8f0',
        brightWhite: '#f8fafc',
      },
      fontFamily: '"Cascadia Code", "JetBrains Mono", "Fira Code", Consolas, monospace',
      fontSize:   14,
      lineHeight: 1.4,
      cursorBlink: true,
      cursorStyle: 'block',
      scrollback:  5000,
      convertEol:  true,
    })

    const fit  = new FitAddon()
    const links = new WebLinksAddon()
    term.loadAddon(fit)
    term.loadAddon(links)
    term.open(containerRef.current)
    fit.fit()

    xtermRef.current  = term
    fitAddon.current  = fit

    // Gửi input từ xterm → Go PTY
    term.onData((data) => {
      TerminalWrite(data).catch(() => {})
    })

    // Resize khi window thay đổi
    const handleResize = () => {
      fit.fit()
      const { rows, cols } = term
      TerminalResize(rows, cols).catch(() => {})
    }
    window.addEventListener('resize', handleResize)

    return () => {
      window.removeEventListener('resize', handleResize)
      term.dispose()
      xtermRef.current = null
    }
  }, [])

  // Lắng nghe output từ Go PTY → xterm
  useEffect(() => {
    const handleOutput = (data: string) => {
      xtermRef.current?.write(data)
    }
    const handleExit = () => {
      xtermRef.current?.write('\r\n\x1b[33m[Process exited]\x1b[0m\r\n')
      setStatus('exited')
    }

    EventsOn('term:output', handleOutput as any)
    EventsOn('term:exit',   handleExit as any)

    return () => {
      EventsOff('term:output')
      EventsOff('term:exit')
    }
  }, [])

  const startTerminal = async (cwd = '') => {
    setError('')
    xtermRef.current?.clear()
    try {
      await TerminalStart(cwd)
      setStatus('running')
      // Sync size ngay sau khi start
      setTimeout(() => {
        fitAddon.current?.fit()
        const term = xtermRef.current
        if (term) TerminalResize(term.rows, term.cols).catch(() => {})
      }, 100)
    } catch (e: any) {
      setError(e?.toString() ?? 'Failed to start terminal')
    }
  }

  const stopTerminal = () => {
    TerminalClose()
    setStatus('exited')
    xtermRef.current?.write('\r\n\x1b[33m[Session closed]\x1b[0m\r\n')
  }

  return (
    <div className="flex flex-col h-full gap-4">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-2xl font-bold text-white">{t.term_title}</h2>
          <p className="text-gray-400 text-sm mt-1">
            {status === 'running'  && <span className="text-green-400"><span className="inline-block w-2 h-2 rounded-full bg-green-400 mr-1.5" />{t.term_running}</span>}
            {status === 'exited'   && <span className="text-yellow-400"><span className="inline-block w-2 h-2 rounded-full bg-yellow-400 mr-1.5" />{t.term_exited}</span>}
            {status === 'idle'     && <span className="text-gray-500"><span className="inline-block w-2 h-2 rounded-full bg-gray-500 mr-1.5" />{t.term_not_started}</span>}
          </p>
        </div>

        <div className="flex gap-2">
          {/* Quick launch buttons */}
          <button
            onClick={() => startTerminal(wwwPath)}
            className="px-3 py-1.5 text-xs rounded-lg bg-blue-500/20 text-blue-400 hover:bg-blue-500/30 transition-colors flex items-center gap-1"
            title={t.term_open_www}
          >
            <FolderOpen size={14} /> {t.term_open_www}
          </button>

          {status === 'running' ? (
            <button
              onClick={stopTerminal}
              className="px-3 py-1.5 text-xs rounded-lg bg-red-500/20 text-red-400 hover:bg-red-500/30 transition-colors flex items-center gap-1"
            >
              <X size={14} /> {t.stop}
            </button>
          ) : (
            <button
              onClick={() => startTerminal('')}
              className="px-3 py-1.5 text-xs rounded-lg bg-green-500/20 text-green-400 hover:bg-green-500/30 transition-colors flex items-center gap-1"
            >
              <Play size={14} /> {t.term_start}
            </button>
          )}
        </div>
      </div>

      {error && (
        <p className="text-red-400 text-xs bg-red-500/10 rounded p-3">{error}</p>
      )}

      {/* Terminal container */}
      <div className="flex-1 bg-[#0a0f1a] border border-[#1e2535] rounded-xl overflow-hidden min-h-0 relative">
        {/* Placeholder khi chưa start */}
        {status === 'idle' && (
          <div className="absolute inset-0 flex flex-col items-center justify-center gap-4 z-10">
            <TerminalIcon size={40} className="text-gray-600" />
            <p className="text-gray-400 text-sm">{t.term_click_start}</p>
            <div className="flex gap-3">
              <button
                onClick={() => startTerminal('')}
                className="px-5 py-2.5 rounded-lg bg-blue-500 text-white hover:bg-blue-600 text-sm font-medium transition-colors"
              >
                {t.term_start_btn}
              </button>
              <button
                onClick={() => startTerminal(wwwPath)}
                className="px-5 py-2.5 rounded-lg bg-[#1e2535] text-gray-300 hover:bg-[#2a3347] text-sm transition-colors"
              >
                {t.term_open_www_btn}
              </button>
            </div>
          </div>
        )}

        {/* xterm.js mount point */}
        <div
          ref={containerRef}
          className="w-full h-full p-2"
          style={{ opacity: status === 'idle' ? 0 : 1 }}
        />
      </div>
    </div>
  )
}
