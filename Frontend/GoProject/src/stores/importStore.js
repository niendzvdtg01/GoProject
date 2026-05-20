import { create } from 'zustand'
import { persist } from 'zustand/middleware'

export const useImportStore = create(
  persist(
    (set) => ({
      taskIds: [],

      addTaskIds(newIds) {
        set((state) => ({ taskIds: [...newIds, ...state.taskIds] }))
      },
    }),
    {
      name: 'import-tasks',
      storage: {
        getItem: (key) => {
          const raw = sessionStorage.getItem(key)
          return raw ? JSON.parse(raw) : null
        },
        setItem: (key, value) => sessionStorage.setItem(key, JSON.stringify(value)),
        removeItem: (key) => sessionStorage.removeItem(key),
      },
    },
  ),
)
