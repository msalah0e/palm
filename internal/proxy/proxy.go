package proxy

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/msalah0e/palm/internal/budget"
	"github.com/msalah0e/palm/internal/vault"
)

// Config holds proxy configuration.
type Config struct {
	Port    int
	LogFile string
	Verbose bool
}

// RequestLog represents a logged API request.
type RequestLog struct {
	Timestamp    time.Time `json:"ts"`
	Method       string    `json:"method"`
	Path         string    `json:"path"`
	Provider     string    `json:"provider"`
	Model        string    `json:"model,omitempty"`
	Status       int       `json:"status"`
	Duration     float64   `json:"duration_ms"`
	InputTokens  int64     `json:"input_tokens,omitempty"`
	OutputTokens int64     `json:"output_tokens,omitempty"`
	Cost         float64   `json:"cost,omitempty"`
}

// Server is the palm proxy server.
type Server struct {
	cfg     Config
	v       vault.Vault
	logFile *os.File
	mu      sync.Mutex
	stats   ProxyStats
}

// ProxyStats tracks real-time proxy statistics.
type ProxyStats struct {
	TotalRequests int64
	TotalTokens   int64
	TotalCost     float64
	StartedAt     time.Time
	ByProvider    map[string]int64
}

// providerRoutes maps path prefixes to upstream targets.
var providerRoutes = map[string]string{
	"/openai/":    "https://api.openai.com",
	"/anthropic/": "https://api.anthropic.com",
	"/google/":    "https://generativelanguage.googleapis.com",
	"/groq/":      "https://api.groq.com",
	"/mistral/":   "https://api.mistral.ai",
	"/ollama/":    "http://localhost:11434",
}

// providerKeys maps provider name to the vault key for auth.
var providerKeys = map[string]string{
	"openai":    "OPENAI_API_KEY",
	"anthropic": "ANTHROPIC_API_KEY",
	"google":    "GOOGLE_API_KEY",
	"groq":      "GROQ_API_KEY",
	"mistral":   "MISTRAL_API_KEY",
}

// New creates a new proxy server.
func New(cfg Config) *Server {
	return &Server{
		cfg: cfg,
		v:   vault.New(),
		stats: ProxyStats{
			StartedAt:  time.Now(),
			ByProvider: make(map[string]int64),
		},
	}
}

// Start begins serving the proxy.
func (s *Server) Start() error {
	// Open log file
	logPath := s.cfg.LogFile
	if logPath == "" {
		dir := os.Getenv("XDG_CONFIG_HOME")
		if dir == "" {
			home, _ := os.UserHomeDir()
			dir = filepath.Join(home, ".config")
		}
		logPath = filepath.Join(dir, "palm", "proxy.jsonl")
	}
	_ = os.MkdirAll(filepath.Dir(logPath), 0o755)

	var err error
	s.logFile, err = os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handleRequest)
	mux.HandleFunc("/palm/status", s.handleStatus)
	mux.HandleFunc("/palm/stats", s.handleStats)

	addr := fmt.Sprintf(":%d", s.cfg.Port)
	log.Printf("palm proxy listening on http://localhost%s\n", addr)
	log.Printf("Routes:")
	for prefix, target := range providerRoutes {
		log.Printf("  http://localhost%s%s → %s", addr, prefix, target)
	}
	log.Printf("\nSet OPENAI_BASE_URL=http://localhost%s/openai/v1 to route through proxy", addr)

	return http.ListenAndServe(addr, mux)
}

func (s *Server) handleRequest(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	// Determine provider from path
	provider, target, trimmedPath := s.resolveProvider(r.URL.Path)
	if provider == "" {
		http.Error(w, "unknown provider — use /openai/, /anthropic/, /google/, etc.", http.StatusBadGateway)
		return
	}

	// Budget check
	if err := budget.CheckBudget(provider); err != nil {
		http.Error(w, fmt.Sprintf("palm proxy: budget exceeded — %v", err), http.StatusPaymentRequired)
		return
	}

	// Parse upstream URL
	upstream, err := url.Parse(target)
	if err != nil {
		http.Error(w, "invalid upstream", http.StatusBadGateway)
		return
	}

	// Create reverse proxy
	proxy := httputil.NewSingleHostReverseProxy(upstream)

	// Inject API key from vault
	if keyName, ok := providerKeys[provider]; ok {
		key := os.Getenv(keyName)
		if key == "" {
			if val, err := s.v.Get(keyName); err == nil {
				key = val
			}
		}
		if key != "" {
			switch provider {
			case "anthropic":
				r.Header.Set("x-api-key", key)
			default:
				r.Header.Set("Authorization", "Bearer "+key)
			}
		}
	}

	// Update request path to strip the provider prefix
	r.URL.Path = trimmedPath
	r.URL.Host = upstream.Host
	r.URL.Scheme = upstream.Scheme
	r.Host = upstream.Host

	// Capture response
	rec := &responseRecorder{ResponseWriter: w}
	proxy.ServeHTTP(rec, r)

	elapsed := time.Since(start)

	// Log the request
	entry := RequestLog{
		Timestamp: start,
		Method:    r.Method,
		Path:      r.URL.Path,
		Provider:  provider,
		Status:    rec.statusCode,
		Duration:  float64(elapsed.Milliseconds()),
	}

	s.mu.Lock()
	s.stats.TotalRequests++
	s.stats.ByProvider[provider]++
	s.mu.Unlock()

	s.writeLog(entry)

	if s.cfg.Verbose {
		log.Printf("[%s] %s %s → %d (%.0fms)", provider, r.Method, r.URL.Path, rec.statusCode, entry.Duration)
	}
}

func (s *Server) resolveProvider(path string) (provider, target, trimmed string) {
	for prefix, t := range providerRoutes {
		if strings.HasPrefix(path, prefix) {
			name := strings.Trim(prefix, "/")
			return name, t, strings.TrimPrefix(path, strings.TrimSuffix(prefix, "/"))
		}
	}
	return "", "", ""
}

func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "running",
		"version": "1.0.0",
		"port":    s.cfg.Port,
		"uptime":  time.Since(s.stats.StartedAt).String(),
	})
}

func (s *Server) handleStats(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(s.stats)
}

func (s *Server) writeLog(entry RequestLog) {
	if s.logFile == nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	_ = json.NewEncoder(s.logFile).Encode(entry)
}

// ReadLogs returns the most recent n log entries.
func ReadLogs(n int) ([]RequestLog, error) {
	dir := os.Getenv("XDG_CONFIG_HOME")
	if dir == "" {
		home, _ := os.UserHomeDir()
		dir = filepath.Join(home, ".config")
	}
	path := filepath.Join(dir, "palm", "proxy.jsonl")

	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	defer f.Close()

	var all []RequestLog
	dec := json.NewDecoder(f)
	for dec.More() {
		var entry RequestLog
		if err := dec.Decode(&entry); err != nil {
			continue
		}
		all = append(all, entry)
	}

	if n > 0 && len(all) > n {
		all = all[len(all)-n:]
	}

	return all, nil
}

// responseRecorder captures the HTTP status code.
type responseRecorder struct {
	http.ResponseWriter
	statusCode int
	body       []byte
}

func (r *responseRecorder) WriteHeader(code int) {
	r.statusCode = code
	r.ResponseWriter.WriteHeader(code)
}

func (r *responseRecorder) Write(b []byte) (int, error) {
	if r.statusCode == 0 {
		r.statusCode = 200
	}
	r.body = append(r.body, b...)
	return r.ResponseWriter.Write(b)
}

// PidFile returns the path to the proxy PID file.
func PidFile() string {
	dir := os.Getenv("XDG_CONFIG_HOME")
	if dir == "" {
		home, _ := os.UserHomeDir()
		dir = filepath.Join(home, ".config")
	}
	return filepath.Join(dir, "palm", "proxy.pid")
}

// IsRunning checks if the proxy is currently running.
func IsRunning() (bool, int) {
	data, err := os.ReadFile(PidFile())
	if err != nil {
		return false, 0
	}
	var pid int
	if _, err := fmt.Sscanf(string(data), "%d", &pid); err != nil {
		return false, 0
	}
	// Check if process exists
	proc, err := os.FindProcess(pid)
	if err != nil {
		return false, 0
	}
	// On Unix, FindProcess always succeeds. Send signal 0 to check.
	if err := proc.Signal(nil); err == nil {
		return true, pid
	}
	// Stale PID file
	_ = os.Remove(PidFile())
	return false, 0
}

// WritePid writes the current process PID to the PID file.
func WritePid() error {
	return os.WriteFile(PidFile(), []byte(fmt.Sprintf("%d", os.Getpid())), 0o644)
}

// We need to use io in responseRecorder but it's not used directly
var _ = io.Discard
