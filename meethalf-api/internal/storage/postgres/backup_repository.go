package postgres

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"meethalf-api/internal/config"
)

const (
	defaultPgDumpPath    = "pg_dump"
	defaultPgRestorePath = "pg_restore"
)

type BackupRepository struct {
	cfg         config.DBConfig
	dumpPath    string
	restorePath string
}

func NewBackupRepository(cfg config.DBConfig) *BackupRepository {
	return &BackupRepository{
		cfg:         cfg,
		dumpPath:    defaultPgDumpPath,
		restorePath: defaultPgRestorePath,
	}
}

func (r *BackupRepository) Dump(ctx context.Context) (io.ReadCloser, error) {
	if r == nil {
		return nil, errors.New("backup repository is nil")
	}

	cfg := withBackupDefaults(r.cfg)
	if strings.TrimSpace(cfg.Name) == "" {
		return nil, errors.New("database name is empty")
	}

	args := []string{
		"--format=custom",
		"--no-owner",
		"--no-privileges",
		"--dbname", cfg.Name,
	}
	if schema := strings.TrimSpace(cfg.Schema); schema != "" {
		args = append(args, "--schema", schema)
	}

	cmd, stdout, stderrBuf, stderrDone, err := startBackupCommand(ctx, cfg, r.dumpPath, args...)
	if err != nil {
		return nil, err
	}

	return &processReadCloser{
		ReadCloser: stdout,
		cmd:        cmd,
		stderr:     stderrBuf,
		stderrDone: stderrDone,
	}, nil
}

func (r *BackupRepository) Restore(ctx context.Context, source io.Reader) error {
	if r == nil {
		return errors.New("backup repository is nil")
	}
	if source == nil {
		return errors.New("backup source is nil")
	}

	cfg := withBackupDefaults(r.cfg)
	if strings.TrimSpace(cfg.Name) == "" {
		return errors.New("database name is empty")
	}

	args := []string{
		"--clean",
		"--if-exists",
		"--no-owner",
		"--no-privileges",
		"--dbname", cfg.Name,
	}
	if schema := strings.TrimSpace(cfg.Schema); schema != "" {
		args = append(args, "--schema", schema)
	}

	if ctx == nil {
		ctx = context.Background()
	}

	cmd := exec.CommandContext(ctx, r.restorePath, args...)
	cmd.Env = append(os.Environ(), pgEnv(cfg)...)
	cmd.Stdin = source
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		message := strings.TrimSpace(stderr.String())
		if message == "" {
			return fmt.Errorf("pg_restore failed: %w", err)
		}
		return fmt.Errorf("pg_restore failed: %w: %s", err, message)
	}

	return nil
}

type processReadCloser struct {
	io.ReadCloser
	cmd        *exec.Cmd
	stderr     *bytes.Buffer
	stderrDone <-chan struct{}
}

func (p *processReadCloser) Close() error {
	if p == nil {
		return nil
	}

	waitErr := error(nil)
	if p.cmd != nil {
		waitErr = p.cmd.Wait()
	}

	if p.stderrDone != nil {
		<-p.stderrDone
	}

	closeErr := error(nil)
	if p.ReadCloser != nil {
		closeErr = p.ReadCloser.Close()
	}

	if waitErr != nil {
		message := ""
		if p.stderr != nil {
			message = strings.TrimSpace(p.stderr.String())
		}
		if message != "" {
			return fmt.Errorf("pg_dump failed: %w: %s", waitErr, message)
		}
		return fmt.Errorf("pg_dump failed: %w", waitErr)
	}

	return closeErr
}

func startBackupCommand(ctx context.Context, cfg config.DBConfig, path string, args ...string) (*exec.Cmd, io.ReadCloser, *bytes.Buffer, <-chan struct{}, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	cmd := exec.CommandContext(ctx, path, args...)
	cmd.Env = append(os.Environ(), pgEnv(cfg)...)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("init pg_dump stdout: %w", err)
	}

	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("init pg_dump stderr: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, nil, nil, nil, fmt.Errorf("start pg_dump: %w", err)
	}

	stderrBuf := &bytes.Buffer{}
	done := make(chan struct{})
	go func() {
		_, _ = io.Copy(stderrBuf, stderrPipe)
		close(done)
	}()

	return cmd, stdout, stderrBuf, done, nil
}

func withBackupDefaults(cfg config.DBConfig) config.DBConfig {
	if cfg.Host == "" {
		cfg.Host = "localhost"
	}
	if cfg.Port == "" {
		cfg.Port = "5432"
	}
	if cfg.User == "" {
		cfg.User = "postgres"
	}
	if cfg.Name == "" {
		cfg.Name = "postgres"
	}
	if cfg.SSLMode == "" {
		cfg.SSLMode = "disable"
	}

	return cfg
}

func pgEnv(cfg config.DBConfig) []string {
	env := []string{
		fmt.Sprintf("PGHOST=%s", cfg.Host),
		fmt.Sprintf("PGPORT=%s", cfg.Port),
		fmt.Sprintf("PGUSER=%s", cfg.User),
		fmt.Sprintf("PGDATABASE=%s", cfg.Name),
		fmt.Sprintf("PGSSLMODE=%s", cfg.SSLMode),
	}
	if cfg.Password != "" {
		env = append(env, fmt.Sprintf("PGPASSWORD=%s", cfg.Password))
	}
	if cfg.ConnectTimeout > 0 {
		env = append(env, fmt.Sprintf("PGCONNECT_TIMEOUT=%d", int(cfg.ConnectTimeout.Seconds())))
	}

	return env
}
