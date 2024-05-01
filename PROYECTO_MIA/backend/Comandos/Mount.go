package Comandos

import (
	"bytes"
	"encoding/binary"
	"os"
	"strconv"
	"strings"
)

//Arreglo para guardar las particiones montadas
var ParticionesMontadas [30]Particiones

type Particiones struct {
	Disco  [1]byte
	Nombre [16]byte
	Id     [4]byte
}

//Función para comprobar los parámetros del comando MOUNT
func ComprobarParametrosMount(parametros []string) string{
	respuesta := ""
	driveletter := ""
	name := ""
	for i := 0; i < len(parametros); i++ {
		datosParametro := strings.Split(parametros[i], "=")
		if Comparar(datosParametro[0], "driveletter") {
			if driveletter == "" {
				driveletter = datosParametro[1]
			} else {
				respuesta += "---[ERROR-MOUNT]: Parámetro \"driveletter\" repetido\n"
				return respuesta
			}
		} else if Comparar(datosParametro[0], "name") {
			if name == "" {
				name = datosParametro[1]
			} else {
				respuesta += "---[ERROR-MOUNT]: Parámetro \"name\" repetido\n"
				return respuesta
			}
		} else {
			respuesta += "---[ERROR-MOUNT]: No se esperaba el parámetro \"" + datosParametro[0] + "\"\n"
			return respuesta
		}
	}
	if driveletter == "" {
		respuesta += "---[ERROR-MOUNT]: Se requiere el parámetro \"driveletter\" para este comando\n"
		return respuesta
	}
	if name == "" {
		respuesta += "---[ERROR-MOUNT]: Se requiere el parámetro \"name\" para este comando\n"
		return respuesta
	}
	if !VerificarArchivoExiste("./MIA/P1/" + driveletter + ".dsk") {
		respuesta += "---[ERROR-MOUNT]: No se encontró el disco\n"
		return respuesta
	}
	respuesta += montarParticion(driveletter, name)
	return respuesta
}

//Función que se encarga de montar una partición del disco en el sistema
func montarParticion(driveletter string, name string) string{
	respuesta := ""
	mbr, res := LeerDisco("./MIA/P1/" + driveletter + ".dsk")
	respuesta += res
	particion, res := BuscarParticion(*mbr, driveletter, name)
	respuesta += res
	if particion == nil {
		respuesta += "---[ERROR-MOUNT]: No existe una partición con el nombre \"" + name + "\"\n"
		return respuesta
	}
	if particion.Part_type == 'E' {
		respuesta += "---[ERROR-MOUNT]: No puede montarse una partición extendida\n"
		return respuesta
	} else if particion.Part_type == 'L' {
		respuesta += "---[ERROR-MOUNT]: No pueden montarse particiones lógicas\n"
		return respuesta
	} else {
		if mbr.Mbr_partition_1.Part_name == particion.Part_name {
			particion = &mbr.Mbr_partition_1
		} else if mbr.Mbr_partition_2.Part_name == particion.Part_name {
			particion = &mbr.Mbr_partition_2
		} else if mbr.Mbr_partition_3.Part_name == particion.Part_name {
			particion = &mbr.Mbr_partition_3
		} else if mbr.Mbr_partition_4.Part_name == particion.Part_name {
			particion = &mbr.Mbr_partition_4
		}
		idParticionMontar := driveletter + strconv.FormatInt(particion.Part_correlative, 10) + "03"
		if VerificarParticionEstaMontada(idParticionMontar) {
			respuesta += "---[ERROR-MOUNT]: La partición \"" + name + "\" ya se encuentra montada\n"
			return respuesta
		}
		fin := 0
		nombreParticion := ""
		seAgregoParticion := false
		for i := 0; i < len(ParticionesMontadas); i++ {
			fin = bytes.IndexByte(ParticionesMontadas[i].Nombre[:], 0)
			nombreParticion = string(ParticionesMontadas[i].Nombre[:fin])
			if nombreParticion == "" {
				seAgregoParticion = true
				//Creación partición montada en la lista
				copy(ParticionesMontadas[i].Disco[:], driveletter)
				copy(ParticionesMontadas[i].Nombre[:], name)
				copy(ParticionesMontadas[i].Id[:], idParticionMontar)
				//Actualización del estado y el ID de la partición en el MBR
				particion.Part_status = '1'
				copy(particion.Part_id[:], idParticionMontar)
				file, err := os.OpenFile(strings.ReplaceAll("./MIA/P1/"+driveletter+".dsk", "\"", ""), os.O_WRONLY, os.ModeAppend)
				defer file.Close()
				if err != nil {
					respuesta += "---[ERROR-MOUNT]: Error al abrir el archivo\n"
					return respuesta
				}
				//Se actualiza el MBR con la información de la partición montada
				file.Seek(0, 0)
				var binario bytes.Buffer
				binary.Write(&binario, binary.BigEndian, mbr)
				EscribirBytes(file, binario.Bytes())
				respuesta += "---[ERROR-MOUNT]: Se montó la partición \"" + idParticionMontar + "\"\n"
				break
			}
		}
		if !seAgregoParticion {
			respuesta += "---[ERROR-MOUNT]: El arreglo de particiones montadas ha llegado a su límite\n"
			return respuesta
		}
	}
	respuesta += MostrarParticionesMontadas()
	return respuesta
}

//Función que retorna true si la partición está en la lista de particiones montadas, y false si no.
func VerificarParticionEstaMontada(idParticionMontar string) bool {
	for i := 0; i < len(ParticionesMontadas); i++ {
		if string(ParticionesMontadas[i].Id[:]) == idParticionMontar {
			return true
		}
	}
	return false
}

//Función que imprime en consola las particiones que han sido montadas
func MostrarParticionesMontadas() string{
	respuesta := ""
	fin := 0
	nombreParticion := ""
	contador := 0
	respuesta += "------------- LISTADO PARTICIONES MONTADAS -------------\n"
	for i := 0; i < len(ParticionesMontadas); i++ {
		fin = bytes.IndexByte(ParticionesMontadas[i].Nombre[:], 0)
		nombreParticion = string(ParticionesMontadas[i].Nombre[:fin])
		if nombreParticion != "" {
			respuesta += "\t* Nombre: " + nombreParticion + " - ID: " + string(ParticionesMontadas[i].Id[:]) + "\n"
			contador++
		}
	}
	if contador == 0 {
		respuesta += "\t* No hay particiones montadas"
	}
	return respuesta
}
