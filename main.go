package main

import (
	"flag"
	"fmt"
	"log"
	"strings"

	"encoding/json"
	"github.com/paypal/gatt"
	"github.com/paypal/gatt/examples/option"
	"io/ioutil"
)

var (
	conffile     = flag.String("config", "config.json", "Configuration file")
	done         = make(chan struct{})
	baseDeviceId string
)

func loadConfiguration() {
	confText, err := ioutil.ReadFile(*conffile)
	if err != nil {
		log.Fatal("Error while reading configuration file: ", err)
	}
	var conf map[string]interface{}
	err = json.Unmarshal(confText, &conf)
	if err != nil {
		log.Fatal("Invalid configuration format: ", err)
	}

	baseDeviceId = conf["device_id"].(string)
}

func onStateChanged(d gatt.Device, s gatt.State) {
	log.Println("State of BT:", s)
	switch s {
	case gatt.StatePoweredOn:
		log.Println("Scanning for ID:", baseDeviceId)
		d.Scan([]gatt.UUID{}, false)
		return
	default:
		d.StopScanning()
	}
}

func onPeriphDiscovered(p gatt.Peripheral, a *gatt.Advertisement, rssi int) {
	if strings.ToUpper(p.ID()) != strings.ToUpper(baseDeviceId) {
		return
	}

	// Stop scanning once we've got the peripheral we're looking for.
	p.Device().StopScanning()

	if len(p.Name()) > 0 {
		log.Println("Device found: ", p.Name())
	} else {
		log.Println("Device found")
	}
	log.Println("	Local Name        =", a.LocalName)
	log.Println("	TX Power Level    =", a.TxPowerLevel)
	log.Println("	Manufacturer Data =", a.ManufacturerData)
	log.Println("	Service Data      =", a.ServiceData)

	p.Device().Connect(p)
}

func getServiceById(p gatt.Peripheral, serviceId Gid, name string) *gatt.Service {
	serv, err := p.DiscoverServices(serviceId.asUUID())
	if (err != nil) || (len(serv) == 0) {
		log.Fatalf("%s service not found", name)
	}
	return serv[0]
}

func checkBatteryLevel(p gatt.Peripheral) {
	// battery has its Id from bt standarts
	serv := getServiceById(p, SERVICE_BATTERY, "Battery")

	log.Printf("Battery found (service=%s)\n", serv.UUID().String())

	// Get battery level characteristic
	cs, err := p.DiscoverCharacteristics(CHAR_BATTERY.asUUID(), serv)
	if err != nil || len(cs) == 0 {
		log.Fatal("Failed to discover battery characteristics, err: %s\n", err)
	}

	var btrchr *gatt.Characteristic

	// read current battery level
	btrchr = cs[0]
	val, err := p.ReadCharacteristic(btrchr)
	if err != nil {
		log.Println("Error reading battery level: ", err)
	}
	log.Printf("Battery level: %d%%\n", val[0])

	// setup notification on battery level
	if (btrchr.Properties() & (gatt.CharNotify | gatt.CharIndicate)) != 0 {
		bnf := func(c *gatt.Characteristic, b []byte, err error) {
			log.Printf("Battery level: %d%%\n", b[0])
		}
		if err := p.SetNotifyValue(btrchr, bnf); err != nil {
			log.Fatal("Failed to subscribe battery characteristic, err: %s\n", err)
		}
	}
	close(done)
}

func onPeriphConnected(p gatt.Peripheral, err error) {
	log.Println("Connection ok")
	defer p.Device().CancelConnection(p)

	checkBatteryLevel(p)
}

func onPeriphDisconnected(p gatt.Peripheral, err error) {
	fmt.Println("Disconnected")
	close(done)
}

func main() {
	flag.Parse()
	loadConfiguration()
	initValidator()

	d, err := gatt.NewDevice(option.DefaultClientOptions...)
	if err != nil {
		log.Fatalf("Failed to open device, err: %s\n", err)
		return
	}

	// Register handlers.
	d.Handle(
		gatt.PeripheralDiscovered(onPeriphDiscovered),
		gatt.PeripheralConnected(onPeriphConnected),
		gatt.PeripheralDisconnected(onPeriphDisconnected),
	)

	d.Init(onStateChanged)
	<-done
	fmt.Println("Done")
}
