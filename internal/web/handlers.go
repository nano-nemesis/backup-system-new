package web

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// ─── Auth handlers ────────────────────────────────────────────────────────────

var loginLimiter = &rateLimiter{}

func (s *Server) handleLoginPage(w http.ResponseWriter, r *http.Request) {
	if s.users.Count() == 0 {
		http.Redirect(w, r, "/setup", http.StatusFound)
		return
	}
	if sess := s.sessionFromRequest(r); sess != nil {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	s.render(w, "login.html", map[string]any{"Error": r.URL.Query().Get("error")})
}

func (s *Server) handleLoginSubmit(w http.ResponseWriter, r *http.Request) {
	ip := remoteIP(r)
	if loginLimiter.limited(ip) {
		s.renderError(w, r, "login.html", "Too many failed attempts. Try again in 5 minutes.")
		return
	}

	username := strings.TrimSpace(r.FormValue("username"))
	password := r.FormValue("password")

	user, ok := s.users.GetByUsername(username)
	if !ok || !s.users.CheckPassword(user, password) {
		loginLimiter.record(ip)
		s.renderError(w, r, "login.html", "Invalid username or password.")
		return
	}

	loginLimiter.reset(ip)
	token, err := s.sessions.Create(user.ID, user.Username, user.Role)
	if err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	s.setSessionCookie(w, token, isHTTPS(r))
	http.Redirect(w, r, "/", http.StatusFound)
}

func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	if c, err := r.Cookie(cookieName); err == nil {
		s.sessions.Delete(c.Value)
	}
	s.clearSessionCookie(w)
	http.Redirect(w, r, "/login", http.StatusFound)
}

func (s *Server) handleSetupPage(w http.ResponseWriter, r *http.Request) {
	if s.users.Count() > 0 {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}
	s.render(w, "setup.html", map[string]any{"Error": r.URL.Query().Get("error")})
}

func (s *Server) handleSetupSubmit(w http.ResponseWriter, r *http.Request) {
	if s.users.Count() > 0 {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}
	username := strings.TrimSpace(r.FormValue("username"))
	password := r.FormValue("password")
	confirm := r.FormValue("confirm")

	if username == "" || password == "" {
		s.renderError(w, r, "setup.html", "Username and password are required.")
		return
	}
	if len(password) < 8 {
		s.renderError(w, r, "setup.html", "Password must be at least 8 characters.")
		return
	}
	if password != confirm {
		s.renderError(w, r, "setup.html", "Passwords do not match.")
		return
	}
	if err := s.users.Create(username, password, RoleAdmin); err != nil {
		s.renderError(w, r, "setup.html", fmt.Sprintf("Failed to create user: %v", err))
		return
	}
	http.Redirect(w, r, "/login", http.StatusFound)
}

// ─── Dashboard ────────────────────────────────────────────────────────────────

func (s *Server) handleDashboard(w http.ResponseWriter, r *http.Request) {
	statuses, err := ReadStatuses(s.cfg)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to read statuses: %v", err), http.StatusInternalServerError)
		return
	}

	var ok, failed, unknown int
	for _, ns := range statuses {
		switch ns.LastStatus {
		case "SUCCESS":
			ok++
		case "ERROR":
			failed++
		default:
			unknown++
		}
	}

	sess := sessionFromCtx(r)
	s.render(w, "dashboard.html", map[string]any{
		"Session":  sess,
		"Nodes":    statuses,
		"Total":    len(statuses),
		"OK":       ok,
		"Failed":   failed,
		"Unknown":  unknown,
	})
}

// ─── Node detail ──────────────────────────────────────────────────────────────

func (s *Server) handleNode(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")

	statuses, err := ReadStatuses(s.cfg)
	if err != nil {
		http.Error(w, "Failed to read statuses", http.StatusInternalServerError)
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
		http.NotFound(w, r)
		return
	}

	logLines, _ := readLogTail(s.cfg.LogDir, name, 80)

	sess := sessionFromCtx(r)
	s.render(w, "node.html", map[string]any{
		"Session":  sess,
		"Node":     found,
		"LogLines": logLines,
	})
}

func readLogTail(logDir, nodeName string, n int) ([]string, error) {
	path := filepath.Join(logDir, safeLogName(nodeName)+".log")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) > n {
		lines = lines[len(lines)-n:]
	}
	return lines, nil
}

// ─── File download ────────────────────────────────────────────────────────────

func (s *Server) handleDownload(w http.ResponseWriter, r *http.Request) {
	nodeType := r.PathValue("type")
	fileName := filepath.Base(r.PathValue("file")) // strip any traversal

	if nodeType != "mikrotik" && nodeType != "database" {
		http.Error(w, "Invalid type", http.StatusBadRequest)
		return
	}

	storageBase, err := filepath.Abs(s.cfg.StorageDir)
	if err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	joined := filepath.Join(storageBase, nodeType, fileName)
	// Ensure the resolved path is still inside storageBase
	if !strings.HasPrefix(joined, storageBase+string(filepath.Separator)) {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, fileName))
	http.ServeFile(w, r, joined)
}

// ─── Admin: user management ───────────────────────────────────────────────────

func (s *Server) handleAdminUsers(w http.ResponseWriter, r *http.Request) {
	sess := sessionFromCtx(r)
	s.render(w, "admin.html", map[string]any{
		"Session": sess,
		"Users":   s.users.All(),
		"Error":   r.URL.Query().Get("error"),
		"Success": r.URL.Query().Get("success"),
	})
}

func (s *Server) handleAdminCreateUser(w http.ResponseWriter, r *http.Request) {
	username := strings.TrimSpace(r.FormValue("username"))
	password := r.FormValue("password")
	role := r.FormValue("role")

	if role != RoleAdmin && role != RoleViewer {
		role = RoleViewer
	}
	if username == "" || password == "" {
		http.Redirect(w, r, "/admin/users?error=Username+and+password+required", http.StatusFound)
		return
	}
	if len(password) < 8 {
		http.Redirect(w, r, "/admin/users?error=Password+must+be+at+least+8+chars", http.StatusFound)
		return
	}
	if err := s.users.Create(username, password, role); err != nil {
		http.Redirect(w, r, "/admin/users?error="+err.Error(), http.StatusFound)
		return
	}
	http.Redirect(w, r, "/admin/users?success=User+created", http.StatusFound)
}

func (s *Server) handleAdminDeleteUser(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	sess := sessionFromCtx(r)

	// Prevent self-deletion
	if sess != nil && sess.UserID == id {
		http.Redirect(w, r, "/admin/users?error=Cannot+delete+your+own+account", http.StatusFound)
		return
	}
	if err := s.users.Delete(id); err != nil {
		http.Redirect(w, r, "/admin/users?error="+err.Error(), http.StatusFound)
		return
	}
	http.Redirect(w, r, "/admin/users?success=User+deleted", http.StatusFound)
}

func (s *Server) handleAdminUpdateRole(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	role := r.FormValue("role")

	if role != RoleAdmin && role != RoleViewer {
		http.Redirect(w, r, "/admin/users?error=Invalid+role", http.StatusFound)
		return
	}
	if err := s.users.UpdateRole(id, role); err != nil {
		http.Redirect(w, r, "/admin/users?error="+err.Error(), http.StatusFound)
		return
	}
	http.Redirect(w, r, "/admin/users?success=Role+updated", http.StatusFound)
}

func (s *Server) handleAdminUpdatePassword(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	password := r.FormValue("password")

	if len(password) < 8 {
		http.Redirect(w, r, "/admin/users?error=Password+must+be+at+least+8+chars", http.StatusFound)
		return
	}
	if err := s.users.UpdatePassword(id, password); err != nil {
		http.Redirect(w, r, "/admin/users?error="+err.Error(), http.StatusFound)
		return
	}
	http.Redirect(w, r, "/admin/users?success=Password+updated", http.StatusFound)
}

// ─── Helpers ──────────────────────────────────────────────────────────────────

func (s *Server) render(w http.ResponseWriter, tmplName string, data any) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := s.tmpl.ExecuteTemplate(w, tmplName, data); err != nil {
		http.Error(w, "Template error: "+err.Error(), http.StatusInternalServerError)
	}
}

func (s *Server) renderError(w http.ResponseWriter, r *http.Request, tmplName, errMsg string) {
	s.render(w, tmplName, map[string]any{"Error": errMsg})
}

// ─── Rate limiter (login) ─────────────────────────────────────────────────────

type loginAttempt struct {
	count     int
	firstSeen time.Time
	blockedAt time.Time
}

type rateLimiter struct {
	mu      sync.Mutex
	entries map[string]*loginAttempt
}

func (rl *rateLimiter) get(ip string) *loginAttempt {
	if rl.entries == nil {
		rl.entries = make(map[string]*loginAttempt)
	}
	if rl.entries[ip] == nil {
		rl.entries[ip] = &loginAttempt{}
	}
	return rl.entries[ip]
}

func (rl *rateLimiter) limited(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	a := rl.get(ip)
	if a.count == 0 {
		return false
	}
	if !a.blockedAt.IsZero() {
		return time.Since(a.blockedAt) < 5*time.Minute
	}
	return false
}

func (rl *rateLimiter) record(ip string) {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	a := rl.get(ip)
	if time.Since(a.firstSeen) > 5*time.Minute {
		a.count = 0
		a.firstSeen = time.Now()
		a.blockedAt = time.Time{}
	}
	if a.count == 0 {
		a.firstSeen = time.Now()
	}
	a.count++
	if a.count >= 5 {
		a.blockedAt = time.Now()
	}
}

func (rl *rateLimiter) reset(ip string) {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	delete(rl.entries, ip)
}
