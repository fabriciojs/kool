package cmd

import (
	"bufio"
	"fmt"
	"io"
	"kool-dev/kool/cmd/shell"
	"strings"
	"time"

	"github.com/gookit/color"
)

// KoolTask holds logic for running kool service as a long task
type KoolTask interface {
	KoolService
	Run([]string) error
}

// DefaultKoolTask holds data for running kool service as a long task
type DefaultKoolTask struct {
	KoolService
	message string
	taskOut shell.OutputWriter
}

// NewKoolTask creates a new kool task
func NewKoolTask(message string, service KoolService) *DefaultKoolTask {
	return &DefaultKoolTask{service, message, shell.NewOutputWriter()}
}

// Run runs task
func (t *DefaultKoolTask) Run(args []string) (err error) {
	var (
		lines chan string
	)

	if !t.IsTerminal() {
		return t.Execute(args)
	}

	originalWriter := t.GetWriter()
	t.taskOut.SetWriter(originalWriter)
	pipeReader, pipeWriter := io.Pipe()

	t.KoolService.SetWriter(pipeWriter)
	defer t.KoolService.SetWriter(originalWriter)

	t.taskOut.Println(fmt.Sprintf("%s ...", t.message))
	t.taskOut.Println("")

	lines = make(chan string)

	chErr := readServiceOutput(pipeReader, lines)
	donePrinting := t.printServiceOutput(lines)

	err = <-t.execService(args)
	pipeWriter.Close()
	<-donePrinting
	outputErr := <-chErr

	if outputErr != nil && outputErr != io.EOF {
		t.taskOut.Error(outputErr)
	}

	var statusMessage string
	if err != nil {
		t.taskOut.Error(err)
		statusMessage = fmt.Sprintf("... %s", color.New(color.Red).Sprint("error"))
	} else {
		statusMessage = fmt.Sprintf("... %s", color.New(color.Green).Sprint("done"))
	}

	t.taskOut.Printf("\r")
	t.taskOut.Println(statusMessage)

	return
}

func (t *DefaultKoolTask) execService(args []string) <-chan error {
	err := make(chan error)

	go func() {
		defer close(err)
		err <- t.Execute(args)
	}()

	return err
}

func readServiceOutput(reader io.Reader, lines chan string) chan error {
	bufReader := bufio.NewReader(reader)
	chErr := make(chan error)

	go func() {
		defer func() {
			close(lines)
			close(chErr)
		}()

		var (
			line string
			err  error
		)

		for err == nil {
			if line, err = bufReader.ReadString('\n'); line != "" {
				lines <- strings.TrimSpace(line)
			}
		}

		chErr <- err
	}()

	return chErr
}

func (t *DefaultKoolTask) printServiceOutput(lines chan string) <-chan bool {
	done := make(chan bool)
	spinChars := [4]byte{'-', '/', '|', '\\'}
	spinPos := 0
	currentSpin := spinChars[spinPos : spinPos+1]

	go func() {
		defer close(done)

	OutputPrint:
		for {
			select {
			case line, ok := <-lines:
				t.taskOut.Printf("\r")
				t.taskOut.Println(">", line)
				t.taskOut.Printf("> %s", currentSpin)

				if !ok {
					fmt.Println("break Output")
					break OutputPrint
				}

			case <-time.After(100 * time.Millisecond):
				spinPos = (spinPos + 1) % 4
				currentSpin = spinChars[spinPos : spinPos+1]
				t.taskOut.Printf("\r... %s", currentSpin)
			}
		}

		fmt.Println("done <- true")
		done <- true
	}()

	return done
}
