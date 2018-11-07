package main

import (
	"github.com/glendc/go-external-ip"
	"github.com/kylegrantlucas/speedtest"
	"log"
	"os"
	"quadstingray/speedtest-influxdb/model"
	"sort"
	"time"
)

func main() {
	settings := model.Parser()

	if (settings.ListServers) {
		listServers()
		os.Exit(0);
	}

	for true {
		log.Println("speed test started")
		stats := runTest(settings)

		if settings.InfluxDbSettings.Use_Influx {
			model.SaveToInfluxDb(stats, settings)
		}

		if settings.ShowMyIp {
			log.Printf("Ping: %3.2f ms | Download: %3.2f Mbps | Upload: %3.2f Mbps | External_Ip: %s", stats.Ping, stats.Down_Mbs, stats.Up_Mbs, stats.External_Ip)
		} else {
			log.Printf("Ping: %3.2f ms | Download: %3.2f Mbps | Upload: %3.2f Mbps", stats.Ping, stats.Down_Mbs, stats.Up_Mbs)
		}
		log.Printf("sleep for %v seconds", settings.Interval)
		time.Sleep(time.Duration(settings.Interval) * time.Second)
	}
}

func listServers() {
	client, err := speedtest.NewDefaultClient()
	if err != nil {
		log.Fatalf("error creating client: %v", err)
	}

	allServers, err := client.HTTPClient.GetServers()
	if err != nil {
		log.Fatalf("error creating client: %v", err)
	}

	sort.Slice(allServers, func(i, j int) bool {
		return allServers[i].Country < allServers[j].Country
	})

	for _, v := range allServers {
		log.Printf("County: %v | Location: %v | ServerId: %v | Sponsor: %v", v.Country, v.Name, v.ID, v.Sponsor)
	}
	os.Exit(0);
}

func runTest(settings model.Settings) model.SpeedTestStatistics {
	client, err := speedtest.NewDefaultClient()
	if err != nil {
		log.Fatalf("error creating client: %v", err)
	}

	// Pass an empty string to select the fastest server
	server, err := client.GetServer(settings.Server)
	if err != nil {
		log.Fatalf("error getting server: %v", err)
	}

	dmbps, err := client.Download(server)
	if err != nil {
		log.Fatalf("error getting download: %v", err)
	}

	umbps, err := client.Upload(server)
	if err != nil {
		log.Fatalf("error getting upload: %v", err)
	}

	consensus := externalip.DefaultConsensus(nil, nil)
	externalIp, err := consensus.ExternalIP()
	if err != nil {
		log.Fatalf("error getting externalIp: %v", err)
	}

	return model.SpeedTestStatistics{server.ID, server.Name + ", " + server.Country, server.Latency, dmbps, umbps, server.Distance, externalIp.String()}
}
