package web

import (
	"context"
	"embed"
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"net"
	"net/http"
	"strings"
	"time"

	"backup-system/config"
)

//go:embed templates/* static/*
var embeddedFS embed.FS

const cookieName = "bs_session"

type Server struct {
	cfg      config.Config
	users    *UserStore
	sessions *SessionStore
	tmpl     *template.Template
	mux      *http.ServeMux
}

func NewServer(cfg config.Config) (*Server, error) {
	if cfg.WebSecret == "" {
		return nil, fmt.Errorf("WEB_SECRET is required")
	}

	users, err := NewUserStore(cfg.WebUsersFile)
	if err != nil {
		return nil, fmt.Errorf("user store: %w", err)
	}

	tmpl, err := loadTemplates()
	if err != nil {
		return nil, fmt.Errorf("templates: %w", err)
	}

	s := &Server{
		cfg:      cfg,
		users:    users,
		sessions: NewSessionStore(),
		tmpl:     tmpl,
		mux:      http.NewServeMux(),
	}
	s.routes()
	return s, nil
}

func (s *Server) routes() {
	staticFS, _ := fs.Sub(embeddedFS, "static")
	s.mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.FS(staticFS))))

	s.mux.HandleFunc("GET /login", s.handleLoginPage)
	s.mux.HandleFunc("POST /login", s.handleLoginSubmit)
	s.mux.HandleFunc("GET /logout", s.handleLogout)
	s.mux.HandleFunc("GET /setup", s.handleSetupPage)
	s.mux.HandleFunc("POST /setup", s.handleSetupSubmit)

	s.mux.HandleFunc("GET /{$}", s.auth(s.handleDashboard))
	s.mux.HandleFunc("GET /nodes/{name}", s.auth(s.handleNode))
	s.mux.HandleFunc("GET /storage/{type}/{file}", s.auth(s.handleDownload))

	s.mux.HandleFunc("GET /admin/users", s.auth(s.admin(s.handleAdminUsers)))
	s.mux.HandleFunc("POST /admin/users", s.auth(s.admin(s.handleAdminCreateUser)))
	s.mux.HandleFunc("POST /admin/users/{id}/delete", s.auth(s.admin(s.handleAdminDeleteUser)))
	s.mux.HandleFunc("POST /admin/users/{id}/role", s.auth(s.admin(s.handleAdminUpdateRole)))
	s.mux.HandleFunc("POST /admin/users/{id}/password", s.auth(s.admin(s.handleAdminUpdatePassword)))
}

func (s *Server) ListenAndServe() error {
	srv := &http.Server{
		Addr:         s.cfg.WebListen,
		Handler:      s.mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	if s.cfg.WebTLSCert != "" && s.cfg.WebTLSKey != "" {
		log.Printf("web: listening on https://%s", s.cfg.WebListen)
		return srv.ListenAndServeTLS(s.cfg.WebTLSCert, s.cfg.WebTLSKey)
	}

	log.Printf("web: listening on http://%s", s.cfg.WebListen)
	return srv.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return nil
}

// auth middleware: redirect to /login if not authenticated, to /setup if no users
func (s *Server) auth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if s.users.Count() == 0 {
			http.Redirect(w, r, "/setup", http.StatusFound)
			return
		}
		sess := s.sessionFromRequest(r)
		if sess == nil {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}
		ctx := context.WithValue(r.Context(), ctxSession, sess)
		next(w, r.WithContext(ctx))
	}
}

// admin middleware: require admin role
func (s *Server) admin(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sess := sessionFromCtx(r)
		if sess == nil || sess.Role != RoleAdmin {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		next(w, r)
	}
}

func (s *Server) sessionFromRequest(r *http.Request) *Session {
	c, err := r.Cookie(cookieName)
	if err != nil {
		return nil
	}
	sess, ok := s.sessions.Get(c.Value)
	if !ok {
		return nil
	}
	return sess
}

func (s *Server) setSessionCookie(w http.ResponseWriter, token string, secure bool) {
	http.SetCookie(w, &http.Cookie{
		Name:     cookieName,
		Value:    token,
		Path:     "/",
		MaxAge:   int(sessionTTL.Seconds()),
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteStrictMode,
	})
}

func (s *Server) clearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     cookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})
}

func isHTTPS(r *http.Request) bool {
	return r.TLS != nil || strings.EqualFold(r.Header.Get("X-Forwarded-Proto"), "https")
}

func remoteIP(r *http.Request) string {
	if fwd := r.Header.Get("X-Forwarded-For"); fwd != "" {
		parts := strings.SplitN(fwd, ",", 2)
		return strings.TrimSpace(parts[0])
	}
	host, _, _ := net.SplitHostPort(r.RemoteAddr)
	return host
}

type ctxKey int

const ctxSession ctxKey = 0

func sessionFromCtx(r *http.Request) *Session {
	s, _ := r.Context().Value(ctxSession).(*Session)
	return s
}

func loadTemplates() (*template.Template, error) {
	funcMap := template.FuncMap{
		"timeAgo": timeAgo,
		"fmtTime": func(t time.Time) string {
			if t.IsZero() {
				return "—"
			}
			return t.UTC().Format("2006-01-02 15:04")
		},
		"logClass": func(line string) string {
			lower := strings.ToLower(line)
			switch {
			case strings.Contains(lower, "error"):
				return "log-error"
			case strings.Contains(lower, "success"):
				return "log-success"
			default:
				return "log-info"
			}
		},
	}
	return template.New("").Funcs(funcMap).ParseFS(embeddedFS, "templates/*.html")
}

func timeAgo(t time.Time) string {
	if t.IsZero() {
		return "never"
	}
	d := time.Since(t)
	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		return fmt.Sprintf("%dm ago", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh ago", int(d.Hours()))
	default:
		return fmt.Sprintf("%dd ago", int(d.Hours()/24))
	}
}
