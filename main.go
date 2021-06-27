package main

import (
	"errors"
	"fmt"
	"log"

	"go.bug.st/serial"
)

// Dataframe

//
// 1DL
// Either UVR1611 0x80 [ 0x01 ... ] (55 bytes)
// or     UVR61-3 0x90 [ 0x01 ... ] (26 bytes)

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
	Data    []byte
}

type DlModus int

const (
	MODUS_1DL DlModus = 1
	MODUS_2DL DlModus = 2
	MODUS_CAN DlModus = 3
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
		return errors.New("Unknown mode")
	}
	fmt.Printf("Mode: %d\n", d.Modus)
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

		for _, i := range df.Data {
			chkSum += i
		}
	}
	return chkSum
}

func (d *Device) Read1DL() (Dataframe, error) {
	var buf [1]byte
	log.Print("Read 1 byte")
	df := Dataframe{}
	count, err := d.Port.Read(buf[0:1])
	if err != nil {
		return df, err
	}
	if count != 1 {
		return df, err
	}
	df.TypeUVF = buf[0]
	switch df.TypeUVF {
	case UVR1611:
		df.Data = make([]byte, SZ_UVR1611)
	case UVR61_3:
		df.Data = make([]byte, SZ_UVR61_3)
	default:
		return df, err
	}
	log.Printf("have to read %d bytes", count)
	lenRead, err := d.Port.Read(df.Data)
	if err != nil {
		return df, err
	}
	if lenRead != len(df.Data) {
		return df, errors.New("not enough")
	}
	return df, nil
}


/*
    Eing"ange 2 byte,  low vor high
    Bitbelegung:
    TTTT TTTT
    VEEE TTTT
    T ... Eingangswert
    V ... Vorzeichen (1=Minus)
    E ... Type (Einheit) des Eingangsparameters:
    x000 xxxx  Eingang unbenutzt
    D001 xxxx  digitaler Pegel (Bit D)
    V010 TTTT  Temperatur (in 1/10 C)
    V011 TTTT  Volumenstrom (in 4 l/h)
    V110 TTTT  Strahlung (in 1 W/m)
    V111 xRRT  Temp. Raumsensor(in 1/10 C)

/* Bearbeitung der Temperatur-Sensoren
void temperaturanz(int regler)
{
  int i, j, anzSensoren = 16;
  UCHAR temp_uvr_typ = 0;
  temp_uvr_typ = uvr_typ;
  for (i = 1; i <= anzSensoren; i++)
    SENS_Art[i] = 0;

  switch (uvr_typ)
  {
  case UVR1611:
    anzSensoren = 16;
    break; /* UVR1611
  case UVR61_3:
    anzSensoren = 6;
    break; /* UVR61-3
  }

  /* vor berechnetemp() die oberen 4 Bit des HighByte auswerten!!!!
  /* Wichtig fuer Minus-Temp.
  j = 1;
  for (i = 1; i <= anzSensoren; i++)
  {
    SENS_Art[i] = eingangsparameter(akt_daten[j + 1]);
    switch (SENS_Art[i])
    {
    case 0:
      SENS[i] = 0;
      break;
    case 1:
      SENS[i] = 0;
      break; // digit. Pegel (AUS)
    case 2:
      SENS[i] = berechnetemp(akt_daten[j], akt_daten[j + 1], SENS_Art[i]);
      break; // Temp.
    case 3:
      SENS[i] = berechnevol(akt_daten[j], akt_daten[j + 1]);
      break;
    case 6:
      SENS[i] = berechnetemp(akt_daten[j], akt_daten[j + 1], SENS_Art[i]);
      break; // Strahlung
    case 7:
      SENS[i] = berechnetemp(akt_daten[j], akt_daten[j + 1], SENS_Art[i]);
      break; // Raumtemp.
    case 9:
      SENS[i] = 1;
      break; // digit. Pegel (EIN)
    case 10:
      SENS[i] = berechnetemp(akt_daten[j], akt_daten[j + 1], SENS_Art[i]);
      break; // Minus-Temperaturen
    case 15:
      SENS[i] = berechnetemp(akt_daten[j], akt_daten[j + 1], SENS_Art[i]);
      break; // Minus-Raumtemp.
    }
    j = j + 2;
  }
  uvr_typ = temp_uvr_typ;
}
*/
func main() {
	d, err := New("/dev/ttyUSB0")
	if err != nil {
		panic(err)
	}
	b, err := d.readCurrentData()
	if err != nil {
		panic(err)
	}
	fmt.Printf("%+v", b)

}
