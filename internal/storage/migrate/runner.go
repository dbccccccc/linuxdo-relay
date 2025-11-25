package migrate

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"gorm.io/gorm"
)

const migrationsTable = "schema_migrations"

var (
	ErrPartialSchema = errors.New("database already contains tables but no migration history; rerun with force")
	ErrSchemaDrift   = errors.New("database schema is ahead of this binary")
)

type Status string

const (
	StatusUnknown      Status = "unknown"
	StatusUnconfigured Status = "unconfigured"
	StatusPartial      Status = "partial"
	StatusPending      Status = "pending"
	StatusReady        Status = "ready"
	StatusDrift        Status = "drift"
)

type Result struct {
	Status  Status   `json:"status"`
	Pending []string `json:"pending"`
	Applied []string `json:"applied"`
	Unknown []string `json:"unknown"`
	Message string   `json:"message"`
}

type Runner struct {
	DB  *gorm.DB
	Dir string
}

func NewRunner(db *gorm.DB, dir string) *Runner {
	if dir == "" {
		dir = "migrations"
	}
	return &Runner{DB: db, Dir: dir}
}

func (r *Runner) Check(ctx context.Context) (*Result, error) {
	files, err := r.loadFiles()
	if err != nil {
		return nil, err
	}

	hasTable, err := r.tableExists(ctx, migrationsTable)
	if err != nil {
		return nil, err
	}

	applied := make([]string, 0)
	appliedSet := make(map[string]struct{})
	if hasTable {
		if err := r.DB.WithContext(ctx).Table(migrationsTable).Select("version").Order("version").Pluck("version", &applied).Error; err != nil {
			return nil, fmt.Errorf("load schema migrations: %w", err)
		}
		for _, v := range applied {
			appliedSet[v] = struct{}{}
		}
	}

	pending := make([]string, 0)
	for _, f := range files {
		if _, ok := appliedSet[f.Version]; !ok {
			pending = append(pending, f.Version)
		}
	}

	res := &Result{
		Pending: pending,
		Applied: applied,
	}

	if !hasTable {
		populated, err := r.hasCoreTables(ctx)
		if err != nil {
			return nil, err
		}
		if populated {
			res.Status = StatusPartial
			res.Message = "database already contains tables but no migration history"
		} else {
			res.Status = StatusUnconfigured
			res.Pending = versions(files)
		}
		return res, nil
	}

	unknown := make([]string, 0)
	fileSet := make(map[string]struct{})
	for _, f := range files {
		fileSet[f.Version] = struct{}{}
	}
	for _, v := range applied {
		if _, ok := fileSet[v]; !ok {
			unknown = append(unknown, v)
		}
	}
	res.Unknown = unknown

	if len(unknown) > 0 {
		res.Status = StatusDrift
		res.Message = "database contains migrations unknown to this build"
		return res, nil
	}

	if len(pending) > 0 {
		res.Status = StatusPending
		res.Message = fmt.Sprintf("%d migrations pending", len(pending))
		return res, nil
	}

	res.Status = StatusReady
	return res, nil
}

func (r *Runner) ApplyPending(ctx context.Context, force bool) (*Result, error) {
	res, err := r.Check(ctx)
	if err != nil {
		return nil, err
	}

	switch res.Status {
	case StatusReady:
		return res, nil
	case StatusDrift:
		return res, ErrSchemaDrift
	case StatusPartial:
		if !force {
			return res, ErrPartialSchema
		}
	}

	if err := r.ensureMigrationsTable(ctx); err != nil {
		return nil, err
	}

	files, err := r.loadFiles()
	if err != nil {
		return nil, err
	}

	appliedSet := make(map[string]struct{}, len(res.Applied))
	for _, v := range res.Applied {
		appliedSet[v] = struct{}{}
	}

	for _, f := range files {
		if _, ok := appliedSet[f.Version]; ok {
			continue
		}
		if err := r.applyFile(ctx, f); err != nil {
			return nil, err
		}
	}

	return r.Check(ctx)
}

func (r *Runner) loadFiles() ([]migrationFile, error) {
	entries, err := os.ReadDir(r.Dir)
	if err != nil {
		return nil, fmt.Errorf("read migrations dir: %w", err)
	}
	files := make([]migrationFile, 0)
	for _, ent := range entries {
		if ent.IsDir() {
			continue
		}
		name := ent.Name()
		if !strings.HasSuffix(name, ".sql") {
			continue
		}
		path := filepath.Join(r.Dir, name)
		files = append(files, migrationFile{Version: name, Path: path})
	}
	sort.Slice(files, func(i, j int) bool {
		return files[i].Version < files[j].Version
	})
	return files, nil
}

func (r *Runner) tableExists(ctx context.Context, name string) (bool, error) {
	var exists bool
	query := "SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema = current_schema() AND table_name = ? LIMIT 1)"
	if err := r.DB.WithContext(ctx).Raw(query, name).Scan(&exists).Error; err != nil {
		return false, fmt.Errorf("check table %s: %w", name, err)
	}
	return exists, nil
}

var coreTables = []string{
	"users",
	"channels",
	"quota_rules",
}

func (r *Runner) hasCoreTables(ctx context.Context) (bool, error) {
	for _, tbl := range coreTables {
		exists, err := r.tableExists(ctx, tbl)
		if err != nil {
			return false, err
		}
		if exists {
			return true, nil
		}
	}
	return false, nil
}

func (r *Runner) ensureMigrationsTable(ctx context.Context) error {
	stmt := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s (
        version TEXT PRIMARY KEY,
        applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
    )`, migrationsTable)
	if err := r.DB.WithContext(ctx).Exec(stmt).Error; err != nil {
		return fmt.Errorf("ensure schema_migrations: %w", err)
	}
	return nil
}

func (r *Runner) applyFile(ctx context.Context, file migrationFile) error {
	contents, err := os.ReadFile(file.Path)
	if err != nil {
		return fmt.Errorf("read migration %s: %w", file.Version, err)
	}
	sql := strings.TrimSpace(string(contents))
	if sql == "" {
		return nil
	}

	tx := r.DB.WithContext(ctx).Begin()
	if err := tx.Error; err != nil {
		return fmt.Errorf("begin migration %s: %w", file.Version, err)
	}

	if err := tx.Exec(sql).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("execute migration %s: %w", file.Version, err)
	}

	if err := tx.Exec("INSERT INTO "+migrationsTable+" (version, applied_at) VALUES (?, ?)", file.Version, time.Now()).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("record migration %s: %w", file.Version, err)
	}

	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("commit migration %s: %w", file.Version, err)
	}

	return nil
}

func versions(files []migrationFile) []string {
	res := make([]string, 0, len(files))
	for _, f := range files {
		res = append(res, f.Version)
	}
	return res
}

type migrationFile struct {
	Version string
	Path    string
}
