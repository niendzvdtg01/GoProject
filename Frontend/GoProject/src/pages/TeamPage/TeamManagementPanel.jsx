import { useState } from 'react'
import { useMutation } from '@tanstack/react-query'
import { useAuthStore } from '../../stores/authStore.js'
import { Button } from '../../shared/components/Button.jsx'
import { Card } from '../../shared/components/Card.jsx'
import { Input } from '../../shared/components/Input.jsx'
import { Select } from '../../shared/components/Select.jsx'
import { getApiErrorMessage } from '../../shared/services/apiError.js'
import { createTeam, addTeamMember, removeTeamMember, deleteTeam } from '../../shared/services/teamsApi.js'

const DEFAULT_CREATE_MEMBER = () => ({ memberName: '', role: 'member' })

export function TeamManagementPanel() {
  const accessToken = useAuthStore((state) => state.accessToken)
  const [createName, setCreateName] = useState('')
  const [createMembers, setCreateMembers] = useState([DEFAULT_CREATE_MEMBER()])
  const [memberTeamName, setMemberTeamName] = useState('')
  const [memberName, setMemberName] = useState('')
  const [memberRole, setMemberRole] = useState('member')
  const [removeTeamName, setRemoveTeamName] = useState('')
  const [removeMemberName, setRemoveMemberName] = useState('')
  const [deleteTeamName, setDeleteTeamName] = useState('')
  const [statusMessage, setStatusMessage] = useState('')
  const [errorMessage, setErrorMessage] = useState('')

  const clearMessages = () => {
    setStatusMessage('')
    setErrorMessage('')
  }

  const createTeamMutation = useMutation({
    mutationFn: ({ teamName, members }) => createTeam(teamName, members),
    onSuccess() {
      setStatusMessage('Tạo đội thành công.')
      setErrorMessage('')
      setCreateName('')
      setCreateMembers([DEFAULT_CREATE_MEMBER()])
    },
    onError(error) {
      setErrorMessage(getApiErrorMessage(error))
    },
  })

  const addMemberMutation = useMutation({
    mutationFn: ({ teamName, memberName, role }) => addTeamMember(teamName, memberName, role),
    onSuccess() {
      setStatusMessage('Thêm thành viên vào đội thành công.')
      setErrorMessage('')
      setMemberName('')
      setMemberRole('member')
    },
    onError(error) {
      setErrorMessage(getApiErrorMessage(error))
    },
  })

  const removeMemberMutation = useMutation({
    mutationFn: ({ teamName, memberName }) => removeTeamMember(teamName, memberName),
    onSuccess() {
      setStatusMessage('Xóa thành viên khỏi đội thành công.')
      setErrorMessage('')
      setRemoveMemberName('')
    },
    onError(error) {
      setErrorMessage(getApiErrorMessage(error))
    },
  })

  const deleteTeamMutation = useMutation({
    mutationFn: (teamName) => deleteTeam(teamName),
    onSuccess() {
      setStatusMessage('Xóa đội thành công.')
      setErrorMessage('')
      setDeleteTeamName('')
    },
    onError(error) {
      setErrorMessage(getApiErrorMessage(error))
    },
  })

  return (
    <div className="grid gap-4">
      <Card className="p-5">
        <div className="flex flex-col gap-3 md:flex-row md:items-center md:justify-between">
          <div>
            <h2 className="text-lg font-extrabold text-slate-950">Team management</h2>
            <p className="mt-1 text-sm text-slate-600">
              Sử dụng các endpoint /api/teams đã được backend triển khai để tạo, thêm và xóa thành viên.
            </p>
          </div>
          <Button type="button" variant="secondary" onClick={clearMessages}>
            Làm mới trạng thái
          </Button>
        </div>
      </Card>

      {statusMessage || errorMessage ? (
        <Card className="p-5">
          {statusMessage ? (
            <p className="rounded-md border border-emerald-200 bg-emerald-50 px-4 py-3 text-sm font-semibold text-emerald-800">
              {statusMessage}
            </p>
          ) : null}
          {errorMessage ? (
            <p className="rounded-md border border-red-200 bg-red-50 px-4 py-3 text-sm font-semibold text-red-700">
              {errorMessage}
            </p>
          ) : null}
        </Card>
      ) : null}

      {!accessToken ? (
        <Card className="p-5">
          <p className="text-sm font-semibold text-slate-700">
            Bạn cần đăng nhập để sử dụng chức năng quản lý đội. Nếu đã đăng nhập, vui lòng đăng xuất rồi đăng nhập lại để làm mới token.
          </p>
        </Card>
      ) : null}

      {accessToken ? (
        <>
          <Card className="p-5">
            <h3 className="text-base font-extrabold text-slate-950">Tạo đội mới</h3>
        <p className="mt-1 text-sm text-slate-600">
          Nhập tên đội và thêm thành viên ngay lúc tạo. Chỉ role manager được phép dùng khi tạo team, backend sẽ tự động gán bạn là owner.
        </p>
        <form
          className="grid gap-4 pt-4"
          onSubmit={(event) => {
            event.preventDefault()
            const members = createMembers
              .filter((member) => member.memberName.trim())
              .map((member) => ({ userID: member.memberName, role: member.role }))
            createTeamMutation.mutate({ teamName: createName, members })
          }}
        >
          <Input
            label="Tên đội"
            value={createName}
            onChange={(event) => setCreateName(event.target.value)}
          />
          <div className="grid gap-3">
            {createMembers.map((member, index) => (
              <div key={index} className="grid gap-3 rounded-lg border border-slate-200 bg-slate-50 p-4 md:grid-cols-[1fr_auto] md:items-end">
                <div className="grid gap-3 md:grid-cols-[1fr_160px]">
                  <Input
                    label={`Tên thành viên ${index + 1}`}
                    value={member.memberName}
                    onChange={(event) => {
                      const newMembers = [...createMembers]
                      newMembers[index] = { ...member, memberName: event.target.value }
                      setCreateMembers(newMembers)
                    }}
                  />
                  <Select
                    label="Vai trò"
                    value={member.role}
                    onChange={(event) => {
                      const newMembers = [...createMembers]
                      newMembers[index] = { ...member, role: event.target.value }
                      setCreateMembers(newMembers)
                    }}
                  >
                    <option value="manager">Manager</option>
                    <option value="member">Member</option>
                  </Select>
                </div>
                <Button
                  type="button"
                  variant="secondary"
                  onClick={() => {
                    setCreateMembers(createMembers.filter((_, memberIndex) => memberIndex !== index))
                  }}
                >
                  Xóa
                </Button>
              </div>
            ))}
          </div>
          <Button
            type="button"
            variant="secondary"
            onClick={() => setCreateMembers([...createMembers, DEFAULT_CREATE_MEMBER()])}
          >
            Thêm thành viên khi tạo
          </Button>
          <Button type="submit" isLoading={createTeamMutation.isLoading}>
            Tạo đội
          </Button>
        </form>
      </Card>

      <Card className="p-5">
        <h3 className="text-base font-extrabold text-slate-950">Thêm thành viên vào đội</h3>
        <p className="mt-1 text-sm text-slate-600">Thêm thành viên vào đội chỉ cần biết tên đăng nhập; quyền global được quản lý ở backend.</p>
        <form
          className="grid gap-4 pt-4"
          onSubmit={(event) => {
            event.preventDefault()
            addMemberMutation.mutate({ teamName: memberTeamName, memberName, role: memberRole })
          }}
        >
          <Input
            label="Tên đội"
            value={memberTeamName}
            onChange={(event) => setMemberTeamName(event.target.value)}
          />
          <div className="grid gap-3 md:grid-cols-[1fr_160px]">
            <Input
              label="Tên thành viên"
              value={memberName}
              onChange={(event) => setMemberName(event.target.value)}
            />
            <Select
              label="Vai trò"
              value={memberRole}
              onChange={(event) => setMemberRole(event.target.value)}
            >
              <option value="member">Member</option>
              <option value="manager">Manager</option>
            </Select>
          </div>
          <Button type="submit" isLoading={addMemberMutation.isLoading}>
            Thêm thành viên
          </Button>
        </form>
      </Card>

      <Card className="p-5">
        <h3 className="text-base font-extrabold text-slate-950">Xóa thành viên khỏi đội</h3>
        <form
          className="grid gap-4 pt-4"
          onSubmit={(event) => {
            event.preventDefault()
            removeMemberMutation.mutate({ teamName: removeTeamName, memberName: removeMemberName })
          }}
        >
          <Input
            label="Tên đội"
            value={removeTeamName}
            onChange={(event) => setRemoveTeamName(event.target.value)}
          />
          <Input
            label="Tên thành viên"
            value={removeMemberName}
            onChange={(event) => setRemoveMemberName(event.target.value)}
          />
          <Button type="submit" isLoading={removeMemberMutation.isLoading} variant="danger">
            Xóa thành viên
          </Button>
        </form>
      </Card>

      <Card className="p-5">
        <h3 className="text-base font-extrabold text-slate-950">Xóa đội</h3>
        <form
          className="grid gap-4 pt-4"
          onSubmit={(event) => {
            event.preventDefault()
            deleteTeamMutation.mutate(deleteTeamName)
          }}
        >
          <Input
            label="Tên đội"
            value={deleteTeamName}
            onChange={(event) => setDeleteTeamName(event.target.value)}
          />
          <Button type="submit" isLoading={deleteTeamMutation.isLoading} variant="danger">
            Xóa đội
          </Button>
        </form>
      </Card>
    </>
      ) : null}
    </div>
  )
}
