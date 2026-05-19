package tests

import (
	"bytes"
	"dungeonGameLib/lib/engine/gamestate"
	"os"
	"os/exec"
	"strings"
	"testing"
)

func TestGolden(t *testing.T) {
	config, err := loadConfig("test_config.json")
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	gameInstance, err := gamestate.CreateGameInstance(config.Floors, config.Monsters, config.OpenAt, config.Duration)
	if err != nil {
		t.Fatal("Failed to create game instance:", err)
	}

	go gameInstance.RunGameLoop()

	eventFile, err := os.Open("test_events.txt")
	if err != nil {
		t.Fatalf("Failed to open events file: %v", err)
	}
	defer eventFile.Close()

	outputs, err := testProcessReader(eventFile, gameInstance.InputEventChan, gameInstance.OutputEventChan)
	if err != nil {
		t.Fatalf("Failed to process events: %v", err)
	}

	outputs = append(outputs, "")
	outputs = append(outputs, "Final report:")
	for _, pair := range compilePlayerDataSafe(gameInstance) {
		outputs = append(outputs, pair.Player.String())
	}

	expectedData, err := os.ReadFile("expected_output.txt")
	if err != nil {
		t.Fatalf("Failed to read expected output: %v", err)
	}

	expectedOutputs := strings.Split(strings.TrimSpace(string(expectedData)), "\n")

	if len(outputs) != len(expectedOutputs) {
		t.Errorf("Output count mismatch: got %d, expected %d", len(outputs), len(expectedOutputs))
		t.Logf("Got outputs:\n%s", strings.Join(outputs, "\n"))
		t.Logf("Expected outputs:\n%s", strings.Join(expectedOutputs, "\n"))
		return
	}

	for i, got := range outputs {
		expected := expectedOutputs[i]
		if got != expected {
			t.Errorf("Output mismatch at line %d:\nGot:      %s\nExpected: %s", i+1, got, expected)
		}
	}
}

func TestGoldenWithConfig(t *testing.T) {
	cmd := exec.Command("go", "run", "../", "test_config.json")

	stdin, err := cmd.StdinPipe()
	if err != nil {
		t.Fatalf("Failed to create stdin pipe: %v", err)
	}

	eventData, err := os.ReadFile("test_events.txt")
	if err != nil {
		t.Fatalf("Failed to read events file: %v", err)
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Start(); err != nil {
		t.Fatalf("Failed to start command: %v", err)
	}

	if _, err := stdin.Write(eventData); err != nil {
		t.Fatalf("Failed to write events to stdin: %v", err)
	}

	stdin.Close()
	if err := cmd.Wait(); err != nil {
		t.Fatalf("Command failed: %v\nStderr: %s", err, stderr.String())
	}

	expectedData, err := os.ReadFile("expected_output.txt")
	if err != nil {
		t.Fatalf("Failed to read expected output: %v", err)
	}

	gotOutput := stdout.String()
	expectedOutput := string(expectedData)

	if gotOutput != expectedOutput {
		if len(gotOutput) != len(expectedOutput) {
			t.Logf("Length mismatch: got=%d, expected=%d", len(gotOutput), len(expectedOutput))
		}
		minLen := len(gotOutput)
		if len(expectedOutput) < minLen {
			minLen = len(expectedOutput)
		}
		for i := 0; i < minLen; i++ {
			if gotOutput[i] != expectedOutput[i] {
				t.Logf("First diff at byte %d: got=%q, expected=%q", i, gotOutput[i], expectedOutput[i])
				break
			}
		}
		t.Errorf("Output mismatch:\nGot:\n%s\nExpected:\n%s", gotOutput, expectedOutput)
	}
}
