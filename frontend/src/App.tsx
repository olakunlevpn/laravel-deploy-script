import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom'
import { Layout } from './components/layout/Layout'
import { Wizard } from './pages/Wizard'
import { Dashboard } from './pages/Dashboard'
import { Actions } from './pages/Actions'
import { Steps } from './pages/Steps'
import { Logs } from './pages/Logs'
import { Env } from './pages/Env'
import { Settings } from './pages/Settings'

export default function App() {
  return (
    <BrowserRouter>
      <Layout>
        <Routes>
          <Route path="/wizard" element={<Wizard />} />
          <Route path="/dashboard" element={<Dashboard />} />
          <Route path="/actions" element={<Actions />} />
          <Route path="/steps" element={<Steps />} />
          <Route path="/logs" element={<Logs />} />
          <Route path="/env" element={<Env />} />
          <Route path="/settings" element={<Settings />} />
          <Route path="*" element={<Navigate to="/wizard" replace />} />
        </Routes>
      </Layout>
    </BrowserRouter>
  )
}
