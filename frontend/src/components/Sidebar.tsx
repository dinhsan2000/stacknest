import React from 'react'
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
} from 'lucide-react'
import { useServiceStore } from '../store/serviceStore'

export type Page = 'dashboard' | 'binaries' | 'database' | 'vhosts' | 'logs' | 'terminal' | 'config' | 'php' | 'settings'

interface Props {
  current: Page
  onChange: (page: Page) => void
}

const navItems: { id: Page; label: string; icon: React.ReactNode }[] = [
  { id: 'dashboard', label: 'Dashboard', icon: <LayoutDashboard size={16} /> },
  { id: 'binaries', label: 'Binaries', icon: <Package size={16} /> },
  { id: 'database', label: 'Database', icon: <Database size={16} /> },
  { id: 'vhosts', label: 'Virtual Hosts', icon: <Globe size={16} /> },
  { id: 'config', label: 'Config Editor', icon: <FileCode size={16} /> },
  { id: 'php', label: 'PHP Versions', icon: <Code2 size={16} /> },
  { id: 'logs', label: 'Log Viewer', icon: <ScrollText size={16} /> },
  { id: 'terminal', label: 'Terminal', icon: <TerminalSquare size={16} /> },
  { id: 'settings', label: 'Settings', icon: <Settings size={16} /> },
]

export default function Sidebar({ current, onChange }: Props) {
  const { services, binaryStatus } = useServiceStore()
  const errorCount = services.filter(s => s.status === 'error').length
  const missingBinaries = binaryStatus.filter(s => !s.versions.some(v => v.installed)).length

  return (
    <aside className="w-52 min-h-screen bg-[#0f1420] border-r border-[#1e2535] flex flex-col py-6 px-3">
      {/* Logo */}
      <div className="px-3 mb-8">
        <h1 className="text-xl font-bold text-white tracking-tight">
          Stack<span className="text-blue-400">nest</span>
        </h1>
        <p className="text-xs text-gray-500 mt-0.5">Dev Environment Manager</p>
      </div>

      {/* Nav */}
      <nav className="flex flex-col gap-0.5">
        {navItems.map(item => (
          <button
            key={item.id}
            onClick={() => onChange(item.id)}
            className={`flex items-center gap-3 px-3 py-2 rounded-lg text-sm font-medium transition-colors text-left w-full
              ${current === item.id
                ? 'bg-blue-500/20 text-blue-400'
                : 'text-gray-400 hover:bg-[#1e2535] hover:text-white'
              }`}
          >
            <span className="shrink-0">{item.icon}</span>
            <span className="flex-1">{item.label}</span>
            {item.id === 'logs' && errorCount > 0 && (
              <span className="bg-red-500 text-white text-xs rounded-full px-1.5 py-0.5 min-w-[18px] text-center">
                {errorCount}
              </span>
            )}
            {item.id === 'binaries' && missingBinaries > 0 && (
              <span className="bg-yellow-500 text-black text-xs rounded-full px-1.5 py-0.5 min-w-[18px] text-center font-medium">
                {missingBinaries}
              </span>
            )}
          </button>
        ))}
      </nav>

      {/* Bottom version */}
      <div className="mt-auto px-3">
        <p className="text-xs text-gray-600">Stacknest v0.1.0</p>
      </div>
    </aside>
  )
}
