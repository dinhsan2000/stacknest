import { Component, ReactNode } from 'react'

interface Props { children: ReactNode }
interface State { error: Error | null }

export default class ErrorBoundary extends Component<Props, State> {
  state: State = { error: null }

  static getDerivedStateFromError(error: Error): State {
    return { error }
  }

  render() {
    if (this.state.error) {
      return (
        <div className="flex flex-col items-center justify-center h-full gap-4 text-center">
          <p className="text-4xl">⚠️</p>
          <p className="text-white font-semibold">Component crashed</p>
          <pre className="text-red-400 text-xs bg-red-500/10 rounded p-3 max-w-lg text-left overflow-auto">
            {this.state.error.message}
          </pre>
          <button
            onClick={() => this.setState({ error: null })}
            className="px-4 py-2 rounded-lg bg-blue-500 text-white text-sm"
          >
            Retry
          </button>
        </div>
      )
    }
    return this.props.children
  }
}
