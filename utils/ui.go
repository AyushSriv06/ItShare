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
	// Define color functions with enhanced styling
	InfoColor    = color.New(color.FgCyan).SprintFunc()
	SuccessColor = color.New(color.FgGreen, color.Bold).SprintFunc()
	ErrorColor   = color.New(color.FgRed, color.Bold).SprintFunc()
	WarningColor = color.New(color.FgYellow, color.Bold).SprintFunc()
	HeaderColor  = color.New(color.FgMagenta, color.Bold).SprintFunc()
	CommandColor = color.New(color.FgBlue, color.Bold).SprintFunc()
	UserColor    = color.New(color.FgGreen, color.Bold).SprintFunc()
	PausedColor  = color.New(color.FgYellow, color.Bold).SprintFunc()
	AccentColor  = color.New(color.FgHiCyan, color.Bold).SprintFunc()
	BorderColor  = color.New(color.FgHiBlack).SprintFunc()
)

type ProgressBar struct {
	Bar       *progressbar.ProgressBar
	IsPaused  bool
	Mutex     sync.Mutex
	TransferId string
}

// CreateProgressBar creates and returns a custom progress bar for file transfers
func CreateProgressBar(size int64, description string) *ProgressBar {
	bar := progressbar.NewOptions64(
		size,
		progressbar.OptionSetDescription(description),
		progressbar.OptionEnableColorCodes(true),
		progressbar.OptionShowBytes(true),
		progressbar.OptionSetWidth(60),
		progressbar.OptionThrottle(50*time.Millisecond),
		progressbar.OptionShowCount(),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "[green]█[reset]",
			SaucerHead:    "[green]█[reset]",
			SaucerPadding: "[dim]░[reset]",
			BarStart:      "[cyan][[reset]",
			BarEnd:        "[cyan]][reset]",
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

// PrintHelp displays all available commands with enhanced formatting
func PrintHelp() {
	fmt.Println(HeaderColor("\n╔════════════════════════════════════════════════════════════════╗"))
	fmt.Println(HeaderColor("║                     ItShare Help Guide                        ║"))
	fmt.Println(HeaderColor("╚════════════════════════════════════════════════════════════════╝"))
	
	fmt.Println(BorderColor("\n┌────────────────────────────────────────────────────────────────┐"))
	fmt.Println(HeaderColor("│                      General Commands                          │"))
	fmt.Println(BorderColor("├────────────────────────────────────────────────────────────────┤"))
	fmt.Printf("│  %s           Show online users and their status        │\n", CommandColor("/status"))
	fmt.Printf("│  %s             Display this help message               │\n", CommandColor("/help"))
	fmt.Printf("│  %s              Disconnect and exit application          │\n", CommandColor("exit"))
	fmt.Println(BorderColor("└────────────────────────────────────────────────────────────────┘"))
	
	fmt.Println(BorderColor("\n┌────────────────────────────────────────────────────────────────┐"))
	fmt.Println(HeaderColor("│                      File Operations                           │"))
	fmt.Println(BorderColor("├────────────────────────────────────────────────────────────────┤"))
	fmt.Printf("│  %s    Browse user's shared files               │\n", CommandColor("/lookup <userId>"))
	fmt.Printf("│  %s Send a file to specific user              │\n", CommandColor("/sendfile <userId> <path>"))
	fmt.Printf("│  %s Send entire folder to user               │\n", CommandColor("/sendfolder <userId> <path>"))
	fmt.Printf("│  %s Download file from user's share          │\n", CommandColor("/download <userId> <fileName>"))
	fmt.Println(BorderColor("└────────────────────────────────────────────────────────────────┘"))
	
	fmt.Println(BorderColor("\n┌────────────────────────────────────────────────────────────────┐"))
	fmt.Println(HeaderColor("│                     Transfer Controls                          │"))
	fmt.Println(BorderColor("├────────────────────────────────────────────────────────────────┤"))
	fmt.Printf("│  %s          Show all active transfers               │\n", CommandColor("/transfers"))
	fmt.Printf("│  %s     Pause an active transfer                 │\n", CommandColor("/pause <transferId>"))
	fmt.Printf("│  %s    Resume a paused transfer                 │\n", CommandColor("/resume <transferId>"))
	fmt.Println(BorderColor("└────────────────────────────────────────────────────────────────┘"))
	
	fmt.Println(BorderColor("\n╔════════════════════════════════════════════════════════════════╗"))
	fmt.Println(AccentColor("║        Type a message and press Enter to chat with everyone!      ║"))
	fmt.Println(BorderColor("╚════════════════════════════════════════════════════════════════╝\n"))
}

// PrintBanner prints the application banner with ItShare branding
func PrintBanner() {
	banner := `
    ____  __  _____ __                    
   /  _/ / /_/ ___// /_  ____ __________ 
   / /  / __/\__ \/ __ \/ __ '/ ___/ _ \
 _/ /  / /_ ___/ / / / / /_/ / /  /  __/
/___/  \__//____/_/ /_/\__,_/_/   \___/ 
                                       
`
	fmt.Println(color.New(color.FgHiCyan, color.Bold).Sprint(banner))
	fmt.Println(AccentColor("════════════════════════════════════════════════════════"))
	fmt.Println(HeaderColor("           🚀 Intelligent File Sharing Network 🚀"))
	fmt.Println(AccentColor("════════════════════════════════════════════════════════\n"))
}

// PrintWelcome displays a welcome message with usage tips
func PrintWelcome() {
	fmt.Println(SuccessColor("Welcome to ItShare!"))
	fmt.Println(InfoColor("Connected to the network successfully"))
	fmt.Println(InfoColor("Type"), CommandColor("/help"), InfoColor("for available commands"))
	fmt.Println(InfoColor("Start chatting or sharing files with other users!\n"))
}

// PrintSeparator prints a visual separator line
func PrintSeparator() {
	fmt.Println(BorderColor("────────────────────────────────────────────────────────────────"))
}

// PrintStatus displays connection status with enhanced formatting
func PrintStatus(message string, statusType string) {
	timestamp := time.Now().Format("15:04:05")
	
	switch statusType {
	case "success":
		fmt.Printf("[%s] %s %s\n", timestamp, SuccessColor("SUCCESS:"), message)
	case "error":
		fmt.Printf("[%s] %s %s\n", timestamp, ErrorColor("ERROR:"), message)
	case "warning":
		fmt.Printf("[%s] %s %s\n", timestamp, WarningColor("WARNING:"), message)
	case "info":
		fmt.Printf("[%s] %s %s\n", timestamp, InfoColor("INFO:"), message)
	default:
		fmt.Printf("[%s] %s\n", timestamp, message)
	}
}