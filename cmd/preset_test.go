package cmd

import (
	"errors"
	"fmt"
	"kool-dev/kool/cmd/compose"
	"kool-dev/kool/cmd/presets"
	"kool-dev/kool/cmd/shell"
	"testing"
)

func newFakeKoolPreset() *KoolPreset {
	return &KoolPreset{
		*newFakeKoolService(),
		&KoolPresetFlags{false},
		&presets.FakeParser{},
		&compose.FakeParser{},
		&shell.FakePromptSelect{},
	}
}

func TestNewKoolPreset(t *testing.T) {
	k := NewKoolPreset()

	if _, ok := k.DefaultKoolService.out.(*shell.DefaultOutputWriter); !ok {
		t.Errorf("unexpected shell.OutputWriter on default KoolPreset instance")
	}

	if _, ok := k.DefaultKoolService.exiter.(*shell.DefaultExiter); !ok {
		t.Errorf("unexpected shell.Exiter on default KoolPreset instance")
	}

	if _, ok := k.DefaultKoolService.in.(*shell.DefaultInputReader); !ok {
		t.Errorf("unexpected shell.InputReader on default KoolPreset instance")
	}

	if k.Flags == nil {
		t.Errorf("Flags not initialized on default KoolPreset instance")
	} else if k.Flags.Override {
		t.Errorf("bad default value for Override flag on default KoolPreset instance")
	}

	if _, ok := k.presetsParser.(*presets.DefaultParser); !ok {
		t.Errorf("unexpected presets.Parser on default KoolPreset instance")
	}

	if _, ok := k.composeParser.(*compose.DefaultParser); !ok {
		t.Errorf("unexpected compose.Parser on default KoolPreset instance")
	}

	if _, ok := k.promptSelect.(*shell.DefaultPromptSelect); !ok {
		t.Errorf("unexpected shell.PromptSelect on default KoolPreset instance")
	}

	if _, ok := k.DefaultKoolService.term.(*shell.DefaultTerminalChecker); !ok {
		t.Errorf("unexpected shell.TerminalChecker on default KoolPreset instance")
	}
}

func TestPresetCommand(t *testing.T) {
	f := newFakeKoolPreset()
	f.presetsParser.(*presets.FakeParser).MockExists = true
	f.presetsParser.(*presets.FakeParser).MockPresetKeys = []string{"kool.yml"}
	f.presetsParser.(*presets.FakeParser).MockPresetKeyContent = "kool.yml content"

	cmd := NewPresetCommand(f)

	cmd.SetArgs([]string{"laravel"})

	if err := cmd.Execute(); err != nil {
		t.Errorf("unexpected error executing preset command; error: %v", err)
	}

	if !f.out.(*shell.FakeOutputWriter).CalledSetWriter {
		t.Error("did not call SetWriter")
	}

	if !f.presetsParser.(*presets.FakeParser).CalledExists {
		t.Error("did not call parser.Exists")
	}

	if !f.presetsParser.(*presets.FakeParser).CalledGetPresetKeyContent["laravel"]["preset_ask_services"] {
		t.Error("did not call parser.GetPresetKeyContent for preset 'laravel' and meta 'preset_ask_services'")
	}

	if !f.out.(*shell.FakeOutputWriter).CalledPrintln {
		t.Error("did not call Println")
	}

	expected := "Preset laravel is initializing!"
	output := f.out.(*shell.FakeOutputWriter).OutLines[0]

	if expected != output {
		t.Errorf("Expecting message '%s', got '%s'", expected, output)
	}

	if !f.presetsParser.(*presets.FakeParser).CalledLookUpFiles {
		t.Error("did not call parser.LookUpFiles")
	}

	if !f.presetsParser.(*presets.FakeParser).CalledGetPresetKeys {
		t.Error("did not call parser.GetPresetKeys")
	}

	if !f.presetsParser.(*presets.FakeParser).CalledGetTemplates {
		t.Error("did not call parser.GetTemplates")
	}

	if val, ok := f.presetsParser.(*presets.FakeParser).CalledGetPresetKeyContent["laravel"]["kool.yml"]; !ok || !val {
		t.Error("failed calling parser.GetPresetKeyContent for preset 'laravel' and file 'kool.yml'")
	}

	if val, ok := f.presetsParser.(*presets.FakeParser).CalledWriteFile["kool.yml"]["kool.yml content"]; !ok || !val {
		t.Error("failed calling parser.WriteFile for file 'kool.yml' with the content 'kool.yml content'")
	}

	if !f.out.(*shell.FakeOutputWriter).CalledSuccess {
		t.Error("did not call Success")
	}

	expected = "Preset laravel initialized!"
	output = fmt.Sprint(f.out.(*shell.FakeOutputWriter).SuccessOutput...)

	if expected != output {
		t.Errorf("Expecting success message '%s', got '%s'", expected, output)
	}
}

func TestInvalidScriptPresetCommand(t *testing.T) {
	f := newFakeKoolPreset()
	cmd := NewPresetCommand(f)

	cmd.SetArgs([]string{"invalid"})

	if err := cmd.Execute(); err != nil {
		t.Errorf("unexpected error executing preset command; error: %v", err)
	}

	if !f.presetsParser.(*presets.FakeParser).CalledExists {
		t.Error("did not call parser.Exists")
	}

	if !f.out.(*shell.FakeOutputWriter).CalledError {
		t.Error("did not call Error")
	}

	expected := "Unknown preset invalid"
	output := f.out.(*shell.FakeOutputWriter).Err.Error()

	if expected != output {
		t.Errorf("expecting error '%s', got '%s'", expected, output)
	}

	if !f.exiter.(*shell.FakeExiter).Exited() {
		t.Error("did not call Exit")
	}
}

func TestExistingFilesPresetCommand(t *testing.T) {
	f := newFakeKoolPreset()
	f.presetsParser.(*presets.FakeParser).MockExists = true
	f.presetsParser.(*presets.FakeParser).MockFoundFiles = []string{"kool.yml"}
	cmd := NewPresetCommand(f)

	cmd.SetArgs([]string{"laravel"})

	if err := cmd.Execute(); err != nil {
		t.Errorf("unexpected error executing preset command; error: %v", err)
	}

	if !f.out.(*shell.FakeOutputWriter).CalledWarning {
		t.Error("did not call Warning")
	}

	expected := "Some preset files already exist. In case you wanna override them, use --override."
	output := fmt.Sprint(f.out.(*shell.FakeOutputWriter).WarningOutput...)

	if output != expected {
		t.Errorf("expecting message '%s', got '%s'", expected, output)
	}

	if !f.exiter.(*shell.FakeExiter).Exited() {
		t.Error("did not call Exit")
	}
}

func TestOverrideFilesPresetCommand(t *testing.T) {
	f := newFakeKoolPreset()
	f.presetsParser.(*presets.FakeParser).MockExists = true
	f.presetsParser.(*presets.FakeParser).MockFoundFiles = []string{"kool.yml"}

	cmd := NewPresetCommand(f)

	cmd.SetArgs([]string{"--override", "laravel"})

	if err := cmd.Execute(); err != nil {
		t.Errorf("unexpected error executing preset command; error: %v", err)
	}

	if f.presetsParser.(*presets.FakeParser).CalledLookUpFiles {
		t.Error("unexpected existing files checking")
	}

	if f.out.(*shell.FakeOutputWriter).CalledWarning {
		t.Error("unexpected existing files Warning")
	}

	if f.exiter.(*shell.FakeExiter).Exited() {
		t.Error("unexpected program Exit")
	}

	if !f.out.(*shell.FakeOutputWriter).CalledSuccess {
		t.Error("did not call Success")
	}
}

func TestWriteErrorPresetCommand(t *testing.T) {
	f := newFakeKoolPreset()
	f.presetsParser.(*presets.FakeParser).MockExists = true
	f.presetsParser.(*presets.FakeParser).MockPresetKeys = []string{"kool.yml"}
	f.presetsParser.(*presets.FakeParser).MockPresetKeyContent = "kool.yml content"
	f.presetsParser.(*presets.FakeParser).MockError = errors.New("write error")

	cmd := NewPresetCommand(f)

	cmd.SetArgs([]string{"laravel"})

	if err := cmd.Execute(); err != nil {
		t.Errorf("unexpected error executing preset command; error: %v", err)
	}

	if !f.out.(*shell.FakeOutputWriter).CalledError {
		t.Error("did not call Error")
	}

	expected := "Failed to write preset file : write error"
	output := f.out.(*shell.FakeOutputWriter).Err.Error()

	if output != expected {
		t.Errorf("expecting error '%s', got '%s'", expected, output)
	}

	if !f.exiter.(*shell.FakeExiter).Exited() {
		t.Error("did not call Exit")
	}
}

func TestNoArgsPresetCommand(t *testing.T) {
	f := newFakeKoolPreset()

	mockAnswer := make(map[string]string)
	mockAnswer["What language do you want to use"] = "php"
	mockAnswer["What preset do you want to use"] = "laravel"

	f.promptSelect.(*shell.FakePromptSelect).MockAnswer = mockAnswer
	f.presetsParser.(*presets.FakeParser).MockLanguages = []string{"php"}
	f.presetsParser.(*presets.FakeParser).MockPresets = []string{"laravel"}
	f.presetsParser.(*presets.FakeParser).MockExists = true

	cmd := NewPresetCommand(f)

	if err := cmd.Execute(); err != nil {
		t.Errorf("unexpected error executing preset command; error: %v", err)
	}

	if !f.promptSelect.(*shell.FakePromptSelect).CalledAsk {
		t.Error("did not call Ask on PromptSelect")
	}

	expected := "Preset laravel is initializing!"
	output := f.out.(*shell.FakeOutputWriter).OutLines[0]

	if expected != output {
		t.Errorf("Expecting message '%s', got '%s'", expected, output)
	}
}

func TestFailingLanguageNoArgsPresetCommand(t *testing.T) {
	f := newFakeKoolPreset()
	f.presetsParser.(*presets.FakeParser).MockLanguages = []string{"php"}
	f.presetsParser.(*presets.FakeParser).MockPresets = []string{"laravel"}

	mockError := make(map[string]error)
	mockError["What language do you want to use"] = errors.New("error prompt select language")

	f.promptSelect.(*shell.FakePromptSelect).MockError = mockError

	cmd := NewPresetCommand(f)

	if err := cmd.Execute(); err != nil {
		t.Errorf("unexpected error executing preset command; error: %v", err)
	}

	if !f.promptSelect.(*shell.FakePromptSelect).CalledAsk {
		t.Error("did not call Ask on PromptSelect")
	}

	if !f.out.(*shell.FakeOutputWriter).CalledError {
		t.Error("did not call Error")
	}

	expected := "error prompt select language"
	output := f.out.(*shell.FakeOutputWriter).Err.Error()

	if output != expected {
		t.Errorf("expecting error '%s', got '%s'", expected, output)
	}

	if !f.exiter.(*shell.FakeExiter).Exited() {
		t.Error("did not call Exit")
	}
}

func TestFailingPresetNoArgsPresetCommand(t *testing.T) {
	f := newFakeKoolPreset()
	f.presetsParser.(*presets.FakeParser).MockLanguages = []string{"php"}
	f.presetsParser.(*presets.FakeParser).MockPresets = []string{"laravel"}

	mockAnswer := make(map[string]string)
	mockAnswer["What language do you want to use"] = "php"

	f.promptSelect.(*shell.FakePromptSelect).MockAnswer = mockAnswer

	mockError := make(map[string]error)
	mockError["What preset do you want to use"] = errors.New("error prompt select preset")

	f.promptSelect.(*shell.FakePromptSelect).MockError = mockError

	cmd := NewPresetCommand(f)

	if err := cmd.Execute(); err != nil {
		t.Errorf("unexpected error executing preset command; error: %v", err)
	}

	if !f.promptSelect.(*shell.FakePromptSelect).CalledAsk {
		t.Error("did not call Ask on PromptSelect")
	}

	if !f.out.(*shell.FakeOutputWriter).CalledError {
		t.Error("did not call Error")
	}

	expected := "error prompt select preset"
	output := f.out.(*shell.FakeOutputWriter).Err.Error()

	if output != expected {
		t.Errorf("expecting error '%s', got '%s'", expected, output)
	}

	if !f.exiter.(*shell.FakeExiter).Exited() {
		t.Error("did not call Exit")
	}
}

func TestCancellingPresetCommand(t *testing.T) {
	f := newFakeKoolPreset()

	mockError := make(map[string]error)
	mockError["What language do you want to use"] = shell.ErrPromptSelectInterrupted

	f.promptSelect.(*shell.FakePromptSelect).MockError = mockError

	cmd := NewPresetCommand(f)

	if err := cmd.Execute(); err != nil {
		t.Errorf("unexpected error executing preset command; error: %v", err)
	}

	if !f.out.(*shell.FakeOutputWriter).CalledWarning {
		t.Error("did not call Warning")
	}

	expected := "Operation Cancelled\n"
	output := fmt.Sprintln(f.out.(*shell.FakeOutputWriter).WarningOutput...)

	if output != expected {
		t.Errorf("expecting warning '%s', got '%s'", expected, output)
	}

	if !f.exiter.(*shell.FakeExiter).Exited() {
		t.Error("did not call Exit")
	}

	if f.exiter.(*shell.FakeExiter).Code() != 0 {
		t.Error("did not call Exit with code 0")
	}
}

func TestNonTTYPresetCommand(t *testing.T) {
	f := newFakeKoolPreset()
	f.DefaultKoolService.term.(*shell.FakeTerminalChecker).MockIsTerminal = false

	cmd := NewPresetCommand(f)

	if err := cmd.Execute(); err != nil {
		t.Errorf("unexpected error executing preset command; error: %v", err)
	}

	if !f.out.(*shell.FakeOutputWriter).CalledError {
		t.Error("did not call Error")
	}

	err := f.out.(*shell.FakeOutputWriter).Err

	if err == nil {
		t.Error("expecting an error, got none")
	} else if err.Error() != "the input device is not a TTY; for non-tty environments, please specify a preset argument" {
		t.Errorf("expecting error 'the input device is not a TTY; for non-tty environments, please specify a preset argument', got %v", err)
	}
}
