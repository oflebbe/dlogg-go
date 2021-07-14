package main

import (
	"math"
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

func TestHeat(t *testing.T) {
	// 03.07.21;09:20:25;  56.3;  63.5;  50.3; ---;  52.7;  57.9;  57.0; ---; ---; ---; ---; ---; ---; ---; ---; 476.0; 1;17; 0; 0; 1; 1; 0; 0; 0; 0; 0; 0; 0; 1; 0; 0; 0; 2.2;54581.6;  ---;  ---;
	df := Dataframe{TypeUVF: 128, RawData: []byte{10, 35, 4, 35, 171, 34, 0, 0, 197, 34, 7, 35, 30, 35, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 125, 48, 13, 2, 30, 128, 128, 128, 1, 36, 48, 0, 0, 149, 25, 54, 0, 0, 1, 10, 60, 0, 1, 0, 1}}
	b := []Dataframe{df}
	sensors := ConvertDataframes(b)
	if len(sensors) != 16 {
		t.Error("length wrong")
	}
	if sensors[0].MeasurementType != Temperature || sensors[0].Value != 77.8 {
		t.Errorf("first value wrong: expected %f, got %f", 77.8, sensors[0].Value)
	}
	if sensors[15].MeasurementType != Volume || sensors[15].Value != 500.0 {
		t.Errorf("last value wrong: expected %f, got %f", 500.0, sensors[15].Value)
	}
	m := ConvertDigitalOutputs(b)
	if len(m) != 15 {
		t.Error("length wrong")
	}
	trues := map[int]int{0: 1, 2: 1, 3: 1, 9: 1}
	for k, v := range m {
		if v != (trues[k] == 1) {
			t.Errorf("output num %d should be false", k)
		}
	}

	rates := ConvertRates(b)
	for k, v := range rates {
		if k == 0 {
			if v != 30 {
				t.Errorf("rates at %d should be 30 but is %d", k, v)
			}
		} else {
			if v != 0 {
				t.Errorf("rates at %d is %d", k, v)
			}
		}
	}
	powers, energies := ConvertHeats(b)

	if powers[0] <= 4.814062 || powers[0] >= 4.814064 {
		t.Errorf("unexpected %f", powers[0])
	}
	if energies[0] != 54654.898438 {
		t.Errorf("unexpected %f", energies[0])
	}
}

func Test_6_7_21(t *testing.T) {
	//06.07.21;10:55:33;  71.0;  63.1;  52.8; ---;  59.5;  65.3;  69.1; ---; ---; ---; ---; ---; ---; ---; ---; 492.0; 1;30; 0; 0; 1; 1; 0; 0; 0; 0; 0; 0; 0; 1; 0; 0; 0; 5.1;54621.7;  ---;  ---;
	df := Dataframe{TypeUVF: 128, RawData: []byte{198, 34, 119, 34, 17, 34, 0, 0, 83, 34, 141, 34, 180, 34, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 125, 48, 13, 2, 30, 128, 128, 128, 1, 247, 51, 0, 0, 73, 24, 54, 0, 0, 1, 0, 1, 0, 1, 0, 1}}
	b := []Dataframe{df}
	sensors := ConvertDataframes(b)
	if len(sensors) != 16 {
		t.Error("length wrong")
	}
	if sensors[0].MeasurementType != Temperature || sensors[0].Value != 71.0 {
		t.Errorf("first value wrong: expected %f, got %f", 56.3, sensors[0].Value)
	}
	if sensors[15].MeasurementType != Volume || sensors[15].Value != 500.0 {
		t.Errorf("last value wrong: expected %f, got %f", 500.0, sensors[15].Value)
	}
	powers, energies := ConvertHeats(b)
	if math.Abs(powers[0]-float32(5.196485)) > 1.0e-5 {
		t.Errorf("unexpected %f, %f", powers[0], 5.196485)
	}
	if energies[0] != 54621.699219 {
		t.Errorf("unexpected %f", energies[0])
	}
}
