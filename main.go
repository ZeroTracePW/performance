package main

import (
	"encoding/json"
	"fmt"
	"os"
	"syscall"
	"time"
)

var (
	user32           = syscall.NewLazyDLL("user32.dll")
	getAsyncKeyState = user32.NewProc("GetAsyncKeyState")
)

// KeyStats holds analytics data about keystrokes
type KeyStats struct {
	TotalKeystrokes     int            `json:"total_keystrokes"`
	KeyFrequency        map[string]int `json:"key_frequency"`
	StartTime           time.Time      `json:"start_time"`
	EndTime             time.Time      `json:"end_time"`
	KeystrokesPerMinute float64        `json:"keystrokes_per_minute"`
}

// keyPressed checks if a specific key is currently pressed
func keyPressed(vKey int) bool {
	val, _, _ := getAsyncKeyState.Call(uintptr(vKey))
	return val != 0
}

// keyToName converts virtual key codes to readable names
func keyToName(vKey int) string {
	// Basic mapping - you can expand this with more keys
	switch vKey {
	case 0x20:
		return "SPACE"
	case 0x0D:
		return "ENTER"
	case 0x08:
		return "BACKSPACE"
	case 0x09:
		return "TAB"
	case 0x10:
		return "SHIFT"
	case 0x11:
		return "CTRL"
	case 0x12:
		return "ALT"
	case 0x1B:
		return "ESC"
	default:
		if vKey >= 0x30 && vKey <= 0x39 {
			return fmt.Sprintf("%d", vKey-0x30)
		}
		if vKey >= 0x41 && vKey <= 0x5A {
			return string(rune(vKey))
		}
		return fmt.Sprintf("0x%X", vKey)
	}
}

// trackKeystrokes monitors keyboard activity and updates statistics
func trackKeystrokes(duration time.Duration) KeyStats {
	stats := KeyStats{
		KeyFrequency: make(map[string]int),
		StartTime:    time.Now(),
	}

	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()
	stop := time.After(duration)

	for {
		select {
		case <-stop:
			stats.EndTime = time.Now()
			elapsed := stats.EndTime.Sub(stats.StartTime).Minutes()
			if elapsed > 0 {
				stats.KeystrokesPerMinute = float64(stats.TotalKeystrokes) / elapsed
			}
			return stats
		case <-ticker.C:
			for key := 0; key < 256; key++ {
				if keyPressed(key) {
					stats.TotalKeystrokes++
					keyName := keyToName(key)
					stats.KeyFrequency[keyName]++
				}
			}
		}
	}
}

// saveStats saves the collected statistics to a JSON file
func saveStats(stats KeyStats, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(stats)
}

// displayStats prints statistics to the console
func displayStats(stats KeyStats) {
	fmt.Println("\nKeyboard Activity Analysis:")
	fmt.Println("---------------------------")
	fmt.Printf("Monitoring Duration: %.2f seconds\n", stats.EndTime.Sub(stats.StartTime).Seconds())
	fmt.Printf("Total Keystrokes: %d\n", stats.TotalKeystrokes)
	fmt.Printf("Keystrokes Per Minute: %.2f\n", stats.KeystrokesPerMinute)

	fmt.Println("\nMost Frequently Pressed Keys:")
	for key, count := range stats.KeyFrequency {
		if count > 5 { // Only show keys pressed more than 5 times
			fmt.Printf("%s: %d\n", key, count)
		}
	}
}

func main() {
	fmt.Println("Starting keyboard activity monitor for 1 minute...")

	for i := 1; i <= 3; i++ {
		fmt.Printf("%d\n", i)
		time.Sleep(1 * time.Second)
	}

	// Track for 1 minute
	stats := trackKeystrokes(60 * time.Second)

	// Display results
	displayStats(stats)

	// Save to file
	err := saveStats(stats, "keystats.json")
	if err != nil {
		fmt.Printf("Error saving stats: %v\n", err)
	} else {
		fmt.Println("\nStatistics saved to keystats.json")
	}
}
