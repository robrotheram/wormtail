
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from '@/components/ui/card'
import { createRoute, defaultRoute, deleteRoute, Route, RouterType, updateRoute } from '@/lib/api'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { useNavigate } from '@tanstack/react-router'
import { Edit, Plus, Save, Trash } from 'lucide-react'
import { Button } from './components/ui/button'
import { useState } from 'react'
import { Input } from './components/ui/input'
import { Label } from './components/ui/label'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from './components/ui/select'


export type RouteCardProps = {
  route?: Route
}
export const RouteDetailCard = ({ route }: RouteCardProps) => {
  const create = route === undefined

  const [editMode, setEditMode] = useState(create)
  const [editedRoute, setEditedRoute] = useState<Route>(route||defaultRoute)
  const queryClient = useQueryClient()
  const navigate = useNavigate({ from: create ? `/` : `/routes/${route.Id}` })
  const update = useMutation({
    mutationFn: create ? createRoute : updateRoute,
    onSuccess: (data) => {
      create ?
        navigate({ to: `/routes/${data.Id}` })
        :
        queryClient.refetchQueries({ queryKey: ['route', { id: route.Id }] })
    },
  })

  const deleteFn = useMutation({
    mutationFn: deleteRoute,
    onSuccess: () => { navigate({ to: `/` }) }
  })

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const { name, value } = e.target
    console.log("TARGET", name, value)
    setEditedRoute((prevRoute) => ({
      ...prevRoute,
      [name]: name === 'Port' || name === 'Listen' ? parseInt(value, 10) || value : value,
    }))
  }

  const handleInputMachineChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const { name, value } = e.target
    setEditedRoute((prevRoute) => ({
      ...prevRoute, Machine: {
        ...prevRoute.Machine,
        [name]: name === 'Port' || name === 'Listen' ? parseInt(value, 10) || value : value,
      }
    }))
  }

  const handleSelectChange = (value: string) => {
    setEditedRoute((prevRoute) => ({
      ...prevRoute,
      Type: value as RouterType,
    }))
  }

  const handleEdit = () => {
    setEditMode(true)
  }

  const handleSave = () => {
    update.mutate(editedRoute)
    setEditMode(false)
  }

  const handleDelete = () => {
    deleteFn.mutate(editedRoute)
  }

  return (
    <Card className="col-span-2">
      <CardHeader className="pb-2">
        <CardTitle className="text-lg">General Information</CardTitle>
        <CardDescription className="text-sm mt-1">
          {editMode
            ? 'Edit the firewall route details below'
            : 'View firewall route details'}
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-2">
        {editMode && (
          <div>
            <Label htmlFor="name">Route Name</Label>
            <Input
              id="name"
              name="Name"
              value={editMode ? editedRoute.Name : route?.Name}
              onChange={handleInputChange}
            />
          </div>
        )}

        <div className="grid grid-cols-2 gap-2">
          <div>
            <Label htmlFor="host">Host</Label>
            <Input
              id="host"
              name="Address"
              value={
                editMode ? editedRoute.Machine.Address : route?.Machine.Address
              }
              onChange={handleInputMachineChange}
              disabled={!editMode}
            />
          </div>
          <div>
            <Label htmlFor="port">Port</Label>
            <Input
              id="port"
              name="Port"
              type="number"
              value={editMode ? editedRoute.Machine.Port : route?.Machine.Port}
              onChange={handleInputMachineChange}
              disabled={!editMode}
            />
          </div>
        </div>
        <div>
          <Label htmlFor="type">Route Type</Label>
          <Select
            name="type"
            disabled={!editMode}
            onValueChange={handleSelectChange}
          >
            <SelectTrigger className="w-full">
              <SelectValue placeholder={editedRoute.Type} />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value={RouterType.HTTP}>{RouterType.HTTP}</SelectItem>
              <SelectItem value={RouterType.TCP}>{RouterType.TCP}</SelectItem>
              <SelectItem value={RouterType.UDP}>{RouterType.UDP}</SelectItem>
            </SelectContent>
          </Select>
        </div>

        {editedRoute.Type === RouterType.TCP && (
          <div>
            <Label htmlFor="listen">Listen</Label>
            <Input
              id="listen"
              name="Port"
              type="number"
              value={editMode ? editedRoute.Port : route?.Port}
              onChange={handleInputChange}
              disabled={!editMode}
            />
          </div>
        )}
      </CardContent>
      <CardFooter
        className={`flex ${editMode && !create ? 'justify-between' : 'justify-end'}`}
      >
        {editMode && !create && (
          <Button onClick={handleDelete} variant="destructive">
            <Trash className="mr-2 h-4 w-4" />
            Delete Route
          </Button>
        )}

        {editMode ? (
          <Button onClick={handleSave}>
            {create ? <Plus className="mr-2 h-4 w-4" /> : <Save className="mr-2 h-4 w-4" />}
            {create ? "Add" : "Update"}
          </Button>
        ) : (
          <Button onClick={handleEdit} className="w-24">
            <Edit className="mr-2 h-4 w-4" />
            Edit
          </Button>
        )}
      </CardFooter>
    </Card>
  )
}