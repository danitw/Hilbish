package main

import (
	"errors"
	"fmt"
	"os"
	"time"

	rt "github.com/arnodel/golua/runtime"
)

type timerType int64

const (
	timerInterval timerType = iota
	timerTimeout
)

type timer struct {
	ticker  *time.Ticker
	ud *rt.UserData
	channel chan struct{}
	fun *rt.Closure
	th *timerHandler
	dur time.Duration
	id int
	typ timerType
	running bool
}

func (t *timer) start() error {
	if t.running {
		return errors.New("timer is already running")
	}

	t.running = true
	t.th.running++
	t.th.wg.Add(1)
	t.ticker = time.NewTicker(t.dur)

	go func() {
		for {
			select {
			case <-t.ticker.C:
				_, err := rt.Call1(l.MainThread(), rt.FunctionValue(t.fun))
				if err != nil {
					fmt.Fprintln(os.Stderr, "Error in function:\n", err)
					t.stop()
				}
				// only run one for timeout
				if t.typ == timerTimeout {
					t.stop()
				}
			case <-t.channel:
				t.ticker.Stop()
				return
			}
		}
	}()

	return nil
}

func (t *timer) stop() error {
	if !t.running {
		return errors.New("timer not running")
	}

	t.channel <- struct{}{}
	t.running = false
	t.th.running--
	t.th.wg.Done()

	return nil
}

func timerStart(thr *rt.Thread, c *rt.GoCont) (rt.Cont, error) {
	if err := c.Check1Arg(); err != nil {
		return nil, err
	}

	t, err := timerArg(c, 0)
	if err != nil {
		return nil, err
	}

	err = t.start()
	if err != nil {
		return nil, err
	}

	return c.Next(), nil
}

func timerStop(thr *rt.Thread, c *rt.GoCont) (rt.Cont, error) {
	if err := c.Check1Arg(); err != nil {
		return nil, err
	}

	t, err := timerArg(c, 0)
	if err != nil {
		return nil, err
	}

	err = t.stop()
	if err != nil {
		return nil, err
	}

	return c.Next(), nil
}
