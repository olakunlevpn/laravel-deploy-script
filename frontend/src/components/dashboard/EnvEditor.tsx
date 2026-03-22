import { useEffect, useState } from 'react'
import { Card } from '../ui/Card'
import { Button } from '../ui/Button'

export function EnvEditor() {
  const [content, setContent] = useState('')
  const [loading, setLoading] = useState(false)
  const [saved, setSaved] = useState(false)
  const [error, setError] = useState('')

  const refresh = () => {
    setLoading(true)
    fetch('/api/env')
      .then(res => res.json())
      .then((res) => { setContent(res.content || ''); setError(res.error || '') })
      .catch(() => setError('Failed to load .env'))
      .finally(() => setLoading(false))
  }

  useEffect(() => { refresh() }, [])

  const handleSave = async () => {
    setLoading(true)
    setSaved(false)
    setError('')
    try {
      const res = await fetch('/api/env', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ content }),
      })
      const data = await res.json()
      if (data.success) {
        setSaved(true)
        setTimeout(() => setSaved(false), 3000)
      } else {
        setError(data.output || 'Failed to save')
      }
    } catch {
      setError('Failed to save .env')
    } finally {
      setLoading(false)
    }
  }

  return (
    <Card title=".env Editor">
      {error && <div className="text-red-400 text-xs mb-2">{error}</div>}
      {saved && <div className="text-green-400 text-xs mb-2">.env saved successfully</div>}
      <textarea
        value={content}
        onChange={(e) => setContent(e.target.value)}
        className="w-full h-64 bg-gray-950 text-gray-300 text-xs p-3 rounded-md font-mono border border-gray-700 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-transparent resize-y"
        spellCheck={false}
      />
      <div className="flex justify-end gap-2 mt-3">
        <Button variant="ghost" onClick={refresh} loading={loading}>Reload</Button>
        <Button onClick={handleSave} loading={loading}>Save .env</Button>
      </div>
    </Card>
  )
}
