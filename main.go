// This example shows how to use 128x64 display over I2C
// Tested on Seeeduino XIAO Expansion Board https://wiki.seeedstudio.com/Seeeduino-XIAO-Expansion-Board/
//
// According to manual, I2C address of the display is 0x78, but that's 8-bit address.
// TinyGo operates on 7-bit addresses and respective 7-bit address would be 0x3C, which we use below.
//
// To learn more about different types of I2C addresses, please see following page
// https://www.totalphase.com/support/articles/200349176-7-bit-8-bit-and-10-bit-I2C-Slave-Addressing

package main

import (
	"machine"
	"math"
	"strconv"

	"image/color"
	"time"

	"tinygo.org/x/drivers/lsm303agr"
	"tinygo.org/x/drivers/ssd1306"
	"tinygo.org/x/drivers/ws2812"
	"tinygo.org/x/tinyfont"
)

const (
	numLEDs = 28 // Adjust this to match your LED strip
)

const (
	MID = iota
	RIGHT
	LEFT
	DOWN
	UP
)

const (
	IDLE = iota
	CENTERING
)

const (
	WHITE = iota
	BLACK
)

var (
	neo                                        machine.Pin = machine.A1
	leds                                       [28]color.RGBA
	ledBytes                                   []byte
	display                                    ssd1306.Device
	data                                       []byte
	circleArc, circleOrientation, circleRadius byte
	ledIndex, heading, offsetHeading           int
	headingDegrees                             float64
	mode                                       = IDLE

	colors = []color.RGBA{
		color.RGBA{255, 255, 255, 255},
		color.RGBA{0, 0, 0, 255},
	}

	gpioPins = []machine.Pin{
		machine.GPIO1,
		machine.GPIO0,
		machine.GPIO16,
		machine.GPIO15,
		machine.GPIO25,
	}
	debounceBtn [5]bool
	pressedBtn  [5]bool
)

func main() {
	waitSerial()

	machine.I2C0.Configure(machine.I2CConfig{
		Frequency: machine.TWI_FREQ_400KHZ,
	})

	display := ssd1306.NewI2C(machine.I2C0)
	display.Configure(ssd1306.Config{
		Address: 0x3C,
		Width:   128,
		Height:  64,
	})

	display.ClearDisplay()

	sensor := lsm303agr.New(machine.I2C0)
	err := sensor.Configure(lsm303agr.Configuration{}) //default settings
	if err != nil {
		for {
			println("Failed to configure", err.Error())
			time.Sleep(time.Second)
		}
	}
	println("Boot up")
	neo.Configure(machine.PinConfig{Mode: machine.PinOutput})

	ws := ws2812.NewWS2812(neo)

	for c := range gpioPins {
		gpioPins[c].Configure(machine.PinConfig{Mode: machine.PinInputPullup})
	}

	ledBytes = make([]byte, 28*3)
	//connect()

	x := int16(0)
	y := int16(0)
	var mx, my int32
	var xf, yf float64
	deltaX := int16(1)
	deltaY := int16(1)
	circleArc = numLEDs
	circleOrientation = byte(randomInt(0, numLEDs*2))
	circleRadius = 250
	for {
		for c := range gpioPins {
			pressedBtn[c] = false
			if !gpioPins[c].Get() {
				if !debounceBtn[c] {
					pressedBtn[c] = true
				}
				print("1")
				debounceBtn[c] = true
			} else {
				print("0")
				debounceBtn[c] = false
			}
		}
		println("----------------")

		mx, _, my, err = sensor.ReadMagneticField()
		if err != nil {
			return
		}
		xf, yf = float64(mx), float64(my)
		headingDegrees = float64((180 / math.Pi) * math.Atan2(yf, xf))

		// Calculate which LED should be lit (assuming LED 0 is at 0 degrees)
		heading = int(math.Round(headingDegrees / (360 / float64(2*numLEDs))))
		heading = numLEDs - 1 - heading - (numLEDs / 2)
		ledIndex = heading + offsetHeading
		if ledIndex < 0 {
			ledIndex += (2 * numLEDs)
		}
		ledIndex %= (2 * numLEDs)

		println("LEDINDEX", ledIndex, heading, offsetHeading)

		//println("Heading:", int(headingDegrees), "degrees, LED:", ledIndex)
		//x, y, _z, _ := sensor.ReadMagneticField()
		//println("X", x, "Y", y, "Z", z)
		//println("------------------------")

		// Clear all LEDs
		for i := range leds {
			leds[i] = color.RGBA{0, 0, 0, 0}
			ledBytes[3*i] = 0
			ledBytes[3*i+1] = 0
			ledBytes[3*i+2] = 0
		}
		if ledIndex >= 0 && ledIndex < numLEDs {
			leds[ledIndex] = color.RGBA{255, 0, 0, 0}
			ledBytes[3*ledIndex] = 255
			ledBytes[3*ledIndex+1] = 0
			ledBytes[3*ledIndex+2] = 0
		}
		ws.WriteColors(leds[:])

		switch mode {
		case IDLE:
			pixel := display.GetPixel(x, y)
			c := colors[WHITE]
			if pixel {
				c = colors[BLACK]
			}
			display.SetPixel(x, y, c)
			display.Display()

			x += deltaX
			y += deltaY

			if x == 0 || x == 127 {
				deltaX = -deltaX
			}

			if y == 0 || y == 63 {
				deltaY = -deltaY
			}
			if pressedBtn[MID] {
				mode = CENTERING
			}
			break
		case CENTERING:
			display.ClearDisplay()
			_, w := tinyfont.LineWidth(&tinyfont.Org01, "CENTERING")
			tinyfont.WriteLine(&display, &tinyfont.Org01, int16(128-w)/2, 40, "CENTERING", colors[WHITE])
			display.Display()
			if pressedBtn[MID] {
				display.ClearDisplay()
				display.Display()
				offsetHeading = (numLEDs / 2) - heading
				mode = IDLE
			}
			break
		}

		// PUBLISH TO MQTT
		data = []byte(strconv.Itoa(ledIndex))
		publishData(orientationTopic, &data)
		publishData(ledsTopic, &ledBytes)
		data = []byte{circleArc, circleOrientation, circleRadius}
		publishData(circlesTopic, &data)

		time.Sleep(50 * time.Millisecond)
	}
}

// Wait for user to open serial console
func waitSerial() {
	for !machine.Serial.DTR() {
		time.Sleep(100 * time.Millisecond)
	}
}
