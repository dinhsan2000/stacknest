import { useState } from 'react'
import { Server, Database, Code2, Zap, Layers } from 'lucide-react'
import { ServiceInfo } from '../types'
import { useServiceStore } from '../store/serviceStore'
import { CheckPortConflict } from '../../wailsjs/go/main/App'
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
}

export default function ServiceRow({ service }: Props) {
  const { startService, stopService, restartService, setServiceEnabled, loading } = useServiceStore()
  const [conflict, setConflict] = useState<ConflictInfo | null>(null)
  const isLoading = loading[service.name]
  const isRunning = service.status === 'running'
  const isActive  = service.status === 'running' || service.status === 'starting'

  const handleStart = async () => {
    const info = await CheckPortConflict(service.port) as ConflictInfo
    if (info.in_use) { setConflict(info); return }
    startService(service.name)
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
        <div className="flex items-center gap-2.5 w-36 shrink-0">
          <ServiceIcon name={service.name} />
          <div>
            <div className="text-sm font-semibold text-white leading-tight">{service.display}</div>
            <div className="text-xs text-gray-600">v{service.version}</div>
          </div>
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
          <code className="text-xs text-blue-400 bg-[#0f1420] px-1.5 py-0.5 rounded font-mono">
            :{service.port}
          </code>
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
          title={service.enabled ? 'Enabled — click to exclude from Start All' : 'Disabled — click to include in Start All'}
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
                Stop
              </button>
              <button
                onClick={() => restartService(service.name)}
                disabled={isLoading}
                className="px-3 py-1.5 rounded-lg text-xs font-medium bg-[#1e2535] text-gray-400 hover:bg-[#2a3347] hover:text-white transition-colors disabled:opacity-40"
              >
                ↺
              </button>
            </>
          ) : (
            <button
              onClick={handleStart}
              disabled={isLoading || service.status === 'starting' || service.status === 'stopping'}
              className="px-3 py-1.5 rounded-lg text-xs font-medium bg-blue-500/15 text-blue-400 hover:bg-blue-500/25 transition-colors disabled:opacity-40"
            >
              {isLoading || service.status === 'starting' ? 'Starting…' : 'Start'}
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
