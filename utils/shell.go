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
	t.TextSize = 11
	t.TextStyle = fyne.TextStyle{Italic: true}
	outputContainer.Add(t)
}

func Shellout(command string, outputContainer *fyne.Container, clear bool) {
	if clear {
		outputContainer.RemoveAll()
	}
	cmd := exec.Command(ShellToUse, "-c", command)
	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()
	AddTextToOutput("Executing "+command, color.RGBA{R: 255, G: 153, B: 0, A: 255}, outputContainer)
	cmd.Start()

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
	cmd.Wait()
	AddTextToOutput("Executed "+command, color.RGBA{R: 255, G: 153, B: 0, A: 255}, outputContainer)
	AddTextToOutput("Done! ", color.RGBA{R: 0, G: 255, B: 0, A: 255}, outputContainer)
}
