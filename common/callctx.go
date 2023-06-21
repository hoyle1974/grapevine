package common

import (
	"fmt"
	"path"
	"runtime"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type CallCtx struct {
	ctx string
	log zerolog.Logger
}

func NewCallCtx(log zerolog.Logger) CallCtx {
	return CallCtx{log: log}
}

func NewCallCtxWithApp(app string) CallCtx {
	log := log.With().Str("app", "tictactoe").Logger()
	return NewCallCtx(log)
}

func (c CallCtx) NewCtx(name string) CallCtx {
	if len(c.ctx) > 0 {
		name = c.ctx + ":" + name
	}
	log := c.log.With().Str("ctx", name).Logger()

	return CallCtx{ctx: name, log: log}
}

// func (c CallCtx) Func(name string) CallCtx {
// 	l := c.log.With().Str("func", "CreateAccount").Logger()

// 	return CallCtx{ctx: c.ctx, log: l}
// }

func (c CallCtx) Log() *zerolog.Logger {
	return &c.log
}

func where(e *zerolog.Event) *zerolog.Event {

	where := ""
	_, file, line, ok := runtime.Caller(2)
	if ok {

		where = fmt.Sprintf("%s:%d", path.Base(file), line)
	}

	return e.Str("where", where)
}

func (c CallCtx) Trace() *zerolog.Event {
	return c.log.Trace()
}

func (c CallCtx) Debug() *zerolog.Event {
	return where(c.log.Debug())
}

func (c CallCtx) Info() *zerolog.Event {
	return c.log.Info()
}

func (c CallCtx) Warn() *zerolog.Event {
	return where(c.log.Warn())
}

func (c CallCtx) Error() *zerolog.Event {
	return where(c.log.Warn())
}

func (c CallCtx) Fatal() *zerolog.Event {
	return where(c.log.Fatal())
}

func (c CallCtx) Panic() *zerolog.Event {
	return where(c.log.Panic())
}

func (c CallCtx) Printf(format string, v ...interface{}) {
	if e := where(c.log.Debug()); e.Enabled() {
		e.CallerSkipFrame(1).Msg(fmt.Sprintf(format, v...))
	}
}
