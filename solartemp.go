package main

import (
	"encoding/json"
	"io"
	"math"
	"net/http"
	"sync"
)

type Values struct {
	Dach       float32
	Oben       float32
	Unten      float32
	Durchfluss float32
}

var values Values
var mutex sync.Mutex

//Handler for a MagicMirror Module
func Handler(w http.ResponseWriter, r *http.Request) {

	mutex.Lock()
	buf, err := json.Marshal(values)
	mutex.Unlock()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.Header().Set("Access-Control-Allow-Origin", "*")
	io.WriteString(w, string(buf))
	// debug.FreeOSMemory()
}

func roundToOneDigit(x float32) float32 {
	return float32(math.Round(float64(x)*10.0) / 10.0)
}

func SetValues(dach, oben, unten float32, durchfluss int8) {
	dach = roundToOneDigit(dach)
	unten = roundToOneDigit(unten)
	oben = roundToOneDigit(oben)

	mutex.Lock()
	values.Dach = dach
	values.Oben = oben
	values.Unten = unten
	values.Durchfluss = float32(durchfluss)
	mutex.Unlock()
}
