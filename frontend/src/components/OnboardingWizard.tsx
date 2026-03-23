import { useEffect, useState } from 'react'
import { useI18n, localeLabels } from '../i18n'
import type { Locale } from '../i18n'
import { useServiceStore } from '../store/serviceStore'
import {
  GetConfig,
  SaveConfig,
  GetVersionCatalog,
  SelectFolder,
  CompleteOnboarding,
} from '../../wailsjs/go/main/App'
import { ServiceCatalog, AppConfig } from '../types'
import {
  Globe,
  FolderOpen,
  Download,
  Check,
  ChevronRight,
  ChevronLeft,
  Server,
  Database,
  Code2,
  Zap,
  Layers,
} from 'lucide-react'

interface Props {
  onComplete: () => void
}

const SERVICE_META: Record<string, { label: string; icon: React.ReactNode; defaultOn: boolean }> = {
  apache: { label: 'Apache', icon: <Server size={20} />, defaultOn: true },
  nginx:  { label: 'Nginx',  icon: <Zap size={20} />,    defaultOn: false },
  mysql:  { label: 'MySQL',  icon: <Database size={20} />, defaultOn: true },
  php:    { label: 'PHP',    icon: <Code2 size={20} />,   defaultOn: true },
  redis:  { label: 'Redis',  icon: <Layers size={20} />,  defaultOn: false },
}

export default function OnboardingWizard({ onComplete }: Props) {
  const { t, locale, setLocale } = useI18n()
  const { downloadBinary, downloadProgress } = useServiceStore()

  const [step, setStep] = useState(1)
  const [rootPath, setRootPath] = useState('')
  const [catalog, setCatalog] = useState<Record<string, ServiceCatalog>>({})
  const [selected, setSelected] = useState<Record<string, string>>({
    apache: '', mysql: '', php: '',
  }) // service -> version
  const [pendingDownloads, setPendingDownloads] = useState<string[]>([]) // ["service@version"]
  const [allDone, setAllDone] = useState(false)

  // Load config and catalog on mount
  useEffect(() => {
    GetConfig().then(cfg => {
      const c = cfg as unknown as AppConfig
      setRootPath(c.root_path)
    })
    GetVersionCatalog().then(cat => {
      const c = cat as Record<string, ServiceCatalog>
      setCatalog(c)
      // Pre-select first version for default services
      const sel: Record<string, string> = {}
      for (const [svc, meta] of Object.entries(SERVICE_META)) {
        if (meta.defaultOn && c[svc]?.versions?.length) {
          sel[svc] = c[svc].versions[0].version
        }
      }
      setSelected(sel)
    })
  }, [])

  // Watch download progress — when all pending are done, advance
  useEffect(() => {
    if (pendingDownloads.length === 0) return
    const allFinished = pendingDownloads.every(key => downloadProgress[key] === undefined)
    if (allFinished && step === 4) {
      setAllDone(true)
    }
  }, [downloadProgress, pendingDownloads, step])

  const handleBrowse = async () => {
    const path = await SelectFolder()
    if (path) setRootPath(path)
  }

  const toggleService = (svc: string) => {
    setSelected(prev => {
      if (prev[svc] !== undefined) {
        const { [svc]: _, ...rest } = prev
        return rest
      }
      const ver = catalog[svc]?.versions?.[0]?.version ?? ''
      return { ...prev, [svc]: ver }
    })
  }

  const handleStartDownloads = async () => {
    // Save config with potentially updated root path
    const cfg = await GetConfig() as unknown as AppConfig
    if (rootPath !== cfg.root_path) {
      cfg.root_path = rootPath
      cfg.bin_path = rootPath + '/bin'
      cfg.data_path = rootPath + '/data'
      cfg.www_path = rootPath + '/www'
      cfg.log_path = rootPath + '/logs'
      await SaveConfig(cfg as any)
    }

    // Start downloads
    const keys: string[] = []
    for (const [svc, version] of Object.entries(selected)) {
      if (version) {
        downloadBinary(svc, version)
        keys.push(`${svc}@${version}`)
      }
    }
    setPendingDownloads(keys)
    if (keys.length === 0) {
      setAllDone(true)
    }
    setStep(4)
  }

  const handleFinish = async () => {
    await CompleteOnboarding()
    onComplete()
  }

  const handleSkip = async () => {
    await CompleteOnboarding()
    onComplete()
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm">
      <div className="bg-[#1e2535] border border-[#2a3347] rounded-2xl shadow-2xl w-full max-w-lg p-8 flex flex-col gap-6">
        {/* Step indicator */}
        <div className="flex items-center justify-center gap-2">
          {[1, 2, 3, 4, 5].map(s => (
            <div
              key={s}
              className={`w-2 h-2 rounded-full transition-colors ${
                s === step ? 'bg-blue-400' : s < step ? 'bg-blue-400/40' : 'bg-gray-600'
              }`}
            />
          ))}
        </div>

        {/* Step 1: Welcome + Language */}
        {step === 1 && (
          <>
            <div className="text-center">
              <h2 className="text-2xl font-bold text-white">{t.onb_welcome}</h2>
              <p className="text-gray-400 text-sm mt-2">{t.onb_welcome_desc}</p>
            </div>

            <div>
              <p className="text-xs text-gray-500 mb-2">{t.onb_select_language}</p>
              <div className="flex gap-2">
                {(['en', 'vi'] as Locale[]).map(loc => (
                  <button
                    key={loc}
                    onClick={() => setLocale(loc)}
                    className={`flex-1 py-2.5 rounded-lg text-sm font-medium transition-colors flex items-center justify-center gap-2 ${
                      locale === loc
                        ? 'bg-blue-500/20 text-blue-400 border border-blue-500/30'
                        : 'bg-[#0f1420] text-gray-400 border border-[#2a3347] hover:border-[#3a4357]'
                    }`}
                  >
                    <Globe size={14} />
                    {localeLabels[loc]}
                  </button>
                ))}
              </div>
            </div>
          </>
        )}

        {/* Step 2: Root Path */}
        {step === 2 && (
          <>
            <div className="text-center">
              <h2 className="text-xl font-bold text-white">{t.onb_root_path}</h2>
              <p className="text-gray-400 text-sm mt-2">{t.onb_root_path_desc}</p>
            </div>

            <div className="flex gap-2">
              <input
                value={rootPath}
                onChange={e => setRootPath(e.target.value)}
                className="flex-1 bg-[#0f1420] border border-[#2a3347] rounded-lg px-3 py-2.5 text-sm text-white font-mono focus:outline-none focus:border-blue-500"
              />
              <button
                onClick={handleBrowse}
                className="px-3 py-2.5 rounded-lg bg-[#0f1420] border border-[#2a3347] text-gray-400 hover:text-white transition-colors"
              >
                <FolderOpen size={16} />
              </button>
            </div>
          </>
        )}

        {/* Step 3: Select Services */}
        {step === 3 && (
          <>
            <div className="text-center">
              <h2 className="text-xl font-bold text-white">{t.onb_select_services}</h2>
              <p className="text-gray-400 text-sm mt-2">{t.onb_select_services_desc}</p>
            </div>

            <div className="flex flex-col gap-2">
              {Object.entries(SERVICE_META).map(([svc, meta]) => {
                const versions = catalog[svc]?.versions ?? []
                const isSelected = selected[svc] !== undefined
                return (
                  <div
                    key={svc}
                    className={`flex items-center gap-3 p-3 rounded-lg border transition-colors cursor-pointer ${
                      isSelected
                        ? 'border-blue-500/30 bg-blue-500/10'
                        : 'border-[#2a3347] bg-[#0f1420] hover:border-[#3a4357]'
                    }`}
                    onClick={() => toggleService(svc)}
                  >
                    <div className={`flex-shrink-0 ${isSelected ? 'text-blue-400' : 'text-gray-500'}`}>
                      {meta.icon}
                    </div>
                    <span className={`text-sm font-medium flex-1 ${isSelected ? 'text-white' : 'text-gray-400'}`}>
                      {meta.label}
                    </span>
                    {isSelected && versions.length > 0 && (
                      <select
                        value={selected[svc]}
                        onClick={e => e.stopPropagation()}
                        onChange={e => setSelected(prev => ({ ...prev, [svc]: e.target.value }))}
                        className="bg-[#1e2535] border border-[#2a3347] rounded px-2 py-1 text-xs text-gray-300 focus:outline-none"
                      >
                        {versions.map(v => (
                          <option key={v.version} value={v.version}>v{v.version}</option>
                        ))}
                      </select>
                    )}
                    <div className={`w-5 h-5 rounded border flex items-center justify-center flex-shrink-0 ${
                      isSelected ? 'bg-blue-500 border-blue-500' : 'border-gray-600'
                    }`}>
                      {isSelected && <Check size={12} className="text-white" />}
                    </div>
                  </div>
                )
              })}
            </div>
          </>
        )}

        {/* Step 4: Downloading */}
        {step === 4 && (
          <>
            <div className="text-center">
              <h2 className="text-xl font-bold text-white">
                {allDone ? t.onb_download_complete : t.onb_downloading}
              </h2>
            </div>

            <div className="flex flex-col gap-3">
              {pendingDownloads.map(key => {
                const [svc, ver] = key.split('@')
                const progress = downloadProgress[key]
                const done = progress === undefined
                const meta = SERVICE_META[svc]
                return (
                  <div key={key} className="flex items-center gap-3 p-3 rounded-lg bg-[#0f1420]">
                    <div className={`flex-shrink-0 ${done ? 'text-green-400' : 'text-gray-400'}`}>
                      {done ? <Check size={20} /> : meta?.icon}
                    </div>
                    <div className="flex-1">
                      <div className="flex items-center justify-between mb-1">
                        <span className="text-sm text-white">{meta?.label} v{ver}</span>
                        <span className="text-xs text-gray-500">
                          {done ? '100%' : progress === 0 ? '...' : `${Math.round(progress ?? 0)}%`}
                        </span>
                      </div>
                      <div className="h-1 bg-[#1e2535] rounded-full overflow-hidden">
                        <div
                          className={`h-full rounded-full transition-all duration-200 ${done ? 'bg-green-500' : 'bg-blue-500'}`}
                          style={{ width: `${done ? 100 : progress ?? 0}%` }}
                        />
                      </div>
                    </div>
                  </div>
                )
              })}
              {pendingDownloads.length === 0 && (
                <p className="text-gray-500 text-center text-sm py-4">{t.onb_download_complete}</p>
              )}
            </div>
          </>
        )}

        {/* Step 5: Done */}
        {step === 5 && (
          <div className="text-center py-4">
            <div className="flex justify-center mb-4">
              <div className="w-16 h-16 rounded-full bg-green-500/20 flex items-center justify-center">
                <Check size={32} className="text-green-400" />
              </div>
            </div>
            <h2 className="text-xl font-bold text-white">{t.onb_done_title}</h2>
            <p className="text-gray-400 text-sm mt-2">{t.onb_done_desc}</p>
          </div>
        )}

        {/* Navigation buttons */}
        <div className="flex items-center justify-between pt-2">
          <button
            onClick={handleSkip}
            className="text-xs text-gray-500 hover:text-gray-300 transition-colors"
          >
            {t.onb_skip}
          </button>

          <div className="flex gap-2">
            {step > 1 && step < 4 && (
              <button
                onClick={() => setStep(s => s - 1)}
                className="px-4 py-2 rounded-lg text-sm bg-[#0f1420] text-gray-400 hover:text-white transition-colors flex items-center gap-1"
              >
                <ChevronLeft size={14} /> {t.onb_back}
              </button>
            )}

            {step < 3 && (
              <button
                onClick={() => setStep(s => s + 1)}
                disabled={step === 2 && !rootPath.trim()}
                className="px-4 py-2 rounded-lg text-sm bg-blue-500 text-white hover:bg-blue-600 transition-colors disabled:opacity-40 flex items-center gap-1"
              >
                {t.onb_next} <ChevronRight size={14} />
              </button>
            )}

            {step === 3 && (
              <button
                onClick={handleStartDownloads}
                className="px-4 py-2 rounded-lg text-sm bg-blue-500 text-white hover:bg-blue-600 transition-colors flex items-center gap-1"
              >
                <Download size={14} /> {t.onb_next}
              </button>
            )}

            {step === 4 && allDone && (
              <button
                onClick={() => setStep(5)}
                className="px-4 py-2 rounded-lg text-sm bg-blue-500 text-white hover:bg-blue-600 transition-colors flex items-center gap-1"
              >
                {t.onb_next} <ChevronRight size={14} />
              </button>
            )}

            {step === 5 && (
              <button
                onClick={handleFinish}
                className="px-5 py-2.5 rounded-lg text-sm bg-green-500 text-white font-medium hover:bg-green-600 transition-colors"
              >
                {t.onb_get_started}
              </button>
            )}
          </div>
        </div>
      </div>
    </div>
  )
}
