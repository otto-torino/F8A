package utils

import (
	"bufio"
	"image/color"
	"os/exec"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
)

const ShellToUse = "bash"

func AddTextToOutput(text string, c color.Color, outputContainer *fyne.Container) {
	t := canvas.NewText(text, c)
	t.TextSize = 18
	t.TextStyle = fyne.TextStyle{Italic: true}
	outputContainer.Add(t)
	Scroll.ScrollToBottom()
}

func Shellout(command string, outputContainer *fyne.Container, clear bool) error {
	if clear {
		outputContainer.RemoveAll()
	}
	cmd := exec.Command(ShellToUse, "-c", command)
	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()
	AddTextToOutput("Executing "+command, color.RGBA{R: 255, G: 153, B: 0, A: 255}, outputContainer)
	err := cmd.Start()
	if err != nil {
		AddTextToOutput(err.Error(), color.RGBA{R: 255, G: 0, B: 0, A: 255}, outputContainer)
		return err
	}

	scanner := bufio.NewScanner(stderr)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		m := scanner.Text()
		AddTextToOutput(m, color.White, outputContainer)
	}

	scanner = bufio.NewScanner(stdout)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		m := scanner.Text()
		AddTextToOutput(m, color.White, outputContainer)
	}
	err = cmd.Wait()
	if err != nil {
		AddTextToOutput(err.Error(), color.RGBA{R: 255, G: 0, B: 0, A: 255}, outputContainer)
		return err
	}

	AddTextToOutput("Executed "+command, color.RGBA{R: 255, G: 153, B: 0, A: 255}, outputContainer)
	AddTextToOutput("Done! ", color.RGBA{R: 0, G: 255, B: 0, A: 255}, outputContainer)

	return nil
}
