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
	ModelName	string
	Baud 		int
	InitWait	time.Duration
	SerWait		time.Duration
	// add in serial options
	port 		*serial.Port
	model 		config
}

type config struct {
	Pins 		int
	AnalogPins	int
	PWM 		[]int
}

func NewArduino(model, port string, baudrate int) (*Arduino, error) {
	arduino := &Arduino{
		ModelName: model,
		Baud: baudrate,
		InitWait: 2*time.Second,
	}
	err := arduino.Connect(port)
	return arduino, err
}

func (arduino *Arduino) Connect(port string) error {
	models := loadModels()

	var err error
	var ok bool
	arduino.model, ok = models[arduino.ModelName]
	if !ok {
		return errors.New(fmt.Sprintf("arduino_serial: Unsupported model name '%s'", arduino.ModelName))
	}

	options := serial.RawOptions
	options.BitRate = arduino.Baud
	options.Mode = serial.MODE_READ_WRITE

	arduino.port, err = options.Open(port)
	if err != nil {
		return errors.New(fmt.Sprintf("Failed to open port: %v", err))
	}

	err = arduino.port.Reset()
	if err != nil {
		return errors.New(fmt.Sprintf("Failed to reset port: %v", err))
	}

	time.Sleep(arduino.InitWait)

	return nil
}

func (arduino *Arduino) Disconnect() error {
	return arduino.port.Close()
}

func (arduino *Arduino) DigitalWrite(pin, val int) error {
	if pin > arduino.model.Pins {
		return errors.New(fmt.Sprintf("arduino_serial: Invalid pin '%d' for model '%s'", pin, arduino.ModelName))
	} else if val != DIGITAL_HIGH && val != DIGITAL_LOW {
		return errors.New(fmt.Sprintf("arduino_serial: Invalid digital write value of '%d'", val))
	}

	_, err := arduino.write(pin, val, fmt.Sprintf("digital_write %d %d", pin, val))
	return err
}

func (arduino *Arduino) AnalogWrite(pin int, val uint8) error {
	if !in(arduino.model.PWM, pin) {
		return errors.New(fmt.Sprintf("arduino_serial: Invalid pwm pin '%d' for model '%s'", pin, arduino.ModelName))
	}

	_, err := arduino.write(pin, int(val), fmt.Sprintf("analog_write %d %d", pin, val))
	return err
}

func (arduino *Arduino) DigitalRead(pin int) (int, error) {
	// can digital read from analog pins by adding the number of digital pins to the desired analog pin
	if pin > arduino.model.Pins + arduino.model.AnalogPins {
		return 0, errors.New(fmt.Sprintf("arduino_serial: Invalid pin '%d' for model '%s'", pin, arduino.ModelName))
	}

	return arduino.read(fmt.Sprintf("digital_read %d", pin), pin)
}

func (arduino *Arduino) AnalogRead(pin int) (float32, error) {
	if pin > arduino.model.AnalogPins {
		return 0, errors.New(fmt.Sprintf("arduino_serial: Invalid pin '%d' for model '%s'", pin, arduino.ModelName))
	}

	val, err := arduino.read(fmt.Sprintf("analog_read %d", pin), pin)
	return float32(val*5)/1023, err
}

func (arduino *Arduino) SetPinMode(pin, mode int) (int, error) {
	if pin > arduino.model.Pins {
		return -1, errors.New(fmt.Sprintf("arduino_serial: Invalid pin '%d' for model '%s'", pin, arduino.ModelName))
	} else if mode != MODE_OUTPUT && mode != MODE_INPUT && mode != MODE_PULLUP {
		return -1, errors.New(fmt.Sprintf("arduino_serial: Invalid pinMode value of '%d'", mode))
	}

	return arduino.write(pin, mode, fmt.Sprintf("set_pin_mode %d %d", pin, mode))
}

func (arduino *Arduino) CustomCommand(command, parameters string) error {
	_, err := arduino.write(-1, -1, fmt.Sprintf("%s %s", command, parameters))
	return err
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

	err = arduino.wait()
	if err != nil {
		return -1, errors.New(fmt.Sprintf("arduino_serial: Failed to get response from arduino: %v", err))
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

func (arduino *Arduino) read(msg string, pin int) (int, error) {
	return arduino.write(pin, -1, msg)
}

func (arduino *Arduino) wait() error {
	available := 0
	for available < MESSAGE_SIZE{
		time.Sleep((baudWait(arduino.Baud) * time.Millisecond) + arduino.SerWait)
		n, err := arduino.port.InputWaiting()
		if err != nil {
			return err
		} else if n <= 0 {
			break
		}
		available += n
	}

	if available != MESSAGE_SIZE {
		return errors.New(fmt.Sprintf("Wrong message size: %d", available))
	}
	return nil
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
	} else if pin > -1 && pin != expected {
		return 0, errors.New(fmt.Sprintf("arduino_serial: Wrong pin readout received from arudino: expected '%d', got '%d'", expected, pin))
	}

	return strconv.Atoi(vals[1])
}

func in(haystack []int, needle int) bool {
	for _, val := range haystack {
		if val == needle {
			return true
		}
	}
	return false
}