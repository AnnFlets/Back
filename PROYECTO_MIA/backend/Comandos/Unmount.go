package Comandos

import (
	"PROYECTO_MIA/Estructuras"
	"bytes"
	"encoding/binary"
	"os"
	"strings"
)

//Función para comprobar los parámetros del comando UNMOUNT
func ComprobarParametrosUnmount(parametros []string) string{
	respuesta := ""
	id := ""
	for i := 0; i < len(parametros); i++ {
		datosParametro := strings.Split(parametros[i], "=")
		if Comparar(datosParametro[0], "id") {
			if id == "" {
				id = datosParametro[1]
			} else {
				respuesta += "---[ERROR-UNMOUNT]: Parámetro \"id\" repetido\n"
				return respuesta
			}
		} else {
			respuesta += "---[ERROR-UNMOUNT]: No se esperaba el parámetro \"" + datosParametro[0] + "\"\n"
			return respuesta
		}
	}
	if id == "" {
		respuesta += "---[ERROR-UNMOUNT]: Se requiere el parámetro \"id\" para este comando\n"
		return respuesta
	}
	respuesta += desmontarParticion(id)
	return respuesta
}

//Función que se encarga de desmontar una partición con ID determinado
func desmontarParticion(id string) string{
	respuesta := ""
	driveletter := string(id[0])
	particion := Particiones{}
	for i := 0; i < len(ParticionesMontadas); i++ {
		if id == string(ParticionesMontadas[i].Id[:]) {
			particion = ParticionesMontadas[i]
			ParticionesMontadas[i].Disco = [1]byte{}
			ParticionesMontadas[i].Nombre = [16]byte{}
			ParticionesMontadas[i].Id = [4]byte{}
			break
		}
	}
	if string(particion.Disco[:]) != driveletter {
		respuesta += "---[ERROR-UNMOUNT]: No hay una partición montada con el id \"" + id + "\"\n"
		return respuesta
	}
	mbr, res := LeerDisco("./MIA/P1/" + driveletter + ".dsk")
	respuesta += res
	var particionMBR *Estructuras.Partition
	if mbr.Mbr_partition_1.Part_name == particion.Nombre {
		particionMBR = &mbr.Mbr_partition_1
	} else if mbr.Mbr_partition_2.Part_name == particion.Nombre {
		particionMBR = &mbr.Mbr_partition_2
	} else if mbr.Mbr_partition_3.Part_name == particion.Nombre {
		particionMBR = &mbr.Mbr_partition_3
	} else if mbr.Mbr_partition_4.Part_name == particion.Nombre {
		particionMBR = &mbr.Mbr_partition_4
	}
	particionMBR.Part_status = '0'
	particionMBR.Part_id = [4]byte{}
	file, err := os.OpenFile(strings.ReplaceAll("./MIA/P1/"+driveletter+".dsk", "\"", ""), os.O_WRONLY, os.ModeAppend)
	defer file.Close()
	if err != nil {
		respuesta += "---[ERROR-UNMOUNT]: Error al abrir el archivo\n"
		return respuesta
	}
	//Se actualiza el MBR con la información de la partición montada
	file.Seek(0, 0)
	var binario bytes.Buffer
	binary.Write(&binario, binary.BigEndian, mbr)
	EscribirBytes(file, binario.Bytes())
	respuesta += "+++[COMANDO-UNMOUNT]: Se desmontó la partición \"" + id + "\"\n"
	respuesta += MostrarParticionesMontadas()
	return respuesta
}

//Función para desmontar las particiones en caso de que se borre el disco
func unmountParticiones(driveletter string){
	for i := 0; i < len(ParticionesMontadas); i++ {
		if driveletter == string(ParticionesMontadas[i].Id[0]) {
			ParticionesMontadas[i].Disco = [1]byte{}
			ParticionesMontadas[i].Nombre = [16]byte{}
			ParticionesMontadas[i].Id = [4]byte{}
		}
	}
}
