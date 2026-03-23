import {
  LayoutDashboard,
  Package,
  Database,
  Globe,
  FileCode,
  ScrollText,
  TerminalSquare,
  Settings,
  Code2,
  Play,
  Square,
  FolderKanban,
} from 'lucide-react'
import { useServiceStore } from '../store/serviceStore'
import { useI18n } from '../i18n'

export type Page = 'dashboard' | 'projects' | 'binaries' | 'database' | 'vhosts' | 'logs' | 'terminal' | 'config' | 'php' | 'settings'

interface Props {
  current: Page
  onNavigate: (page: Page) => void
}

export default function Sidebar({ current, onNavigate }: Props) {
  const { services, startAll, stopAll } = useServiceStore()
  const { t } = useI18n()
  const runningCount = services.filter(s => s.status === 'running').length

  const navItems: { id: Page; label: string; icon: React.ReactNode }[] = [
    { id: 'dashboard', label: t.nav_dashboard, icon: <LayoutDashboard size={16} /> },
    { id: 'projects', label: t.nav_projects, icon: <FolderKanban size={16} /> },
    { id: 'binaries', label: t.nav_binaries, icon: <Package size={16} /> },
    { id: 'database', label: t.nav_database, icon: <Database size={16} /> },
    { id: 'vhosts', label: t.nav_vhosts, icon: <Globe size={16} /> },
    { id: 'config', label: t.nav_config, icon: <FileCode size={16} /> },
    { id: 'php', label: t.nav_php, icon: <Code2 size={16} /> },
    { id: 'logs', label: t.nav_logs, icon: <ScrollText size={16} /> },
    { id: 'terminal', label: t.nav_terminal, icon: <TerminalSquare size={16} /> },
    { id: 'settings', label: t.nav_settings, icon: <Settings size={16} /> },
  ]

  return (
    <aside className="w-56 bg-[#0a0f1a] border-r border-[#1e2535] flex flex-col py-4 px-3 gap-1">
      {/* Brand */}
      <div className="px-3 mb-4">
        <h1 className="text-lg font-bold text-white tracking-tight">
          <span className="text-blue-400">Stack</span>nest
        </h1>
        <p className="text-[10px] text-gray-600 mt-0.5">v0.1.0</p>
      </div>

      {/* Navigation */}
      {navItems.map(item => (
        <button
          key={item.id}
          onClick={() => onNavigate(item.id)}
          className={`flex items-center gap-3 px-3 py-2 rounded-lg text-sm transition-colors
            ${current === item.id
              ? 'bg-blue-500/15 text-blue-400 font-medium'
              : 'text-gray-400 hover:text-white hover:bg-[#1e2535]'
            }`}
        >
          {item.icon}
          <span>{item.label}</span>
        </button>
      ))}

      {/* Spacer */}
      <div className="flex-1" />

      {/* Quick actions */}
      <div className="flex flex-col gap-1 border-t border-[#1e2535] pt-3 mt-2">
        <button
          onClick={startAll}
          className="flex items-center gap-2 px-3 py-2 text-sm text-green-400 hover:bg-green-500/10 rounded-lg transition-colors"
        >
          <Play size={12} /> {t.dash_start_all}
        </button>
        <button
          onClick={stopAll}
          disabled={runningCount === 0}
          className="flex items-center gap-2 px-3 py-2 text-sm text-red-400 hover:bg-red-500/10 rounded-lg transition-colors disabled:opacity-30"
        >
          <Square size={12} /> {t.dash_stop_all}
        </button>
      </div>
    </aside>
  )
}
