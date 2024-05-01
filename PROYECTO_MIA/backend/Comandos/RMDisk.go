package Comandos

import (
	"os"
	"strings"
)

//Función para comprobar los parámetros del comando RMDISK
func ComprobarParametrosRMDISK(parametros []string) string{
	respuesta := ""
	driveletter := ""
	datosParametro := strings.Split(parametros[0], "=")
	if Comparar(datosParametro[0], "driveletter") {
		driveletter = datosParametro[1]
	} else {
		respuesta += "---[ERROR-RMDISK]: No se esperaba el parámetro \"" + datosParametro[0] + "\"\n"
		return respuesta
	}
	if driveletter == "" {
		respuesta += "---[ERROR-RMDISK]: Se requiere el valor para el parámetro \"driveletter\"\n"
		return respuesta
	}
	if !VerificarArchivoExiste("./MIA/P1/" + driveletter + ".dsk") {
		respuesta += "---[ERROR-RMDISK]: No se encontró el disco solicitado\n"
		return respuesta
	}
	respuesta += eliminarDisco(driveletter)
	return respuesta
}

//Función para eliminar el disco
func eliminarDisco(driveletter string) string{
	respuesta := ""
	err := os.Remove("./MIA/P1/" + driveletter + ".dsk")
	if err != nil {
		respuesta += "---[ERROR-RMDISK]: No pudo eliminarse el disco\n"
		return respuesta
	}
	unmountParticiones(driveletter)
	respuesta += "+++[COMANDO-RMDISK]: Disco \"" + driveletter + "\" eliminado con éxito\n"
	return respuesta
}