package logger

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

// DailyHandlerOptions 配置脚手架自定义 slog 处理器。
type DailyHandlerOptions struct {
	ProjectName   string
	BaseDir       string
	Dir           string
	Level         slog.Level
	AddSource     bool
	Console       bool
	ConsoleOut    io.Writer
	RetentionDays int
}

// DailyHandler 将制表符分隔的日志行写入按日期划分的文件，并可同步输出到控制台。
type DailyHandler struct {
	core   *dailyCore
	attrs  []slog.Attr
	groups []string
}

type dailyCore struct {
	mu            sync.Mutex
	projectName   string
	baseDir       string
	dir           string
	level         slog.Level
	addSource     bool
	console       bool
	consoleOut    io.Writer
	retentionDays int
	currentDate   string
	file          *os.File
}

// NewDailyHandler 打开当天日志文件并初始化处理器。
func NewDailyHandler(opts DailyHandlerOptions) (*DailyHandler, error) {
	if opts.ProjectName == "" {
		opts.ProjectName = "app"
	}
	if opts.Dir == "" {
		opts.Dir = "logs"
	}
	if opts.ConsoleOut == nil {
		opts.ConsoleOut = os.Stdout
	}
	if opts.RetentionDays < 0 {
		return nil, fmt.Errorf("日志保留天数不能小于 0")
	}

	core := &dailyCore{
		projectName:   opts.ProjectName,
		baseDir:       opts.BaseDir,
		dir:           opts.Dir,
		level:         opts.Level,
		addSource:     opts.AddSource,
		console:       opts.Console,
		consoleOut:    opts.ConsoleOut,
		retentionDays: opts.RetentionDays,
	}
	if err := core.rotate(time.Now()); err != nil {
		return nil, err
	}

	return &DailyHandler{core: core}, nil
}

// Enabled 判断指定级别的日志是否需要输出。

func (h *DailyHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.core.level
}

// Handle 串行化写入，保证日期轮转和文件追加顺序稳定。

func (h *DailyHandler) Handle(_ context.Context, record slog.Record) error {
	if !h.Enabled(context.Background(), record.Level) {
		return nil
	}

	h.core.mu.Lock()
	defer h.core.mu.Unlock()

	rotateAt := record.Time
	if rotateAt.IsZero() {
		rotateAt = time.Now()
	}
	if err := h.core.rotate(rotateAt); err != nil {
		return err
	}

	line := h.format(record)
	if _, err := h.core.file.Write(line); err != nil {
		return err
	}
	if h.core.console {
		_, _ = h.core.consoleOut.Write(line)
	}
	return nil
}

// WithAttrs 返回携带固定结构化属性的新处理器。

func (h *DailyHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	next := *h
	next.attrs = append([]slog.Attr{}, h.attrs...)
	next.attrs = appendQualifiedAttrs(next.attrs, h.groups, attrs)
	return &next
}

// WithGroup 返回将后续属性归入指定分组的新处理器。
func (h *DailyHandler) WithGroup(name string) slog.Handler {
	if name == "" {
		return h
	}
	next := *h
	next.groups = append(append([]string{}, h.groups...), name)
	return &next
}

// Close 关闭当前日志文件；重复调用不会产生错误。

func (h *DailyHandler) Close() error {
	h.core.mu.Lock()
	defer h.core.mu.Unlock()
	if h.core.file == nil {
		return nil
	}
	err := h.core.file.Close()
	h.core.file = nil
	return err
}

// format 生成紧凑、便于文本检索的日志行，同时保留结构化属性。
func (h *DailyHandler) format(record slog.Record) []byte {
	var buf bytes.Buffer
	when := record.Time
	if !when.IsZero() {
		buf.WriteString(when.Local().Format("2006-01-02 15:04:05.000"))
	}
	buf.WriteByte('\t')
	buf.WriteString(record.Level.String())
	buf.WriteByte('\t')

	if h.core.addSource && record.PC != 0 {
		buf.WriteString(h.source(record.PC))
		buf.WriteByte('\t')
	}

	buf.WriteString(record.Message)
	attrs := make([]slog.Attr, 0, len(h.attrs)+record.NumAttrs())
	attrs = append(attrs, h.attrs...)
	record.Attrs(func(attr slog.Attr) bool {
		attrs = appendQualifiedAttr(attrs, h.groups, attr)
		return true
	})
	if len(attrs) > 0 {
		buf.WriteByte('\t')
		h.appendAttrs(&buf, attrs)
	}
	buf.WriteByte('\n')
	return buf.Bytes()
}

func (h *DailyHandler) appendAttrs(buf *bytes.Buffer, attrs []slog.Attr) {
	for i, attr := range attrs {
		if i > 0 {
			buf.WriteByte(' ')
		}
		buf.WriteString(attr.Key)
		buf.WriteByte('=')
		buf.WriteString(attrValue(attr.Value))
	}
}

func appendQualifiedAttrs(dst []slog.Attr, groups []string, attrs []slog.Attr) []slog.Attr {
	for _, attr := range attrs {
		dst = appendQualifiedAttr(dst, groups, attr)
	}
	return dst
}

func appendQualifiedAttr(dst []slog.Attr, groups []string, attr slog.Attr) []slog.Attr {
	attr.Value = attr.Value.Resolve()
	if attr.Equal(slog.Attr{}) {
		return dst
	}

	if attr.Value.Kind() == slog.KindGroup {
		nestedGroups := groups
		if attr.Key != "" {
			nestedGroups = append(append([]string{}, groups...), attr.Key)
		}
		return appendQualifiedAttrs(dst, nestedGroups, attr.Value.Group())
	}

	if len(groups) > 0 {
		attr.Key = strings.Join(groups, ".") + "." + attr.Key
	}
	return append(dst, attr)
}

func (h *DailyHandler) source(pc uintptr) string {
	if pc == 0 {
		return "unknown:0"
	}
	frame, _ := runtime.CallersFrames([]uintptr{pc}).Next()
	file := frame.File
	if h.core.baseDir != "" {
		if rel, err := filepath.Rel(h.core.baseDir, file); err == nil && rel != "." && !strings.HasPrefix(rel, "..") {
			file = rel
		}
	}
	return fmt.Sprintf("%s:%d", file, frame.Line)
}

// attrValue 针对常见 slog 值类型做轻量格式化，避免回退成 JSON。
func attrValue(value slog.Value) string {
	switch value.Kind() {
	case slog.KindString:
		return strconv.Quote(value.String())
	case slog.KindInt64:
		return strconv.FormatInt(value.Int64(), 10)
	case slog.KindUint64:
		return strconv.FormatUint(value.Uint64(), 10)
	case slog.KindFloat64:
		return strconv.FormatFloat(value.Float64(), 'f', -1, 64)
	case slog.KindBool:
		return strconv.FormatBool(value.Bool())
	case slog.KindDuration:
		return value.Duration().String()
	case slog.KindTime:
		return value.Time().Format(time.RFC3339)
	case slog.KindAny:
		return fmt.Sprintf("%v", value.Any())
	default:
		return value.String()
	}
}

// rotate 在日志记录跨越本地日期边界时切换文件。
func (c *dailyCore) rotate(now time.Time) error {
	return c.rotateWithCleanup(now, c.cleanupExpiredLogs)
}

// rotateWithCleanup 在日期变化时切换到当天的日志文件，并执行过期日志清理。
func (c *dailyCore) rotateWithCleanup(now time.Time, cleanup func(string, time.Time) error) error {
	// 当前文件已经对应当天时，无需重复打开日志文件。
	date := now.Local().Format("2006-01-02")
	if c.file != nil && c.currentDate == date {
		return nil
	}

	// 日志目录支持相对路径；目录不存在时自动创建。
	dir := filepath.Join(c.baseDir, c.dir)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("创建日志目录失败: %w", err)
	}

	// 以追加模式打开当天的日志文件，避免覆盖已有内容。
	path := filepath.Join(dir, fmt.Sprintf("%s-%s.log", c.projectName, date))
	file, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("打开日志文件失败: %w", err)
	}

	// 新文件打开成功后再替换并关闭旧文件，防止打开失败时丢失当前文件句柄。
	previousFile := c.file
	c.currentDate = date
	c.file = file
	if previousFile != nil {
		_ = previousFile.Close()
	}

	// 清理失败只记录错误，不影响当天日志继续写入。
	if err := cleanup(dir, now); err != nil {
		c.reportCleanupError(now, err)
	}
	return nil
}

// reportCleanupError 将日志清理失败信息写入当前日志文件，并同步输出到控制台。
func (c *dailyCore) reportCleanupError(now time.Time, err error) {
	// 使用固定的 ERROR 级别格式，并对错误文本加引号、转义特殊字符。
	line := []byte(fmt.Sprintf(
		"%s\tERROR\t清理过期日志失败\terror=%s\n",
		now.Local().Format("2006-01-02 15:04:05.000"),
		strconv.Quote(err.Error()),
	))

	// 日志文件可用时记录清理错误；写入失败不再递归上报。
	if c.file != nil {
		_, _ = c.file.Write(line)
	}
	// 无论普通日志是否开启控制台输出，都尽量让清理错误可见。
	if c.consoleOut != nil {
		_, _ = c.consoleOut.Write(line)
	}
}

// cleanupExpiredLogs 仅删除当前项目早于保留窗口的按日日志文件。
func (c *dailyCore) cleanupExpiredLogs(dir string, now time.Time) error {
	if c.retentionDays == 0 {
		return nil
	}

	localNow := now.Local()
	startOfToday := time.Date(localNow.Year(), localNow.Month(), localNow.Day(), 0, 0, 0, 0, localNow.Location())
	cutoff := startOfToday.AddDate(0, 0, -(c.retentionDays - 1))
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("读取日志目录失败: %w", err)
	}

	prefix := c.projectName + "-"
	for _, entry := range entries {
		name := entry.Name()
		if !strings.HasPrefix(name, prefix) || !strings.HasSuffix(name, ".log") {
			continue
		}

		dateText := strings.TrimSuffix(strings.TrimPrefix(name, prefix), ".log")
		if len(dateText) != len("2006-01-02") {
			continue
		}
		logDate, err := time.ParseInLocation("2006-01-02", dateText, localNow.Location())
		if err != nil || logDate.Format("2006-01-02") != dateText || !logDate.Before(cutoff) {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			return fmt.Errorf("读取日志文件信息 %s 失败: %w", name, err)
		}
		if !info.Mode().IsRegular() {
			continue
		}
		if err := os.Remove(filepath.Join(dir, name)); err != nil {
			return fmt.Errorf("删除过期日志文件 %s 失败: %w", name, err)
		}
	}
	return nil
}
