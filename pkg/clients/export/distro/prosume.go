package distro

import (
	"fmt"
	"os/exec"
)

const (
	prosumeScriptPath     = "/etc/init.d/prosume-cli.sh"
	ProsumeOpStart        = "start"
	ProsumeOpStop         = "stop"
	ProsumeOpGenerateKeys = "generateKeys"
)

func ProsumeClientExec(operation string) (string, error) {
	out, err := exec.Command("sh", prosumeScriptPath, operation).CombinedOutput()
	output := string(out)
	if err != nil {
		err = fmt.Errorf("Prosume %s operation failed: %s", operation, err.Error())
	}
	return output, err
}
