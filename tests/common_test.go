package tests

import (
	"bufio"
	"dungeonGameLib/lib/engine/event"
	"dungeonGameLib/lib/engine/gamestate"
	"dungeonGameLib/lib/parser"
	"fmt"
	"io"
	"strings"
)

func testProcessReader(r io.Reader, inputEventChan chan event.GameEvent, outputEventChan chan event.GameOutputEvent) ([]string, error) {
	var outputs []string
	reader := bufio.NewReader(r)
	lineNum := 0

	for {
		lineNum++
		line, err := reader.ReadString('\n')

		if err != nil {
			if err == io.EOF {
				if line == "" {
					break
				}
			} else {
				return nil, fmt.Errorf("line %d: read error: %w", lineNum, err)
			}
		}

		line = strings.TrimRight(line, "\r\n")
		if line == "" {
			continue
		}

		event, err := parser.ParseLine(line)
		if err != nil {
			return nil, fmt.Errorf("line %d: parse error: %w", lineNum, err)
		}

		inputEventChan <- event

		for {
			output := <-outputEventChan
			if output.Error == nil {
				outputs = append(outputs, output.Event.String())
			}

			if !output.ReadOneMore {
				break
			}
		}
	}

	return outputs, nil
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if strings.Contains(s, item) {
			return true
		}
	}
	return false
}

func compilePlayerDataSafe(gameInstance *gamestate.GameInstance) []gamestate.PlayerAndIdPair {
	drainDone := make(chan struct{})
	go func() {
		for {
			select {
			case <-gameInstance.OutputEventChan:
			case <-drainDone:
				return
			}
		}
	}()
	result := gameInstance.CompilePlayerData()
	close(drainDone)
	return result
}

func loadConfig(path string) (*parser.GameConfig, error) {
	return parser.LoadConfigFromFile(path)
}
