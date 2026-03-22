import { useEffect, useState } from 'react'
import { clearLogs } from '../../api/client'
import { Card } from '../ui/Card'
import { Button } from '../ui/Button'

export function LogViewer() {
  const [lines, setLines] = useState<string[]>([])
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')
  const [search, setSearch] = useState('')
  const [logSource, setLogSource] = useState<'laravel' | 'nginx-access' | 'nginx-error'>('laravel')

  const refresh = () => {
    setLoading(true)
    setError('')
    const endpoint = logSource === 'laravel'
      ? '/api/logs/laravel'
      : `/api/logs/${logSource}`
    fetch(endpoint)
      .then(res => res.json())
      .then((res) => { setLines(res.lines || []); if (res.error) setError(res.error) })
      .catch(() => setError('Failed to fetch logs'))
      .finally(() => setLoading(false))
  }

  useEffect(() => { refresh() }, [logSource])

  const handleClear = async () => {
    setLoading(true)
    try {
      const res = await clearLogs()
      if (res.success) setLines([])
    } catch {
      setError('Failed to clear log.')
    } finally {
      setLoading(false)
    }
  }

  const filteredLines = search
    ? lines.filter(line => line.toLowerCase().includes(search.toLowerCase()))
    : lines

  return (
    <Card title="Log Viewer">
      <div className="flex gap-2 mb-3">
        {[
          { key: 'laravel' as const, label: 'Laravel' },
          { key: 'nginx-access' as const, label: 'Nginx Access' },
          { key: 'nginx-error' as const, label: 'Nginx Error' },
        ].map((tab) => (
          <button
            key={tab.key}
            onClick={() => { setLogSource(tab.key); setSearch('') }}
            className={`px-3 py-1 text-xs rounded-md transition-colors ${
              logSource === tab.key
                ? 'bg-indigo-600 text-white'
                : 'bg-gray-800 text-gray-400 hover:bg-gray-700'
            }`}
          >
            {tab.label}
          </button>
        ))}
      </div>

      <div className="flex justify-end gap-2 mb-3">
        <Button variant="ghost" onClick={refresh} loading={loading}>Refresh</Button>
        {logSource === 'laravel' && (
          <Button variant="danger" onClick={handleClear}>Clear Log</Button>
        )}
      </div>

      {error && <div className="text-red-400 text-xs mb-2">{error}</div>}

      <input
        type="text"
        placeholder="Search logs..."
        value={search}
        onChange={(e) => setSearch(e.target.value)}
        className="w-full bg-gray-800 border border-gray-700 rounded-md px-3 py-1.5 text-sm text-gray-100 placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-transparent mb-3"
      />

      {search && lines.length > 0 && (
        <div className="text-xs text-gray-500 mb-2">{filteredLines.length} of {lines.length} lines match</div>
      )}

      <pre className="bg-gray-950 text-gray-300 text-xs p-4 rounded-md overflow-auto max-h-96 font-mono">
        {filteredLines.length > 0 ? filteredLines.join('\n') : (search ? 'No matching lines.' : 'Log is empty.')}
      </pre>
    </Card>
  )
}
