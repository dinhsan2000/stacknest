import { useState, useRef, useEffect } from 'react'
import { Server, Database, Code2, Zap, Layers, ChevronDown, RotateCw } from 'lucide-react'
import { ServiceInfo } from '../types'
import { useServiceStore } from '../store/serviceStore'
import { CheckPortConflict } from '../../wailsjs/go/main/App'
import { useI18n } from '../i18n'
import type { Page } from './Sidebar'
import type { NavContext } from '../App'
import PortConflictModal from './PortConflictModal'

const statusDot: Record<string, string> = {
  running:  'bg-green-500',
  stopped:  'bg-gray-500',
  starting: 'bg-yellow-400 animate-pulse',
  stopping: 'bg-orange-400 animate-pulse',
  error:    'bg-red-500',
}

const statusText: Record<string, string> = {
  running:  'text-green-400',
  stopped:  'text-gray-500',
  starting: 'text-yellow-400',
  stopping: 'text-orange-400',
  error:    'text-red-400',
}

const ServiceIcon = ({ name }: { name: string }) => {
  const sz = 16
  switch (name) {
    case 'apache': return <Server   size={sz} className="text-orange-400" />
    case 'nginx':  return <Zap     size={sz} className="text-green-400"  />
    case 'mysql':  return <Database size={sz} className="text-blue-400"  />
    case 'php':    return <Code2   size={sz} className="text-purple-400" />
    case 'redis':  return <Layers  size={sz} className="text-red-400"   />
    default:       return <Server   size={sz} className="text-gray-400"  />
  }
}

interface ConflictInfo {
  port: number
  pid: number
  process: string
  in_use: boolean
}

interface Props {
  service: ServiceInfo
  onNavigate?: (page: Page, ctx?: NavContext) => void
}

export default function ServiceRow({ service, onNavigate }: Props) {
  const { startService, stopService, restartService, setServiceEnabled, setActiveVersion, loading, binaryStatus } = useServiceStore()
  const { t } = useI18n()
  const [conflict, setConflict] = useState<ConflictInfo | null>(null)
  const [versionOpen, setVersionOpen] = useState(false)
  const [switching, setSwitching] = useState(false)
  const dropdownRef = useRef<HTMLDivElement>(null)
  const isLoading = loading[service.name]
  const isRunning = service.status === 'running'
  const isActive  = service.status === 'running' || service.status === 'starting'

  // Find installed versions for this service
  const svcVersions = binaryStatus.find(b => b.service === service.name)
  const installedVersions = svcVersions?.versions.filter(v => v.installed) ?? []
  const hasMultipleVersions = installedVersions.length > 1

  // Close dropdown on outside click
  useEffect(() => {
    if (!versionOpen) return
    const handler = (e: MouseEvent) => {
      if (dropdownRef.current && !dropdownRef.current.contains(e.target as Node)) {
        setVersionOpen(false)
      }
    }
    document.addEventListener('mousedown', handler)
    return () => document.removeEventListener('mousedown', handler)
  }, [versionOpen])

  const handleStart = async () => {
    const info = await CheckPortConflict(service.port) as ConflictInfo
    if (info.in_use) { setConflict(info); return }
    startService(service.name)
  }

  const handleSwitchVersion = async (version: string) => {
    setVersionOpen(false)
    setSwitching(true)
    try {
      await setActiveVersion(service.name, version)
    } finally {
      setSwitching(false)
    }
  }

  return (
    <>
      <div className={`flex items-center gap-4 px-4 py-3 rounded-xl border transition-colors
        ${service.status === 'error'
          ? 'border-red-500/30 bg-red-500/5'
          : isActive
            ? 'border-green-500/20 bg-[#1a2035]'
            : 'border-[#2a3347] bg-[#161b27] hover:border-[#3a4557]'
        }`}
      >
        {/* Service name */}
        <div className="flex items-center gap-2.5 w-28 shrink-0">
          <ServiceIcon name={service.name} />
          <span className="text-sm font-semibold text-white">{service.display}</span>
        </div>

        {/* Version */}
        <div className="w-28 shrink-0 relative" ref={dropdownRef}>
          {hasMultipleVersions ? (
            <>
              <button
                onClick={() => setVersionOpen(!versionOpen)}
                disabled={switching}
                title={t.dash_switch_version}
                className={`inline-flex items-center gap-1 px-2.5 py-1 rounded-lg text-xs font-mono transition-colors ${
                  switching
                    ? 'bg-yellow-500/10 text-yellow-400'
                    : 'bg-[#0f1420] text-blue-400 hover:bg-blue-500/15 hover:text-blue-300'
                } disabled:opacity-50`}
              >
                {switching ? t.dash_switching : `v${service.version}`}
                <ChevronDown size={12} className={`transition-transform ${versionOpen ? 'rotate-180' : ''}`} />
              </button>
              {versionOpen && (
                <div className="absolute top-full left-0 mt-1 z-50 bg-[#1e2535] border border-[#2a3347] rounded-lg shadow-xl py-1 min-w-[130px]">
                  {installedVersions.map(v => (
                    <button
                      key={v.version}
                      onClick={() => !v.active && handleSwitchVersion(v.version)}
                      className={`w-full text-left px-3 py-1.5 text-xs font-mono transition-colors flex items-center justify-between ${
                        v.active
                          ? 'text-green-400 bg-green-500/10 cursor-default'
                          : 'text-gray-300 hover:bg-[#2a3347] hover:text-white'
                      }`}
                    >
                      v{v.version}
                      {v.active && <span className="text-[10px] text-green-500">●</span>}
                    </button>
                  ))}
                </div>
              )}
            </>
          ) : (
            <span className="inline-flex px-2.5 py-1 rounded-lg text-xs font-mono bg-[#0f1420] text-gray-500">
              v{service.version}
            </span>
          )}
        </div>

        {/* Status */}
        <div className="flex items-center gap-1.5 w-24 shrink-0">
          <span className={`w-1.5 h-1.5 rounded-full shrink-0 ${statusDot[service.status]}`} />
          <span className={`text-xs capitalize ${statusText[service.status]}`}>
            {service.status}
          </span>
        </div>

        {/* Port */}
        <div className="w-16 shrink-0">
          <button
            onClick={() => onNavigate?.('settings', { highlightPort: service.name })}
            title={t.dash_edit_port}
            className="text-xs text-blue-400 bg-[#0f1420] px-1.5 py-0.5 rounded font-mono hover:bg-blue-500/15 hover:text-blue-300 transition-colors cursor-pointer"
          >
            :{service.port}
          </button>
        </div>

        {/* PID */}
        <div className="w-24 shrink-0 text-xs">
          {service.pid > 0
            ? <code className="text-gray-500 bg-[#0f1420] px-1.5 py-0.5 rounded font-mono">PID {service.pid}</code>
            : <span className="text-gray-700">—</span>
          }
        </div>

        {/* Error message — chiếm phần còn lại */}
        <div className="flex-1 min-w-0">
          {service.error && (
            <p className="text-xs text-red-400 truncate" title={service.error}>
              {service.error}
            </p>
          )}
        </div>

        {/* Enabled toggle */}
        <button
          onClick={() => setServiceEnabled(service.name, !service.enabled)}
          title={service.enabled ? t.dash_enabled_tip : t.dash_disabled_tip}
          className={`relative w-9 h-5 rounded-full transition-colors duration-200 shrink-0 focus:outline-none ${
            service.enabled ? 'bg-blue-500' : 'bg-gray-600'
          }`}
        >
          <span className={`absolute top-0.5 left-0.5 w-4 h-4 rounded-full bg-white shadow-sm transition-transform duration-200 ${
            service.enabled ? 'translate-x-4' : 'translate-x-0'
          }`} />
        </button>

        {/* Actions */}
        <div className="flex items-center gap-2 shrink-0">
          {isRunning ? (
            <>
              <button
                onClick={() => stopService(service.name)}
                disabled={isLoading}
                className="px-3 py-1.5 rounded-lg text-xs font-medium bg-red-500/15 text-red-400 hover:bg-red-500/25 transition-colors disabled:opacity-40"
              >
                {t.stop}
              </button>
              <button
                onClick={() => restartService(service.name)}
                disabled={isLoading}
                className="px-3 py-1.5 rounded-lg text-xs font-medium bg-[#1e2535] text-gray-400 hover:bg-[#2a3347] hover:text-white transition-colors disabled:opacity-40"
              >
                <RotateCw size={14} />
              </button>
            </>
          ) : (
            <button
              onClick={handleStart}
              disabled={isLoading || service.status === 'starting' || service.status === 'stopping'}
              className="px-3 py-1.5 rounded-lg text-xs font-medium bg-blue-500/15 text-blue-400 hover:bg-blue-500/25 transition-colors disabled:opacity-40"
            >
              {isLoading || service.status === 'starting' ? t.dash_starting : t.start}
            </button>
          )}
        </div>
      </div>

      {conflict && (
        <PortConflictModal
          conflict={conflict}
          serviceName={service.name}
          onClose={() => setConflict(null)}
          onResolved={() => { setConflict(null); startService(service.name) }}
        />
      )}
    </>
  )
}
