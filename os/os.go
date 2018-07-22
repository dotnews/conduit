package os

import (
	"io/ioutil"
	"os/exec"
)

const (
	sh   = "sh"
	flag = "-c"
)

// Run bash command from stdin and return stdout
func Run(c string, in []byte) ([]byte, error) {
	cmd := exec.Command(sh, flag, c)

	stdIn, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}

	stdOut, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	if err = cmd.Start(); err != nil {
		return nil, err
	}

	if _, err = stdIn.Write(in); err != nil {
		return nil, err
	}

	if err = stdIn.Close(); err != nil {
		return nil, err
	}

	out, err := ioutil.ReadAll(stdOut)
	if err != nil {
		return nil, err
	}

	if err = cmd.Wait(); err != nil {
		return nil, err
	}

	return out, nil
}
