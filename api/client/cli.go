package client

import (
	"flag"
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"
)

type ControllerCli struct {
	in     io.ReadCloser
	out    io.Writer
	err    io.Writer
	proto  string
	addr   string
	scheme string
}

func CreateNewClient(in io.ReadCloser, out, err io.Writer, proto, addr string) (cli *ControllerCli) {
	var scheme = "http"

	cli = &ControllerCli{
		in:     in,
		out:    out,
		err:    err,
		proto:  proto,
		addr:   addr,
		scheme: scheme,
	}
	return
}

func (ctrlCli *ControllerCli) subCmd(name, signature, description string) (flags *flag.FlagSet) {
	flags = flag.NewFlagSet(name, flag.ContinueOnError)
	flags.Usage = func() {
		fmt.Fprintf(ctrlCli.err, "\nUsage: controller %s %s\n\n%s\n\n", name, signature, description)
		flags.PrintDefaults()
		os.Exit(2)
	}
	return
}

func (ctrlCli *ControllerCli) Cmd(args ...string) error {
	if len(args) < 1 {
		return ctrlCli.CmdHelp(args...)
	} else {
		method, exists := ctrlCli.getMethod(args[0])
		if !exists {
			return ctrlCli.CmdHelp(args...)
		} else {
			return method(args[1:]...)
		}
	}
}
func (ctrlCli *ControllerCli) getMethod(name string) (func(...string) error, bool) {
	if len(name) == 0 {
		return nil, false
	}
	methodName := "Cmd" + strings.ToUpper(name[:1]) + strings.ToLower(name[1:])
	method := reflect.ValueOf(ctrlCli).MethodByName(methodName)
	if !method.IsValid() {
		return nil, false
	}
	return method.Interface().(func(...string) error), true
}
