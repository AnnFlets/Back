package Comandos

import (
	"PROYECTO_MIA/Estructuras"
	"bytes"
	"encoding/binary"
	"log"
	"os"
	"strings"
	"unsafe"
)

/*
Función para comparar y verificar que el lexema ingresado por el usuario concuerde con el esperado
Retorna true si coinciden; false si no
*/
func Comparar(lexemaRecibido string, lexemaEsperado string) bool {
	if strings.ToUpper(lexemaRecibido) == strings.ToUpper(lexemaEsperado) {
		return true
	}
	return false
}

/*
Función que comprueba la existencia de un archivo
Retorna true si existe; false si no
*/
func VerificarArchivoExiste(ruta string) bool {
	var _, err = os.Stat(ruta)
	if os.IsNotExist(err) {
		return false
	}
	return true
}

/*
Función para leer una cantidad determinada de bytes de un archivo específico
Retorna un arreglo de bytes, que contiene los bytes leídos
*/
func LeerBytes(file *os.File, cantidad int) []byte {
	bytesLeidos := make([]byte, cantidad)
	_, err := file.Read(bytesLeidos)
	if err != nil {
		log.Fatal(err)
	}
	return bytesLeidos
}

//Método para escribir un arreglo de bytes determinado en un archivo específico
func EscribirBytes(file *os.File, bytesEscribir []byte) {
	_, err := file.Write(bytesEscribir)
	if err != nil {
		log.Fatal(err)
	}
}

//Método que retorna una referencia del MBR del disco
func LeerDisco(path string) (*Estructuras.MBR, string) {
	respuesta := ""
	mbr := Estructuras.MBR{}
	file, err1 := os.Open(strings.ReplaceAll(path, "\"", ""))
	defer file.Close()
	if err1 != nil {
		respuesta += "---[ERROR-LEER_DISCO]: No pudo abrirse el archivo"
		return nil, respuesta
	}
	file.Seek(0, 0)
	datos := LeerBytes(file, int(unsafe.Sizeof(Estructuras.MBR{})))
	buffer := bytes.NewBuffer(datos)
	err2 := binary.Read(buffer, binary.BigEndian, &mbr)
	if err2 != nil {
		respuesta += "---[ERROR-LEER_DISCO]: No pudo leerse el archivo"
		return nil, respuesta
	}
	var mbrRef *Estructuras.MBR = &mbr
	return mbrRef, respuesta
}
