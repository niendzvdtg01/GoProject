import { useState } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import {
  getFolders, createFolder, updateFolder, deleteFolder,
  getFolderNotes, createNote
} from '../../shared/services/foldersApi.js'
import { updateNote, deleteNote } from '../../shared/services/notesApi.js'
import { shareAsset, revokeAccess } from '../../shared/services/sharingApi.js'
import { getApiErrorMessage } from '../../shared/services/apiError.js'
import { Button } from '../../shared/components/Button.jsx'
import { Card } from '../../shared/components/Card.jsx'

export default function TeamWorkspace() {
  const { teamName } = useParams()
  const navigate = useNavigate()
  const qc = useQueryClient()

  // Active tab: 'assets' | 'share'
  const [tab, setTab] = useState('assets')

  // Folder UI state
  const [expandedFolder, setExpandedFolder] = useState(null)
  const [showNewFolder, setShowNewFolder] = useState(false)
  const [newFolderName, setNewFolderName] = useState('')
  const [editingFolder, setEditingFolder] = useState(null) // { id, name }
  const [editFolderName, setEditFolderName] = useState('')

  // Note UI state
  const [showNewNote, setShowNewNote] = useState(null) // folderId or null
  const [newNote, setNewNote] = useState({ title: '', content: '' })
  const [editingNote, setEditingNote] = useState(null) // { id, title, content }

  // Share UI state
  const [shareForm, setShareForm] = useState({ email: '', assetType: 'folder', assetId: '', permissionType: 'read' })
  const [revokeForm, setRevokeForm] = useState({ email: '', assetType: 'folder', assetId: '' })
  const [shareMsg, setShareMsg] = useState('')
  const [revokeMsg, setRevokeMsg] = useState('')

  // Queries
  const { data: folders = [], isLoading: foldersLoading } = useQuery({
    queryKey: ['folders'],
    queryFn: getFolders,
    select: d => d.folders ?? d,
  })

  const { data: folderNotes = [] } = useQuery({
    queryKey: ['notes', expandedFolder],
    queryFn: () => getFolderNotes(expandedFolder),
    enabled: expandedFolder !== null,
    select: d => d.notes ?? d,
  })

  const invalidateFolders = () => qc.invalidateQueries({ queryKey: ['folders'] })
  const invalidateNotes = (fid) => qc.invalidateQueries({ queryKey: ['notes', fid] })

  // Folder mutations
  const createFolderMut = useMutation({
    mutationFn: () => createFolder(newFolderName),
    onSuccess: () => { invalidateFolders(); setShowNewFolder(false); setNewFolderName('') },
  })
  const updateFolderMut = useMutation({
    mutationFn: () => updateFolder(editingFolder.id, editFolderName),
    onSuccess: () => { invalidateFolders(); setEditingFolder(null) },
  })
  const deleteFolderMut = useMutation({
    mutationFn: (id) => deleteFolder(id),
    onSuccess: () => { invalidateFolders(); if (expandedFolder === deleteFolderMut.variables) setExpandedFolder(null) },
  })

  // Note mutations
  const createNoteMut = useMutation({
    mutationFn: () => createNote(showNewNote, newNote.title, newNote.content),
    onSuccess: () => { invalidateNotes(showNewNote); setShowNewNote(null); setNewNote({ title: '', content: '' }) },
  })
  const updateNoteMut = useMutation({
    mutationFn: () => updateNote(editingNote.id, editingNote.title, editingNote.content),
    onSuccess: () => { invalidateNotes(expandedFolder); setEditingNote(null) },
  })
  const deleteNoteMut = useMutation({
    mutationFn: (id) => deleteNote(id),
    onSuccess: () => invalidateNotes(expandedFolder),
  })

  // Share mutations
  const shareMut = useMutation({
    mutationFn: () => {
      const payload = { email: shareForm.email, permission_type: shareForm.permissionType }
      if (shareForm.assetType === 'folder') payload.folder_id = Number(shareForm.assetId)
      else payload.note_id = Number(shareForm.assetId)
      return shareAsset(payload)
    },
    onSuccess: () => setShareMsg('Chia sẻ thành công!'),
    onError: (e) => setShareMsg(getApiErrorMessage(e)),
  })
  const revokeMut = useMutation({
    mutationFn: () => {
      const payload = { email: revokeForm.email }
      if (revokeForm.assetType === 'folder') payload.folder_id = Number(revokeForm.assetId)
      else payload.note_id = Number(revokeForm.assetId)
      return revokeAccess(payload)
    },
    onSuccess: () => setRevokeMsg('Đã thu hồi quyền truy cập!'),
    onError: (e) => setRevokeMsg(getApiErrorMessage(e)),
  })

  return (
    <div className="p-6 max-w-4xl mx-auto">
      {/* Header */}
      <div className="flex items-center gap-3 mb-6">
        <button onClick={() => navigate('/dashboard')} className="text-slate-400 hover:text-slate-600 text-sm">
          ← Dashboard
        </button>
        <span className="text-slate-300">/</span>
        <h1 className="text-xl font-bold text-slate-800">{teamName}</h1>
      </div>

      {/* Tabs */}
      <div className="flex gap-1 mb-6 border-b border-slate-200">
        {[['assets', 'Folders & Notes'], ['share', 'Chia sẻ']].map(([key, label]) => (
          <button
            key={key}
            onClick={() => setTab(key)}
            className={`px-4 py-2 text-sm font-medium border-b-2 -mb-px transition-colors ${
              tab === key ? 'border-sky-500 text-sky-600' : 'border-transparent text-slate-500 hover:text-slate-700'
            }`}
          >
            {label}
          </button>
        ))}
      </div>

      {/* ASSETS TAB */}
      {tab === 'assets' && (
        <div>
          <div className="flex justify-between items-center mb-4">
            <h2 className="text-base font-semibold text-slate-700">Folders của bạn</h2>
            <Button variant="primary" onClick={() => setShowNewFolder(true)} className="text-sm py-1 px-3">
              + New Folder
            </Button>
          </div>

          {/* New Folder Form */}
          {showNewFolder && (
            <Card className="p-4 mb-4 flex gap-2 items-center">
              <input
                autoFocus
                className="flex-1 border border-slate-300 rounded px-3 py-1.5 text-sm outline-none focus:ring-2 focus:ring-sky-400"
                placeholder="Tên folder..."
                value={newFolderName}
                onChange={e => setNewFolderName(e.target.value)}
                onKeyDown={e => e.key === 'Enter' && createFolderMut.mutate()}
              />
              <Button onClick={() => createFolderMut.mutate()} isLoading={createFolderMut.isPending} className="text-sm py-1 px-3">Tạo</Button>
              <Button variant="ghost" onClick={() => { setShowNewFolder(false); setNewFolderName('') }} className="text-sm py-1 px-3">Huỷ</Button>
            </Card>
          )}

          {foldersLoading && <p className="text-slate-400 text-sm">Đang tải...</p>}
          {!foldersLoading && folders.length === 0 && (
            <p className="text-slate-400 text-sm text-center py-8">Chưa có folder nào. Tạo folder đầu tiên!</p>
          )}

          {/* Folder List */}
          <div className="space-y-3">
            {folders.map(folder => (
              <Card key={folder.id} className="overflow-hidden">
                <div className="flex items-center gap-3 p-4">
                  <button
                    onClick={() => setExpandedFolder(expandedFolder === folder.id ? null : folder.id)}
                    className="flex-1 text-left flex items-center gap-2"
                  >
                    <span className="text-slate-400">{expandedFolder === folder.id ? '▾' : '▸'}</span>
                    {editingFolder?.id === folder.id ? (
                      <input
                        autoFocus
                        className="border border-slate-300 rounded px-2 py-0.5 text-sm outline-none focus:ring-2 focus:ring-sky-400"
                        value={editFolderName}
                        onChange={e => setEditFolderName(e.target.value)}
                        onKeyDown={e => e.key === 'Enter' && updateFolderMut.mutate()}
                        onClick={e => e.stopPropagation()}
                      />
                    ) : (
                      <span className="font-medium text-slate-700">{folder.name}</span>
                    )}
                  </button>
                  <div className="flex gap-1">
                    {editingFolder?.id === folder.id ? (
                      <>
                        <Button variant="ghost" className="text-xs py-0.5 px-2" onClick={() => updateFolderMut.mutate()} isLoading={updateFolderMut.isPending}>Lưu</Button>
                        <Button variant="ghost" className="text-xs py-0.5 px-2" onClick={() => setEditingFolder(null)}>Huỷ</Button>
                      </>
                    ) : (
                      <>
                        <button onClick={() => { setEditingFolder(folder); setEditFolderName(folder.name) }} className="text-xs text-slate-400 hover:text-sky-500 px-2 py-0.5">Sửa</button>
                        <button onClick={() => deleteFolderMut.mutate(folder.id)} className="text-xs text-slate-400 hover:text-red-500 px-2 py-0.5">Xoá</button>
                      </>
                    )}
                  </div>
                </div>

                {/* Notes inside folder */}
                {expandedFolder === folder.id && (
                  <div className="border-t border-slate-100 bg-slate-50 px-4 py-3">
                    <div className="flex justify-between items-center mb-2">
                      <span className="text-xs text-slate-400 uppercase tracking-wider">Ghi chú</span>
                      <button onClick={() => { setShowNewNote(folder.id); setNewNote({ title: '', content: '' }) }} className="text-xs text-sky-600 hover:text-sky-700">+ Thêm ghi chú</button>
                    </div>

                    {/* New Note Form */}
                    {showNewNote === folder.id && (
                      <div className="bg-white rounded border border-slate-200 p-3 mb-3 space-y-2">
                        <input
                          autoFocus
                          className="w-full border border-slate-300 rounded px-2 py-1 text-sm outline-none focus:ring-2 focus:ring-sky-400"
                          placeholder="Tiêu đề..."
                          value={newNote.title}
                          onChange={e => setNewNote(n => ({ ...n, title: e.target.value }))}
                        />
                        <textarea
                          className="w-full border border-slate-300 rounded px-2 py-1 text-sm outline-none focus:ring-2 focus:ring-sky-400 resize-none"
                          placeholder="Nội dung..."
                          rows={3}
                          value={newNote.content}
                          onChange={e => setNewNote(n => ({ ...n, content: e.target.value }))}
                        />
                        <div className="flex gap-2">
                          <Button className="text-xs py-1 px-3" onClick={() => createNoteMut.mutate()} isLoading={createNoteMut.isPending}>Lưu</Button>
                          <Button variant="ghost" className="text-xs py-1 px-3" onClick={() => setShowNewNote(null)}>Huỷ</Button>
                        </div>
                      </div>
                    )}

                    {/* Note List */}
                    {folderNotes.length === 0 && showNewNote !== folder.id && (
                      <p className="text-xs text-slate-400 py-2">Chưa có ghi chú nào.</p>
                    )}
                    <div className="space-y-2">
                      {folderNotes.map(note => (
                        <div key={note.id} className="bg-white rounded border border-slate-200 p-3">
                          {editingNote?.id === note.id ? (
                            <div className="space-y-2">
                              <input
                                autoFocus
                                className="w-full border border-slate-300 rounded px-2 py-1 text-sm outline-none focus:ring-2 focus:ring-sky-400"
                                value={editingNote.title}
                                onChange={e => setEditingNote(n => ({ ...n, title: e.target.value }))}
                              />
                              <textarea
                                className="w-full border border-slate-300 rounded px-2 py-1 text-sm outline-none focus:ring-2 focus:ring-sky-400 resize-none"
                                rows={3}
                                value={editingNote.content}
                                onChange={e => setEditingNote(n => ({ ...n, content: e.target.value }))}
                              />
                              <div className="flex gap-2">
                                <Button className="text-xs py-1 px-3" onClick={() => updateNoteMut.mutate()} isLoading={updateNoteMut.isPending}>Lưu</Button>
                                <Button variant="ghost" className="text-xs py-1 px-3" onClick={() => setEditingNote(null)}>Huỷ</Button>
                              </div>
                            </div>
                          ) : (
                            <div className="flex justify-between items-start">
                              <div>
                                <p className="text-sm font-medium text-slate-700">{note.title}</p>
                                <p className="text-xs text-slate-500 mt-0.5 line-clamp-2">{note.content}</p>
                              </div>
                              <div className="flex gap-1 ml-3 shrink-0">
                                <button onClick={() => setEditingNote(note)} className="text-xs text-slate-400 hover:text-sky-500">Sửa</button>
                                <button onClick={() => deleteNoteMut.mutate(note.id)} className="text-xs text-slate-400 hover:text-red-500">Xoá</button>
                              </div>
                            </div>
                          )}
                        </div>
                      ))}
                    </div>
                  </div>
                )}
              </Card>
            ))}
          </div>
        </div>
      )}

      {/* SHARE TAB */}
      {tab === 'share' && (
        <div className="grid md:grid-cols-2 gap-6">
          {/* Share form */}
          <Card className="p-5">
            <h3 className="font-semibold text-slate-700 mb-4">Chia sẻ tài nguyên</h3>
            <div className="space-y-3">
              <div>
                <label className="text-xs text-slate-500 mb-1 block">Email người nhận</label>
                <input className="w-full border border-slate-300 rounded px-3 py-1.5 text-sm outline-none focus:ring-2 focus:ring-sky-400" value={shareForm.email} onChange={e => setShareForm(f => ({ ...f, email: e.target.value }))} placeholder="user@example.com" />
              </div>
              <div>
                <label className="text-xs text-slate-500 mb-1 block">Loại tài nguyên</label>
                <select className="w-full border border-slate-300 rounded px-3 py-1.5 text-sm outline-none focus:ring-2 focus:ring-sky-400" value={shareForm.assetType} onChange={e => setShareForm(f => ({ ...f, assetType: e.target.value, assetId: '' }))}>
                  <option value="folder">Folder</option>
                  <option value="note">Note</option>
                </select>
              </div>
              <div>
                <label className="text-xs text-slate-500 mb-1 block">ID tài nguyên</label>
                <select className="w-full border border-slate-300 rounded px-3 py-1.5 text-sm outline-none focus:ring-2 focus:ring-sky-400" value={shareForm.assetId} onChange={e => setShareForm(f => ({ ...f, assetId: e.target.value }))}>
                  <option value="">-- Chọn --</option>
                  {shareForm.assetType === 'folder'
                    ? folders.map(f => <option key={f.id} value={f.id}>{f.name} (#{f.id})</option>)
                    : folderNotes.map(n => <option key={n.id} value={n.id}>{n.title} (#{n.id})</option>)
                  }
                </select>
              </div>
              <div>
                <label className="text-xs text-slate-500 mb-1 block">Quyền truy cập</label>
                <select className="w-full border border-slate-300 rounded px-3 py-1.5 text-sm outline-none focus:ring-2 focus:ring-sky-400" value={shareForm.permissionType} onChange={e => setShareForm(f => ({ ...f, permissionType: e.target.value }))}>
                  <option value="read">Chỉ đọc (Read)</option>
                  <option value="write">Đọc &amp; Ghi (Write)</option>
                </select>
              </div>
              <Button onClick={() => { setShareMsg(''); shareMut.mutate() }} isLoading={shareMut.isPending} disabled={!shareForm.email || !shareForm.assetId}>
                Chia sẻ
              </Button>
              {shareMsg && <p className={`text-sm ${shareMsg.includes('thành công') ? 'text-emerald-600' : 'text-red-500'}`}>{shareMsg}</p>}
            </div>
          </Card>

          {/* Revoke form */}
          <Card className="p-5">
            <h3 className="font-semibold text-slate-700 mb-4">Thu hồi quyền truy cập</h3>
            <div className="space-y-3">
              <div>
                <label className="text-xs text-slate-500 mb-1 block">Email người dùng</label>
                <input className="w-full border border-slate-300 rounded px-3 py-1.5 text-sm outline-none focus:ring-2 focus:ring-sky-400" value={revokeForm.email} onChange={e => setRevokeForm(f => ({ ...f, email: e.target.value }))} placeholder="user@example.com" />
              </div>
              <div>
                <label className="text-xs text-slate-500 mb-1 block">Loại tài nguyên</label>
                <select className="w-full border border-slate-300 rounded px-3 py-1.5 text-sm outline-none focus:ring-2 focus:ring-sky-400" value={revokeForm.assetType} onChange={e => setRevokeForm(f => ({ ...f, assetType: e.target.value, assetId: '' }))}>
                  <option value="folder">Folder</option>
                  <option value="note">Note</option>
                </select>
              </div>
              <div>
                <label className="text-xs text-slate-500 mb-1 block">ID tài nguyên</label>
                <select className="w-full border border-slate-300 rounded px-3 py-1.5 text-sm outline-none focus:ring-2 focus:ring-sky-400" value={revokeForm.assetId} onChange={e => setRevokeForm(f => ({ ...f, assetId: e.target.value }))}>
                  <option value="">-- Chọn --</option>
                  {revokeForm.assetType === 'folder'
                    ? folders.map(f => <option key={f.id} value={f.id}>{f.name} (#{f.id})</option>)
                    : folderNotes.map(n => <option key={n.id} value={n.id}>{n.title} (#{n.id})</option>)
                  }
                </select>
              </div>
              <Button variant="danger" onClick={() => { setRevokeMsg(''); revokeMut.mutate() }} isLoading={revokeMut.isPending} disabled={!revokeForm.email || !revokeForm.assetId}>
                Thu hồi
              </Button>
              {revokeMsg && <p className={`text-sm ${revokeMsg.includes('thành công') ? 'text-emerald-600' : 'text-red-500'}`}>{revokeMsg}</p>}
            </div>
          </Card>
        </div>
      )}
    </div>
  )
}
