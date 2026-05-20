import { useState, useRef } from 'react'
import { useMutation, useQuery } from '@tanstack/react-query'
import { startImport, getImportTask } from '../../shared/services/usersApi.js'
import { getApiErrorMessage } from '../../shared/services/apiError.js'
import { useImportStore } from '../../stores/importStore.js'
import { Card } from '../../shared/components/Card.jsx'
import { Button } from '../../shared/components/Button.jsx'
import { Toast } from '../../shared/components/Toast.jsx'

const STATUS_CONFIG = {
  pending:    { label: 'Đang chờ',    cls: 'bg-amber-100 text-amber-700' },
  processing: { label: 'Đang xử lý', cls: 'bg-sky-100 text-sky-700' },
  completed:  { label: 'Hoàn tất',   cls: 'bg-emerald-100 text-emerald-700' },
  failed:     { label: 'Thất bại',   cls: 'bg-red-100 text-red-700' },
}

function ImportTaskRow({ taskId }) {
  const [showErrors, setShowErrors] = useState(false)

  const { data: task } = useQuery({
    queryKey: ['import-task', taskId],
    queryFn: () => getImportTask(taskId),
    refetchInterval: (query) => {
      const status = query.state.data?.status
      if (status === 'completed' || status === 'failed') return false
      return 2000
    },
    staleTime: 0,
  })

  const sc = STATUS_CONFIG[task?.status] ?? STATUS_CONFIG.pending

  let errors = []
  if (task?.error_log) {
    try { errors = JSON.parse(task.error_log) } catch { /* raw string, ignore */ }
  }

  const progress =
    task?.status === 'processing' && task?.total_rows > 0
      ? Math.round((task.processed_rows / task.total_rows) * 100)
      : null

  return (
    <div className="p-4 bg-slate-50 rounded-lg border border-slate-200 space-y-2">
      <div className="flex items-center gap-3">
        <div className="flex-1 min-w-0">
          <p className="text-sm font-medium text-slate-800 truncate">
            {task?.file_name ?? `Task #${taskId}`}
          </p>
          {task?.status === 'completed' && (
            <p className="text-xs text-slate-500 mt-0.5">
              {task.succeeded} thành công · {task.failed} thất bại
            </p>
          )}
        </div>

        <div className="flex items-center gap-2 shrink-0">
          {(task?.status === 'pending' || task?.status === 'processing') && (
            <span className="inline-block w-3 h-3 rounded-full bg-sky-500 animate-pulse" />
          )}
          <span className={`text-xs font-semibold px-2.5 py-1 rounded-full ${sc.cls}`}>
            {sc.label}
          </span>
        </div>
      </div>

      {progress !== null && (
        <div>
          <div className="h-1.5 bg-slate-200 rounded-full overflow-hidden">
            <div
              className="h-full bg-sky-500 transition-all duration-500 rounded-full"
              style={{ width: `${progress}%` }}
            />
          </div>
          <p className="text-xs text-slate-500 mt-0.5">
            {task.processed_rows}/{task.total_rows} hàng ({progress}%)
          </p>
        </div>
      )}

      {task?.status === 'completed' && task?.failed > 0 && errors.length > 0 && (
        <div>
          <button
            onClick={() => setShowErrors(v => !v)}
            className="text-xs text-red-600 font-medium hover:underline"
          >
            {showErrors ? 'Ẩn lỗi' : `Xem ${errors.length} lỗi`}
          </button>
          {showErrors && (
            <div className="mt-1 max-h-32 overflow-y-auto space-y-1">
              {errors.map((e, i) => (
                <div key={i} className="text-xs text-red-600 bg-red-50 px-2 py-1 rounded">
                  {typeof e === 'object' ? `${e.email ?? e.row ?? i + 1}: ${e.error}` : e}
                </div>
              ))}
            </div>
          )}
        </div>
      )}
    </div>
  )
}

export default function ImportPage() {
  const [files, setFiles] = useState([])
  const [toast, setToast] = useState(null)
  const fileRef = useRef()

  const { taskIds, addTaskIds } = useImportStore()

  const mutation = useMutation({
    mutationFn: (fileList) =>
      Promise.all(Array.from(fileList).map(f => startImport(f))),
    onSuccess: (results) => {
      const newIds = results.map(r => r.task_id)
      addTaskIds(newIds)
      setFiles([])
      if (fileRef.current) fileRef.current.value = ''
      setToast({
        type: 'info',
        message: `Đã gửi ${newIds.length} file. Đang xử lý, vui lòng chờ...`,
      })
    },
    onError: (err) => {
      setToast({ type: 'error', message: getApiErrorMessage(err) })
    },
  })

  const handleFileChange = (e) => {
    setFiles(Array.from(e.target.files))
  }

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
          File CSV phải có định dạng:{' '}
          <code className="bg-slate-100 px-1 rounded">username,email,password,role</code>
          <br />
          Role hợp lệ:{' '}
          <code className="bg-slate-100 px-1 rounded">member</code> hoặc{' '}
          <code className="bg-slate-100 px-1 rounded">manager</code>. Có thể chọn nhiều file cùng lúc.
        </p>

        <div className="flex items-center gap-3">
          <input
            ref={fileRef}
            type="file"
            accept=".csv"
            multiple
            onChange={handleFileChange}
            className="block text-sm text-slate-500 file:mr-3 file:py-2 file:px-4 file:rounded file:border-0 file:text-sm file:font-medium file:bg-sky-50 file:text-sky-700 hover:file:bg-sky-100"
          />
          <Button
            onClick={() => mutation.mutate(files)}
            isLoading={mutation.isPending}
            disabled={files.length === 0 || mutation.isPending}
          >
            {files.length > 1 ? `Import ${files.length} file` : 'Import'}
          </Button>
        </div>

        {files.length > 0 && (
          <div className="mt-3 flex flex-wrap gap-1">
            {files.map((f, i) => (
              <span key={i} className="text-xs bg-sky-50 text-sky-700 px-2 py-0.5 rounded-full border border-sky-200">
                {f.name}
              </span>
            ))}
          </div>
        )}
      </Card>

      {taskIds.length > 0 && (
        <div>
          <h2 className="text-sm font-semibold text-slate-600 mb-3">
            Lịch sử import ({taskIds.length} task)
          </h2>
          <div className="space-y-3">
            {taskIds.map(id => (
              <ImportTaskRow key={id} taskId={id} />
            ))}
          </div>
        </div>
      )}
    </div>
  )
}
