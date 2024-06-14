package signal_handler

import (
	"context"
	"log"
	"os"
	"os/signal"
	"reflect"
	"runtime"
	"strings"
	"sync"
)

const structure string = "SignalHandler"

type SignalHandler struct {
	signalChan    chan os.Signal
	callbackFuncs []func(os.Signal)
	logger        *log.Logger
	ctx           context.Context
	cancel        context.CancelFunc
	running       bool
	wg            *sync.WaitGroup
}

func (sh *SignalHandler) watch() {
	const function string = "watch"
	sh.logger.Printf("%s %s: %s\n", structure, function, "started")
	var sig os.Signal
	sh.logger.Printf("%s %s: %s\n", structure, function, "waiting for a signal or a context cancel")
	select {
	case sig = <-sh.signalChan:
		sh.logger.Printf("%s %s: got the signal %s\n", structure, function, sig.String())
		var fName string
		for _, f := range sh.callbackFuncs {
			fName = runtime.FuncForPC(reflect.ValueOf(f).Pointer()).Name()
			sh.logger.Printf("%s %s: execute the callback function named %s\n", structure, function, fName)
			f(sig)
			sh.logger.Printf("%s %s: callback function named %s executed\n", structure, function, fName)
		}
		sh.cancel()
	case <-sh.ctx.Done():
		sh.logger.Printf("%s %s: local context cancelled\n", structure, function)
		signal.Reset()
	}
	sh.logger.Printf("%s %s: stopped\n", structure, function)
	sh.wg.Done()
}
func (sh *SignalHandler) RegisterCallback(f func(os.Signal)) error {
	if sh.running {
		return ErrAlreadyStarted
	} else {
		sh.callbackFuncs = append(sh.callbackFuncs, f)
		sh.logger.Printf("%s %s: callback named %s registered\n", structure, "RegisterCallback", runtime.FuncForPC(reflect.ValueOf(f).Pointer()).Name())
		return nil
	}
}
func (sh *SignalHandler) StartOn(signals []os.Signal) error {
	if sh.running {
		return ErrAlreadyStarted
	} else {
		sh.wg.Add(1)
		sh.running = true
		var sigsName string
		for _, s := range signals {
			sigsName += s.String() + ", "
		}
		sigsName = strings.Trim(sigsName, ", ")
		sh.logger.Printf("%s %s: starting to watch over the signals %s\n", structure, "StartOn", sigsName)
		sh.ctx, sh.cancel = context.WithCancel(context.Background())
		signal.Notify(sh.signalChan, signals...)
		go sh.watch()
		return nil
	}
}
func (sh *SignalHandler) Stop() error {
	if !sh.running {
		return ErrNotRunning
	} else {
		sh.logger.Printf("%s Stop: cancelling the local context\n", structure)
		sh.cancel()
		sh.running = false
		return nil
	}
}
func (sh *SignalHandler) Wait() {
	sh.wg.Wait()
}

func NewSignalHandler(logger *log.Logger) *SignalHandler {
	return &SignalHandler{
		signalChan:    make(chan os.Signal, 1),
		callbackFuncs: make([]func(os.Signal), 0),
		logger:        logger,
		ctx:           nil,
		cancel:        nil,
		running:       false,
		wg:            &sync.WaitGroup{},
	}
}
