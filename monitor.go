//go:build ignore

package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

var (
	logDir = "/MIKROTIK-LOG/"

	botToken = os.Getenv("TELEGRAM_BOT_TOKEN")
	chatID   = os.Getenv("TELEGRAM_CHAT_ID")
	openai   = os.Getenv("OPENAI_API_KEY")

	cooldown = 30 * time.Minute
	maxAI    = 20

	lastAlert = make(map[string]time.Time)
	state     = make(map[string]string)
	cnt       = make(map[string]int)
	cache     = make(map[string]string)

	aiUsed int
	mu     sync.Mutex
)

func tg(msg string) {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", botToken)
	http.PostForm(url, map[string][]string{"chat_id": {chatID}, "text": {msg}})
}

func classify(e string) (string, string) {
	e = strings.ToLower(e)
	switch {
	case strings.Contains(e, "timeout"):
		return "NETWORK", "cek firewall / koneksi"
	case strings.Contains(e, "route"):
		return "ROUTING", "cek route / vpn"
	case strings.Contains(e, "denied"):
		return "AUTH", "cek user/password"
	default:
		return "UNKNOWN", "analisa ai"
	}
}

func ai(e string) string {
	if openai == "" || aiUsed >= maxAI {
		return "-"
	}

	body, _ := json.Marshal(map[string]interface{}{
		"model": "gpt-4o-mini",
		"messages": []map[string]string{
			{"role": "user", "content": e},
		},
	})

	req, _ := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(body))
	req.Header.Set("Authorization", "Bearer "+openai)
	req.Header.Set("Content-Type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return "ai error"
	}
	defer res.Body.Close()

	var r map[string]interface{}
	json.NewDecoder(res.Body).Decode(&r)

	aiUsed++

	if c, ok := r["choices"].([]interface{}); ok {
		return c[0].(map[string]interface{})["message"].(map[string]interface{})["content"].(string)
	}

	return "-"
}

func process(node, line string) {
	mu.Lock()
	defer mu.Unlock()

	if !strings.Contains(line, "ERROR") && !strings.Contains(line, "SUCCESS") {
		return
	}

	now := time.Now()

	if t, ok := lastAlert[node]; ok && now.Sub(t) < cooldown {
		return
	}

	if strings.Contains(line, "ERROR") {
		state[node] = "ERROR"
		cnt[node]++

		cat, act := classify(line)

		insight := ""
		if cat == "UNKNOWN" || cnt[node] >= 3 {
			if v, ok := cache[line]; ok {
				insight = v
			} else {
				insight = ai(line)
				cache[line] = insight
			}
		}

		tg(fmt.Sprintf("🚨 %s\n%s\n%s\n%s", node, cat, act, insight))
		lastAlert[node] = now
	}

	if strings.Contains(line, "SUCCESS") && state[node] == "ERROR" {
		state[node] = "OK"
		tg(fmt.Sprintf("✅ %s NORMAL", node))
	}
}

func watch(f string) {
	file, _ := os.Open(f)
	defer file.Close()

	file.Seek(0, os.SEEK_END)
	r := bufio.NewReader(file)

	node := strings.TrimSuffix(filepath.Base(f), ".log")

	for {
		line, err := r.ReadString('\n')
		if err != nil {
			time.Sleep(time.Second)
			continue
		}
		process(node, line)
	}
}

func main() {
	w, _ := fsnotify.NewWatcher()
	defer w.Close()

	w.Add(logDir)

	files, _ := filepath.Glob(logDir + "*.log")
	for _, f := range files {
		go watch(f)
	}

	go func() {
		for e := range w.Events {
			if e.Op&fsnotify.Create == fsnotify.Create {
				go watch(e.Name)
			}
		}
	}()

	select {}
}
