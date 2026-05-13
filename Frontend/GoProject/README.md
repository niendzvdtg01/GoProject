# Frontend

React 19 SPA, built with Vite. Communicates with the Go backend at `http://localhost:8080/api`.

## Stack

| | |
|---|---|
| UI Framework | React 19 + Vite |
| Routing | React Router DOM v7 |
| Server State | TanStack Query v5 |
| Client State | Zustand v5 (persisted to localStorage) |
| Forms | React Hook Form + Zod |
| HTTP | Axios |
| Styling | Tailwind CSS v4 |

## Run

```bash
npm install
npm run dev      # http://localhost:5173
npm run build    # production build
```

Set `VITE_API_URL` in `.env` if the backend is not at `http://localhost:8080/api`.

## Project Structure

```
src/
├── App.jsx                        # Route definitions
├── main.jsx                       # App entry, QueryClient, BrowserRouter
│
├── pages/
│   ├── LoginPage/                 # Email + password login
│   ├── DashboardPage/             # Team cards — click to open workspace
│   ├── TeamPage/
│   │   ├── index.jsx              # Routes to Workspace or Management panel
│   │   ├── TeamManagementPanel.jsx # Create team, add/remove members
│   │   └── TeamWorkspace.jsx      # Folders, Notes, Share — per team
│   ├── ImportPage/                # CSV bulk user import (manager only)
│   ├── ProfilePage/               # Current user info
│   └── NotFoundPage/
│
├── shared/
│   ├── components/
│   │   ├── Button.jsx             # variants: primary, secondary, ghost, danger
│   │   ├── Card.jsx
│   │   ├── Input.jsx
│   │   ├── Select.jsx
│   │   ├── Table.jsx
│   │   ├── RoleBadge.jsx
│   │   ├── LoadingSkeleton.jsx
│   │   └── EmptyState.jsx
│   │
│   ├── services/                  # API client functions
│   │   ├── axios.js               # Axios instance (base URL, auth interceptor)
│   │   ├── interceptor.js         # 401 → auto logout
│   │   ├── authApi.js             # login, register, logout
│   │   ├── usersApi.js            # getUsers, importUsers
│   │   ├── teamsApi.js            # getTeams, createTeam, addMember, removeMember, deleteTeam
│   │   ├── foldersApi.js          # getFolders, createFolder, updateFolder, deleteFolder, getFolderNotes, createNote
│   │   ├── notesApi.js            # updateNote, deleteNote
│   │   ├── sharingApi.js          # shareAsset, revokeAccess
│   │   └── apiError.js            # extract error message from axios error
│   │
│   ├── hooks/
│   │   ├── useTeams.js            # TanStack Query — GET /teams
│   │   ├── useUsers.js            # TanStack Query — GET /users
│   │   └── useLogout.js           # POST /auth/logout + clear session
│   │
│   ├── layouts/
│   │   ├── DashboardLayout.jsx    # Sidebar nav + Outlet
│   │   └── AuthLayout.jsx         # Two-column auth screen
│   │
│   ├── utils/
│   │   ├── ProtectedRoute.jsx     # Redirect to /login if not authenticated
│   │   ├── PublicRoute.jsx        # Redirect to /dashboard if already authenticated
│   │   ├── storage.js             # localStorage helpers
│   │   ├── formatDate.js          # Vietnamese date formatting
│   │   └── user.js                # Normalize user object (camelCase / snake_case)
│   │
│   └── constants/
│       ├── routes.js              # ROUTES constant
│       └── roles.js               # USER_ROLES, ROLE_LABELS
│
└── stores/
    ├── authStore.js               # Zustand: accessToken + user (persisted)
    └── uiStore.js                 # Zustand: sidebar, theme (prepared)
```

## Pages & Routing

| Route | Component | Access |
|---|---|---|
| `/` | redirect → `/dashboard` | — |
| `/login` | LoginPage | Public (redirects if logged in) |
| `/dashboard` | DashboardPage | Auth required |
| `/teams` | TeamManagementPanel | Auth required |
| `/teams/:teamName` | TeamWorkspace | Auth required |
| `/import` | ImportPage | Manager only |
| `/profile` | ProfilePage | Auth required |
| `*` | NotFoundPage | — |

## Feature Walkthrough

### Dashboard
Shows a grid of team cards for all teams the current user belongs to (created or joined). Clicking a card opens the team workspace.

### Team Workspace (`/teams/:teamName`)
Two tabs:

**Folders & Notes**
- List all folders owned by the current user
- Create, rename, delete folders
- Expand a folder to see notes inside
- Create, edit, delete notes (title + content)

**Chia sẻ (Share)**
- Share a folder or note with another user by email (Read or Write permission)
- Folder sharing automatically shares all notes inside (inheritance)
- Revoke access from a user

### Team Management (`/teams`)
Manager-only operations:
- Create a new team with optional initial members
- Add a member to an existing team (by username + role)
- Remove a member
- Delete a team

### Import Users (`/import`) — Manager only
Upload a `.csv` file to batch-create users. The backend processes the file with a concurrent worker pool.

CSV format:
```
username,email,password,role
alice,alice@example.com,secret123,member
bob,bob@example.com,secret456,manager
```
`role` column is optional — defaults to `member`.

The result shows how many accounts were created successfully and lists any failures with their error messages.

## Auth Flow

1. User logs in at `/login` — backend returns a JWT
2. Token is stored in `localStorage` as `goproject.accessToken`
3. Every axios request adds `Authorization: Bearer <token>` via request interceptor
4. On 401 response, the interceptor clears the session and redirects to `/login`
5. Logout calls `POST /auth/logout` to revoke the token server-side, then clears local state
