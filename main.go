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
	SPEED   = 16
)

const (
	MID = iota
	RIGHT
	LEFT
	DOWN
	UP
	HAND
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
	jx                                             machine.ADC
	jy                                             machine.ADC
	leds                                           [28]color.RGBA
	ledBytes                                       []byte
	display                                        ssd1306.Device
	data                                           []byte
	circleArc, circleOrientation                   byte
	ledIndex, heading, offsetHeading, circleRadius int
	headingRads, offsetHeadingRads                 float64
	mode                                           = IDLE
	game                                           = CIRCLE

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
		machine.GPIO26,
	}
	debounceBtn [6]bool
	pressedBtn  [6]bool
)

func main() {
	waitSerial()

	machine.InitADC()

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

	_, w := tinyfont.LineWidth(&tinyfont.Org01, "BOOT UP...")
	tinyfont.WriteLine(&display, &tinyfont.Org01, int16(128-w)/2, 40, "BOOT UP...", colors[WHITE])
	display.Display()

	sensor := lsm303agr.New(machine.I2C0)
	err := sensor.Configure(lsm303agr.Configuration{}) //default settings
	if err != nil {
		display.ClearDisplay()
		_, w := tinyfont.LineWidth(&tinyfont.Org01, "FAILED")
		tinyfont.WriteLine(&display, &tinyfont.Org01, int16(128-w)/2, 40, "FAILED", colors[WHITE])
		display.Display()
		for {
			println("Failed to configure", err.Error())
			time.Sleep(time.Second)
		}
	}

	neo.Configure(machine.PinConfig{Mode: machine.PinOutput})

	ws := ws2812.NewWS2812(neo)

	for c := range gpioPins {
		gpioPins[c].Configure(machine.PinConfig{Mode: machine.PinInputPullup})
	}

	//jx := machine.ADC{machine.A2}
	jx = machine.ADC{machine.A3}
	jy = machine.ADC{machine.A2}
	jx.Configure(machine.ADCConfig{})
	jy.Configure(machine.ADCConfig{})

	ledBytes = make([]byte, numLEDs*3)

	display.ClearDisplay()
	_, w = tinyfont.LineWidth(&tinyfont.Org01, "CONNECTING")
	tinyfont.WriteLine(&display, &tinyfont.Org01, int16(128-w)/2, 40, "CONNECTING", colors[WHITE])
	display.Display()

	connect()

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

	display.ClearDisplay()
	_, w = tinyfont.LineWidth(&tinyfont.Org01, "CONNECTED")
	tinyfont.WriteLine(&display, &tinyfont.Org01, int16(128-w)/2, 40, "CONNECTED", colors[WHITE])
	display.Display()

	for {
		for c := range gpioPins {
			pressedBtn[c] = false
			if !gpioPins[c].Get() {
				if !debounceBtn[c] {
					pressedBtn[c] = true
				}
				debounceBtn[c] = true
			} else {
				debounceBtn[c] = false
			}
		}

		mx, _, my, err = sensor.ReadMagneticField()
		if err != nil {
			return
		}
		xf, yf = float64(mx), float64(my)
		headingRads = math.Atan2(yf, xf)

		// Calculate which LED should be lit (assuming LED 0 is at 0 degrees)
		heading = int((float64(numLEDs) * headingRads) / math.Pi)
		heading = numLEDs - 1 - heading - (numLEDs / 2)
		ledIndex = heading + offsetHeading
		if ledIndex < 0 {
			ledIndex += (2 * numLEDs)
		}
		ledIndex %= (2 * numLEDs)

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
				if !success {
					game = GAMEOVER
				}
				circleArc++
				circleRadius = 300
				circleOrientation = byte(randomInt(0, numLEDs*2))
			}

			data = []byte(strconv.Itoa(int(circleArc)))
			//println("DATA", data, circleArc, int(circleArc), strconv.Itoa(int(circleArc)))
			publishData(circlesArcTopic, &data)
			data = []byte(strconv.Itoa(int(circleOrientation)))
			publishData(circlesOrientationTopic, &data)
			data = []byte{
				byte(circleRadius >> 24),
				byte(circleRadius >> 16),
				byte(circleRadius >> 8),
				byte(circleRadius),
			}
			publishData(circlesRadiusTopic, &data)

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
			mapx = px
			mapy = py
			if jy.Get() < 1000 {
				px += int(SPEED * math.Sin(offsetHeadingRads-headingRads))
				py -= int(SPEED * math.Cos(offsetHeadingRads-headingRads))
			} else if jy.Get() > 64000 {
				px -= int(SPEED * math.Sin(offsetHeadingRads-headingRads))
				py += int(SPEED * math.Cos(offsetHeadingRads-headingRads))
			}
			if jx.Get() < 1000 {
				py -= int(SPEED * math.Sin(offsetHeadingRads-headingRads))
				px += int(SPEED * math.Cos(offsetHeadingRads-headingRads))
			} else if jx.Get() > 64000 {
				py += int(SPEED * math.Sin(offsetHeadingRads-headingRads))
				px -= int(SPEED * math.Cos(offsetHeadingRads-headingRads))
			}

			if maze[py/TILESIZE][px/TILESIZE] {
				px = mapx
				py = mapy
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
			println(int((headingRads*180)/math.Pi), int((offsetHeadingRads*180)/math.Pi), int(((offsetHeadingRads-headingRads)*180)/math.Pi), heading, offsetHeading, ledIndex)
			//printTile(px, py)

			data = []byte{
				byte(px >> 24),
				byte(px >> 16),
				byte(px >> 8),
				byte(px),
				byte(py >> 24),
				byte(py >> 16),
				byte(py >> 8),
				byte(py),
			}
			publishData(mazeTopic, &data)

			break
		}

		ws.WriteColors(leds[:])

		// PUBLISH TO MQTT
		data = []byte(strconv.Itoa(ledIndex))
		publishData(orientationTopic, &data)
		publishData(ledsTopic, &ledBytes)

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
			if pressedBtn[UP] {
				mode = CENTERING
			}
			break
		case CENTERING:
			display.ClearDisplay()
			_, w := tinyfont.LineWidth(&tinyfont.Org01, "CENTERING")
			tinyfont.WriteLine(&display, &tinyfont.Org01, int16(128-w)/2, 40, "CENTERING", colors[WHITE])
			display.Display()
			if pressedBtn[UP] {
				display.ClearDisplay()
				display.Display()
				offsetHeading = (numLEDs / 2) - heading
				offsetHeadingRads = headingRads
				mode = IDLE
			}
			break
		}

		time.Sleep(50 * time.Millisecond)
	}
}

// Wait for user to open serial console
func waitSerial() {
	for !machine.Serial.DTR() {
		time.Sleep(100 * time.Millisecond)
	}
}
