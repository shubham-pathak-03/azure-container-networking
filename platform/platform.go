package platform

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"

	utilexec "k8s.io/utils/exec"
)

type Platform struct {
	exec utilexec.Interface
}

func New(exec utilexec.Interface) *Platform {
	return &Platform{
		exec: exec,
	}
}

func (Platform) CheckIfFileExists(filepath string) (bool, error) {
	_, err := os.Stat(filepath)
	if err == nil {
		return true, nil
	}

	if os.IsNotExist(err) {
		return false, nil
	}

	return true, err
}

func (p Platform) CreateDirectory(dirPath string) error {
	if dirPath == "" {
		log.Printf("dirPath is empty, nothing to create.")
		return nil
	}

	isExist, err := p.CheckIfFileExists(dirPath)
	if err != nil {
		log.Printf("CheckIfFileExists returns err:%v", err)
		return err
	}

	if !isExist {
		err = os.Mkdir(dirPath, os.ModePerm)
	}

	return err
}

func (p Platform) ExecuteCommand(command string) (string, error) {
	log.Printf("[Azure-Utils] %s", command)
	out, err := p.exec.Command("sh", "-c", command).CombinedOutput()
	if err != nil {
		return "", err
	}

	return string(out), nil
}

// readFileByLines reads file line by line and return array of lines.
func readFileByLines(filename string) ([]string, error) {
	var lineStrArr []string

	f, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("Error opening %s file error %v", filename, err)
	}

	defer f.Close()

	r := bufio.NewReader(f)

	for {
		lineStr, err := r.ReadString('\n')
		if err != nil {
			if err != io.EOF {
				return nil, fmt.Errorf("Error reading %s file error %v", filename, err)
			}

			lineStrArr = append(lineStrArr, lineStr)
			break
		}

		lineStrArr = append(lineStrArr, lineStr)
	}

	return lineStrArr, nil
}
