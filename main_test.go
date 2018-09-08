package seriard

import (
	"fmt"
	"os"
	"testing"
)

func TestSpeed(t *testing.T) {
	/*bauds := []int{
		BAUD_600,
		BAUD_1200,
		BAUD_2400,
		BAUD_4800,
		BAUD_9600,
		BAUD_19200,
		BAUD_38400,
		BAUD_57600,
		BAUD_115200,
	}*/
	baud := BAUD_115200
	arduino, err := NewArduino(MODEL_UNO, os.Getenv("RGB_PORT"), baud)
	check(err, -1)
	for i := 0; i < 500; i++ {
		err = arduino.DigitalWrite(5, DIGITAL_LOW)
		check(err, i)
		_, err = arduino.AnalogRead(1)
		check(err, i)
	}
	err = arduino.Disconnect()
	check(err, -1)
}

func TestFunc(t *testing.T) {
	arduino, err := NewArduino(MODEL_UNO, os.Getenv("RGB_PORT"), BAUD_115200)
	check(err, -1)
	err = arduino.AnalogWrite(9, 128)
	check(err, 1)
	err = arduino.AnalogWrite(10, 64)
	check(err, 2)
	err = arduino.AnalogWrite(11, 230)
	check(err, 3)
	err = arduino.CustomCommand("set_rgb_mode", "fade")
	check(err, 4)
	err = arduino.CustomCommand("set_speed", "15")
	check(err, 5)
	err = arduino.Disconnect()
	check(err, -2)
}

func check(err error, i int) {
	if err != nil {
		fmt.Println(i)
		panic(err)
	}
}