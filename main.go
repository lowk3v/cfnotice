package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"
)

type DNSChecker struct {
	StoragePath string
	ZoneId      string
}

func NewDNSChecker(storagePath, zoneId string) *DNSChecker {
	return &DNSChecker{StoragePath: storagePath, ZoneId: zoneId}
}

func (checker *DNSChecker) LoadPreviousRecords() ([]DNSRecord, error) {
	file, err := os.Open(checker.StoragePath)
	if err != nil {
		if os.IsNotExist(err) {
			return []DNSRecord{}, errors.New(fmt.Sprintf("The %s file is not exist", checker.StoragePath))
		}
		return nil, err
	}
	defer file.Close()

	var records map[string][]DNSRecord
	err = json.NewDecoder(file).Decode(&records)
	return records[checker.ZoneId], err
}

func (checker *DNSChecker) SaveRecords(zoneId string, records []DNSRecord) error {
	data, err := json.MarshalIndent(map[string][]DNSRecord{
		zoneId: records,
	}, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(checker.StoragePath, data, 0644)
}

func (checker *DNSChecker) DetectChanges(current, previous []DNSRecord) (old, added, removed []DNSRecord) {
	prevMap := make(map[string]DNSRecord)
	for _, record := range previous {
		prevMap[record.ID] = record
	}

	for _, record := range current {
		if _, exists := prevMap[record.ID]; !exists {
			added = append(added, record)
		} else {
			old = append(old, record)
		}
		delete(prevMap, record.ID)
	}

	for _, record := range prevMap {
		removed = append(removed, record)
	}

	return old, added, removed
}

type Config struct {
	CfAPIKey        string
	CfCookie        string
	PollingInterval int
	StoragePath     string
	ZoneId          string
	ZoneNo          int
	Debug           bool
}

func (c *Config) printDebug(isError bool, msg string) {
	if !c.Debug {
		return
	}
	if isError {
		_, _ = fmt.Fprintf(os.Stderr, msg)
	} else {
		fmt.Println(msg)
	}
}

func option() *Config {
	var storagePath string
	flag.StringVar(&storagePath, "s", os.Getenv("CF_NOTICE_PATH"), "Load a specific storage path or set the CF_NOTICE_PATH "+
		"environment. Default is ~/.config/cf-notice.json")

	var interval int
	flag.IntVar(&interval, "i", 0, "an interval to re-check. Disable by default")

	var zoneId string
	flag.StringVar(&zoneId, "zid", "", "Cloudflare zone id")

	var zoneNo int
	flag.IntVar(&zoneNo, "zno", 0, "Cloudflare zone number. Default is 0")

	var cookie string
	flag.StringVar(&cookie, "c", os.Getenv("CF_COOKIE"), "a cookie (file or string) of your Cloudflare access. Default is empty")

	var apiKey string
	flag.StringVar(&apiKey, "k", os.Getenv("CF_API_KEY"), "an API Key of your Cloudflare access. Default is empty")

	var debug bool
	flag.BoolVar(&debug, "debug", false, "Enable debugging. Default is false")

	flag.Parse()

	if storagePath == "" {
		storagePath = os.Getenv("HOME") + "/.config/cf-notice.json"
	}

	if _, err := os.Stat(cookie); err == nil {
		// load the cookie as a file
		f, _ := os.ReadFile(cookie)
		cookie = strings.TrimSpace(string(f))
	}

	if apiKey == "" && cookie == "" {
		_, _ = fmt.Fprintf(os.Stderr, "Cookie and Api key not found ")
		os.Exit(0)
	}

	return &Config{
		CfAPIKey:        apiKey,
		CfCookie:        cookie,
		PollingInterval: interval,
		StoragePath:     storagePath,
		ZoneId:          zoneId,
		ZoneNo:          zoneNo,
		Debug:           debug,
	}
}

func run(cfg *Config, checker *DNSChecker, api *CloudflareAPI) {
	fmt.Println("Checking for DNS changes...")

	currentRecords, err := api.ListDNSRecords(cfg.ZoneId)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Failed to fetch DNS records: %v", err)
		return
	}

	previousRecords, _ := checker.LoadPreviousRecords()
	if err != nil {
		cfg.printDebug(false, fmt.Sprint("Load previous record error, create a new one"))
	}
	notChange, added, removed := checker.DetectChanges(currentRecords, previousRecords)

	if len(added) > 0 || len(removed) > 0 || len(notChange) > 0 {
		fmt.Printf("Changes detected! Added: %d, Removed: %d\n", len(added), len(removed))

		for _, record := range notChange {
			fmt.Printf("Not Change: %s\t%s\t%s\n", record.Name, record.Type, record.Content)
		}

		for _, record := range added {
			fmt.Printf("Added: %s\t%s\t%s\n", record.Name, record.Type, record.Content)
		}

		for _, record := range removed {
			fmt.Printf("Removed: %s\t%s\t%s\n", record.Name, record.Type, record.Content)
		}

		err := checker.SaveRecords(cfg.ZoneId, currentRecords)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "err: %v", err)
		}
	}
}

func main() {
	cfg := option()

	cfg.printDebug(false, fmt.Sprintf("Loading storage at %s\n", cfg.StoragePath))

	if cfg.PollingInterval == 0 {
		cfg.printDebug(false, fmt.Sprint("Disable running interval"))

	} else {
		cfg.printDebug(false, fmt.Sprintf("Set interval is %d s\n", cfg.PollingInterval))
	}

	api := NewCloudflareAPI(cfg.CfAPIKey, cfg.CfCookie)

	if cfg.ZoneId == "" {
		zones, err := api.ListZones()
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Failed to fetch zones: %v", err)
			return
		}
		cfg.ZoneId = zones[cfg.ZoneNo].ID

		fmt.Println("All available zones:")
		for _, z := range zones {
			fmt.Printf("[%s] %s\n", z.ID, z.Name)
		}
		fmt.Printf("Picked the zone: [%s]%s\n", zones[cfg.ZoneNo].ID, zones[cfg.ZoneNo].Name)
	}

	checker := NewDNSChecker(cfg.StoragePath, cfg.ZoneId)

	if cfg.PollingInterval > 0 {
		// run and schedule to run
		ticker := time.NewTicker(time.Duration(cfg.PollingInterval) * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			run(cfg, checker, api)
		}
	} else {
		// run one time
		run(cfg, checker, api)
	}
}
