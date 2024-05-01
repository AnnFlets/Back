package Comandos

import (
	"PROYECTO_MIA/Estructuras"
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"math/big"
	"os"
	"strconv"
	"strings"
	"time"
)

//Función para comprobar los parámetros del comando MKDISK
func ComprobarParametrosMKDISK(parametros []string) string{
	respuesta := ""
	path := ""
	letras := []string{"A", "B", "C", "D", "E", "F", "G", "H", "I", "J", "K", "L", "M", "N", "O", "P", "Q", "R", "S", "T", "U", "V", "W", "X", "Y", "Z"}
	letraDisponible := false
	for i := 0; i < len(letras); i++ {
		path = "./MIA/P1/" + letras[i] + ".dsk"
		if !VerificarArchivoExiste(path) {
			letraDisponible = true
			break
		}
	}
	if !letraDisponible {
		respuesta += "---[ERROR-MKDISK]: Ya no pueden crearse más discos\n"
		return respuesta
	}
	size := ""
	fit := ""
	unit := ""
	for i := 0; i < len(parametros); i++ {
		datosParametro := strings.Split(parametros[i], "=")
		if Comparar(datosParametro[0], "size") {
			if size == "" {
				size = datosParametro[1]
			} else {
				respuesta += "---[ERROR-MKDISK]: Parámetro \"size\" repetido\n"
				return respuesta
			}
		} else if Comparar(datosParametro[0], "fit") {
			if fit == "" {
				fit = datosParametro[1]
			} else {
				respuesta += "---[ERROR-MKDISK]: Parámetro \"fit\" repetido\n"
				return respuesta
			}
		} else if Comparar(datosParametro[0], "unit") {
			if unit == "" {
				unit = datosParametro[1]
			} else {
				respuesta += "---[ERROR-MKDISK]: Parámetro \"unit\" repetido\n"
				return respuesta
			}
		} else {
			respuesta += "---[ERROR-MKDISK]: No se esperaba el parámetro \"" + datosParametro[0] + "\"\n"
			return respuesta
		}
	}
	if size == "" {
		respuesta += "---[ERROR-MKDISK]: Se requiere el parámetro \"size\" para este comando\n"
		return respuesta
	}
	if fit == "" {
		fit = "FF"
	}
	if unit == "" {
		unit = "M"
	}
	sizeInt, err := strconv.Atoi(size)
	if err != nil {
		respuesta += "---[ERROR-MKDISK]: El parámetro \"size\" debe ser un valor entero\n"
		return respuesta
	}
	if sizeInt <= 0 {
		respuesta += "---[ERROR-MKDISK]: El parámetro \"size\" debe ser un valor mayor a 0\n"
		return respuesta
	}
	if !Comparar(fit, "BF") {
		if !Comparar(fit, "FF") {
			if !Comparar(fit, "WF") {
				respuesta += "---[ERROR-MKDISK]: El parámetro \"fit\" posee un valor no esperado\n"
				return respuesta
			}
		}
	}
	if !Comparar(unit, "K") {
		if !Comparar(unit, "M") {
			respuesta += "---[ERROR-MKDISK]: El parámetro \"unit\" posee un valor no esperado\n"
			return respuesta
		}
	}
	if Comparar(unit, "K") {
		sizeInt = 1024 * sizeInt
	} else if Comparar(unit, "M") {
		sizeInt = 1024 * 1024 * sizeInt
	}
	respuesta += crearDisco(path, sizeInt, fit)
	return respuesta
}

//Función para crear el disco con los atributos solicitados
func crearDisco(path string, size int, fit string) string{
	respuesta := ""
	//Se crea y se asignan los datos al MBR
	var mbr = Estructuras.NuevoMBR()
	mbr.Mbr_tamano = int64(size)
	fecha := time.Now().String()
	copy(mbr.Mbr_fecha_creacion[:], fecha)
	aleatorio, _ := rand.Int(rand.Reader, big.NewInt(999999999))
	entero, _ := strconv.Atoi(aleatorio.String())
	mbr.Mbr_disk_signature = int64(entero)
	copy(mbr.Dsk_fit[:], string(fit[0]))
	mbr.Mbr_partition_1 = Estructuras.NuevaPartition()
	mbr.Mbr_partition_2 = Estructuras.NuevaPartition()
	mbr.Mbr_partition_3 = Estructuras.NuevaPartition()
	mbr.Mbr_partition_4 = Estructuras.NuevaPartition()
	//Se crea el disco
	file, err := os.Create(path)
	defer file.Close()
	if err != nil {
		respuesta += "---[ERROR-MKDISK]: No pudo crerse el disco\n"
		return respuesta
	}
	var vacio int8 = 0
	refVacio := &vacio
	tamanio := int64(size) - 1
	var binario1 bytes.Buffer
	//Escribe la representación binaria de refVacio en &binario1, utilizando el orden BigEndian (escribe el primer byte "0" del archivo)
	binary.Write(&binario1, binary.BigEndian, refVacio)
	EscribirBytes(file, binario1.Bytes())
	//Se mueve el puntero la cantidad de bytes de "tamanio" desde el inicio del archivo (llena el archivo binario con 0)
	file.Seek(tamanio, 0)
	var binario2 bytes.Buffer
	binary.Write(&binario2, binary.BigEndian, refVacio)
	EscribirBytes(file, binario2.Bytes())
	//Se mueve el puntero al inicio del archivo
	file.Seek(0, 0)
	var binario3 bytes.Buffer
	//Escribe la información del MBR en el archivo
	binary.Write(&binario3, binary.BigEndian, mbr)
	EscribirBytes(file, binario3.Bytes())
	nombreDisco := strings.Split(path, "/")
	respuesta += "+++[COMANDO-MKDISK]: Disco \"" + nombreDisco[len(nombreDisco)-1] + "\" creado con éxito\n"
	return respuesta
}
