//go:build !influxdb

package main

func initInfluxdb() func(sensors []Sensors, digitalOuts []bool, rates []int8, powers, energies []float32) {
	return func(sensors []Sensors, digitalOuts []bool, rates []int8, powers, energies []float32) {}
}
