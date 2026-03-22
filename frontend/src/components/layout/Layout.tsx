import { ReactNode, useState } from 'react'
import { useLocation } from 'react-router-dom'
import { Sidebar } from './Sidebar'

const pageTitles: Record<string, string> = {
  '/wizard': 'Deployment Setup',
  '/dashboard': 'Dashboard',
  '/actions': 'Actions',
  '/steps': 'Deploy Steps',
  '/logs': 'Logs',
  '/env': 'Environment',
  '/settings': 'Settings',
}

interface LayoutProps {
  children: ReactNode
}

export function Layout({ children }: LayoutProps) {
  const [mobileOpen, setMobileOpen] = useState(false)
  const location = useLocation()
  const title = pageTitles[location.pathname] || 'Deploy Panel'
  const isWizard = location.pathname === '/wizard'

  return (
    <div className="h-full">
      <Sidebar mobileOpen={mobileOpen} onClose={() => setMobileOpen(false)} />

      <div className="xl:pl-64 flex flex-col h-full">
        {/* Top bar */}
        <div className="sticky top-0 z-40 flex h-12 shrink-0 items-center gap-x-4 border-b border-white/5 bg-gray-950/80 px-4 backdrop-blur-md sm:px-6">
          <button
            type="button"
            onClick={() => setMobileOpen(true)}
            className="-m-2.5 p-2.5 text-gray-400 hover:text-white xl:hidden"
          >
            <svg viewBox="0 0 20 20" fill="currentColor" className="size-5">
              <path fillRule="evenodd" clipRule="evenodd" d="M2 4.75A.75.75 0 0 1 2.75 4h14.5a.75.75 0 0 1 0 1.5H2.75A.75.75 0 0 1 2 4.75ZM2 10a.75.75 0 0 1 .75-.75h14.5a.75.75 0 0 1 0 1.5H2.75A.75.75 0 0 1 2 10Zm0 5.25a.75.75 0 0 1 .75-.75h14.5a.75.75 0 0 1 0 1.5H2.75a.75.75 0 0 1-.75-.75Z" />
            </svg>
          </button>
          <div className="h-5 w-px bg-white/10 xl:hidden" />
          <h1 className="text-sm font-medium text-white">{title}</h1>
        </div>

        {/* Page content — wizard gets centered, other pages scroll normally */}
        {isWizard ? (
          <div className="flex-1 flex items-center justify-center overflow-y-auto">
            <div className="w-full">{children}</div>
          </div>
        ) : (
          <div className="flex-1 overflow-y-auto">
            {children}
          </div>
        )}
      </div>
    </div>
  )
}
