package utils

import (
	"bufio"
	"errors"
	"image/color"
	"os/exec"
	"strings"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"github.com/spf13/viper"
)

const ShellToUse = "bash"

// buildShellCommand prepends the optional "shell.init" snippet from the
// settings file to the command. This lets commands run in a non-interactive
// shell (e.g. when the app is launched from the desktop menu) find tools
// that are only set up by shell init scripts, such as nvm-managed yarn.
func buildShellCommand(command string) string {
	initSnippet := strings.TrimSpace(viper.GetString("shell.init"))
	if initSnippet == "" {
		return command
	}
	return initSnippet + " && " + command
}

// shellCommand returns an exec.Cmd running the command through the shell,
// including the configured init snippet.
func shellCommand(command string) *exec.Cmd {
	return exec.Command(ShellToUse, "-c", buildShellCommand(command))
}

func ExecCommand(command string) (string, error) {
	cmd := shellCommand(command)
	output, err := cmd.CombinedOutput()
	return string(output), err
}

func AddTextToOutput(text string, c color.Color, outputContainer *fyne.Container) {
	t := canvas.NewText(text, c)
	t.TextSize = 14
	t.TextStyle = fyne.TextStyle{Italic: true}
	outputContainer.Add(t)
	Scroll.ScrollToBottom()
}

func Shellout(description string, command string, outputContainer *fyne.Container, clear bool) error {
	if clear {
		outputContainer.RemoveAll()
	}
	cmd := shellCommand(command)
	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()
	AddTextToOutput("Executing: "+description, color.RGBA{R: 255, G: 153, B: 0, A: 255}, outputContainer)
	AddTextToOutput("Command: "+command, color.RGBA{R: 135, G: 206, B: 250, A: 255}, outputContainer)
	err := cmd.Start()
	if err != nil {
		AddTextToOutput(err.Error(), color.RGBA{R: 255, G: 0, B: 0, A: 255}, outputContainer)
		return err
	}

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			m := scanner.Text()
			AddTextToOutput(m, color.White, outputContainer)
		}
	}()

	go func() {
		defer wg.Done()
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			m := scanner.Text()
			AddTextToOutput(m, color.White, outputContainer)
		}
	}()

	wg.Wait()
	err = cmd.Wait()
	if err != nil {
		AddTextToOutput(err.Error(), color.RGBA{R: 255, G: 0, B: 0, A: 255}, outputContainer)
		return err
	}

	AddTextToOutput("Executed "+command, color.RGBA{R: 255, G: 153, B: 0, A: 255}, outputContainer)
	AddTextToOutput("Done! ", color.RGBA{R: 0, G: 255, B: 0, A: 255}, outputContainer)

	return nil
}

func Shell(command string) (*[]string, error) {
	output := []string{}
	cmd := shellCommand(command)
	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()
	err := cmd.Start()
	if err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(stderr)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		m := scanner.Text()
		return nil, errors.New(m)
	}

	scanner = bufio.NewScanner(stdout)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		m := scanner.Text()
		output = append(output, m)
	}
	err = cmd.Wait()
	if err != nil {
		return nil, err
	}

	return &output, nil
}
