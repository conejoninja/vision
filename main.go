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
	NORTH = iota
	CIRCLE
	MAZE
	GAMEOVER
)

const (
	WHITE = iota
	BLACK
	RED
	GREEN
	BLUE
)

var (
	neo                                            machine.Pin = machine.A1
	leds                                           [28]color.RGBA
	ledBytes                                       []byte
	display                                        ssd1306.Device
	data                                           []byte
	circleArc, circleOrientation                   byte
	ledIndex, heading, offsetHeading, circleRadius int
	headingRads, offsetHeadingRads                 float64
	mode                                           = IDLE
	game                                           = MAZE

	colors = []color.RGBA{
		color.RGBA{255, 255, 255, 255},
		color.RGBA{0, 0, 0, 255},
		color.RGBA{255, 0, 0, 255},
		color.RGBA{0, 255, 0, 255},
		color.RGBA{0, 0, 255, 255},
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

	ledBytes = make([]byte, numLEDs*3)
	//connect()

	x := int16(0)
	y := int16(0)
	var mx, my int32
	var xf, yf float64
	px = 720
	py = 360
	deltaX := int16(1)
	deltaY := int16(1)
	circleArc = numLEDs
	circleOrientation = byte(randomInt(0, numLEDs*2))
	circleRadius = 300
	for {
		for c := range gpioPins {
			pressedBtn[c] = false
			if !gpioPins[c].Get() {
				if !debounceBtn[c] {
					pressedBtn[c] = true
				}
				//print("1")
				debounceBtn[c] = true
			} else {
				//print("0")
				debounceBtn[c] = false
			}
		}
		//println("----------------")

		mx, _, my, err = sensor.ReadMagneticField()
		if err != nil {
			return
		}
		xf, yf = float64(mx), float64(my)
		headingRads = math.Atan2(yf, xf)

		// Calculate which LED should be lit (assuming LED 0 is at 0 degrees)
		//heading = int(math.Round(headingDegrees / (360 / float64(2*numLEDs))))
		heading = int((float64(numLEDs) * headingRads) / math.Pi)
		heading = numLEDs - 1 - heading - (numLEDs / 2)
		ledIndex = heading + offsetHeading
		if ledIndex < 0 {
			ledIndex += (2 * numLEDs)
		}
		ledIndex %= (2 * numLEDs)
		//println("LEDINDEX", ledIndex, heading, offsetHeading)

		// Clear all LEDs
		for i := range leds {
			leds[i] = colors[BLACK]
			ledBytes[3*i] = 0
			ledBytes[3*i+1] = 0
			ledBytes[3*i+2] = 0
		}

		switch game {
		case NORTH:
			if ledIndex >= 0 && ledIndex < numLEDs {
				leds[ledIndex] = colors[RED]
				ledBytes[3*ledIndex] = 255
				ledBytes[3*ledIndex+1] = 0
				ledBytes[3*ledIndex+2] = 0
			}
			break
		case CIRCLE:

			brightness := 300 - circleRadius
			if brightness < 0 {
				brightness = 0
			} else if brightness > 255 {
				brightness = 255
			}
			// gamma correction
			brightness = int(math.Pow(float64(brightness)/255, 2.5) * 255)

			//println("BRIGHTNESS", brightness)
			c := color.RGBA{0, 0, byte(brightness), 255}
			success := true
			for i := byte(0); i < circleArc; i++ {
				idx := i + circleOrientation + byte(heading+offsetHeading)
				if idx < 0 {
					idx += numLEDs * 2
				}
				idx %= (numLEDs * 2)
				if idx > 0 && idx < numLEDs {
					leds[idx] = c
					if idx == 13 {
						success = false
					}
				}
			}
			leds[13] = colors[RED]

			circleRadius--
			if circleRadius < 56 {
				circleRadius = 300
				if !success {
					game = GAMEOVER
				}
			}

			break
		case GAMEOVER:
			for i := 0; i < 5; i++ {
				for i := range leds {
					leds[i] = colors[RED]
				}
				ws.WriteColors(leds[:])
				time.Sleep(600 * time.Millisecond)
				for i := range leds {
					leds[i] = colors[BLACK]
				}
				ws.WriteColors(leds[:])
				time.Sleep(600 * time.Millisecond)
			}
			game = CIRCLE
			break
		case MAZE:
			if pressedBtn[DOWN] {
				px += int(30 * math.Sin(offsetHeadingRads-headingRads))
				py -= int(30 * math.Cos(offsetHeadingRads-headingRads))
			}
			if pressedBtn[UP] {
				py += 30
			}
			if pressedBtn[LEFT] {
				px += 30
			}
			if pressedBtn[RIGHT] {
				px -= 30
			}
			if px < 0 {
				px = 0
			} else if px > 9600 {
				px = 9600
			}
			if py < 0 {
				py = 0
			} else if py > 9600 {
				py = 9600
			}

			for i := float64(0); i < numLEDs; i++ {
				brightness := 300 - castRay(offsetHeadingRads-headingRads-i*(math.Pi/float64(numLEDs)))
				// gamma correction
				brightness = int(math.Pow(float64(brightness)/255, 2.6) * 255)
				c := color.RGBA{0, 0, byte(brightness), 255}
				leds[int(i)] = c
			}
			println(int((headingRads*180)/math.Pi), int((offsetHeadingRads*180)/math.Pi), int(((offsetHeadingRads-headingRads)*180)/math.Pi), heading, offsetHeading)
			printTile(px, py)

			break
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
				offsetHeadingRads = headingRads
				mode = IDLE
			}
			break
		}

		// PUBLISH TO MQTT
		/*data = []byte(strconv.Itoa(ledIndex))
		publishData(orientationTopic, &data)
		publishData(ledsTopic, &ledBytes)
		data = []byte{circleArc, circleOrientation, circleRadius}
		publishData(circlesTopic, &data)*/

		time.Sleep(50 * time.Millisecond)
	}
}

// Wait for user to open serial console
func waitSerial() {
	for !machine.Serial.DTR() {
		time.Sleep(100 * time.Millisecond)
	}
}
