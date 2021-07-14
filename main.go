package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"go.bug.st/serial"
)

// Dataframe

//
// 1DL
// Either UVR1611 0x80
//				  16 2xbyte pairs for value
//				  2 bytes for bits of output

// or     UVR61-3 0x90 [ 0x01 ... ] (26 bytes)
// 				  6 2 byte pairs for values
// 				  1 bytes for 6 bits of output

// 2DL
// Either a UVR1611 or UVR61-3 followed by either UVR1611 or UVR61-3followed by one chksum

const (
	UVR1611 byte = 0x80
	UVR61_3 byte = 0x90
)

const SZ_UVR1611 = 55
const SZ_UVR61_3 = 26

type Device struct {
	Port  serial.Port
	Modus DlModus
}

type Dataframe struct {
	TypeUVF byte
	RawData []byte
}

type Sensors struct {
	MeasurementType UnitType
	Value           float32
}

type LoggEntry struct {
	Sensors     []Sensors
	Outputs     []bool
	RateOutputs []int8
	Power       []float32
	Energy      []float32
}

type DlModus int

const (
	MODUS_1DL DlModus = 1
	MODUS_2DL DlModus = 2
	MODUS_CAN DlModus = 3
)

type UnitType int

const (
	Zero            UnitType = 0
	Digital         UnitType = 1
	Temperature     UnitType = 2
	Volume          UnitType = 3
	Radiation       UnitType = 6
	Roomtemperature UnitType = 7
)

/* Modulmoduskennung abfragen */
func (d *Device) getModulmodus() (err error) {
	sendbuf := [1]byte{0x81}
	var rcvbuf [1]byte

	outNum, err := d.Port.Write(sendbuf[:])

	log.Printf("Wrote\n")
	if err != nil {
		return err
	}
	if outNum != 1 {
		return errors.New("could not write")
	}
	log.Printf("Read\n")
	inNum, err := d.Port.Read(rcvbuf[:])
	log.Printf("Read finish\n")
	if err != nil {
		return err
	}
	if inNum != 1 {
		return errors.New("could not read")
	}
	switch rcvbuf[0] {
	case 0xA8:
		d.Modus = MODUS_1DL
	case 0xD1:
		d.Modus = MODUS_2DL
	case 0xDC:
		d.Modus = MODUS_CAN
	default:
		return errors.New("unknown mode")
	}
	return nil
}

func New(device string) (*Device, error) {
	mode := &serial.Mode{
		BaudRate: 115200,
		DataBits: 8,
		Parity:   serial.NoParity,
	}
	port, err := serial.Open(device, mode)
	if err != nil {
		return nil, err
	}
	log.Printf("Open Succeeded")
	d := &Device{Port: port}
	if err != d.getModulmodus() {
		return nil, err
	}
	log.Printf("Module Mode known")
	return d, nil
}

func (d *Device) readCurrentData() ([]Dataframe, error) {
	sendBuf := []byte{0xab}
	len, err := d.Port.Write(sendBuf)
	if err != nil {
		return nil, err
	}
	if len != 1 {
		return nil, errors.New("couldn't write")
	}
	log.Print("Data Read issued")

	dfs := []Dataframe{}
	buf, err := d.Read1DL()
	if err != nil {
		return dfs, err
	}
	dfs = append(dfs, buf)
	if d.Modus == MODUS_2DL {
		buf, err = d.Read1DL()
		if err != nil {
			return dfs, err
		}
		dfs = append(dfs, buf)
	}
	var chksum [1]byte
	d.Port.Read(chksum[:])
	log.Print("Chksum read")
	if calcChksum(dfs) != chksum[0] {
		return nil, errors.New("chksum does not match")
	}
	log.Print("Chksum matched")
	return dfs, nil
}

func calcChksum(dfs []Dataframe) byte {
	var chkSum byte = 0
	for _, df := range dfs {
		chkSum += df.TypeUVF

		for _, i := range df.RawData {
			chkSum += i
		}
	}
	return chkSum
}

func (d *Device) Read1DL() (Dataframe, error) {
	var buf [1]byte
	log.Print("Read 1 byte")
	df := Dataframe{}
	count, err := d.Port.Read(buf[:])
	if err != nil {
		return df, err
	}
	if count != 1 {
		return df, err
	}
	df.TypeUVF = buf[0]
	switch df.TypeUVF {
	case UVR1611:
		df.RawData = make([]byte, SZ_UVR1611)
	case UVR61_3:
		df.RawData = make([]byte, SZ_UVR61_3)
	default:
		return df, err
	}
	log.Printf("have to read %d bytes", count)
	lenRead, err := d.Port.Read(df.RawData)
	if err != nil {
		return df, err
	}
	if lenRead != len(df.RawData) {
		return df, errors.New("not enough")
	}
	return df, nil
}

// Value Hi/Lo
func ConvertDataframes(dfs []Dataframe) []Sensors {
	var vals []Sensors
	for _, df := range dfs {
		var numSensors int
		switch df.TypeUVF {
		case UVR1611:
			numSensors = 16
		case UVR61_3:
			numSensors = 6
		}
		vals = append(vals, make([]Sensors, numSensors)...)
		for i := 0; i < numSensors; i++ {
			sensorType := (df.RawData[2*i+1] >> 4) & 0x07
			vals[i].MeasurementType = UnitType(sensorType)
			vals[i].Value = Convert2Bytes(df.RawData[2*i : 2*i+2])

			switch vals[i].MeasurementType {

			case Zero: // not configured
				if vals[i].Value != 0.0 {
					panic("expect zero")
				}
			case Digital:
				if df.RawData[2*i+1]&0x80 == 0x80 {
					vals[i].Value = 1
				} else {
					vals[i].Value = 0
				}
			case Temperature:
				// Temperature in 10 °C
				vals[i].Value /= 10.0

			case Radiation:
				// Radiation W/m^2

			case Volume:
				// Volume in 1/4 l/h
				vals[i].Value *= 4.0
			default:
				panic("Unknown Measurement Type")
			}
		}
	}
	return vals
}

func Convert2Bytes(d []byte) float32 {
	lo := d[0]
	hi := d[1]
	value20 := int16(lo) | (int16(hi&0x0f) << 8)
	if hi&0x80 == 0x80 {
		uv20 := uint16(value20) | 0xf000
		value20 = int16(uv20)
	}
	return float32(value20)
}

func ConvertDigitalOutputs(dfs []Dataframe) []bool {
	var output []bool

	for _, df := range dfs {
		var numOutputs int
		var bits uint16

		switch df.TypeUVF {

		case UVR1611:
			numOutputs = 15
			bits = uint16(df.RawData[32]) | (uint16(df.RawData[33]) << 8)
		case UVR61_3:
			numOutputs = 3
			bits = uint16(df.RawData[12])
		default:
			panic("should not happen")
		}
		tmpOutputs := make([]bool, numOutputs)
		for i := 0; i < numOutputs; i++ {
			tmpOutputs[i] = bits&0x01 == 0x01
			bits >>= 1
		}
		output = append(output, tmpOutputs...)
	}
	return output
}

func ConvertRates(dfs []Dataframe) []int8 {
	var output []int8
	for _, df := range dfs {
		var numOutputs int

		switch df.TypeUVF {

		case UVR1611:
			numOutputs = 4
		case UVR61_3:
			numOutputs = 1
		default:
			panic("should not happen")
		}
		tmpOutputs := make([]int8, numOutputs)
		for i := 0; i < numOutputs; i++ {
			b := df.RawData[34+i]
			if b&0x80 == 0 {
				tmpOutputs[i] = int8(b & 0x1f)
			}
		}
		output = append(output, tmpOutputs...)
	}
	return output
}

func ConvertHeats(dfs []Dataframe) ([]float32, []float32) {
	var power []float32
	var energy []float32
	for _, df := range dfs {
		heatEnabled := df.RawData[38]
		if heatEnabled&0x1 == 0x1 {
			p, e := ConvertPowerEnergyBytes(df.RawData[39 : 39+8])
			power = append(power, p)
			energy = append(energy, e)
		}
		if heatEnabled&0x2 == 0x2 {
			p, e := ConvertPowerEnergyBytes(df.RawData[47 : 47+8])
			power = append(power, p)
			energy = append(energy, e)
		}
	}
	return power, energy
}

// Convert to Power in 1/10 kW , Energy in 1/10 kWh
func ConvertPowerEnergyBytes(b []byte) (float32, float32) {
	lowlow := b[0]
	var val int32
	for i := 3; i >= 1; i-- {
		val <<= 8
		val |= int32(b[i])
	}
	power := (float32(val) + float32(lowlow)/256.0) / 10.0

	energy := float32(int16(b[5])<<8|int16(b[4])) / 10.0
	energy += float32(int16(b[7])<<8|int16(b[6])) * 1000.
	return power, energy
}

var b2i = map[bool]int{false: 0, true: 1}

// Convert to CSV (Same as dlogg-linux)
func printCSV(sensors []Sensors, digitalOuts []bool, rates []int8, powers []float32, energies []float32) string {

	builder := &strings.Builder{}
	t := time.Now()
	builder.WriteString(t.Format("02.01.06;15:04:05;"))

	for _, v := range sensors {
		if v.MeasurementType == 0 {
			builder.WriteString(" ---;")
		} else {
			fmt.Fprintf(builder, " %.1f;", v.Value)
		}
	}

	for i := 0; i < 2; i++ {
		v := b2i[digitalOuts[i]]
		fmt.Fprintf(builder, " %d;%d;", v, rates[i])
	}
	for i := 2; i < 5; i++ {
		v := b2i[digitalOuts[i]]
		fmt.Fprintf(builder, " %d;", v)
	}
	for i := 5; i < 7; i++ {
		v := b2i[digitalOuts[i]]
		fmt.Fprintf(builder, " %d;%d;", v, rates[i-4])
	}
	for i := 7; i < 12; i++ {
		v := b2i[digitalOuts[i]]
		fmt.Fprintf(builder, " %d;", v)
	}
	for i := 0; i < len(powers); i++ {
		fmt.Fprintf(builder, " %.1f;%.1f;", powers[i], energies[i])
	}
	builder.WriteString(" ---; ---;")

	builder.WriteString("\n")
	return builder.String()
}

// Layout

func loop(d *Device) {
	for {
		b, err := d.readCurrentData()
		if err != nil {
			panic(err)
		}

		sensors := ConvertDataframes(b)
		digitalOuts := ConvertDigitalOutputs(b)
		rates := ConvertRates(b)
		powers, energies := ConvertHeats(b)
		//logg := LoggEntry{Sensors: sensors, Outputs: digitalOuts, RateOutputs: rates, Power: powers, Energy: energies}
		//buf, _ := json.Marshal(logg)
		//fmt.Printf("%s\n", string(buf))
		t := time.Now()
		filename := t.Format("E0601.csv")
		f, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Fprintf(f, "%s", printCSV(sensors, digitalOuts, rates, powers, energies))
		f.Close()
		time.Sleep(time.Second * 30)
	}
}

func main() {
	d, err := New("/dev/ttyUSB0")
	if err != nil {
		panic(err)
	}
	loop(d)
}
