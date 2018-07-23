package os

import (
	"io/ioutil"
	"os/exec"

	"github.com/golang/glog"
)

const (
	sh   = "sh"
	flag = "-c"
)

// Run command in dir from stdin and return stdout
func Run(cmd, dir string, in []byte) ([]byte, error) {
	c := exec.Command(sh, flag, cmd)
	c.Dir = dir

	stdIn, err := c.StdinPipe()
	if err != nil {
		glog.Errorf("Failed getting StdinPipe: '%s' %v", cmd, err)
		return nil, err
	}

	stdOut, err := c.StdoutPipe()
	if err != nil {
		glog.Errorf("Failed getting StdoutPipe: '%s' %v", cmd, err)
		return nil, err
	}

	stdErr, err := c.StderrPipe()
	if err != nil {
		glog.Errorf("Failed getting StderrPipe: '%s' %v", cmd, err)
		return nil, err
	}

	if err = c.Start(); err != nil {
		glog.Errorf("Failed starting command: '%s' %v", cmd, err)
		return nil, err
	}

	if _, err = stdIn.Write(in); err != nil {
		glog.Errorf("Failed writing to stdIn: '%s' %v", cmd, err)
		return nil, err
	}

	if err = stdIn.Close(); err != nil {
		glog.Errorf("Failed closing stdIn: '%s' %v", cmd, err)
		return nil, err
	}

	out, err := ioutil.ReadAll(stdOut)
	if err != nil {
		glog.Errorf("Failed reading from stdOut: '%s' %v", cmd, err)
		return nil, err
	}

	e, err := ioutil.ReadAll(stdErr)
	if err != nil {
		glog.Errorf("Failed reading from stdErr: '%s' %v", cmd, err)
		return nil, err
	}

	if len(e) > 0 {
		glog.Errorf("Failed executing command: '%s' %s", cmd, string(e))
	}

	if err = c.Wait(); err != nil {
		glog.Errorf("Failed waiting for command: '%s' %v", cmd, err)
		return nil, err
	}

	return out, nil
}
