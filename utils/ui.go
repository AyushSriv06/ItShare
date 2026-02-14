package utils

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/fatih/color"
	"github.com/schollz/progressbar/v3"
)

var (
	// Define color functions
	InfoColor    = color.New(color.FgCyan).SprintFunc()
	SuccessColor = color.New(color.FgGreen).SprintFunc()
	ErrorColor   = color.New(color.FgRed).SprintFunc()
	WarningColor = color.New(color.FgYellow).SprintFunc()
	HeaderColor  = color.New(color.FgMagenta, color.Bold).SprintFunc()
	CommandColor = color.New(color.FgBlue, color.Bold).SprintFunc()
	UserColor    = color.New(color.FgGreen, color.Bold).SprintFunc()
	PausedColor  = color.New(color.FgYellow, color.Bold).SprintFunc()
)

type ProgressBar struct {
	Bar        *progressbar.ProgressBar
	IsPaused   bool
	Mutex      sync.Mutex
	TransferId string
}

// CreateProgressBar creates and returns a custom progress bar for file transfers
func CreateProgressBar(size int64, description string) *ProgressBar {
	bar := progressbar.NewOptions64(
		size,
		progressbar.OptionSetDescription(description),
		progressbar.OptionEnableColorCodes(true),
		progressbar.OptionShowBytes(true),
		progressbar.OptionSetWidth(50),
		progressbar.OptionThrottle(65*time.Millisecond),
		progressbar.OptionShowCount(),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "[green]=[reset]",
			SaucerHead:    "[green]>[reset]",
			SaucerPadding: " ",
			BarStart:      "[",
			BarEnd:        "]",
		}),
		progressbar.OptionOnCompletion(func() {
			fmt.Fprint(os.Stdout, "\n")
		}),
	)

	return &ProgressBar{
		Bar:      bar,
		IsPaused: false,
	}
}

func (pb *ProgressBar) Write(p []byte) (n int, err error) {
	pb.Mutex.Lock()
	defer pb.Mutex.Unlock()

	if pb.IsPaused {
		return len(p), nil
	}

	return pb.Bar.Write(p)
}

func (pb *ProgressBar) SetPaused(paused bool) {
	pb.Mutex.Lock()
	defer pb.Mutex.Unlock()

	pb.IsPaused = paused

	description := pb.Bar.String()
	if paused {
		pb.Bar.Describe(fmt.Sprintf("%s %s", description, PausedColor("[PAUSED]")))
	} else {
		pb.Bar.Describe(description)
	}
}

func (pb *ProgressBar) GetTransferId() string {
	return pb.TransferId
}

func (pb *ProgressBar) SetTransferId(id string) {
	pb.TransferId = id
}

// PrintHelp displays all available commands
func PrintHelp() {
	fmt.Println(HeaderColor("\n📚 ItShare Help - Available Commands 📚"))
	fmt.Println(InfoColor("------------------------------------------------"))

	fmt.Println(HeaderColor("\n🔍 Server Discovery:"))
	fmt.Printf("  %s - Servers are automatically discovered via UDP broadcast\n", InfoColor("• Auto-discovery"))
	fmt.Printf("  %s - No need to manually share IP addresses\n", InfoColor("• Network scanning"))
	fmt.Printf("  %s - Fallback to manual entry if needed\n", InfoColor("• Manual override"))

	fmt.Println(HeaderColor("\n🌐 General Commands:"))
	fmt.Printf("  %s - Show online users\n", CommandColor("/status"))
	fmt.Printf("  %s - Show this help message\n", CommandColor("/help"))
	fmt.Printf("  %s - Disconnect and exit\n", CommandColor("exit"))

	fmt.Println(HeaderColor("\n🏠 Room Commands:"))
	fmt.Printf("  %s - Create a new room with selected users\n", CommandColor("/createroom"))
	fmt.Printf("  %s - Join a specific room\n", CommandColor("/joinroom <roomID>"))
	fmt.Printf("  %s - Leave current room\n", CommandColor("/leaveroom"))
	fmt.Printf("  %s - List your rooms\n", CommandColor("/rooms"))

	fmt.Println(HeaderColor("\n📁 File Operations:"))
	fmt.Printf("  %s - Browse user's shared files\n", CommandColor("/lookup <userId>"))
	fmt.Printf("  %s - Send a file to user\n", CommandColor("/sendfile <userId> <filePath>"))
	fmt.Printf("  %s - Send a folder to user\n", CommandColor("/sendfolder <userId> <folderPath>"))
	fmt.Printf("  %s - Download a file from user\n", CommandColor("/download <userId> <fileName>"))

	fmt.Println(HeaderColor("\n📡 Transfer Controls:"))
	fmt.Printf("  %s - Show all active transfers\n", CommandColor("/transfers"))
	fmt.Printf("  %s - Pause an active transfer\n", CommandColor("/pause <transferId>"))
	fmt.Printf("  %s - Resume a paused transfer\n", CommandColor("/resume <transferId>"))

	fmt.Println(InfoColor("------------------------------------------------"))
	fmt.Println(InfoColor("Type a message and press Enter to send to current room or everyone\n"))
}

func PrintBanner() {
	banner := `
  ___ _   ___ _                 
 |_ _| |_/ __| |_  __ _ _ _ ___ 
  | ||  _\__ \ ' \/ _' | '_/ -_)
 |___|\__|___/_||_\__,_|_| \___|
                                
`
	fmt.Println(color.New(color.FgCyan, color.Bold).Sprint(banner))
}
