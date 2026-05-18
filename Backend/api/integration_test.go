package routing_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"
	"sync"
	"testing"
	"time"

	routing "backend/api"
	"backend/internal/middleware"
	"backend/internal/repository"
	"backend/internal/service"
	"backend/package/utils"

	"github.com/gin-gonic/gin"
	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/bcrypt"
)

var registerOnce sync.Once

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	registerOnce.Do(func() { utils.RegisterValidators() })
	os.Exit(m.Run())
}

// --- test harness ---

type testServer struct {
	engine *gin.Engine
	mock   sqlmock.Sqlmock
	auth   *middleware.AuthMiddleware
}

func newTestServer(t *testing.T) (*testServer, func()) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}

	userRepo := repository.NewUserRepository(db)
	teamRepo := repository.NewTeamRepository(db)
	teamMemberRepo := repository.NewTeamMemberRepository(db)
	folderRepo := repository.NewFolderRepository(db)
	noteRepo := repository.NewNoteRepository(db)
	permRepo := repository.NewPermissionRepository(db)
	importTaskRepo := repository.NewImportTaskRepository(db)

	auth := middleware.NewAuthMiddleware("test-secret")
	authSvc := service.NewAuthService(userRepo, auth)
	userSvc := service.NewUserService(userRepo, auth, importTaskRepo)
	teamSvc := service.NewTeamManagementService(teamRepo, teamMemberRepo, userRepo)
	folderSvc := service.NewFolderService(folderRepo, userRepo, permRepo, teamMemberRepo)
	noteSvc := service.NewNoteService(noteRepo, folderRepo, userRepo, permRepo, teamMemberRepo)
	sharingSvc := service.NewSharing(noteRepo, folderRepo, userRepo, permRepo)

	engine := routing.SetupRouter(auth, authSvc, userSvc, userRepo, teamSvc, folderSvc, noteSvc, sharingSvc)

	ts := &testServer{engine: engine, mock: mock, auth: auth}
	return ts, func() { db.Close() }
}

func (ts *testServer) bearerToken(userID, username, role string) string {
	tok, _ := ts.auth.GenerateToken(userID, username, role)
	return "Bearer " + tok
}

func doRequest(t *testing.T, engine *gin.Engine, method, path string, body any, authHeader string) *httptest.ResponseRecorder {
	t.Helper()
	var b []byte
	if body != nil {
		var err error
		b, err = json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal body: %v", err)
		}
	}
	req := httptest.NewRequest(method, path, bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	if authHeader != "" {
		req.Header.Set("Authorization", authHeader)
	}
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)
	return w
}

// --- auth tests ---

func TestIntegration_Login_Success(t *testing.T) {
	ts, cleanup := newTestServer(t)
	defer cleanup()

	hash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.MinCost)
	ts.mock.ExpectQuery("SELECT user_id, username, email, password_hash").
		WithArgs("alice@example.com").
		WillReturnRows(sqlmock.NewRows([]string{"user_id", "username", "email", "password_hash", "role", "created_at"}).
			AddRow("uid-1", "alice", "alice@example.com", string(hash), "member", time.Now()))

	w := doRequest(t, ts.engine, http.MethodPost, "/api/auth/login",
		map[string]string{"email": "alice@example.com", "password": "password123"}, "")

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d — body: %s", w.Code, w.Body.String())
	}
	var resp map[string]any
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["token"] == "" || resp["token"] == nil {
		t.Error("expected token in response")
	}
}

func TestIntegration_Login_WrongPassword(t *testing.T) {
	ts, cleanup := newTestServer(t)
	defer cleanup()

	hash, _ := bcrypt.GenerateFromPassword([]byte("correct-password"), bcrypt.MinCost)
	ts.mock.ExpectQuery("SELECT user_id, username, email, password_hash").
		WithArgs("alice@example.com").
		WillReturnRows(sqlmock.NewRows([]string{"user_id", "username", "email", "password_hash", "role", "created_at"}).
			AddRow("uid-1", "alice", "alice@example.com", string(hash), "member", time.Now()))

	w := doRequest(t, ts.engine, http.MethodPost, "/api/auth/login",
		map[string]string{"email": "alice@example.com", "password": "wrong-password"}, "")

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestIntegration_Login_MissingFields(t *testing.T) {
	ts, cleanup := newTestServer(t)
	defer cleanup()

	w := doRequest(t, ts.engine, http.MethodPost, "/api/auth/login",
		map[string]string{"email": "alice@example.com"}, "")

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestIntegration_Logout_Success(t *testing.T) {
	ts, cleanup := newTestServer(t)
	defer cleanup()

	token := ts.bearerToken("uid-1", "alice", "member")
	w := doRequest(t, ts.engine, http.MethodPost, "/api/auth/logout", nil, token)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d — body: %s", w.Code, w.Body.String())
	}
}

func TestIntegration_AuthRequired_NoToken(t *testing.T) {
	ts, cleanup := newTestServer(t)
	defer cleanup()

	for _, path := range []string{"/api/users", "/api/teams", "/api/folders"} {
		w := doRequest(t, ts.engine, http.MethodGet, path, nil, "")
		if w.Code != http.StatusUnauthorized {
			t.Errorf("GET %s: expected 401, got %d", path, w.Code)
		}
	}
}

// --- user tests ---

func TestIntegration_Register_Success(t *testing.T) {
	ts, cleanup := newTestServer(t)
	defer cleanup()

	ts.mock.ExpectExec(regexp.QuoteMeta("INSERT INTO users (user_id, username, email, password_hash, role)")).
		WillReturnResult(sqlmock.NewResult(1, 1))

	w := doRequest(t, ts.engine, http.MethodPost, "/api/users/register",
		map[string]string{"username": "bob", "email": "bob@example.com", "password": "password123", "role": "member"}, "")

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d — body: %s", w.Code, w.Body.String())
	}
	var resp map[string]any
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["token"] == nil {
		t.Error("expected token in response")
	}
}

func TestIntegration_Register_DuplicateEmail(t *testing.T) {
	ts, cleanup := newTestServer(t)
	defer cleanup()

	ts.mock.ExpectExec(regexp.QuoteMeta("INSERT INTO users (user_id, username, email, password_hash, role)")).
		WillReturnError(&mysql.MySQLError{Number: 1062, Message: "Duplicate entry"})

	w := doRequest(t, ts.engine, http.MethodPost, "/api/users/register",
		map[string]string{"username": "bob", "email": "bob@example.com", "password": "password123", "role": "member"}, "")

	if w.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d", w.Code)
	}
}

func TestIntegration_Register_InvalidRole(t *testing.T) {
	ts, cleanup := newTestServer(t)
	defer cleanup()

	// DTO validates role via oneof=manager member, so invalid role fails at binding
	w := doRequest(t, ts.engine, http.MethodPost, "/api/users/register",
		map[string]string{"username": "bob", "email": "bob@example.com", "password": "password123", "role": "superadmin"}, "")

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestIntegration_ListUsers_Success(t *testing.T) {
	ts, cleanup := newTestServer(t)
	defer cleanup()

	now := time.Now()
	ts.mock.ExpectQuery("SELECT user_id, username, email, role, created_at").
		WillReturnRows(sqlmock.NewRows([]string{"user_id", "username", "email", "role", "created_at"}).
			AddRow("uid-1", "alice", "alice@example.com", "member", now).
			AddRow("uid-2", "bob", "bob@example.com", "manager", now))

	token := ts.bearerToken("uid-1", "alice", "member")
	w := doRequest(t, ts.engine, http.MethodGet, "/api/users", nil, token)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d — body: %s", w.Code, w.Body.String())
	}
	var resp map[string]any
	json.Unmarshal(w.Body.Bytes(), &resp)
	users, ok := resp["users"].([]any)
	if !ok || len(users) != 2 {
		t.Errorf("expected 2 users in response, got: %v", resp["users"])
	}
}

// --- team tests ---

func TestIntegration_CreateTeam_AsManager(t *testing.T) {
	ts, cleanup := newTestServer(t)
	defer cleanup()

	ts.mock.ExpectExec("INSERT INTO teams").
		WithArgs("alpha").
		WillReturnResult(sqlmock.NewResult(1, 1))
	ts.mock.ExpectExec("INSERT INTO team_members").
		WithArgs(1, "mgr-1", "OWNER").
		WillReturnResult(sqlmock.NewResult(1, 1))

	token := ts.bearerToken("mgr-1", "manager-user", "manager")
	w := doRequest(t, ts.engine, http.MethodPost, "/api/teams",
		map[string]any{"teamName": "alpha", "members": []any{}}, token)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d — body: %s", w.Code, w.Body.String())
	}
	var resp map[string]any
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["teamID"] == nil {
		t.Error("expected teamID in response")
	}
}

func TestIntegration_CreateTeam_AsMember_Forbidden(t *testing.T) {
	ts, cleanup := newTestServer(t)
	defer cleanup()

	token := ts.bearerToken("uid-1", "alice", "member")
	w := doRequest(t, ts.engine, http.MethodPost, "/api/teams",
		map[string]any{"teamName": "alpha"}, token)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", w.Code)
	}
}

func TestIntegration_ListTeams_Success(t *testing.T) {
	ts, cleanup := newTestServer(t)
	defer cleanup()

	now := time.Now()
	ts.mock.ExpectQuery("SELECT t.team_id, t.team_name").
		WithArgs("uid-1").
		WillReturnRows(sqlmock.NewRows([]string{"team_id", "team_name", "created_at", "updated_at"}).
			AddRow(1, "alpha", now, now))

	token := ts.bearerToken("uid-1", "alice", "member")
	w := doRequest(t, ts.engine, http.MethodGet, "/api/teams", nil, token)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d — body: %s", w.Code, w.Body.String())
	}
	var resp map[string]any
	json.Unmarshal(w.Body.Bytes(), &resp)
	teams, ok := resp["teams"].([]any)
	if !ok || len(teams) != 1 {
		t.Errorf("expected 1 team, got: %v", resp["teams"])
	}
}

func TestIntegration_DeleteTeam_AsMember_Forbidden(t *testing.T) {
	ts, cleanup := newTestServer(t)
	defer cleanup()

	token := ts.bearerToken("uid-1", "alice", "member")
	w := doRequest(t, ts.engine, http.MethodDelete, "/api/teams/alpha", nil, token)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", w.Code)
	}
}

// --- folder tests ---

func TestIntegration_CreateFolder_Success(t *testing.T) {
	ts, cleanup := newTestServer(t)
	defer cleanup()

	now := time.Now()
	ts.mock.ExpectExec("INSERT INTO folders").
		WithArgs("uid-1", "Work").
		WillReturnResult(sqlmock.NewResult(5, 1))
	ts.mock.ExpectQuery("SELECT folder_id, owner_id, name, created_at, updated_at").
		WithArgs(5).
		WillReturnRows(sqlmock.NewRows([]string{"folder_id", "owner_id", "name", "created_at", "updated_at"}).
			AddRow(5, "uid-1", "Work", now, now))

	token := ts.bearerToken("uid-1", "alice", "member")
	w := doRequest(t, ts.engine, http.MethodPost, "/api/folders",
		map[string]string{"name": "Work"}, token)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d — body: %s", w.Code, w.Body.String())
	}
}

func TestIntegration_ListFolders_Success(t *testing.T) {
	ts, cleanup := newTestServer(t)
	defer cleanup()

	now := time.Now()
	ts.mock.ExpectQuery("SELECT folder_id, owner_id, name, created_at, updated_at").
		WithArgs("uid-1").
		WillReturnRows(sqlmock.NewRows([]string{"folder_id", "owner_id", "name", "created_at", "updated_at"}).
			AddRow(1, "uid-1", "Work", now, now).
			AddRow(2, "uid-1", "Personal", now, now))

	token := ts.bearerToken("uid-1", "alice", "member")
	w := doRequest(t, ts.engine, http.MethodGet, "/api/folders", nil, token)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d — body: %s", w.Code, w.Body.String())
	}
	var resp map[string]any
	json.Unmarshal(w.Body.Bytes(), &resp)
	folders, ok := resp["folders"].([]any)
	if !ok || len(folders) != 2 {
		t.Errorf("expected 2 folders, got: %v", resp["folders"])
	}
}

func TestIntegration_GetFolder_NotFound(t *testing.T) {
	ts, cleanup := newTestServer(t)
	defer cleanup()

	ts.mock.ExpectQuery("SELECT folder_id, owner_id, name, created_at, updated_at").
		WithArgs(99).
		WillReturnRows(sqlmock.NewRows([]string{"folder_id", "owner_id", "name", "created_at", "updated_at"}))

	token := ts.bearerToken("uid-1", "alice", "member")
	w := doRequest(t, ts.engine, http.MethodGet, "/api/folders/99", nil, token)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d — body: %s", w.Code, w.Body.String())
	}
}

func TestIntegration_GetFolder_Forbidden(t *testing.T) {
	ts, cleanup := newTestServer(t)
	defer cleanup()

	now := time.Now()
	// Folder owned by uid-2, not by uid-1
	ts.mock.ExpectQuery("SELECT folder_id, owner_id, name, created_at, updated_at").
		WithArgs(3).
		WillReturnRows(sqlmock.NewRows([]string{"folder_id", "owner_id", "name", "created_at", "updated_at"}).
			AddRow(3, "uid-2", "Secret", now, now))
	// permission check: no permission
	ts.mock.ExpectQuery("SELECT").
		WillReturnRows(sqlmock.NewRows([]string{}))
	// team manager check
	ts.mock.ExpectQuery("SELECT COUNT").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	token := ts.bearerToken("uid-1", "alice", "member")
	w := doRequest(t, ts.engine, http.MethodGet, "/api/folders/3", nil, token)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d — body: %s", w.Code, w.Body.String())
	}
}

func TestIntegration_DeleteFolder_InvalidID(t *testing.T) {
	ts, cleanup := newTestServer(t)
	defer cleanup()

	token := ts.bearerToken("uid-1", "alice", "member")
	w := doRequest(t, ts.engine, http.MethodDelete, "/api/folders/not-a-number", nil, token)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}
