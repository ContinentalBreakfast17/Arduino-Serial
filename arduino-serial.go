package seriard

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/mikepb/go-serial"
)

type Arduino struct {
	port 		*serial.Port
	model 		config
}

type config struct {
	Name 		string
	Pins 		int
	AnalogPins	int
	PWM 		[]int
}

func Connect(model, port string, baudrate int) (*Arduino, error) {
	loadModels()

	var err error
	var ok bool
	arduino := Arduino{}
	arduino.model, ok = models[model]
	if !ok {
		return nil, errors.New(fmt.Sprintf("arduino_serial: Unsupported model name '%s'", model))
	}

	options := serial.RawOptions
	options.BitRate = baudrate // make sure this will cause serial.Open to error with improper baudrate
	options.Mode = serial.MODE_READ_WRITE

	arduino.port, err = options.Open(port)
	if err != nil {
		return nil, err
	}

	err = arduino.port.Reset()
	if err != nil {
		return nil, err
	}

	time.Sleep(2 * time.Second)

	return &arduino, nil
}

func (arduino *Arduino) Disconnect() error {
	return arduino.port.Close()
}

func (arduino *Arduino) DigitalWrite(pin, val int) (int, error) {
	if pin > arduino.model.Pins {
		return -1, errors.New(fmt.Sprintf("arduino_serial: Invalid pin '%d' for model '%s'", pin, arduino.model.Name))
	} else if val != DIGITAL_HIGH && val != DIGITAL_LOW {
		return -1, errors.New(fmt.Sprintf("arduino_serial: Invalid digital write value of '%d'", val))
	}

	return arduino.write(pin, val, fmt.Sprintf("digital_write %d %d", pin, val))
}

func (arduino *Arduino) AnalogWrite(pin int, val uint8) (int, error) {
	if !in(arduino.model.PWM, pin+1) {
		return -1, errors.New(fmt.Sprintf("arduino_serial: Invalid pwm pin '%d' for model '%s'", pin, arduino.model.Name))
	}

	return arduino.write(pin, int(val), fmt.Sprintf("analog_write %d %d", pin, val))
}

func (arduino *Arduino) write(pin, val int, msg string) (int, error) {
	raw_msg := make([]byte, MESSAGE_SIZE)
	for i := 0; i < MESSAGE_SIZE; i++ {
		raw_msg[i] = ' '
	}
	copy(raw_msg[:], msg)

	_, err := arduino.port.Write(raw_msg)
	if err != nil {
		return -1, err
	}

	time.Sleep(75 * time.Millisecond)
	n, err := arduino.port.InputWaiting()
	if err != nil {
		return -1, err
	} else if n != MESSAGE_SIZE {
		return -1, errors.New(fmt.Sprintf("arduino_serial: Need a larger timeout fool - only %d bytes waiting", n))
	}

	resp, err := arduino.getResponse(pin)
	if err != nil {
		return -1, err
	}

	if val > -1 && val != resp {
		return -1, errors.New(fmt.Sprintf("arduino_serial: Wrong value readout received from arduino: expected '%d', got '%d'", val, resp))
	}

	return resp, nil
}

func (arduino *Arduino) DigitalRead(pin int) (int, error) {
	// can digital read from analog pins by adding the number of digital pins to the desired analog pin
	if pin > arduino.model.Pins + arduino.model.AnalogPins {
		return 0, errors.New(fmt.Sprintf("arduino_serial: Invalid pin '%d' for model '%s'", pin, arduino.model.Name))
	}

	return arduino.read(fmt.Sprintf("digital_read %d", pin), pin)
}

func (arduino *Arduino) AnalogRead(pin int) (float32, error) {
	if pin > arduino.model.AnalogPins {
		return 0, errors.New(fmt.Sprintf("arduino_serial: Invalid pin '%d' for model '%s'", pin, arduino.model.Name))
	}

	val, err := arduino.read(fmt.Sprintf("analog_read %d", pin), pin)
	return float32(val*5)/1023, err
}

func (arduino *Arduino) read(msg string, pin int) (int, error) {
	return arduino.write(pin, -1, msg)
}

// should change to return and int slice for functions with more/less than 2 vals
// should also have an assert function rather than writing it multiple times
func (arduino *Arduino) getResponse(expected int) (int, error) {
	raw_resp := make([]byte, MESSAGE_SIZE)
	_, err := arduino.port.Read(raw_resp)
	if err != nil {
		return 0, err
	}

	resp := strings.TrimRight(string(raw_resp), " ")
	vals := strings.Split(resp, " ")
	if len(vals) != 2 {
		return 0, errors.New(fmt.Sprintf("arduino_serial: Invalid response from arduino recieved: %s", resp))
	}

	pin, err := strconv.Atoi(vals[0])
	if err != nil {
		return 0, err 
	} else if pin != expected {
		return 0, errors.New(fmt.Sprintf("arduino_serial: Wrong pin readout received from arudino: expected '%d', got '%d'", expected, pin))
	}

	return strconv.Atoi(vals[1])
}

func (arduino *Arduino) SetPinMode(pin, mode int) (int, error) {
	if pin > arduino.model.Pins {
		return -1, errors.New(fmt.Sprintf("arduino_serial: Invalid pin '%d' for model '%s'", pin, arduino.model.Name))
	} else if mode != MODE_OUTPUT && mode != MODE_INPUT && mode != MODE_PULLUP {
		return -1, errors.New(fmt.Sprintf("arduino_serial: Invalid pinMode value of '%d'", mode))
	}

	return arduino.write(pin, mode, fmt.Sprintf("set_pin_mode %d %d", pin, mode))
}

func in(haystack []int, needle int) bool {
	for _, val := range haystack {
		if val == needle {
			return true
		}
	}
	return false
}