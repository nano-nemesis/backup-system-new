package backup

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"backup-system/config"
	"backup-system/internal/database"
	"backup-system/internal/mikrotik"
	"backup-system/internal/notifier"
	sshclient "backup-system/internal/ssh"
	"backup-system/internal/utils"
)

type Runner struct {
	cfg         config.Config
	logger      *utils.Logger
	storage     *Storage
	nodeLogger  *NodeLogger
	mysql       *database.MySQLBackup
	mikrotik    *mikrotik.Backup
	telegram    *notifier.Telegram
}

type Result struct {
	NodeName  string
	NodeType  string
	Path      string
	Size      int64
	Duration  time.Duration
	Err       error
}

func NewRunner(cfg config.Config, logger *utils.Logger) *Runner {
	ssh := sshclient.New(cfg.SSHTimeout)

	return &Runner{
		cfg:      cfg,
		logger:   logger,
		storage:  NewStorage(cfg.StorageDir),
		nodeLogger: NewNodeLogger(cfg.LogDir),
		mysql:    database.NewMySQLBackup(ssh),
		mikrotik: mikrotik.NewBackup(ssh),
		telegram: notifier.NewTelegram(cfg.TelegramBotToken, cfg.TelegramChatID),
	}
}

func (r *Runner) Run(ctx context.Context) error {
	nodes, err := config.LoadNodes(r.cfg.NodesPath)
	if err != nil {
		return err
	}
	nodes = r.applyDefaults(nodes)

	now := time.Now()
	if err := r.storage.PurgeOlderThan(config.NodeTypeDatabase, r.cfg.RetentionDays, now); err != nil {
		r.logger.Printf("retention warning for database backups: %v", err)
	}
	if err := r.storage.PurgeOlderThan(config.NodeTypeMikrotik, r.cfg.RetentionDays, now); err != nil {
		r.logger.Printf("retention warning for mikrotik backups: %v", err)
	}

	results := r.runParallel(ctx, nodes)
	return r.report(results)
}

func (r *Runner) applyDefaults(nodes []config.Node) []config.Node {
	for i := range nodes {
		if nodes[i].Port == 0 && r.cfg.SSHPort > 0 {
			nodes[i].Port = r.cfg.SSHPort
		}
		if nodes[i].Port == 0 {
			nodes[i].Port = 22
		}
	}
	return nodes
}

func (r *Runner) runParallel(ctx context.Context, nodes []config.Node) []Result {
	results := make([]Result, 0, len(nodes))
	resultsCh := make(chan Result, len(nodes))

	var wg sync.WaitGroup
	sem := r.makeSemaphore()

	for _, node := range nodes {
		node := node
		wg.Add(1)
		go func() {
			defer wg.Done()

			if sem != nil {
				select {
				case sem <- struct{}{}:
					defer func() { <-sem }()
				case <-ctx.Done():
					resultsCh <- Result{NodeName: node.Name, NodeType: node.Type, Err: ctx.Err()}
					return
				}
			}

			resultsCh <- r.backupNode(ctx, node)
		}()
	}

	go func() {
		wg.Wait()
		close(resultsCh)
	}()

	for result := range resultsCh {
		results = append(results, result)
	}

	return results
}

func (r *Runner) backupNode(ctx context.Context, node config.Node) Result {
	start := time.Now()

	select {
	case <-ctx.Done():
		return Result{NodeName: node.Name, NodeType: node.Type, Duration: time.Since(start), Err: ctx.Err()}
	default:
	}

	var result Result
	result.NodeName = node.Name
	result.NodeType = node.Type

	switch node.Type {
	case config.NodeTypeDatabase:
		result = r.backupDatabase(node)
	case config.NodeTypeMikrotik:
		result = r.backupMikrotik(node)
	default:
		result.Err = fmt.Errorf("unsupported node type %q", node.Type)
	}

	result.NodeName = node.Name
	result.NodeType = node.Type
	result.Duration = time.Since(start)

	if result.Err != nil {
		r.logger.Printf("backup failed for %s (%s): %v", node.Name, node.Type, result.Err)
		r.logNodeEvent(node.Name, "ERROR", fmt.Sprintf("backup %s failed: %v", node.Type, result.Err))
		_ = r.telegram.Send(fmt.Sprintf("BACKUP FAILED\nnode: %s\ntype: %s\nerror: %v", node.Name, node.Type, result.Err))
		return result
	}

	r.logger.Printf("backup completed for %s (%s): %s (%s)", node.Name, node.Type, result.Path, utils.FormatBytes(result.Size))
	r.logNodeEvent(node.Name, "SUCCESS", fmt.Sprintf("backup %s completed: file=%s size=%s duration=%s", node.Type, result.Path, utils.FormatBytes(result.Size), result.Duration.Round(time.Millisecond)))
	_ = r.telegram.Send(fmt.Sprintf("BACKUP SUCCESS\nnode: %s\ntype: %s\nfile: %s\nsize: %s", node.Name, node.Type, result.Path, utils.FormatBytes(result.Size)))
	return result
}

func (r *Runner) backupDatabase(node config.Node) Result {
	stream, waitCh, err := r.mysql.Stream(node)
	if err != nil {
		return Result{Err: err}
	}
	defer stream.Close()

	filename := database.BackupFilename(node.Name, time.Now())
	path, size, err := r.storage.Save(config.NodeTypeDatabase, filename, stream)
	if err != nil {
		return Result{Err: err}
	}

	if err := <-waitCh; err != nil {
		return Result{Err: err}
	}

	return Result{Path: path, Size: size}
}

func (r *Runner) backupMikrotik(node config.Node) Result {
	fetchResult, err := r.mikrotik.Fetch(node)
	if err != nil {
		return Result{Err: err}
	}

	filename := mikrotik.BackupFilename(node.Name, time.Now())
	path, size, err := r.storage.Save(config.NodeTypeMikrotik, filename, fetchResult.Reader)
	if err != nil {
		return Result{Err: err}
	}

	if fetchResult.UsedPasswordFallback {
		r.logger.Printf("mikrotik backup for %s used SSH password fallback after key authentication failed", node.Name)
		r.logNodeEvent(node.Name, "INFO", "mikrotik backup used SSH password fallback after key authentication failed")
	}

	return Result{Path: path, Size: size}
}

func (r *Runner) report(results []Result) error {
	var successCount int
	var failures []string

	for _, result := range results {
		if result.Err != nil {
			failures = append(failures, fmt.Sprintf("%s (%s): %v", result.NodeName, result.NodeType, result.Err))
			continue
		}
		successCount++
	}

	summary := fmt.Sprintf("backup run finished: %d success, %d failed", successCount, len(failures))
	r.logger.Println(summary)

	if len(failures) > 0 {
		return fmt.Errorf("%s: %s", summary, strings.Join(failures, "; "))
	}

	return nil
}

func (r *Runner) makeSemaphore() chan struct{} {
	if r.cfg.MaxParallelBackups <= 0 {
		return nil
	}
	return make(chan struct{}, r.cfg.MaxParallelBackups)
}

func (r *Runner) logNodeEvent(nodeName, status, message string) {
	if err := r.nodeLogger.Log(nodeName, status, message); err != nil {
		r.logger.Printf("node log write failed for %s: %v", nodeName, err)
	}
}
