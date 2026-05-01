import { RegisterForm } from '../../features/auth/components/RegisterForm.jsx'
import { AuthLayout } from '../../shared/layouts/AuthLayout.jsx'

export function RegisterPage() {
  return (
    <AuthLayout>
      <RegisterForm />
    </AuthLayout>
  )
}
