import ProtectedRoute from '@/Protected'
import { RouteList } from '@/RouteList'
import { createLazyFileRoute } from '@tanstack/react-router'

export const Route = createLazyFileRoute('/')({
  component: () => <ProtectedRoute><RouteList/></ProtectedRoute>,
})
