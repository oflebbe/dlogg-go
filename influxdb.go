package main

import (
	"fmt"
	"os"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
)

func initInfluxdb() func(sensors []Sensors, digitalOuts []bool, rates []int8, powers, energies []float32) {
	influxdb := os.Getenv("INFLUX_URL")
	if influxdb == "" {
		panic("Need to set INFLUX_URL Database url")
	}
	influxtoken := os.Getenv("INFLUX_TOKEN")
	if influxtoken == "" {
		panic("Need to set INFLUX_TOKEN access token")
	}
	influxbucket := os.Getenv("INFLUX_BUCKET")
	if influxbucket == "" {
		panic("Need to set INFLUX_BUCKET bucket name")
	}

	client := influxdb2.NewClient(influxdb, influxtoken)
	writeAPI := client.WriteAPI(influxbucket, "solar")
	influxErrors := writeAPI.Errors()

	return func(sensors []Sensors, digitalOuts []bool, rates []int8, powers, energies []float32) {

		influx := map[string]interface{}{}

		for k, v := range sensors {
			switch v.MeasurementType {
			case Temperature:
				influx[fmt.Sprintf("temperature_%d", k)] = v.Value
			case Volume:
				influx[fmt.Sprintf("volume_%d", k)] = v.Value
			case Digital:
				influx[fmt.Sprintf("digital:%d", k)] = v.Value
			}
		}

		for i := 0; i < len(powers); i++ {
			influx[fmt.Sprintf("power_%d", i)] = powers[i]
			influx[fmt.Sprintf("energy_%d", i)] = energies[i]
		}

		p := influxdb2.NewPoint(
			"solar",
			map[string]string{
				"name": "solar",
			},
			influx,
			time.Now())
		// write asynchronously
		select {
		case errors := <-influxErrors:
			fmt.Println(errors)
		default:
		}
		writeAPI.WritePoint(p)
	}
}
