package seriard

const (
	MODEL_UNO 		= "Arduino Uno"
)

const (
	BAUD_9600 		= 9600
)

const (
	DIGITAL_LOW		= 0
	DIGITAL_HIGH	= 1
	MODE_INPUT 		= 0
	MODE_OUTPUT 	= 1
	MODE_PULLUP		= 2
)

const MESSAGE_SIZE 	= 32

var models = make(map[string]config)

func loadModels() {
	if(len(models) > 0)  {
		return
	}

	models[MODEL_UNO] = config{
		Name: 		MODEL_UNO,
		Pins: 		14,
		AnalogPins: 6,
		PWM: 		[]int{3, 5, 6, 9, 10, 11},
	}
}