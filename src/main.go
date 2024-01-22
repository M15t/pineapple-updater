package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"runtime"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"

	"github.com/cavaliergopher/grab/v3"
)

// custom const
const (
	GHEndpoint        = "https://github.com/pineappleEA/pineapple-src/"
	PAReleaseEndpoint = "https://pineappleEA.github.io/"

	DefaultPath = "C:/yuzu"
)

func main() {
	// Create log file or open for append
	logfile, err := os.OpenFile("log.txt", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer logfile.Close()

	// Set log output to the log file
	log.SetOutput(logfile)

	// * create new app
	a := app.NewWithID(strconv.Itoa(os.Getpid()))
	// * create new window
	w := a.NewWindow("PinEApple Updater")
	// * set icon
	w.SetIcon(resourceIconPng)

	log.Printf("Downloading available versions from %s", PAReleaseEndpoint)

	// * download list of versions and links
	versions := make([]int, 0)
	links := make(map[int]string)
	platform := runtime.GOOS  // windows or linux
	if platform == "darwin" { // force windows
		platform = "windows"
		// log.Fatal("macOS is not supported")
	}
	if err := list(platform, &versions, links); err != nil {
		log.Fatal(err)
	}

	// * create mainUI
	w.SetContent(mainUI(versions, links))
	w.Resize(fyne.NewSize(500, 450))
	w.Show()

	// * run app
	a.Run()
}

func list(platform string, versions *[]int, links map[int]string) error {
	// Download site into resp
	resp, err := http.Get(PAReleaseEndpoint)
	if err != nil {
		return fmt.Errorf("Could not download site: %v", err)
	}
	defer resp.Body.Close()

	// Read response body through scanner
	scanner := bufio.NewScanner(resp.Body)
	for i := 0; scanner.Scan(); i++ {
		line := scanner.Text()

		// Check if the line contains "EA" followed by a space and a number
		if match, _ := regexp.MatchString("EA [0-9]", line); match {
			// Extract version number
			versionPattern := regexp.MustCompile("EA [0-9]+")
			versionString := versionPattern.FindString(line)
			versionString = regexp.MustCompile("[0-9]+").FindString(versionString)
			version, _ := strconv.Atoi(versionString)

			// * Mapping link
			switch platform {
			case "windows":
				links[version] = fmt.Sprintf("%sreleases/download/EA-%d/Windows-Yuzu-EA-%d.zip", GHEndpoint, version, version)
			case "linux":
				links[version] = fmt.Sprintf("%sreleases/download/EA-%d/Linux-Yuzu-EA-%d.AppImage", GHEndpoint, version, version)
			}

			*versions = append(*versions, version)

			// sample
			// https://github.com/pineappleEA/pineapple-src/releases/download/EA-4079/Linux-Yuzu-EA-4079.AppImage
			// https://github.com/pineappleEA/pineapple-src/releases/download/EA-4079/Windows-Yuzu-EA-4079.zip
		} else if line == "</html>" {
			// Exit loop when the </html> tag is encountered
			break
		}
	}

	if len(*versions) == 0 {
		return fmt.Errorf("Could not obtain the list of files")
	}

	return nil
}

func download(link string) error {
	// * ping download link
	ping, err := http.Get(link)
	if err != nil {
		return fmt.Errorf("Failed to download from GitHub: %v", err)
	}
	defer ping.Body.Close()

	if ping.StatusCode != http.StatusOK {
		return fmt.Errorf("no download link found, Anonfiles or GitHub seems to be having issues")
	}

	//TODO: figure out proper way to set the path for windows
	dst := fyne.CurrentApp().Preferences().StringWithFallback("path", DefaultPath)
	fmt.Println("Download to:", dst)
	req, err := grab.NewRequest(dst, link)
	if err != nil {
		return fmt.Errorf("Failed to create download request: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	req = req.WithContext(ctx)
	resp := grab.DefaultClient.Do(req)

	//TODO: figure out why the mainUI is unresponsive when the downloadUI is open
	go downloadUI(resp, cancel)

	// check for errors
	if err := resp.Err(); err != nil && err.Error() != "context canceled" {
		return fmt.Errorf("Download failed: %v", err)
	}

	return nil
}
