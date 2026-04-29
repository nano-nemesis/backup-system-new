import { createBrowserRouter } from 'react-router-dom'
import AuthGuard from './AuthGuard'
import LoginPage from '@/pages/LoginPage'
import SetupPage from '@/pages/SetupPage'
import DashboardPage from '@/pages/DashboardPage'
import NodeDetailPage from '@/pages/NodeDetailPage'
import AdminPage from '@/pages/AdminPage'

export const router = createBrowserRouter([
  {
    path: '/login',
    element: <LoginPage />,
  },
  {
    path: '/setup',
    element: <SetupPage />,
  },
  {
    element: <AuthGuard />,
    children: [
      { path: '/', element: <DashboardPage /> },
      { path: '/nodes/:name', element: <NodeDetailPage /> },
      { path: '/admin', element: <AdminPage /> },
    ],
  },
])
