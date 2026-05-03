import { RegisterForm } from './RegisterForm.jsx'
import { AuthLayout } from '../../shared/layouts/AuthLayout.jsx'

export function RegisterPage() {
  return (
    <AuthLayout>
      <RegisterForm />
    </AuthLayout>
  )
}
