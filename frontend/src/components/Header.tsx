import { useServiceStore } from '../store/serviceStore'
import { HideWindow } from '../../wailsjs/go/main/App'
import { useI18n, tt } from '../i18n'

export default function Header() {
  const { services } = useServiceStore()
  const { t } = useI18n()
  const runningCount = services.filter(s => s.status === 'running').length

  return (
    <div className="flex items-center justify-between px-6 py-3 border-b border-[#1e2535] bg-[#0a0f1a]">
      {/* Status summary */}
      <div className="flex items-center gap-3">
        <div className={`w-2 h-2 rounded-full ${runningCount > 0 ? 'bg-green-500' : 'bg-gray-600'}`} />
        <span className="text-sm text-gray-400">
          {runningCount > 0
            ? tt(t.header_services_running, { count: runningCount, s: runningCount > 1 ? 's' : '' })
            : t.header_all_stopped}
        </span>
      </div>

      {/* Minimize to tray */}
      <button
        onClick={() => HideWindow()}
        title={t.header_minimize}
        className="text-gray-500 hover:text-gray-300 text-xs px-3 py-1.5 rounded-lg hover:bg-[#1e2535] transition-colors"
      >
        {t.header_minimize}
      </button>
    </div>
  )
}
