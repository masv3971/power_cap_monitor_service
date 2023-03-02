package main

import (
	"errors"
	"fmt"
	"os"
	"os/signal"
	"strconv"
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
	inValue     int
	result      string
	termChan    chan os.Signal
	ticker      *time.Ticker
	heartbeat   *time.Ticker
}

func main() {
	c := *&client{
		inFilePath:  "/sys/devices/virtual/powercap/intel-rapl-mmio/intel-rapl-mmio:0/constraint_0_power_limit_uw",
		outFilePath: "/tmp/i3status/power_cap",
		termChan:    make(chan os.Signal, 1),
		ticker:      time.NewTicker(4 * time.Second),
		heartbeat:   time.NewTicker(24 * time.Hour),
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

	fmt.Println("Service started")

	signal.Notify(c.termChan, syscall.SIGINT, syscall.SIGTERM)
	<-c.termChan
	c.quit()
}

func (c *client) loadFile() error {
	var err error
	dat, err := os.ReadFile(c.inFilePath)
	if err != nil {
		return err
	}
	c.inValue, err = strconv.Atoi(strings.TrimSpace(string(dat)))
	if err != nil {
		return err
	}

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
	switch c.inValue {
	case 28000000:
		c.result = "5"
	case 8000000:
		c.result = "4"
	case 64000000:
		c.result = "3"
	case 15000000:
		c.result = "2"
	case 7500000:
		c.result = "1"
	default:
		c.result = "0"
		fmt.Printf("Unknown value: %d",c.inValue)

	}
}

func (c *client) quit() {
	fmt.Println("Quitting...")
	c.ticker.Stop()
	c.heartbeat.Stop()
	os.Exit(0)
}
