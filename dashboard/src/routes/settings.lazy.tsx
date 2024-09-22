import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Switch } from '@/components/ui/switch'
import { Dashboard, getDashboardConfig, getTSConfig, Tailsale, updateDashboardConfig, updateTSConfig } from '@/lib/api'
import ProtectedRoute from '@/Protected'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { createLazyFileRoute } from '@tanstack/react-router'
import { Save, Edit } from 'lucide-react'
import { useState } from 'react'

type TailScaleFormProps = {
    config: Tailsale
}
const TailScaleForm = ({ config }: TailScaleFormProps) => {
    const [editMode, setEditMode] = useState(false)
    const [editedRoute, setEditedRoute] = useState<Tailsale>(config)
    const queryClient = useQueryClient()
    const update = useMutation({
        mutationFn: updateTSConfig,
        onSuccess: (data) => {
            console.log(data)
            queryClient.setQueryData(['settings', 'tailscale'], data)
        },
    })

    const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
        const { name, value } = e.target
        console.log("TARGET", name, value)
        setEditedRoute((prevRoute) => ({
            ...prevRoute,
            [name]: value,
        }))
    }

    const handleEdit = () => {
        setEditMode(true)
        setEditedRoute(config)
    }

    const handleSave = () => {
        update.mutate(editedRoute)
        setEditMode(false)
    }
    return <>
        <CardContent>
            <div className="grid grid-cols-2 gap-2">
                <div>
                    <Label htmlFor="APIKey">Tailscale API Key</Label>
                    <Input
                        id="APIKey"
                        name="APIKey"
                        type="password"
                        value={editMode ? editedRoute.APIKey : config.APIKey}
                        onChange={handleInputChange}
                        disabled={!editMode}
                    />
                </div>
                <div>
                    <Label htmlFor="Hostname">Tailscale Hostname</Label>
                    <Input
                        id="Hostname"
                        name="Hostname"
                        type="text"
                        value={editMode ? editedRoute.Hostname : config.Hostname}
                        onChange={handleInputChange}
                        disabled={!editMode}
                    />
                </div>
            </div>
        </CardContent>
        <CardFooter
            className={`flex justify-end`}
        >


            {editMode ? (
                <Button onClick={handleSave} className="w-24">
                    <Save className="mr-2 h-4 w-4" />
                    Save
                </Button>
            ) : (
                <Button onClick={handleEdit} className="w-24">
                    <Edit className="mr-2 h-4 w-4" />
                    Edit
                </Button>
            )}
        </CardFooter>
    </>
}


const DashboardForm = () => {
    const {data} = useQuery({
        queryKey: ['settings', 'dashboard'],
        queryFn: getDashboardConfig,
    })
    const [editMode, setEditMode] = useState(false)
    const [editedRoute, setEditedRoute] = useState<Dashboard>(data||{Enabled:true, Username:"", Password:""})
    const queryClient = useQueryClient()
    const update = useMutation({
        mutationFn: updateDashboardConfig,
        onSuccess: (data) => {
            console.log(data)
            queryClient.setQueryData(['settings', 'dashboard'], data)
        },
    })

    const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
        const { name, value } = e.target
        console.log("TARGET", name, value)
        setEditedRoute((prevRoute) => ({
            ...prevRoute,
            [name]: value,
        }))
    }

    const handleEdit = () => {
        data && setEditedRoute(data)
        setEditMode(true)
    }

    const handleSave = () => {
        update.mutate(editedRoute)
        setEditMode(false)
    }

    const handleEnabled = (value:boolean) => {
        update.mutate(editedRoute)
        setEditedRoute((prevRoute) => ({...prevRoute, Enabled: value}))
    }

    return <>
        <CardContent>
            <div className="grid grid-cols-2 gap-2">
                <div>
                    <Label htmlFor="APIKey">Username</Label>
                    <Input
                        id="Username"
                        name="Username"
                        type="text"
                        value={editMode ? editedRoute.Username : data?.Username}
                        onChange={handleInputChange}
                        disabled={!editMode}
                    />
                </div>
                <div>
                    <Label htmlFor="Hostname">Password</Label>
                    <Input
                        id="Password"
                        name="Password"
                        type="password"
                        value={editMode ? editedRoute.Password : data?.Password}
                        onChange={handleInputChange}
                        disabled={!editMode}
                    />
                </div>
                <div className='flex items-center gap-2'>
                    <Label htmlFor="Hostname">Enabled: </Label>
                    <Switch
                      checked={editMode ? editedRoute.Enabled : data?.Enabled}
                      onCheckedChange={handleEnabled}
                      disabled={!editMode}
                    />
                </div>


            </div>
        </CardContent>
        <CardFooter
            className={`flex justify-end`}
        >


            {editMode ? (
                <Button onClick={handleSave} className="w-24">
                    <Save className="mr-2 h-4 w-4" />
                    Save
                </Button>
            ) : (
                <Button onClick={handleEdit} className="w-24">
                    <Edit className="mr-2 h-4 w-4" />
                    Edit
                </Button>
            )}
        </CardFooter>
    </>
}



const SettingComponent = () => {
    const { isPending, error, data, isLoading } = useQuery({
        queryKey: ['settings', 'tailscale'],
        queryFn: getTSConfig,
    })

    if (isPending || isLoading) {
        return "LOADING"
    } else if (error) {
        console.log(error)
        return JSON.stringify(error)
    }

    return <div className='flex flex-col space-y-6'>
        <div>
            <h1 className='text-3xl'>Settings</h1>
            <h2 className='text-xl'>Manage connection</h2>
        </div>
        <Card>
            <CardHeader>
                <CardTitle>Dashboard</CardTitle>
                <CardDescription>Manage the connection</CardDescription>
            </CardHeader>
            <DashboardForm/>
        </Card>
        <Card>
            <CardHeader>
                <CardTitle>Tailscale</CardTitle>
                <CardDescription>Manage the connection</CardDescription>
            </CardHeader>
            <TailScaleForm config={data} />
        </Card>
    </div>
}


export const Route = createLazyFileRoute('/settings')({
    component: () => <ProtectedRoute><SettingComponent /></ProtectedRoute>,
})
