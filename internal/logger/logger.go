package logger

import (
	"fmt"
	"time"

	"github.com/fatih/color"
)

var (
	yellow = color.New(color.FgYellow, color.Bold).SprintFunc()
	green  = color.New(color.FgGreen, color.Bold).SprintFunc()
	red    = color.New(color.FgRed, color.Bold).SprintFunc()
	cyan   = color.New(color.FgCyan, color.Bold).SprintFunc()
)

func Init(format string, args ...any) {
	fmt.Printf("%s %s\n", yellow("[INIT]"), fmt.Sprintf(format, args...))
}

func Live(ok bool, format string, args ...any) {
	prefix := green("[LIVE]")
	if !ok {
		prefix = red("[LIVE]")
	}
	fmt.Printf("%s %s\n", prefix, fmt.Sprintf(format, args...))
}

func Test(ok bool, format string, args ...any) {
	prefix := green("[TEST]")
	if !ok {
		prefix = red("[TEST]")
	}
	fmt.Printf("%s %s\n", prefix, fmt.Sprintf(format, args...))
}

func Agent(format string, args ...any) {
	fmt.Printf("%s %s\n", cyan("[AGENT]"), fmt.Sprintf(format, args...))
}

func AgentLine(line string) {
	fmt.Printf("%s %s\n", cyan("[AGENT]>"), line)
}

func Restart(format string, args ...any) {
	fmt.Printf("%s %s\n", yellow("[RESTART]"), fmt.Sprintf(format, args...))
}

func Stop(format string, args ...any) {
	fmt.Printf("%s %s\n", yellow("[STOP]"), fmt.Sprintf(format, args...))
}

func Error(format string, args ...any) {
	fmt.Printf("%s %s\n", red("[ERROR]"), fmt.Sprintf(format, args...))
}

// LiveThrottle logs liveness progress at most once every 5 seconds.
type LiveThrottle struct {
	last time.Time
}

func (t *LiveThrottle) Log(elapsed time.Duration) {
	now := time.Now()
	if now.Sub(t.last) >= 5*time.Second {
		t.last = now
		Live(false, "waiting... %ds elapsed", int(elapsed.Seconds()))
	}
}
