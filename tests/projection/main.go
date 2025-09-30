package main

import (
	"machine"
	"math"
	"time"

	"tinygo.org/x/drivers/lsm303agr"
)

var (
	minX, maxX, minY, maxY, minZ, maxZ int32
)

func main() {
	// Configurar I2C
	machine.I2C0.Configure(machine.I2CConfig{
		Frequency: machine.TWI_FREQ_400KHZ,
	})

	// Crear el dispositivo LSM303AGR
	sensor := lsm303agr.New(machine.I2C0)

	// Configurar el sensor
	err := sensor.Configure(lsm303agr.Configuration{})
	if err != nil {
		println("Error configurando LSM303AGR: ", err)
		return
	}

	println("LSM303AGR inicializado. Leyendo norte magnético...")

	minX, maxX, minY, maxY, minZ, maxZ = calibrateMagnetometer(sensor)

	for {
		// Obtener el norte magnético
		heading, err := getMagneticHeading(sensor)
		if err != nil {
			println("Error leyendo heading: ", err)
			time.Sleep(1 * time.Second)
			continue
		}

		// Convertir a dirección cardinal
		direction := getCardinalDirection(heading)

		println("Norte magnético: ", heading, direction)

		time.Sleep(500 * time.Millisecond)
	}
}

// Obtiene el rumbo magnético en grados (0-360°)
func getMagneticHeading(sensor *lsm303agr.Device) (float64, error) {
	// Leer el campo magnético
	x, y, _, err := applyCalibratedReading(sensor, minX, maxX, minY, maxY, minZ, maxZ)
	if err != nil {
		return 0, err
	}

	// Calcular el rumbo usando atan2 con X e Y
	// Nota: Dependiendo de la orientación del sensor, puede que necesites
	// ajustar qué ejes usar o invertir algunos valores
	heading := math.Atan2(float64(y), float64(x))

	// Convertir de radianes a grados
	heading = heading * 180.0 / math.Pi

	// Normalizar a 0-360 grados
	if heading < 0 {
		heading += 360
	}

	return heading, nil
}

// Versión mejorada que compensa la inclinación usando el acelerómetro
func getTiltCompensatedHeading(sensor *lsm303agr.Device) (float64, error) {
	// Leer magnetómetro
	mx, my, mz, err := sensor.ReadMagneticField()
	if err != nil {
		return 0, err
	}

	// Leer acelerómetro para compensar inclinación
	ax, ay, az, err := sensor.ReadAcceleration()
	if err != nil {
		return 0, err
	}

	// Normalizar vectores del acelerómetro
	norm := math.Sqrt(float64(ax*ax + ay*ay + az*az))
	if norm == 0 {
		return getMagneticHeading(sensor) // Fallback al método simple
	}

	axNorm := float64(ax) / norm
	ayNorm := float64(ay) / norm
	azNorm := float64(az) / norm

	// Calcular ángulos de pitch y roll
	pitch := math.Asin(-axNorm)
	roll := math.Atan2(ayNorm, azNorm)

	// Compensar la inclinación del magnetómetro
	mxComp := float64(mx)*math.Cos(pitch) + float64(mz)*math.Sin(pitch)
	myComp := float64(mx)*math.Sin(roll)*math.Sin(pitch) +
		float64(my)*math.Cos(roll) -
		float64(mz)*math.Sin(roll)*math.Cos(pitch)

	// Calcular heading compensado
	heading := math.Atan2(myComp, mxComp)

	// Convertir a grados
	heading = heading * 180.0 / math.Pi

	// Normalizar a 0-360 grados
	if heading < 0 {
		heading += 360
	}

	return heading, nil
}

// Convierte grados a dirección cardinal
func getCardinalDirection(degrees float64) string {
	// Normalizar a 0-360
	for degrees < 0 {
		degrees += 360
	}
	for degrees >= 360 {
		degrees -= 360
	}

	directions := []string{
		"N", "NNE", "NE", "ENE",
		"E", "ESE", "SE", "SSE",
		"S", "SSW", "SW", "WSW",
		"W", "WNW", "NW", "NNW",
	}

	// Cada dirección cubre 22.5 grados
	index := int((degrees+11.25)/22.5) % 16
	return directions[index]
}

// Calibración simple del magnetómetro
func calibrateMagnetometer(sensor *lsm303agr.Device) (int32, int32, int32, int32, int32, int32) {
	println("Iniciando calibración del magnetómetro...")
	println("Rota el sensor en todas las direcciones durante 30 segundos...")

	var minX, minY, minZ int32 = 32767, 32767, 32767
	var maxX, maxY, maxZ int32 = -32768, -32768, -32768

	startTime := time.Now()
	for time.Since(startTime) < 30*time.Second {
		x, y, z, err := sensor.ReadMagneticField()
		if err != nil {
			continue
		}

		if x < minX {
			minX = x
		}
		if x > maxX {
			maxX = x
		}
		if y < minY {
			minY = y
		}
		if y > maxY {
			maxY = y
		}
		if z < minZ {
			minZ = z
		}
		if z > maxZ {
			maxZ = z
		}

		time.Sleep(50 * time.Millisecond)
	}

	println("Calibración completada!")
	println("X:", minX, "a", maxX)
	println("Y:", minY, "a", maxY)
	println("Z:", minZ, "a", maxZ)

	return minX, maxX, minY, maxY, minZ, maxZ
}

// Aplica calibración a las lecturas del magnetómetro
func applyCalibratedReading(sensor *lsm303agr.Device, minX, maxX, minY, maxY, minZ, maxZ int32) (float64, float64, float64, error) {
	x, y, z, err := sensor.ReadMagneticField()
	if err != nil {
		return 0, 0, 0, err
	}

	// Aplicar offset y escala
	xCal := float64(x - (maxX+minX)/2)
	yCal := float64(y - (maxY+minY)/2)
	zCal := float64(z - (maxZ+minZ)/2)

	// Normalizar al rango más pequeño para mantener la forma esférica
	xRange := float64(maxX - minX)
	yRange := float64(maxY - minY)
	zRange := float64(maxZ - minZ)

	minRange := math.Min(xRange, math.Min(yRange, zRange))

	if xRange > 0 {
		xCal = xCal * minRange / xRange
	}
	if yRange > 0 {
		yCal = yCal * minRange / yRange
	}
	if zRange > 0 {
		zCal = zCal * minRange / zRange
	}

	return xCal, yCal, zCal, nil
}
