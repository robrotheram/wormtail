import ProtectedRoute from '@/Protected'
import { RouteList } from '../../RouteList'
import { createFileRoute } from '@tanstack/react-router'

export const Route = createFileRoute('/routes/')({
    component: () => <ProtectedRoute><RouteList/></ProtectedRoute>,
})
