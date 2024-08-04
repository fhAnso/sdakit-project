package lib

import (
	"errors"
	"fmt"
	"net"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/hashicorp/go-version"
)

const na = "n/a"

func DefaultOutputName(hostname string) string {
	currentTime := time.Now()
	formatTime := currentTime.Format("2006-01-02_15-04-05")
	outputFile := fmt.Sprintf("%s-%s.txt", formatTime, hostname)
	return outputFile
}

func CreateOutputDir() error {
	outputDir := "output"
	if _, err := os.Stat(outputDir); os.IsNotExist(err) {
		err := os.MkdirAll(outputDir, 0755)
		if err != nil {
			return errors.New("unable to create output directory: " + outputDir)
		}
	}
	return nil
}

func GetCurrentLocalVersion(failHandler *VersionHandler) string {
	var (
		versionPath string
		content     []byte
		err         error
	)
	if runtime.GOOS == "windows" {
		versionPath = "..\\version.txt"
	} else if runtime.GOOS == "linux" {
		versionPath = "../version.txt"
	}
	if _, err := os.Stat(versionPath); errors.Is(err, os.ErrNotExist) {
		versionPath = "version.txt"
	}
	content, err = os.ReadFile(versionPath)
	version := string(content)
	TestVersionFail(*failHandler, &version, err)
	return version
}

func VersionCompare() {
	failHandler := &VersionHandler{}
	repo := GetCurrentRepoVersion(failHandler)
	local := GetCurrentLocalVersion(failHandler)
	if repo == na || local == na || local == "" {
		return
	}
	parseRepoVersion, _ := version.NewVersion(repo)
	parseLocalVersion, _ := version.NewVersion(local)
	if repo != local && parseLocalVersion.LessThan(parseRepoVersion) {
		fmt.Printf("[*] An update is available! %s->%s\n", local, repo)
	}
}

func IsInExclude(httpCode string, list []string) bool {
	for _, code := range list {
		if httpCode == code {
			return true
		}
	}
	return false
}

func EditDbEntries(hostname string) []string {
	entries := make([]string, 0, len(Db))
	for idx, entry := range Db {
		endpoint := strings.Replace(entry, "HOST", hostname, 1)
		fmt.Printf("\n%d. Entry: %s\n ===[ %s\n", idx+1, entry, endpoint)
		entries = append(entries, endpoint)
	}
	fmt.Printf("\n[*] Using %d endpoints\n", len(entries))
	return entries
}

func RequestIpAddresses(subdomain string) string {
	ips, err := net.LookupIP(subdomain)
	if err != nil {
		// Lookup failed, leave blank
		return ""
	}
	var results []string
	for _, ip := range ips {
		results = append(results, ip.String())
	}
	result := fmt.Sprintf("(%s)", strings.Join(results, ", "))
	return result
}
