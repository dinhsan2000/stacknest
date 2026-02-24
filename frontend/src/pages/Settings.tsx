import { useEffect, useState } from 'react'
import { useServiceStore } from '../store/serviceStore'
import { SaveConfig } from '../../wailsjs/go/main/App'
import { AppConfig } from '../types'

export default function Settings() {
  const { config, fetchConfig } = useServiceStore()
  const [form, setForm] = useState<AppConfig | null>(null)
  const [saved, setSaved] = useState(false)

  useEffect(() => { fetchConfig() }, [])
  useEffect(() => { if (config) setForm(config) }, [config])

  const handleSave = async () => {
    if (!form) return
    await SaveConfig(form as any)
    setSaved(true)
    setTimeout(() => setSaved(false), 2000)
  }

  if (!form) return <p className="text-gray-400">Loading...</p>

  return (
    <div className="flex flex-col gap-6 max-w-2xl">
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
            <div key={svc} className="flex flex-col gap-1">
              <label className="text-xs text-gray-400 capitalize">{svc} port</label>
              <input
                type="number"
                value={form[svc].port}
                onChange={e => setForm(f => f ? {
                  ...f,
                  [svc]: { ...f[svc], port: parseInt(e.target.value) }
                } : f)}
                className="bg-[#0f1420] border border-[#2a3347] rounded-lg px-3 py-2 text-sm text-white focus:outline-none focus:border-blue-500"
              />
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
        className="self-start px-6 py-2.5 rounded-lg bg-blue-500 text-white hover:bg-blue-600 font-medium transition-colors"
      >
        {saved ? 'Saved!' : 'Save Settings'}
      </button>
    </div>
  )
}
