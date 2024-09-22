import { createFileRoute } from '@tanstack/react-router'
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'

import {
  LineChart,
  Line,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  Legend,
  ResponsiveContainer,
} from 'recharts'
import {
  PlayCircle,
  StopCircle,
  Activity,
} from 'lucide-react'
import { getRoute, RouterStatus, startRoute, stopRoute } from '../../lib/api'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { formatBytes, formatXAxis } from '@/lib/utils'
import { RouteCardProps, RouteDetailCard } from '@/RouteCard'
import ProtectedRoute from '@/Protected'

const RouteFormComponent = () => {
  const { routesId } = Route.useParams()
  const { isPending, error, data, isLoading } = useQuery({
    queryKey: ['route', { id: routesId }],
    queryFn: () => getRoute(routesId),
  })


  if (isPending || isLoading) {
    return 'LOADING'
  } else if (error) {
    console.log(error)
    return JSON.stringify(error)
  }



  const route = data
  return (
    <div className="container mx-auto p-2 space-y-6">
      <h1 className="text-2xl font-bold mb-6">Route: {route.Name}</h1>

      <div className="grid grid-cols-1 gap-0 space-y-6 md:grid-cols-3 md:gap-6  md:space-y-0">
        <RouteDetailCard route={data} />
        <RouteStatusCard route={data} />
      </div>

      <Card>
        <CardHeader className="pb-2">
          <CardTitle className="text-lg">Traffic Statistics</CardTitle>
        </CardHeader>
        <CardContent className="space-y-2">
          <div className="flex justify-between">
            <p>
              <strong>Total Sent:</strong> {formatBytes(route.Stats.Total.Sent)}
            </p>
            <p>
              <strong>Total Received:</strong>{' '}
              {formatBytes(route.Stats.Total.Received)}
            </p>
          </div>
          {route.Stats.Points.length > 0 &&
            <ResponsiveContainer width="100%" height={200}>
              <LineChart data={route.Stats.Points}>
                <CartesianGrid strokeDasharray="3 3" />
                <XAxis
                  dataKey="Timestamp"
                  tickFormatter={formatXAxis}
                  domain={['dataMin', 'dataMax']}
                />
                <YAxis tickFormatter={formatBytes} />
                <Tooltip
                  contentStyle={{ backgroundColor: '#000' }}
                  formatter={formatBytes}
                  labelFormatter={formatXAxis}
                />
                <Legend />
                <Line
                  type="monotone"
                  dataKey="Value.Sent"
                  stroke="#8884d8"
                  name="Sent"
                />
                <Line
                  type="monotone"
                  dataKey="Value.Received"
                  stroke="#82ca9d"
                  name="Received"
                />
              </LineChart>
            </ResponsiveContainer>
          }
        </CardContent>
      </Card>
    </div>
  )
}



const RouteStatusCard = ({ route }: RouteCardProps) => {
  if (!route) {
    return
  }
  const queryClient = useQueryClient()
  const update = useMutation({
    mutationFn: route?.Status === RouterStatus.RUNNING ? stopRoute : startRoute,
    onSuccess: () => {
      queryClient.setQueryData(['route', { id: route?.Id }], { ...route, Status: route?.Status === RouterStatus.RUNNING ? RouterStatus.STOPPED : RouterStatus.RUNNING })
    },
  })
  const toggleStatus = () => {
    update.mutate(route?.Id)
  }
  return (
    <Card>
      <CardContent className="py-8 md:py-0 flex flex-col items-center space-y-4 h-full justify-center">
        <div className="flex items-center space-x-2">
          <Activity
            className={`h-5 w-5 ${route?.Status === RouterStatus.RUNNING ? 'text-green-500' : 'text-red-500'}`}
          />
          <Badge
            variant={
              route?.Status === RouterStatus.RUNNING ? 'success' : 'destructive'
            }
            className="text-xs px-2 py-1"
          >
            {route?.Status}
          </Badge>
        </div>
        <Button
          onClick={toggleStatus}
          variant={
            route?.Status === RouterStatus.RUNNING ? 'destructive' : 'default'
          }
          className="w-full"
        >
          {route?.Status === RouterStatus.RUNNING ? (
            <>
              <StopCircle className="mr-2 h-4 w-4" />
              Stop
            </>
          ) : (
            <>
              <PlayCircle className="mr-2 h-4 w-4" />
              Start
            </>
          )}
        </Button>
      </CardContent>
    </Card>
  )
}

export const Route = createFileRoute('/routes/$routesId')({
  component: ()=> <ProtectedRoute><RouteFormComponent/></ProtectedRoute>,
})
