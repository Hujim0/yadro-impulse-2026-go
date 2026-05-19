package main

import (
	"bufio"
	"dungeonGameLib/lib/engine/event"
	"dungeonGameLib/lib/engine/gamestate"
	"dungeonGameLib/lib/parser"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatalf("usage: %s <config-file.json>", os.Args[0])
	}

	config, err := parser.LoadConfigFromFile(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}

	gameInstance, err := gamestate.CreateGameInstance(config.Floors, config.Monsters, config.OpenAt, config.Duration)
	if err != nil {
		log.Fatal(err)
	}

	go gameInstance.RunGameLoop()

	defer func() {
		fmt.Println("\nFinal report:")
		for _, pair := range gameInstance.CompilePlayerData() {
			fmt.Println(pair.Player)
		}
	}()

	if err := processReader(os.Stdin, gameInstance.InputEventChan, gameInstance.OutputEventChan); err != nil {
		log.Fatal(err)
	}
}

func processReader(r io.Reader, inputEventChan chan event.GameEvent, outputEvenChan chan event.GameOutputEvent) error {
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
				return fmt.Errorf("line %d: read error: %w", lineNum, err)
			}
		}

		line = strings.TrimRight(line, "\r\n")
		if line == "" {
			continue
		}

		event, err := parser.ParseLine(line)
		if err != nil {
			return fmt.Errorf("line %d: parse error: %w", lineNum, err)
		}

		inputEventChan <- event

		for {
			output := <-outputEvenChan
			if output.Error == nil {
				fmt.Println(output.Event)
			}

			if !output.ReadOneMore {
				break
			}
		}
	}

	return nil
}
