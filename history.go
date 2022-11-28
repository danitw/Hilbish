package main

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	rt "github.com/arnodel/golua/runtime"
)

type luaHistory struct{}

func (h *luaHistory) Write(line string) (int, error) {
	histWrite := hshMod.Get(rt.StringValue("history")).AsTable().Get(rt.StringValue("add"))
	ln, err := rt.Call1(l.MainThread(), histWrite, rt.StringValue(line))

	var num int64
	if ln.Type() == rt.IntType {
		num = ln.AsInt()
	}

	return int(num), err
}

func (h *luaHistory) GetLine(idx int) (string, error) {
	histGet := hshMod.Get(rt.StringValue("history")).AsTable().Get(rt.StringValue("get"))
	lcmd, err := rt.Call1(l.MainThread(), histGet, rt.IntValue(int64(idx)))

	var cmd string
	if lcmd.Type() == rt.StringType {
		cmd = lcmd.AsString()
	}

	return cmd, err
}

func (h *luaHistory) Len() int {
	histSize := hshMod.Get(rt.StringValue("history")).AsTable().Get(rt.StringValue("size"))
	ln, _ := rt.Call1(l.MainThread(), histSize)

	var num int64
	if ln.Type() == rt.IntType {
		num = ln.AsInt()
	}

	return int(num)
}

func (h *luaHistory) Dump() interface{} {
	// hilbish.history interface already has all function, this isnt used in readline
	return nil
}

type fileHistory struct {
	f     *os.File
	items []string
}

func newFileHistory(path string) *fileHistory {
	dir := filepath.Dir(path)

	err := os.MkdirAll(dir, 0755)
	if err != nil {
		panic(err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if !errors.Is(err, fs.ErrNotExist) {
			panic(err)
		}
	}

	itms := []string{""}
	lines := strings.Split(string(data), "\n")
	for i, l := range lines {
		if i == len(lines) - 1 {
			continue
		}
		itms = append(itms, l)
	}
	f, err := os.OpenFile(path, os.O_APPEND | os.O_WRONLY | os.O_CREATE, 0755)
	if err != nil {
		panic(err)
	}

	fh := &fileHistory{
		items: itms,
		f: f,
	}

	return fh
}

func (h *fileHistory) Write(line string) (int, error) {
	if line == "" {
		return len(h.items), nil
	}

	_, err := h.f.WriteString(line + "\n")
	if err != nil {
		return 0, err
	}
	h.f.Sync()

	h.items = append(h.items, line)
	return len(h.items), nil
}

func (h *fileHistory) GetLine(idx int) (string, error) {
	if len(h.items) == 0 {
		return "", nil
	}
	if idx == -1 { // this should be fixed readline side
		return "", nil
	}
	return h.items[idx], nil
}

func (h *fileHistory) Len() int {
	return len(h.items)
}

func (h *fileHistory) Dump() interface{} {
	return h.items
}

func (h *fileHistory) clear() {
	h.items = []string{}
	h.f.Truncate(0)
	h.f.Sync()
}
