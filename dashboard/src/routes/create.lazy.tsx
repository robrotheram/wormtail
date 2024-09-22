import ProtectedRoute from '@/Protected'
import { RouteDetailCard } from '@/RouteCard'
import { createLazyFileRoute } from '@tanstack/react-router'

export const Route = createLazyFileRoute('/create')({
  component: () => <ProtectedRoute><RouteDetailCard /></ProtectedRoute>,
})
