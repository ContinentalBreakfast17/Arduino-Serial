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
	fmt.Println("Starting", baud)
	arduino, err := NewArduino(MODEL_UNO, os.Getenv("RGB_PORT"), baud)
	check(err, -1)
	for i := 0; i < 500; i++ {
		_, err = arduino.DigitalWrite(5, DIGITAL_LOW)
		check(err, i)
		_, err = arduino.AnalogRead(1)
		check(err, i)
	}
	err = arduino.Disconnect()
	check(err, -1)
	fmt.Println("Finished", baud)
}

func check(err error, i int) {
	if err != nil {
		fmt.Println(i)
		panic(err)
	}
}