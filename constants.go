package seriard

import (
	"math"
	"time"
)

const (
	MODEL_UNO 		= "Arduino Uno"
)

const (
	BAUD_600		= 600
	BAUD_1200		= 1200
	BAUD_2400		= 2400
	BAUD_4800		= 4800
	BAUD_9600 		= 9600
	BAUD_19200		= 19200
	BAUD_38400		= 38400
	BAUD_57600		= 57600
	BAUD_115200		= 115200
)

const (
	DIGITAL_LOW		= 0
	DIGITAL_HIGH	= 1
	MODE_INPUT 		= 0
	MODE_OUTPUT 	= 1
	MODE_PULLUP		= 2
)

const MESSAGE_SIZE 	= 32


func loadModels() map[string]config {
	models := make(map[string]config)

	models[MODEL_UNO] = config{
		Pins: 		14,
		AnalogPins: 6,
		PWM: 		[]int{3, 5, 6, 9, 10, 11},
	}

	return models
}

func baudWait(baud int) time.Duration {
	return time.Duration(math.Ceil(639508.2332 / float64(baud) + 5.033987607))
}