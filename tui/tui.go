package tui

import (
	"crowl4dead/models"
	"fmt"
	"os"
	"time"

	"github.com/charmbracelet/huh"
)

func truncate(s string, maxLen int) string {
	if len(s) > maxLen {
		return s[:maxLen-3] + "..."
	}
	return s
}

func RunTUI(results []models.Result, timer time.Duration) {
	for {
		var filter string
		fmt.Println("time:", timer)
		form := huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Title("Select Filter").
					Options(
						huh.NewOption("All Links", "all"),
						huh.NewOption("Alive Links", "alive"),
						huh.NewOption("Dead Links", "dead"),
						huh.NewOption("Outbound Links", "outbound"),
						huh.NewOption("Exit", "exit"),
					).
					Value(&filter),
			),
		)

		err := form.Run()
		if err != nil {
			fmt.Printf("Error running form: %v\n", err)
			os.Exit(1)
		}

		if filter == "exit" {
			fmt.Println("Exiting...")
			break
		}
		fmt.Println("\nFiltered Results:")
		fmt.Println("=================")
		count := 0
		for _, res := range results {
			if filter == "all" || res.Status == filter {
				count++
				fmt.Printf("URL:    %s\nStatus: %s\nSource: %s\n\n", truncate(res.Link.URL, 60), res.Status, truncate(res.Link.Source, 60))
			}
		}
		fmt.Printf("Total: %d results\n", count)
		fmt.Println("=================")
	}
}
