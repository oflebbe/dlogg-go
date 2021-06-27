package main

import (
	"testing"
)

func TestCalcTemp(t *testing.T) {

	f := Convert2Bytes([]byte{0x0, 0x0})
	if f != 0.0 {
		t.Error("not zero")
	}
	f2 := Convert2Bytes([]byte{0x1, 0x0})
	if f2 != 1 {
		t.Error("not one")
	}

	f3 := Convert2Bytes([]byte{0xff, 0x8f})
	if f3 != -1 {
		t.Error("not minus one")
	}
	f4 := Convert2Bytes([]byte{0xff, 0xff})
	if f4 != -1 {
		t.Error("not minus one")
	}

}

func TestConversion(t *testing.T) {
	df := Dataframe{TypeUVF: 128, RawData: []byte{22, 34, 35, 35, 189, 34, 0, 0, 165, 33, 168, 34, 243, 33, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 48, 0, 0, 0, 128, 128, 128, 1, 0, 0, 0, 0, 71, 20, 54, 0, 0, 1, 0, 1, 0, 1, 0, 1}}
	m := ConvertDataframes([]Dataframe{df})
	if len(m) != 16 {
		t.Error("length wrong")
	}
	if m[0].MeasurementType != 2 || m[0].Value != 53.4 {
		t.Error("first value wrong")
	}
	if m[15].MeasurementType != Volume || m[15].Value != 0 {
		t.Error("last value wrong")
	}
}

func TestOutput(t *testing.T) {
	df := Dataframe{TypeUVF: 128, RawData: []byte{22, 34, 35, 35, 189, 34, 0, 0, 165, 33, 168, 34, 243, 33, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 48, 0, 0, 0, 128, 128, 128, 1, 0, 0, 0, 0, 71, 20, 54, 0, 0, 1, 0, 1, 0, 1, 0, 1}}

	b := []Dataframe{df}
	m := ConvertDigitalOutputs(b)
	if len(m) != 15 {
		t.Error("length wrong")
	}
	for _, v := range m {
		if v {
			t.Error("rates are zero")
		}
	}

	rates := ConvertRates(b)
	for _, v := range rates {
		if v != 0 {
			t.Error("rates are zero")
		}
	}

	powers, energies := ConvertHeats(b)

	if len(powers) != 1 || len(energies) != 1 {
		t.Error("only one heatmeasurement available")
	}

	if powers[0] != 0.0 {
		t.Errorf("unexpected %f", powers[0])
	}
	if energies[0] != 54519.101562 {
		t.Errorf("unexpected %f", energies[0])
	}

}
