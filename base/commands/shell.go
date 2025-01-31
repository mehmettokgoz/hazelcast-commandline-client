//go:build std || shell

package commands

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"sync"

	"github.com/hazelcast/hazelcast-commandline-client/clc"
	"github.com/hazelcast/hazelcast-commandline-client/clc/cmd"
	"github.com/hazelcast/hazelcast-commandline-client/clc/paths"
	"github.com/hazelcast/hazelcast-commandline-client/clc/shell"
	puberrors "github.com/hazelcast/hazelcast-commandline-client/errors"
	"github.com/hazelcast/hazelcast-commandline-client/internal"
	"github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
	"github.com/hazelcast/hazelcast-commandline-client/internal/str"
	"github.com/hazelcast/hazelcast-commandline-client/internal/terminal"
)

const banner = `Hazelcast CLC %s (c) 2023 Hazelcast Inc.
		
* Participate in our survey at: https://forms.gle/rPFywdQjvib1QCe49
* Type 'help' for help information. Prefix non-SQL commands with \
		
%s%s

`

type ShellCommand struct {
	shortcuts map[string]struct{}
	mu        sync.RWMutex
}

func (cm *ShellCommand) Init(cc plug.InitContext) error {
	cc.SetCommandUsage("shell")
	help := "Start the interactive shell"
	cc.SetCommandHelp(help, help)
	cc.Hide()
	cm.mu.Lock()
	cm.shortcuts = map[string]struct{}{
		`\di`:   {},
		`\dm`:   {},
		`\dm+`:  {},
		`\exit`: {},
	}
	cm.mu.Unlock()
	return nil
}

func (cm *ShellCommand) Exec(context.Context, plug.ExecContext) error {
	return nil
}

func (cm *ShellCommand) ExecInteractive(ctx context.Context, ec plug.ExecContext) error {
	if len(ec.Args()) > 0 {
		return puberrors.ErrNotAvailable
	}
	m, err := ec.(*cmd.ExecContext).Main().Clone(cmd.ModeInteractive)
	if err != nil {
		return fmt.Errorf("cloning Main: %w", err)
	}
	var cfgText, logText string
	if !terminal.IsPipe(ec.Stdin()) {
		cfgPath := ec.ConfigPath()
		if cfgPath != "" {
			cfgText = fmt.Sprintf("Configuration : %s\n", cfgPath)
		}
		logPath := ec.Props().GetString(clc.PropertyLogPath)
		if logPath != "" {
			logLevel := strings.ToUpper(ec.Props().GetString(clc.PropertyLogLevel))
			logText = fmt.Sprintf("Log %9s : %s", logLevel, logPath)
		}
		check.I2(fmt.Fprintf(ec.Stdout(), banner, internal.Version, cfgText, logText))
		if err = MaybePrintNewVersionNotification(ctx, ec); err != nil {
			ec.Logger().Error(err)
		}
	}
	endLineFn := makeEndLineFunc()
	textFn := makeTextFunc(m, ec, func(shortcut string) bool {
		cm.mu.RLock()
		_, ok := cm.shortcuts[shortcut]
		cm.mu.RUnlock()
		return ok
	})
	path := paths.Join(paths.Home(), "shell.history")
	if terminal.IsPipe(ec.Stdin()) {
		sio := clc.IO{
			Stdin:  ec.Stdin(),
			Stderr: ec.Stderr(),
			Stdout: ec.Stdout(),
		}
		sh := shell.NewOneshotShell(endLineFn, sio, textFn)
		sh.SetCommentPrefix("--")
		return sh.Run(ctx)
	}
	promptFn := func() string {
		cfgPath := ec.ConfigPath()
		if cfgPath == "" {
			return "> "
		}
		// Best effort for absolute path
		p, err := filepath.Abs(cfgPath)
		if err == nil {
			cfgPath = p
		}
		pd := paths.ParentDir(cfgPath)
		return fmt.Sprintf("%s> ", str.MaybeShorten(pd, 12))
	}
	sh, err := shell.New(promptFn, " ... ", path, ec.Stdout(), ec.Stderr(), ec.Stdin(), endLineFn, textFn)
	if err != nil {
		return err
	}
	sh.SetCommentPrefix("--")
	defer sh.Close()
	return sh.Start(ctx)
}

func init() {
	check.Must(plug.Registry.RegisterCommand("shell", &ShellCommand{}))
}
