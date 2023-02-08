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

const ()

type values struct {
	high   string
	middle string
	low    string
	na     string
}

type client struct {
	inFilePath  string
	outFilePath string
	inValue     string
	result      string
	termChan    chan os.Signal
	ticker      *time.Ticker
	heartbeat   *time.Ticker
	values      values
	results     values
}

func main() {
	c := *&client{
		inFilePath:  "/sys/devices/virtual/powercap/intel-rapl-mmio/intel-rapl-mmio:0/constraint_0_power_limit_uw",
		outFilePath: "/tmp/i3status/power_cap",
		termChan:    make(chan os.Signal, 1),
		ticker:      time.NewTicker(4 * time.Second),
		heartbeat:   time.NewTicker(0 * time.Hour),
		values: values{
			high:   "64000000",
			middle: "15000000",
			low:    "7500000",
		},
		results: values{
			high:   "H",
			middle: "M",
			low:    "L",
			na:     "N/A",
		},
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
			}
		}
	}()

	fmt.Println("Started service")

	signal.Notify(c.termChan, syscall.SIGINT, syscall.SIGTERM)
	<-c.termChan
	c.quit()
}

func (c *client) loadFile() error {
	dat, err := os.ReadFile(c.inFilePath)
	if err != nil {
		return err
	}
	c.inValue = strings.TrimSpace(string(dat))
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

	f, err := os.Create(c.outFilePath)
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
	case c.inValue == c.values.high:
		c.result = c.results.high
	case c.inValue == c.values.middle:
		c.result = c.results.middle
	case c.inValue == c.values.low:
		c.result = c.results.low
	default:
		c.result = c.results.na
	}
}

func (c *client) quit() {
	fmt.Println("Quitting...")
	c.ticker.Stop()
	c.heartbeat.Stop()
	os.Exit(0)
}
