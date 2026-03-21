import { useState, useEffect } from 'react'
import { useServiceStore } from './store/serviceStore'
import Sidebar, { type Page } from './components/Sidebar'
import Header from './components/Header'
import ErrorBoundary from './components/ErrorBoundary'
import Binaries from './pages/Binaries'
import Dashboard from './pages/Dashboard'
import VirtualHosts from './pages/VirtualHosts'
import LogViewer from './pages/LogViewer'
import Terminal from './pages/Terminal'
import Database from './pages/Database'
import ConfigEditor from './pages/ConfigEditor'
import Settings from './pages/Settings'
import PHPSwitcher from './pages/PHPSwitcher'

function App() {
  const [page, setPage] = useState<Page>('dashboard')
  const { initEventListeners } = useServiceStore()

  useEffect(() => {
    initEventListeners()
  }, [])

  const renderPage = () => {
    switch (page) {
      case 'dashboard': return <Dashboard />
      case 'binaries': return <Binaries />
      case 'vhosts': return <VirtualHosts />
      case 'logs': return <LogViewer />
      case 'terminal': return <Terminal />
      case 'database': return <Database />
      case 'config': return <ConfigEditor />
      case 'php': return <PHPSwitcher />
      case 'settings': return <Settings />
      default: return <Dashboard />
    }
  }

  return (
    <div className="flex h-screen bg-[#0f1420] text-white font-sans overflow-hidden">
      <Sidebar current={page} onNavigate={setPage} />
      <div className="flex-1 flex flex-col overflow-hidden">
        <Header />
        {/* flex flex-col để child có thể dùng flex-1 và h-full */}
        <main className="flex-1 flex flex-col p-8 overflow-auto min-h-0">
          <ErrorBoundary key={page}>
            {renderPage()}
          </ErrorBoundary>
        </main>
      </div>
    </div>
  )
}

export default App
