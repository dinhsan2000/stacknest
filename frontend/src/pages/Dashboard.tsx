import { useEffect } from 'react'
import { useServiceStore } from '../store/serviceStore'
import ServiceRow from '../components/ServiceCard'

export default function Dashboard() {
  const { services, fetchServices, startAll, stopAll } = useServiceStore()

  useEffect(() => {
    fetchServices()
  }, [])

  const runningCount = services.filter(s => s.status === 'running').length
  const allRunning   = runningCount === services.length && services.length > 0

  return (
    <div className="flex flex-col gap-6 max-w-4xl">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-2xl font-bold text-white">Dashboard</h2>
          <p className="text-gray-500 text-sm mt-1">
            {runningCount} of {services.length} services running
          </p>
        </div>

        <div className="flex gap-2">
          <button
            onClick={stopAll}
            disabled={runningCount === 0}
            className="px-4 py-2 rounded-lg text-sm font-medium bg-[#1e2535] text-gray-400 hover:bg-[#2a3347] hover:text-white transition-colors disabled:opacity-30"
          >
            Stop All
          </button>
          <button
            onClick={startAll}
            disabled={allRunning}
            className="px-4 py-2 rounded-lg text-sm font-medium bg-blue-500/20 text-blue-400 hover:bg-blue-500/30 transition-colors disabled:opacity-30"
          >
            Start All
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
      <div className="flex items-center gap-4 px-4 text-xs font-medium text-gray-600 uppercase tracking-wider">
        <span className="w-36 shrink-0">Service</span>
        <span className="w-24 shrink-0">Status</span>
        <span className="w-16 shrink-0">Port</span>
        <span className="w-24 shrink-0">PID</span>
        <span className="flex-1" />
        <span className="w-8 shrink-0 text-center">On</span>
        <span className="w-24 shrink-0 text-right pr-1">Actions</span>
      </div>

      {/* Service rows */}
      <div className="flex flex-col gap-1.5">
        {services.map(svc => (
          <ServiceRow key={svc.name} service={svc} />
        ))}
        {services.length === 0 && (
          <p className="text-gray-600 text-sm text-center py-10">Loading services…</p>
        )}
      </div>
    </div>
  )
}
