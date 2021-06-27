package main

import (
	"fmt"
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
	m := ConvertDigitalOutputs([]Dataframe{df})
	if len(m) != 15 {
		t.Error("length wrong")
	}
	fmt.Printf("%v\n", m)
}
