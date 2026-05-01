import { LoginForm } from '../../features/auth/components/LoginForm.jsx'
import { AuthLayout } from '../../shared/layouts/AuthLayout.jsx'

export function LoginPage() {
  return (
    <AuthLayout>
      <LoginForm />
    </AuthLayout>
  )
}
