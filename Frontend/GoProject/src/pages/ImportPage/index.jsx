import { useState, useRef } from 'react'
import { useMutation } from '@tanstack/react-query'
import { importUsers } from '../../shared/services/usersApi.js'
import { getApiErrorMessage } from '../../shared/services/apiError.js'
import { Card } from '../../shared/components/Card.jsx'
import { Button } from '../../shared/components/Button.jsx'
import { Toast } from '../../shared/components/Toast.jsx'

export default function ImportPage() {
  const [file, setFile] = useState(null)
  const [result, setResult] = useState(null)
  const [toast, setToast] = useState(null)
  const fileRef = useRef()

  const mutation = useMutation({
    mutationFn: () => importUsers(file),
    onSuccess: (data) => {
      setResult(data)
      const type = data.failed === 0 ? 'success' : data.succeeded === 0 ? 'error' : 'warning'
      const message =
        data.failed === 0
          ? `Import hoàn tất: ${data.succeeded} người dùng được tạo thành công.`
          : `Import hoàn tất: ${data.succeeded} thành công, ${data.failed} thất bại.`
      setToast({ type, message })
    },
    onError: (err) => {
      setResult(null)
      setToast({ type: 'error', message: getApiErrorMessage(err) })
    },
  })

  return (
    <div className="max-w-2xl mx-auto p-6">
      <Toast
        message={toast?.message}
        type={toast?.type}
        onClose={() => setToast(null)}
      />

      <h1 className="text-2xl font-bold text-slate-800 mb-6">Import Users từ CSV</h1>
      <Card className="p-6 mb-6">
        <p className="text-sm text-slate-500 mb-4">
          File CSV phải có định dạng: <code className="bg-slate-100 px-1 rounded">username,email,password,role</code><br/>
          Role hợp lệ: <code className="bg-slate-100 px-1 rounded">member</code> hoặc <code className="bg-slate-100 px-1 rounded">manager</code>. Nếu không có role, mặc định là <code className="bg-slate-100 px-1 rounded">member</code>.
        </p>
        <div className="flex items-center gap-3">
          <input
            ref={fileRef}
            type="file"
            accept=".csv"
            onChange={(e) => { setFile(e.target.files[0]); setResult(null) }}
            className="block text-sm text-slate-500 file:mr-3 file:py-2 file:px-4 file:rounded file:border-0 file:text-sm file:font-medium file:bg-sky-50 file:text-sky-700 hover:file:bg-sky-100"
          />
          <Button
            onClick={() => mutation.mutate()}
            isLoading={mutation.isPending}
            disabled={!file || mutation.isPending}
          >
            Import
          </Button>
        </div>
      </Card>

      {result && (
        <Card className="p-6">
          <div className="flex gap-6 mb-4">
            <div className="text-center">
              <p className="text-3xl font-bold text-emerald-600">{result.succeeded}</p>
              <p className="text-sm text-slate-500">Thành công</p>
            </div>
            <div className="text-center">
              <p className="text-3xl font-bold text-red-500">{result.failed}</p>
              <p className="text-sm text-slate-500">Thất bại</p>
            </div>
          </div>
          {result.errors?.length > 0 && (
            <div>
              <p className="text-sm font-medium text-slate-700 mb-2">Chi tiết lỗi:</p>
              <div className="space-y-1 max-h-64 overflow-y-auto">
                {result.errors.map((e, i) => (
                  <div key={i} className="text-sm text-red-600 bg-red-50 px-3 py-1 rounded">
                    {e.email}: {e.error}
                  </div>
                ))}
              </div>
            </div>
          )}
        </Card>
      )}
    </div>
  )
}
