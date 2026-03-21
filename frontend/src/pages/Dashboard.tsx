import { useEffect } from 'react'
import { useServiceStore } from '../store/serviceStore'
import ServiceRow from '../components/ServiceCard'
import { useI18n, tt } from '../i18n'
import type { Page } from '../components/Sidebar'
import type { NavContext } from '../App'

interface Props {
  onNavigate: (page: Page, ctx?: NavContext) => void
}

export default function Dashboard({ onNavigate }: Props) {
  const { services, fetchServices, fetchBinaryStatus, startAll, stopAll } = useServiceStore()
  const { t } = useI18n()

  useEffect(() => {
    fetchServices()
    fetchBinaryStatus()
  }, [])

  const runningCount = services.filter(s => s.status === 'running').length
  const allRunning = runningCount === services.length && services.length > 0

  return (
    <div className="flex flex-col gap-6 max-w-4xl">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-2xl font-bold text-white">{t.dash_title}</h2>
          <p className="text-gray-500 text-sm mt-1">
            {tt(t.dash_running_count, { running: runningCount, total: services.length })}
          </p>
        </div>

        <div className="flex gap-2">
          <button
            onClick={stopAll}
            disabled={runningCount === 0}
            className="px-4 py-2 rounded-lg text-sm font-medium bg-[#1e2535] text-gray-400 hover:bg-[#2a3347] hover:text-white transition-colors disabled:opacity-30"
          >
            {t.dash_stop_all}
          </button>
          <button
            onClick={startAll}
            disabled={allRunning}
            className="px-4 py-2 rounded-lg text-sm font-medium bg-blue-500/20 text-blue-400 hover:bg-blue-500/30 transition-colors disabled:opacity-30"
          >
            {t.dash_start_all}
          </button>
        </div>
      </div>

      {/* Progress bar */}
      <div className="h-0.5 bg-[#1e2535] rounded-full overflow-hidden">
        <div
          className="h-full bg-green-500 rounded-full transition-all duration-700"
          style={{ width: `${services.length ? (runningCount / services.length) * 100 : 0}%` }}
        />
      </div>

      {/* Table header */}
      <div className="flex items-center gap-4 px-4 text-xs font-medium text-gray-400 uppercase tracking-wider">
        <span className="w-28 shrink-0">{t.dash_service}</span>
        <span className="w-28 shrink-0">{t.dash_version}</span>
        <span className="w-24 shrink-0">{t.dash_status}</span>
        <span className="w-16 shrink-0">{t.dash_port}</span>
        <span className="w-24 shrink-0">{t.dash_pid}</span>
        <span className="flex-1" />
        <span className="w-8 shrink-0 text-center">{t.dash_on}</span>
        <span className="w-24 shrink-0 text-right pr-1">{t.dash_actions}</span>
      </div>

      {/* Service rows */}
      <div className="flex flex-col gap-1.5">
        {services.map(svc => (
          <ServiceRow key={svc.name} service={svc} onNavigate={onNavigate} />
        ))}
        {services.length === 0 && (
          <p className="text-gray-600 text-sm text-center py-10">{t.dash_loading}</p>
        )}
      </div>
    </div>
  )
}
