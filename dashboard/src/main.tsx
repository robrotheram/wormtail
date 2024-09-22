import { StrictMode } from 'react'
import ReactDOM from 'react-dom/client'
import { RouterProvider, createRouter } from '@tanstack/react-router'
import "./main.css"
// Import the generated route tree
import { routeTree } from './routeTree.gen'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { AuthProvider } from './AuthContext'

// Create a new router instance
const router = createRouter({ routeTree })

// Register the router instance for type safety
declare module '@tanstack/react-router' {
   interface Register {
      router: typeof router
   }
}

const queryClient = new QueryClient()

// Render the app
const rootElement = document.getElementById('root')!
if (!rootElement.innerHTML) {
   const root = ReactDOM.createRoot(rootElement)
   root.render(
      <StrictMode>
         <QueryClientProvider client={queryClient}>
            <AuthProvider>
               <RouterProvider router={router} />
            </AuthProvider>
         </QueryClientProvider>
      </StrictMode>,
   )
}