
import { Badge } from '@/components/ui/badge'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { getRoutes } from '@/lib/api'
import { useQuery } from '@tanstack/react-query'
import { Link, useNavigate } from '@tanstack/react-router'
import { Button } from './components/ui/button'
import { PlusIcon } from './Icons'

export const RouteList = () => {
    const navigate = useNavigate({ from: '/' })
    const { isPending, error, data, isLoading } = useQuery({
      queryKey: ['repoData'],
      queryFn: getRoutes,
    })
  
    if (isPending || isLoading) {
      return "LOADING"
    } else if (error) {
      console.log(error)
      return JSON.stringify(error)
    }

    data.sort((a, b) => {
      return a.Name.toLowerCase().localeCompare(b.Name.toLowerCase());
    });
  
    return <Card>
        <CardHeader className='flex flex-row justify-between'>
          <div className='space-y-1.5 flex flex-col'>
          <CardTitle>Routes</CardTitle>
          <CardDescription>Manage your load balancer routes.</CardDescription>
          </div>
          <Link to='/create' className='float-right'>
            <Button><PlusIcon className='mr-2'/>New Route</Button>
          </Link>
        </CardHeader>
        <CardContent>
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Route Name</TableHead>
                <TableHead>External Port</TableHead>
                <TableHead>Interal Host</TableHead>
                <TableHead>Interal Port</TableHead>
                <TableHead>Status</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {
                data.map(route => {
                return  <TableRow onClick={()=> navigate({to: `/routes/${route.Id}`}) } className='cursor-pointer' key={route.Id}>
                  <TableCell className="font-medium">{route.Name}</TableCell>
                  <TableCell>{route.Type}</TableCell>
                  <TableCell>{route.Machine.Address}</TableCell>
                  <TableCell>{route.Machine.Port}</TableCell>
                  <TableCell>
                    <Badge variant="secondary">{route.Status}</Badge>
                  </TableCell>
                </TableRow>
              })}
            </TableBody>
          </Table>
        </CardContent>
      </Card>
  }