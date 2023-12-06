// yourapp/pythonintegration/runner.go

package pythonintegration

import (
	"errors"
	"os"
	"os/exec"
)

func getProjectPath() (string, error) {
	path := os.Getenv("PROJECT_PATH")
	if path == "" {
		return "", errors.New("PROJECT_PATH environment variable is not set")
	}
	return path, nil
}

func RunPythonScript(scriptName string, args ...string) error {
	projectPath, err := getProjectPath()
	if err != nil {
		return err
	}

	cmd := exec.Command("python", append([]string{projectPath + "/" + scriptName}, args...)...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return errors.New(string(output))
	}

	return nil
}
