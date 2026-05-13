import { useNavigate } from 'react-router-dom'
import { useTeams } from '../../shared/hooks/useTeams.js'
import { useAuthStore } from '../../stores/authStore.js'
import { Card } from '../../shared/components/Card.jsx'
import { LoadingSkeleton } from '../../shared/components/LoadingSkeleton.jsx'
import { EmptyState } from '../../shared/components/EmptyState.jsx'

export function DashboardPage() {
  const { data: teams = [], isLoading, error } = useTeams()
  const user = useAuthStore(s => s.user)
  const navigate = useNavigate()

  return (
    <div className="p-6">
      <div className="mb-6">
        <h1 className="text-2xl font-bold text-slate-800">Xin chào, {user?.username} 👋</h1>
        <p className="text-slate-500 text-sm mt-1">Chọn một team để bắt đầu làm việc</p>
      </div>

      {isLoading && <LoadingSkeleton rows={3} />}
      {error && <p className="text-red-500 text-sm">Không thể tải danh sách team.</p>}

      {!isLoading && teams.length === 0 && (
        <EmptyState
          title="Bạn chưa thuộc team nào"
          description="Liên hệ manager để được thêm vào team, hoặc tạo team mới."
        />
      )}

      {!isLoading && teams.length > 0 && (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          {teams.map(team => (
            <button
              key={team.team_id}
              onClick={() => navigate(`/teams/${encodeURIComponent(team.team_name)}`)}
              className="text-left"
            >
              <Card className="p-5 hover:border-sky-400 hover:shadow-md transition-all cursor-pointer border border-transparent">
                <div className="flex items-start justify-between">
                  <div>
                    <h3 className="font-semibold text-slate-800 text-base">{team.team_name}</h3>
                    <p className="text-xs text-slate-400 mt-1">
                      Tạo lúc: {new Date(team.created_at).toLocaleDateString('vi-VN')}
                    </p>
                  </div>
                  <span className="text-2xl">👥</span>
                </div>
                <p className="text-xs text-sky-600 mt-3 font-medium">Mở workspace →</p>
              </Card>
            </button>
          ))}
        </div>
      )}
    </div>
  )
}
