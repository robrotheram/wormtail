import React, { useState } from 'react'
import { useAuth } from './AuthContext'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Label } from '@/components/ui/label'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'
import { useMutation } from '@tanstack/react-query'
import { login as api } from "./lib/api"
import { useNavigate } from '@tanstack/react-router'



export const LoginPage: React.FC = () => {
    const [username, setUsername] = useState('')
    const [password, setPassword] = useState('')
    const { login } = useAuth()
    const navigate = useNavigate()
    const authenticate = useMutation({
      mutationFn: api,
      onSuccess: (data) => {
        login(data.authorization_token)
        navigate({ to: '/' })
      },
      onError: ()=>{
        alert('Invalid credentials')
      }
    })
  
    const handleSubmit = async (e: React.FormEvent) => {
      e.preventDefault()
      authenticate.mutate({
        username,
        password
      })
    }
  
    return (
      <Card className="col-span-2 m-10">
        <CardHeader className="pb-2 flex flex-row items-center gap-4">
          <img src='/logo.png' className='w-20'/>
          <CardTitle className="text-3xl">WarpTail Login</CardTitle>
        </CardHeader>
        <CardContent>
          <form onSubmit={handleSubmit} className='flex flex-col space-y-4'>
            <div>
              <Label>Username: </Label>
              <Input
                type="text"
                value={username}
                onChange={(e) => setUsername(e.target.value)}
              />
            </div>
            <div>
              <Label>Password: </Label>
              <Input
                type="password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
              />
            </div>
  
            <Button type="submit" className='w-24'>Login</Button>
          </form>
        </CardContent>
      </Card>
    )
  }
  