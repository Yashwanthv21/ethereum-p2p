package log

import (
	"fmt"
	"os"
	"time"

	"github.com/go-stack/stack"
)

const timeKey = "t"
const lvlKey = "lvl"
const msgKey = "msg"
const errorKey = "LOG15_ERROR"

type Lvl int

const (
	LvlCrit Lvl = iota
	LvlError
	LvlWarn
	LvlInfo
	LvlDebug
	LvlTrace
	LvlSql
	LvlMessageRx
	LvlMessageTx
	LvlNeighbors
	LvlTask
	LvlPeer
	LvlHello
	LvlDiscProto
	LvlDiscPeer
	LvlStatus
	LvlDaoFork
)

// Aligned returns a 5-character string containing the name of a Lvl.
func (l Lvl) AlignedString() string {
	switch l {
	case LvlDaoFork:
		return "DAOFORK"
	case LvlStatus:
		return "STATUS"
	case LvlDiscPeer:
		return "DISCPEER"
	case LvlDiscProto:
		return "DISCPROTO"
	case LvlHello:
		return "HELLO"
	case LvlPeer:
		return "PEER"
	case LvlTask:
		return "TASK"
	case LvlNeighbors:
		return "NEIGHBORS"
	case LvlMessageTx:
		return "MSGTX"
	case LvlMessageRx:
		return "MSGRX"
	case LvlSql:
		return "SQL"
	case LvlTrace:
		return "TRACE"
	case LvlDebug:
		return "DEBUG"
	case LvlInfo:
		return "INFO"
	case LvlWarn:
		return "WARN"
	case LvlError:
		return "ERROR"
	case LvlCrit:
		return "CRIT"
	default:
		panic("bad level")
	}
}

// Strings returns the name of a Lvl.
func (l Lvl) String() string {
	switch l {
	case LvlDaoFork:
		return "daofork"
	case LvlStatus:
		return "status"
	case LvlDiscPeer:
		return "disc-peer"
	case LvlDiscProto:
		return "disc-proto"
	case LvlHello:
		return "hello"
	case LvlPeer:
		return "peer"
	case LvlTask:
		return "task"
	case LvlNeighbors:
		return "neighbors"
	case LvlMessageTx:
		return "message-sent"
	case LvlMessageRx:
		return "message-received"
	case LvlSql:
		return "sql"
	case LvlTrace:
		return "trce"
	case LvlDebug:
		return "dbug"
	case LvlInfo:
		return "info"
	case LvlWarn:
		return "warn"
	case LvlError:
		return "eror"
	case LvlCrit:
		return "crit"
	default:
		panic("bad level")
	}
}

// Returns the appropriate Lvl from a string name.
// Useful for parsing command line args and configuration files.
func LvlFromString(lvlString string) (Lvl, error) {
	switch lvlString {
	case "daofork":
		return LvlDaoFork, nil
	case "status":
		return LvlStatus, nil
	case "disc-peer":
		return LvlDiscPeer, nil
	case "disc-proto":
		return LvlDiscProto, nil
	case "hello":
		return LvlHello, nil
	case "peer":
		return LvlPeer, nil
	case "task":
		return LvlTask, nil
	case "neighbors":
		return LvlNeighbors, nil
	case "message-sent", "message-tx", "msg-tx":
		return LvlMessageTx, nil
	case "message-received", "message-rx", "msg-rx":
		return LvlMessageRx, nil
	case "sql":
		return LvlSql, nil
	case "trace", "trce":
		return LvlTrace, nil
	case "debug", "dbug":
		return LvlDebug, nil
	case "info":
		return LvlInfo, nil
	case "warn":
		return LvlWarn, nil
	case "error", "eror":
		return LvlError, nil
	case "crit":
		return LvlCrit, nil
	default:
		return LvlDebug, fmt.Errorf("Unknown level: %v", lvlString)
	}
}

// A Record is what a Logger asks its handler to write
type Record struct {
	Time     time.Time
	Lvl      Lvl
	Msg      string
	Ctx      []interface{}
	Call     stack.Call
	KeyNames RecordKeyNames
}

type RecordKeyNames struct {
	Time string
	Msg  string
	Lvl  string
}

// A Logger writes key/value pairs to a Handler
type Logger interface {
	// New returns a new Logger that has this logger's context plus the given context
	New(ctx ...interface{}) Logger

	// GetHandler gets the handler associated with the logger.
	GetHandler() Handler

	// SetHandler updates the logger to write records to the specified handler.
	SetHandler(h Handler)

	SetGlogger(glogger *GlogHandler)

	GetGlogger() *GlogHandler

	// Log a message at the given level with context key/value pairs
	DaoFork(t time.Time, connInfoCtx []interface{}, support bool)
	Status(t time.Time, connInfoCtx []interface{}, helloStr string, statusStr string)
	DiscPeer(t time.Time, connInfoCtx []interface{}, discReason string)
	DiscProto(t time.Time, connInfoCtx []interface{}, discReason string)
	Hello(t time.Time, connInfoCtx []interface{}, helloStr string)
	Peer(msg string, connInfoCtx []interface{})
	Task(msg string, taskInfoCtx []interface{})
	Neighbors(t time.Time, ctx []interface{}, neighbor interface{})
	MessageTx(t time.Time, msgType string, size uint32, connInfoCtx []interface{}, err error)
	MessageRx(t time.Time, msgType string, size uint32, connInfoCtx []interface{}, err error)
	Sql(msg string, ctx ...interface{})
	Trace(msg string, ctx ...interface{})
	Debug(msg string, ctx ...interface{})
	Info(msg string, ctx ...interface{})
	Warn(msg string, ctx ...interface{})
	Error(msg string, ctx ...interface{})
	Crit(msg string, ctx ...interface{})
}

type logger struct {
	ctx     []interface{}
	h       *swapHandler
	datadir string
	glogger *GlogHandler
}

func (l *logger) write(msg string, lvl Lvl, ctx []interface{}) {
	l.h.Log(&Record{
		Time: time.Now(),
		Lvl:  lvl,
		Msg:  msg,
		Ctx:  newContext(l.ctx, ctx),
		Call: stack.Caller(2),
		KeyNames: RecordKeyNames{
			Time: timeKey,
			Msg:  msgKey,
			Lvl:  lvlKey,
		},
	})
}

func (l *logger) New(ctx ...interface{}) Logger {
	child := &logger{ctx: newContext(l.ctx, ctx), h: new(swapHandler)}
	child.SetHandler(l.h)
	return child
}

func newContext(prefix []interface{}, suffix []interface{}) []interface{} {
	normalizedSuffix := normalize(suffix)
	newCtx := make([]interface{}, len(prefix)+len(normalizedSuffix))
	n := copy(newCtx, prefix)
	copy(newCtx[n:], normalizedSuffix)
	return newCtx
}

func (l *logger) writeTimeMsgType(lvl Lvl, t time.Time, msgType string, size uint32, connInfoCtx []interface{}, err error) {
	ctx := []interface{}{
		"size", size,
	}
	ctx = append(ctx, connInfoCtx...)
	if err != nil {
		ctx = append(ctx, "err", err)
	}
	l.write(fmt.Sprintf("%.6f|%s", float64(t.UnixNano())/1e9, msgType), lvl, ctx)
}

func (l *logger) writeTime(lvl Lvl, t time.Time, ctx []interface{}) {
	l.write(fmt.Sprintf("%.6f", float64(t.UnixNano())/1e9), lvl, ctx)
}

func (l *logger) DaoFork(t time.Time, connInfoCtx []interface{}, support bool) {
	ctx := append(connInfoCtx,
		"support", support,
	)
	l.writeTime(LvlDaoFork, t, ctx)
}

func (l *logger) Status(t time.Time, connInfoCtx []interface{}, helloStr string, statusStr string) {
	ctx := append(connInfoCtx,
		"hello", helloStr,
		"status", statusStr,
	)
	l.writeTime(LvlStatus, t, ctx)
}

func (l *logger) DiscPeer(t time.Time, connInfoCtx []interface{}, discReason string) {
	ctx := append(connInfoCtx,
		"discReason", discReason,
	)
	l.writeTime(LvlDiscPeer, t, ctx)
}

func (l *logger) DiscProto(t time.Time, connInfoCtx []interface{}, discReason string) {
	ctx := append(connInfoCtx,
		"discReason", discReason,
	)
	l.writeTime(LvlDiscProto, t, ctx)
}

func (l *logger) Hello(t time.Time, connInfoCtx []interface{}, helloStr string) {
	ctx := append(connInfoCtx,
		"hello", helloStr,
	)
	l.writeTime(LvlHello, t, ctx)
}

func (l *logger) Peer(msg string, connInfoCtx []interface{}) {
	l.write(msg, LvlPeer, connInfoCtx)
}

func (l *logger) Task(msg string, taskInfoCtx []interface{}) {
	l.write(msg, LvlTask, taskInfoCtx)
}

func (l *logger) Neighbors(t time.Time, connInfoCtx []interface{}, neighbor interface{}) {
	ctx := append(connInfoCtx,
		"neighbor", neighbor,
	)
	l.writeTime(LvlNeighbors, t, ctx)
}

func (l *logger) MessageTx(t time.Time, msgType string, size uint32, connInfoCtx []interface{}, err error) {
	l.writeTimeMsgType(LvlMessageTx, t, msgType, size, connInfoCtx, err)
}

func (l *logger) MessageRx(t time.Time, msgType string, size uint32, connInfoCtx []interface{}, err error) {
	l.writeTimeMsgType(LvlMessageRx, t, msgType, size, connInfoCtx, err)
}

func (l *logger) Sql(msg string, ctx ...interface{}) {
	l.write(msg, LvlSql, ctx)
}

func (l *logger) Trace(msg string, ctx ...interface{}) {
	l.write(msg, LvlTrace, ctx)
}

func (l *logger) Debug(msg string, ctx ...interface{}) {
	l.write(msg, LvlDebug, ctx)
}

func (l *logger) Info(msg string, ctx ...interface{}) {
	l.write(msg, LvlInfo, ctx)
}

func (l *logger) Warn(msg string, ctx ...interface{}) {
	l.write(msg, LvlWarn, ctx)
}

func (l *logger) Error(msg string, ctx ...interface{}) {
	l.write(msg, LvlError, ctx)
}

func (l *logger) Crit(msg string, ctx ...interface{}) {
	l.write(msg, LvlCrit, ctx)
	os.Exit(1)
}

func (l *logger) GetHandler() Handler {
	return l.h.Get()
}

func (l *logger) SetHandler(h Handler) {
	l.h.Swap(h)
}

func (l *logger) GetGlogger() *GlogHandler {
	return l.glogger
}

func (l *logger) SetGlogger(glogger *GlogHandler) {
	l.glogger = glogger
}

func normalize(ctx []interface{}) []interface{} {
	// if the caller passed a Ctx object, then expand it
	if len(ctx) == 1 {
		if ctxMap, ok := ctx[0].(Ctx); ok {
			ctx = ctxMap.toArray()
		}
	}

	// ctx needs to be even because it's a series of key/value pairs
	// no one wants to check for errors on logging functions,
	// so instead of erroring on bad input, we'll just make sure
	// that things are the right length and users can fix bugs
	// when they see the output looks wrong
	if len(ctx)%2 != 0 {
		ctx = append(ctx, nil, errorKey, "Normalized odd number of arguments by adding nil")
	}

	return ctx
}

// Lazy allows you to defer calculation of a logged value that is expensive
// to compute until it is certain that it must be evaluated with the given filters.
//
// Lazy may also be used in conjunction with a Logger's New() function
// to generate a child logger which always reports the current value of changing
// state.
//
// You may wrap any function which takes no arguments to Lazy. It may return any
// number of values of any type.
type Lazy struct {
	Fn interface{}
}

// Ctx is a map of key/value pairs to pass as context to a log function
// Use this only if you really need greater safety around the arguments you pass
// to the logging functions.
type Ctx map[string]interface{}

func (c Ctx) toArray() []interface{} {
	arr := make([]interface{}, len(c)*2)

	i := 0
	for k, v := range c {
		arr[i] = k
		arr[i+1] = v
		i += 2
	}

	return arr
}
