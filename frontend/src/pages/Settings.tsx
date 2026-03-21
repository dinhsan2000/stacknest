import { useEffect, useState, useRef } from 'react'
import { useServiceStore } from '../store/serviceStore'
import { SaveConfig } from '../../wailsjs/go/main/App'
import { AppConfig } from '../types'

interface Props {
  highlightPort?: string
}

export default function Settings({ highlightPort }: Props) {
  const { config, fetchConfig } = useServiceStore()
  const [form, setForm] = useState<AppConfig | null>(null)
  const [saved, setSaved] = useState(false)
  const [portErrors, setPortErrors] = useState<Record<string, string>>({})
  const [highlighted, setHighlighted] = useState<string | undefined>(highlightPort)
  const portRefs = useRef<Record<string, HTMLDivElement | null>>({})

  useEffect(() => { fetchConfig() }, [])
  useEffect(() => { if (config) setForm(config) }, [config])

  // Auto-scroll and highlight the target port input
  useEffect(() => {
    if (!highlightPort) return
    setHighlighted(highlightPort)
    // Wait for DOM to render
    const timer = setTimeout(() => {
      portRefs.current[highlightPort]?.scrollIntoView({ behavior: 'smooth', block: 'center' })
    }, 100)
    // Clear highlight after animation
    const clearTimer = setTimeout(() => setHighlighted(undefined), 2500)
    return () => { clearTimeout(timer); clearTimeout(clearTimer) }
  }, [highlightPort])

  const validatePort = (svc: string, value: number): string => {
    if (isNaN(value) || value < 1 || value > 65535) {
      return 'Port must be 1–65535'
    }
    // Check for duplicate ports
    if (form) {
      const allPorts: Record<string, number> = {
        apache: form.apache.port,
        nginx: form.nginx.port,
        mysql: form.mysql.port,
        php: form.php.port,
        redis: form.redis.port,
        [svc]: value, // override with current value
      }
      for (const [name, port] of Object.entries(allPorts)) {
        if (name !== svc && port === value) {
          return `Conflicts with ${name} port`
        }
      }
    }
    return ''
  }

  const handlePortChange = (svc: string, value: string) => {
    const port = parseInt(value) || 0
    setForm(f => f ? {
      ...f,
      [svc]: { ...f[svc as keyof AppConfig] as any, port }
    } : f)

    const err = validatePort(svc, port)
    setPortErrors(prev => {
      if (err) return { ...prev, [svc]: err }
      const { [svc]: _, ...rest } = prev
      return rest
    })
  }

  const hasPortErrors = Object.keys(portErrors).length > 0

  const handleSave = async () => {
    if (!form || hasPortErrors) return
    await SaveConfig(form as any)
    setSaved(true)
    setTimeout(() => setSaved(false), 2000)
  }

  if (!form) return <p className="text-gray-400">Loading...</p>

  return (
    <div className="flex flex-col gap-6 max-w-4xl">
      <h2 className="text-2xl font-bold text-white">Settings</h2>

      {/* Paths */}
      <section className="bg-[#1e2535] border border-[#2a3347] rounded-xl p-5 flex flex-col gap-4">
        <h3 className="text-white font-semibold">Paths</h3>

        {([
          { key: 'root_path', label: 'Root Path', hint: undefined },
          { key: 'data_path', label: 'Data Path', hint: 'MySQL and service data (e.g. databases)' },
          { key: 'www_path', label: 'WWW Path', hint: undefined },
          { key: 'log_path', label: 'Log Path', hint: undefined },
        ] as const).map(({ key, label, hint }) => (
          <div key={key} className="flex flex-col gap-1">
            <label className="text-xs text-gray-400">{label}</label>
            <input
              value={form[key]}
              onChange={e => setForm(f => f ? { ...f, [key]: e.target.value } : f)}
              className="bg-[#0f1420] border border-[#2a3347] rounded-lg px-3 py-2 text-sm text-white focus:outline-none focus:border-blue-500"
            />
            {hint && <p className="text-xs text-gray-500">{hint}</p>}
          </div>
        ))}
      </section>

      {/* Service Ports */}
      <section className="bg-[#1e2535] border border-[#2a3347] rounded-xl p-5 flex flex-col gap-4">
        <h3 className="text-white font-semibold">Service Ports</h3>
        <div className="grid grid-cols-2 gap-3">
          {(['apache', 'nginx', 'mysql', 'php', 'redis'] as const).map(svc => (
            <div
              key={svc}
              ref={el => { portRefs.current[svc] = el }}
              className={`flex flex-col gap-1 rounded-lg p-2 -m-2 transition-colors duration-700 ${
                highlighted === svc ? 'bg-blue-500/10 ring-1 ring-blue-500/40' : ''
              }`}
            >
              <label className="text-xs text-gray-400 capitalize">{svc} port</label>
              <input
                type="number"
                min={1}
                max={65535}
                value={form[svc].port}
                onChange={e => handlePortChange(svc, e.target.value)}
                autoFocus={highlighted === svc}
                className={`bg-[#0f1420] border rounded-lg px-3 py-2 text-sm text-white focus:outline-none ${portErrors[svc]
                    ? 'border-red-500 focus:border-red-500'
                    : highlighted === svc
                      ? 'border-blue-500'
                      : 'border-[#2a3347] focus:border-blue-500'
                  }`}
              />
              {portErrors[svc] && (
                <p className="text-xs text-red-400">{portErrors[svc]}</p>
              )}
            </div>
          ))}
        </div>
      </section>

      {/* General */}
      <section className="bg-[#1e2535] border border-[#2a3347] rounded-xl p-5 flex flex-col gap-4">
        <h3 className="text-white font-semibold">General</h3>
        <label className="flex items-center gap-3 cursor-pointer">
          <input
            type="checkbox"
            checked={form.auto_start}
            onChange={e => setForm(f => f ? { ...f, auto_start: e.target.checked } : f)}
            className="accent-blue-500 w-4 h-4"
          />
          <span className="text-sm text-gray-300">Auto-start services on launch</span>
        </label>
      </section>

      <button
        onClick={handleSave}
        disabled={hasPortErrors}
        className={`self-start px-6 py-2.5 rounded-lg font-medium transition-colors ${hasPortErrors
            ? 'bg-gray-600 text-gray-400 cursor-not-allowed'
            : 'bg-blue-500 text-white hover:bg-blue-600'
          }`}
      >
        {saved ? '✓ Saved!' : hasPortErrors ? 'Fix errors first' : 'Save Settings'}
      </button>
    </div>
  )
}
