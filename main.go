package main

import (
	"errors"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

const (
	high   string = "64000000"
	middle string = "15000000"
	low    string = "7500000"

	highResult   string = "H"
	middleResult string = "M"
	lowResult    string = "L"
	naResult     string = "N/A"

	path           string = "/sys/devices/virtual/powercap/intel-rapl-mmio/intel-rapl-mmio:0/constraint_0_power_limit_uw"
	pathResultFile string = "/tmp/i3status/power_cap"
)

type client struct {
	inFilePath  string
	outFilePath string
	value       string
	result      string
	termChan    chan os.Signal
	ticker      *time.Ticker
	heartbeat   *time.Ticker
}

func (c *client) loadFile() error {
	dat, err := os.ReadFile(c.inFilePath)
	if err != nil {
		return err
	}
	c.value = strings.TrimSpace(string(dat))
	return nil
}

func (c *client) writeValueToFile() error {
	dirPath := "/tmp/i3status"
	if _, err := os.Stat(dirPath); errors.Is(err, os.ErrNotExist) {
		err := os.Mkdir(dirPath, os.ModePerm)
		if err != nil {
			return err
		}
	}

	f, err := os.Create(pathResultFile)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.WriteString(c.result)
	if err != nil {
		return err
	}

	return nil
}

func (c *client) makeResult() {
	switch {
	case c.value == high:
		c.result = highResult
	case c.value == middle:
		c.result = middleResult
	case c.value == low:
		c.result = lowResult
	default:
		c.result = naResult
	}
}

func main() {
	c := *&client{
		inFilePath:  path,
		outFilePath: pathResultFile,
		termChan:    make(chan os.Signal, 1),
		ticker:      time.NewTicker(4 * time.Second),
		heartbeat:   time.NewTicker(30 * time.Minute),
	}
	fmt.Printf("\nreading from: %q\nwriting to: %q\n", c.inFilePath, c.outFilePath)

	go func() {
		for {
			select {
			case <-c.ticker.C:
				c.loadFile()
				c.makeResult()
				c.writeValueToFile()
			case <-c.heartbeat.C:
				fmt.Println("I'm alive!")
				fmt.Printf("Latest value from file %q\n", c.value)
				fmt.Printf("Latest result %q\n", c.result)
			}
		}
	}()

	//go func() {
	//	for {
	//		select {}
	//	}
	//}()

	fmt.Println("Started service")

	signal.Notify(c.termChan, syscall.SIGINT, syscall.SIGTERM)
	<-c.termChan
	c.quit()
}

func (c *client) quit() {
	fmt.Println("Quitting...")
	c.ticker.Stop()
	c.heartbeat.Stop()
	os.Exit(0)
}
