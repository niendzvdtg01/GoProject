import { BrowserRouter } from 'react-router-dom'
import { QueryProvider } from './QueryProvider.jsx'
import { setupApiInterceptors } from '../../shared/services/interceptor.js'

setupApiInterceptors()

export function AppProviders({ children }) {
  return (
    <QueryProvider>
      <BrowserRouter>{children}</BrowserRouter>
    </QueryProvider>
  )
}
