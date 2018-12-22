package main

import (
	"fmt"
	"log"
	"time"

	nsq "github.com/bitly/go-nsq"
	wemo "github.com/danward79/go.wemo"
	"golang.org/x/net/context"
)

type WashingMachine string

const (
	Idle     WashingMachine = "idle"
	Washing  WashingMachine = "washing"
	Finished WashingMachine = "finished"
)

type WashingMachineInsightThreshhold int

const (
	startedWashing  WashingMachineInsightThreshhold = 100000
	finishedWashing WashingMachineInsightThreshhold = 100
)

func main() {
	washingMachine := Idle
	publishToNsq("wemo:started_monitoring")

	// you can either create a device directly OR use the
	// #Discover/#DiscoverAll methods to find devices
	device := &wemo.Device{Host: "192.168.2.3:49153"}

	ctx := context.Background()

	// retrieve device info
	deviceInfo, _ := device.FetchDeviceInfo(ctx)
	fmt.Printf("Found => %+v\n", deviceInfo)

	// device controls
	//	device.On()
	//	device.Off()
	//	device.Toggle()
	//	device.GetBinaryState()

	tickChan := time.NewTicker(time.Second * 10).C

	insightParams := device.GetInsightParams()
	initPower := insightParams.Power
	powerChangedChan := make(chan int)

	go func() {
		time.Sleep(time.Second * 1)
		for {
			select {
			case newPower := <-powerChangedChan:
				fmt.Printf("wemo:power:%v\n", newPower)
				powerMeasurement := WashingMachineInsightThreshhold(newPower)
				if washingMachine == Idle && powerMeasurement > startedWashing {
					washingMachine = Washing
					publishToNsq("wemo:started_washing")
				} else if washingMachine == Washing && powerMeasurement < finishedWashing {
					washingMachine = Finished
					publishToNsq("wemo:finished_washing")
				}
				fmt.Printf("wemo:%v\n", washingMachine)
			}
		}
	}()

	for {
		select {
		case <-tickChan:
			insightParams = device.GetInsightParams()
			if insightParams != nil {
				fmt.Printf("New insights '%v' \n", insightParams.Power)
				if insightParams.Power != initPower {
					initPower = insightParams.Power
					powerChangedChan <- insightParams.Power
				} else {
				}
			} else {
				log.Panic("Could not fetch device data")
			}
		case <-powerChangedChan:
			fmt.Printf("Power changed: '%v'\n", initPower)
		}
	}

	//	device.BinaryState() // returns 0 or 1
}

func publishToNsq(msg string) {
	config := nsq.NewConfig()
	w, _ := nsq.NewProducer("192.168.2.17:4150", config)
	err := w.Publish("wemo_monitor_washingmachine", []byte(msg))
	if err != nil {
		log.Panic("Could not connect")
	}
	w.Stop()
	fmt.Println(msg)
}
