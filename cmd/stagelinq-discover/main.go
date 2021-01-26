package main

import (
	"log"
	"time"

	"github.com/icedream/go-stagelinq"
)

const (
	appName    = "Icedream StagelinQ Receiver"
	appVersion = "0.0.0"
	timeout    = 4 * time.Second
)

var stateValues = []string{
	stagelinq.EngineDeck1Play,
	stagelinq.EngineDeck1PlayState,
	stagelinq.EngineDeck1CurrentBPM,
	stagelinq.EngineDeck1TrackArtistName,
	stagelinq.EngineDeck1TrackSongName,
	stagelinq.MixerCH1faderPosition,  

	stagelinq.EngineDeck2Play,
	stagelinq.EngineDeck2PlayState,
	stagelinq.EngineDeck2CurrentBPM,
	stagelinq.EngineDeck2TrackArtistName,
	stagelinq.EngineDeck2TrackSongName,
	stagelinq.MixerCH2faderPosition, 
	
	stagelinq.EngineDeck3Play,
	stagelinq.EngineDeck3PlayState,
	stagelinq.EngineDeck3CurrentBPM,
	stagelinq.EngineDeck3TrackArtistName,
	stagelinq.EngineDeck3TrackSongName,
	stagelinq.MixerCH3faderPosition, 

	stagelinq.EngineDeck4Play,
	stagelinq.EngineDeck4PlayState,
	stagelinq.EngineDeck4CurrentBPM,
	stagelinq.EngineDeck4TrackArtistName,
	stagelinq.EngineDeck4TrackSongName,
	stagelinq.MixerCH4faderPosition, 
}

func makeStateMap() map[string]bool {
	retval := map[string]bool{}
	for _, value := range stateValues {
		retval[value] = false
	}
	return retval
}

func allStateValuesReceived(v map[string]bool) bool {
	for _, value := range v {
		if !value {
			return false
		}
	}
	return true
}

func main() {
	listener, err := stagelinq.ListenWithConfiguration(&stagelinq.ListenerConfiguration{
		DiscoveryTimeout: timeout,
		SoftwareName:     appName,
		SoftwareVersion:  appVersion,
		Name:             "testing",
	})
	if err != nil {
		panic(err)
	}
	defer listener.Close()

	listener.AnnounceEvery(time.Second)

	deadline := time.After(timeout)
	foundDevices := []*stagelinq.Device{}

	log.Printf("Listening for devices for %s", timeout)

discoveryLoop:
	for {
		select {
		case <-deadline:
			break discoveryLoop
		default:
			device, deviceState, err := listener.Discover(timeout)
			if err != nil {
				log.Printf("WARNING: %s", err.Error())
				continue discoveryLoop
			}
			if device == nil {
				continue
			}
			// ignore device leaving messages since we do a one-off list
			if deviceState != stagelinq.DevicePresent {
				continue discoveryLoop
			}
			// check if we already found this device before
			for _, foundDevice := range foundDevices {
				if foundDevice.IsEqual(device) {
					continue discoveryLoop
				}
			}
			foundDevices = append(foundDevices, device)
			log.Printf("%s %q %q %q", device.IP.String(), device.Name, device.SoftwareName, device.SoftwareVersion)

			// discover provided services
			log.Println("\tattempting to connect to this device…")
			deviceConn, err := device.Connect(listener.Token(), []*stagelinq.Service{})
			if err != nil {
				log.Printf("WARNING: %s", err.Error())
			} else {
				defer deviceConn.Close()
				log.Println("\trequesting device data services…")
				services, err := deviceConn.RequestServices()
				if err != nil {
					log.Printf("WARNING: %s", err.Error())
					continue
				}

				for _, service := range services {
					log.Printf("\toffers %s at port %d", service.Name, service.Port)
					switch service.Name {
					case "StateMap":
						stateMapTCPConn, err := device.Dial(service.Port)
						defer stateMapTCPConn.Close()
						if err != nil {
							log.Printf("WARNING: %s", err.Error())
							continue
						}
						stateMapConn, err := stagelinq.NewStateMapConnection(stateMapTCPConn, listener.Token())
						if err != nil {
							log.Printf("WARNING: %s", err.Error())
							continue
						}

						m := makeStateMap()
						for _, stateValue := range stateValues {
							stateMapConn.Subscribe(stateValue)
						}
						for state := range stateMapConn.StateC() {
							log.Printf("\t%s = %s", state.Name, state.Value)
							m[state.Name] = true
							if allStateValuesReceived(m) {
								break
							}
						}
						select {
						case err := <-stateMapConn.ErrorC():
							log.Printf("WARNING: %s", err.Error())
						default:
						}
						stateMapTCPConn.Close()
					}
				}

				log.Println("\tend of list of device data services")
			}
		}
	}

	log.Printf("Found devices: %d", len(foundDevices))
}
