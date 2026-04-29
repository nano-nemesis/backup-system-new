package web

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"
)

// apiRoutes registers all JSON API endpoints under /api/*.
func (s *Server) apiRoutes() {
	s.mux.HandleFunc("GET /api/setup/status", s.apiSetupStatus)
	s.mux.HandleFunc("POST /api/setup", s.apiSetup)
	s.mux.HandleFunc("POST /api/login", s.apiLogin)
	s.mux.HandleFunc("POST /api/logout", s.apiLogout)
	s.mux.HandleFunc("GET /api/me", s.apiAuth(s.apiMe))
	s.mux.HandleFunc("GET /api/nodes", s.apiAuth(s.apiNodes))
	s.mux.HandleFunc("GET /api/nodes/{name}", s.apiAuth(s.apiNode))
	s.mux.HandleFunc("GET /api/admin/users", s.apiAuth(s.apiAdminOnly(s.apiAdminUsers)))
	s.mux.HandleFunc("POST /api/admin/users", s.apiAuth(s.apiAdminOnly(s.apiAdminCreateUser)))
	s.mux.HandleFunc("DELETE /api/admin/users/{id}", s.apiAuth(s.apiAdminOnly(s.apiAdminDeleteUser)))
	s.mux.HandleFunc("PATCH /api/admin/users/{id}/role", s.apiAuth(s.apiAdminOnly(s.apiAdminUpdateRole)))
	s.mux.HandleFunc("PATCH /api/admin/users/{id}/password", s.apiAuth(s.apiAdminOnly(s.apiAdminUpdatePassword)))
}

// ─── JSON helpers ─────────────────────────────────────────────────────────────

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v) //nolint:errcheck
}

func writeAPIError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

// ─── Middleware ────────────────────────────────────────────────────────────────

func (s *Server) apiAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if s.users.Count() == 0 {
			writeAPIError(w, http.StatusServiceUnavailable, "setup required")
			return
		}
		sess := s.sessionFromRequest(r)
		if sess == nil {
			writeAPIError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		ctx := context.WithValue(r.Context(), ctxSession, sess)
		next(w, r.WithContext(ctx))
	}
}

func (s *Server) apiAdminOnly(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sess := sessionFromCtx(r)
		if sess == nil || sess.Role != RoleAdmin {
			writeAPIError(w, http.StatusForbidden, "forbidden")
			return
		}
		next(w, r)
	}
}

// ─── Setup ────────────────────────────────────────────────────────────────────

func (s *Server) apiSetupStatus(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]bool{"needs_setup": s.users.Count() == 0})
}

func (s *Server) apiSetup(w http.ResponseWriter, r *http.Request) {
	if s.users.Count() > 0 {
		writeAPIError(w, http.StatusConflict, "already configured")
		return
	}
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	req.Username = strings.TrimSpace(req.Username)
	if req.Username == "" || req.Password == "" {
		writeAPIError(w, http.StatusBadRequest, "username and password are required")
		return
	}
	if len(req.Password) < 8 {
		writeAPIError(w, http.StatusBadRequest, "password must be at least 8 characters")
		return
	}
	if err := s.users.Create(req.Username, req.Password, RoleAdmin); err != nil {
		writeAPIError(w, http.StatusInternalServerError, "failed to create user")
		return
	}
	writeJSON(w, http.StatusCreated, map[string]bool{"ok": true})
}

// ─── Auth ─────────────────────────────────────────────────────────────────────

func (s *Server) apiLogin(w http.ResponseWriter, r *http.Request) {
	ip := remoteIP(r)
	if loginLimiter.limited(ip) {
		writeAPIError(w, http.StatusTooManyRequests, "too many failed attempts, try again in 5 minutes")
		return
	}
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	req.Username = strings.TrimSpace(req.Username)
	user, ok := s.users.GetByUsername(req.Username)
	if !ok || !s.users.CheckPassword(user, req.Password) {
		loginLimiter.record(ip)
		writeAPIError(w, http.StatusUnauthorized, "invalid username or password")
		return
	}
	loginLimiter.reset(ip)
	token, err := s.sessions.Create(user.ID, user.Username, user.Role)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal error")
		return
	}
	s.setSessionCookie(w, token, isHTTPS(r))
	writeJSON(w, http.StatusOK, map[string]any{
		"user": map[string]string{
			"id":       user.ID,
			"username": user.Username,
			"role":     user.Role,
		},
	})
}

func (s *Server) apiLogout(w http.ResponseWriter, r *http.Request) {
	if c, err := r.Cookie(cookieName); err == nil {
		s.sessions.Delete(c.Value)
	}
	s.clearSessionCookie(w)
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (s *Server) apiMe(w http.ResponseWriter, r *http.Request) {
	sess := sessionFromCtx(r)
	writeJSON(w, http.StatusOK, map[string]any{
		"user": map[string]string{
			"id":       sess.UserID,
			"username": sess.Username,
			"role":     sess.Role,
		},
	})
}

// ─── Nodes ────────────────────────────────────────────────────────────────────

type apiNodeDTO struct {
	Name        string       `json:"name"`
	Type        string       `json:"type"`
	Host        string       `json:"host"`
	LastStatus  string       `json:"last_status"`
	LastTime    *time.Time   `json:"last_time"`
	LastMsg     string       `json:"last_msg"`
	BackupFiles []apiFileDTO `json:"backup_files"`
}

type apiFileDTO struct {
	Name    string    `json:"name"`
	RelPath string    `json:"rel_path"`
	Size    int64     `json:"size"`
	SizeStr string    `json:"size_str"`
	ModTime time.Time `json:"mod_time"`
}

func toNodeDTO(ns NodeStatus) apiNodeDTO {
	dto := apiNodeDTO{
		Name:       ns.Name,
		Type:       ns.Type,
		Host:       ns.Host,
		LastStatus: ns.LastStatus,
		LastMsg:    ns.LastMsg,
	}
	if !ns.LastTime.IsZero() {
		t := ns.LastTime
		dto.LastTime = &t
	}
	dto.BackupFiles = make([]apiFileDTO, len(ns.BackupFiles))
	for i, f := range ns.BackupFiles {
		dto.BackupFiles[i] = apiFileDTO{
			Name:    f.Name,
			RelPath: f.RelPath,
			Size:    f.Size,
			SizeStr: f.SizeStr,
			ModTime: f.ModTime,
		}
	}
	return dto
}

func (s *Server) apiNodes(w http.ResponseWriter, r *http.Request) {
	statuses, err := ReadStatuses(s.cfg)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "failed to read node statuses")
		return
	}
	var ok, failed, unknown int
	nodes := make([]apiNodeDTO, len(statuses))
	for i, ns := range statuses {
		nodes[i] = toNodeDTO(ns)
		switch ns.LastStatus {
		case "SUCCESS":
			ok++
		case "ERROR":
			failed++
		default:
			unknown++
		}
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"nodes": nodes,
		"stats": map[string]int{
			"total":   len(statuses),
			"ok":      ok,
			"failed":  failed,
			"unknown": unknown,
		},
	})
}

func (s *Server) apiNode(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	statuses, err := ReadStatuses(s.cfg)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "failed to read node statuses")
		return
	}
	var found *NodeStatus
	for i := range statuses {
		if statuses[i].Name == name {
			found = &statuses[i]
			break
		}
	}
	if found == nil {
		writeAPIError(w, http.StatusNotFound, "node not found")
		return
	}
	logs, _ := readLogTail(s.cfg.LogDir, name, 100)
	if logs == nil {
		logs = []string{}
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"node": toNodeDTO(*found),
		"logs": logs,
	})
}

// ─── Admin: users ─────────────────────────────────────────────────────────────

type apiUserDTO struct {
	ID        string    `json:"id"`
	Username  string    `json:"username"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
}

func (s *Server) apiAdminUsers(w http.ResponseWriter, r *http.Request) {
	all := s.users.All()
	out := make([]apiUserDTO, len(all))
	for i, u := range all {
		out[i] = apiUserDTO{ID: u.ID, Username: u.Username, Role: u.Role, CreatedAt: u.CreatedAt}
	}
	writeJSON(w, http.StatusOK, map[string]any{"users": out})
}

func (s *Server) apiAdminCreateUser(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Role     string `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	req.Username = strings.TrimSpace(req.Username)
	if req.Role != RoleAdmin && req.Role != RoleViewer {
		req.Role = RoleViewer
	}
	if req.Username == "" || req.Password == "" {
		writeAPIError(w, http.StatusBadRequest, "username and password are required")
		return
	}
	if len(req.Password) < 8 {
		writeAPIError(w, http.StatusBadRequest, "password must be at least 8 characters")
		return
	}
	if err := s.users.Create(req.Username, req.Password, req.Role); err != nil {
		writeAPIError(w, http.StatusConflict, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, map[string]bool{"ok": true})
}

func (s *Server) apiAdminDeleteUser(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	sess := sessionFromCtx(r)
	if sess != nil && sess.UserID == id {
		writeAPIError(w, http.StatusBadRequest, "cannot delete your own account")
		return
	}
	if err := s.users.Delete(id); err != nil {
		writeAPIError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (s *Server) apiAdminUpdateRole(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var req struct {
		Role string `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Role != RoleAdmin && req.Role != RoleViewer {
		writeAPIError(w, http.StatusBadRequest, "invalid role")
		return
	}
	if err := s.users.UpdateRole(id, req.Role); err != nil {
		writeAPIError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (s *Server) apiAdminUpdatePassword(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var req struct {
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if len(req.Password) < 8 {
		writeAPIError(w, http.StatusBadRequest, "password must be at least 8 characters")
		return
	}
	if err := s.users.UpdatePassword(id, req.Password); err != nil {
		writeAPIError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}
