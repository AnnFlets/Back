package Comandos

import (
	"PROYECTO_MIA/Estructuras"
	"bytes"
	"encoding/binary"
	"os"
	"strconv"
	"strings"
	"unsafe"
)

//Variable que indica la posición de inicio de la partición a insertar
var comienzoPartInsertar int

//Estructura para almacenar información de las particiones en el disco
type Transition struct {
	Partition int
	Start     int
	End       int
	Before    int
	After     int
}

//Función para comprobar los parámetros del comando FDISK
func ComprobarParametrosFDisk(parametros []string) string{
	respuesta := ""
	size := ""
	driveletter := ""
	name := ""
	unit := ""
	tipo := ""
	fit := ""
	eliminar := ""
	agregar := ""
	for i := 0; i < len(parametros); i++ {
		datosParametro := strings.Split(parametros[i], "=")
		if Comparar(datosParametro[0], "size") {
			if size == "" {
				size = datosParametro[1]
			} else {
				respuesta += "---[ERROR-FDISK]: Parámetro \"size\" repetido\n"
				return respuesta
			}
		} else if Comparar(datosParametro[0], "driveletter") {
			if driveletter == "" {
				driveletter = datosParametro[1]
			} else {
				respuesta += "---[ERROR-FDISK]: Parámetro \"driveletter\" repetido\n"
				return respuesta
			}
		} else if Comparar(datosParametro[0], "name") {
			if name == "" {
				name = strings.ReplaceAll(datosParametro[1], "\"", "")
			} else {
				respuesta += "---[ERROR-FDISK]: Parámetro \"name\" repetido\n"
				return respuesta
			}
		} else if Comparar(datosParametro[0], "unit") {
			if unit == "" {
				unit = datosParametro[1]
			} else {
				respuesta += "---[ERROR-FDISK]: Parámetro \"unit\" repetido\n"
				return respuesta
			}
		} else if Comparar(datosParametro[0], "type") {
			if tipo == "" {
				tipo = datosParametro[1]
			} else {
				respuesta += "---[ERROR-FDISK]: Parámetro \"type\" repetido\n"
				return respuesta
			}
		} else if Comparar(datosParametro[0], "fit") {
			if fit == "" {
				fit = datosParametro[1]
			} else {
				respuesta += "---[ERROR-FDISK]: Parámetro \"fit\" repetido\n"
				return respuesta
			}
		} else if Comparar(datosParametro[0], "delete") {
			if eliminar == "" {
				eliminar = datosParametro[1]
			} else {
				respuesta += "---[ERROR-FDISK]: Parámetro \"delete\" repetido\n"
				return respuesta
			}
		} else if Comparar(datosParametro[0], "add") {
			if agregar == "" {
				agregar = datosParametro[1]
			} else {
				respuesta += "---[ERROR-FDISK]: Parámetro \"add\" repetido\n"
				return respuesta
			}
		} else {
			respuesta += "---[ERROR-FDISK]: No se esperaba el parámetro \"" + datosParametro[0] + "\"\n"
			return respuesta
		}
	}
	if driveletter == "" {
		respuesta += "---[ERROR-FDISK]: Se requiere el parámetro \"driveletter\"\n"
		return respuesta
	}
	if !VerificarArchivoExiste("./MIA/P1/" + driveletter + ".dsk") {
		respuesta += "---[ERROR-FDISK]: El disco \"" + driveletter + "\" no existe\n"
		return respuesta
	}
	if name == "" {
		respuesta += "---[ERROR-FDISK]: Se requiere el parámetro \"name\"\n"
		return respuesta
	}
	if agregar == "" && eliminar == "" && size == "" {
		respuesta += "---[ERROR-FDISK]: Se requiere el parámetro \"size\", \"delete\" o \"add\"\n"
		return respuesta
	}
	if agregar != "" && eliminar != "" {
		respuesta += "---[ERROR-FDISK]: Se tienen los parámetros \"add\" y \"delete\" en la misma instrucción\n"
		return respuesta
	}
	if eliminar != "" {
		if !Comparar(eliminar, "full") {
			respuesta += "---[ERROR-FDISK]: Valor inválido. El parámetro \"delete\" espera el valor \"full\"\n"
			return respuesta
		} else {
			respuesta += eliminarParticion(name, driveletter)
			return respuesta
		}
	}
	if unit == "" {
		unit = "K"
	}
	if tipo == "" {
		tipo = "P"
	}
	if fit == "" {
		fit = "WF"
	}
	if agregar != "" {
		agregarInt, err2 := strconv.Atoi(agregar)
		if err2 != nil {
			respuesta += "---[ERROR-FDISK]: El parámetro \"add\" debe ser un valor entero\n"
			return respuesta
		}
		if agregarInt == 0 {
			respuesta += "---[ERROR-FDISK]: El valor de \"add\" debe ser distinto de 0\n"
			return respuesta
		}
		respuesta += administrarEspacioParticion(agregarInt, driveletter, name, unit)
		return respuesta
	}
	sizeInt, err1 := strconv.Atoi(size)
	if err1 != nil {
		respuesta += "---[ERROR-FDISK]: El parámetro \"size\" debe ser un valor entero\n"
		return respuesta
	}
	if sizeInt <= 0 {
		respuesta += "---[ERROR-FDISK]: El parámetro \"size\" debe ser un valor mayor a 0\n"
		return respuesta
	}
	if !Comparar(unit, "B") {
		if !Comparar(unit, "K") {
			if !Comparar(unit, "M") {
				respuesta += "---[ERROR-FDISK]: El parámetro \"unit\" posee un valor no esperado\n"
				return respuesta
			}
		}
	}
	if Comparar(unit, "K") {
		sizeInt = sizeInt * 1024
	} else if Comparar(unit, "M") {
		sizeInt = sizeInt * 1024 * 1024
	}
	if !Comparar(tipo, "P") {
		if !Comparar(tipo, "E") {
			if !Comparar(tipo, "L") {
				respuesta += "---[ERROR-FDISK]: El parámetro \"type\" posee un valor no esperado\n"
				return respuesta
			}
		}
	}
	if !Comparar(fit, "BF") {
		if !Comparar(fit, "FF") {
			if !Comparar(fit, "WF") {
				respuesta += "---[ERROR-FDISK]: El parámetro \"fit\" posee un valor no esperado\n"
				return respuesta
			}
		}
	}
	if agregar == "" && eliminar == "" {
		respuesta += crearParticion(sizeInt, driveletter, name, tipo, fit)
	}
	return respuesta
}

//Función que se encarga de eliminar una partición
func eliminarParticion(name string, driveletter string) string{
	respuesta := ""
	mbr, res := LeerDisco("./MIA/P1/" + driveletter + ".dsk")
	respuesta += res
	particionBuscada, res := BuscarParticion(*mbr, driveletter, name)
	respuesta += res
	if particionBuscada == nil {
		respuesta += "---[ERROR-FDISK]: No existe una partición con el nombre \"" + name + "\"\n"
		return respuesta
	}
	var particionEliminar *Estructuras.Partition
	if particionBuscada.Part_type == 'P' || particionBuscada.Part_type == 'E' {
		if mbr.Mbr_partition_1.Part_name == particionBuscada.Part_name {
			particionEliminar = &mbr.Mbr_partition_1
		} else if mbr.Mbr_partition_2.Part_name == particionBuscada.Part_name {
			particionEliminar = &mbr.Mbr_partition_2
		} else if mbr.Mbr_partition_3.Part_name == particionBuscada.Part_name {
			particionEliminar = &mbr.Mbr_partition_3
		} else if mbr.Mbr_partition_4.Part_name == particionBuscada.Part_name {
			particionEliminar = &mbr.Mbr_partition_4
		}
		file, err := os.OpenFile(strings.ReplaceAll("./MIA/P1/"+driveletter+".dsk", "\"", ""), os.O_RDWR, 0644)
		defer file.Close()
		if err != nil {
			respuesta += "---[ERROR-FDISK]: Error al abrir el archivo\n"
			return respuesta
		}
		//Escribir los 0 en el espacio de la partición
		/*
		inicio := particionEliminar.Part_start
		file.Seek(inicio, 0)
		var vacio int8 = 0
		refVacio := &vacio
		var binario2 bytes.Buffer
		binary.Write(&binario2, binary.BigEndian, refVacio)
		for i := 0; i < int(particionEliminar.Part_s); i++ {
			EscribirBytes(file, binario2.Bytes())
		}
		*/
		particionEliminar.Part_status = '0'
		particionEliminar.Part_type = 'P'
		particionEliminar.Part_fit = 'W'
		particionEliminar.Part_start = -1
		particionEliminar.Part_s = 0
		particionEliminar.Part_name = [16]byte{}
		particionEliminar.Part_correlative = -1
		particionEliminar.Part_id = [4]byte{}
		//Se actualiza la información del MBR
		file.Seek(0, 0)
		var binario1 bytes.Buffer
		binary.Write(&binario1, binary.BigEndian, mbr)
		EscribirBytes(file, binario1.Bytes())
		respuesta += "+++[COMANDO-FDISK]: Partición \"" + name + "\" eliminada con éxito\n"
	} else {
		particionExtendida := Estructuras.NuevaPartition()
		if mbr.Mbr_partition_1.Part_type == 'E' {
			particionExtendida = mbr.Mbr_partition_1
		} else if mbr.Mbr_partition_2.Part_type == 'E' {
			particionExtendida = mbr.Mbr_partition_2
		} else if mbr.Mbr_partition_3.Part_type == 'E' {
			particionExtendida = mbr.Mbr_partition_3
		} else if mbr.Mbr_partition_4.Part_type == 'E' {
			particionExtendida = mbr.Mbr_partition_4
		}
		arregloEBR, res := ObtenerParticionesLogicas(particionExtendida, driveletter)
		respuesta += res
		fin := 0
		nombreParticion := ""
		for i := 0; i < len(arregloEBR); i++ {
			fin = bytes.IndexByte(arregloEBR[i].Part_name[:], 0)
			nombreParticion = string(arregloEBR[i].Part_name[:fin])
			if nombreParticion == name {
				file, err := os.OpenFile(strings.ReplaceAll("./MIA/P1/"+driveletter+".dsk", "\"", ""), os.O_RDWR, 0644)
				defer file.Close()
				if err != nil {
					respuesta += "---[ERROR-FDISK]: Error al abrir el archivo\n"
					return respuesta
				}
				//Escribir los 0 en el espacio de la partición
				inicio := arregloEBR[i].Part_start
				file.Seek(inicio, 0)
				var vacio int8 = 0
				refVacio := &vacio
				var binario2 bytes.Buffer
				binary.Write(&binario2, binary.BigEndian, refVacio)
				for j := 0; j < int(arregloEBR[i].Part_s); j++ {
					EscribirBytes(file, binario2.Bytes())
				}
				//Se actualiza la información del EBR
				arregloEBR[i].Part_fit = 'W'
				arregloEBR[i].Part_s = 0
				arregloEBR[i].Part_name = [16]byte{}
				file.Seek(arregloEBR[i].Part_start-int64(unsafe.Sizeof(Estructuras.EBR{})), 0)
				var binario bytes.Buffer
				binary.Write(&binario, binary.BigEndian, arregloEBR[i])
				EscribirBytes(file, binario.Bytes())
				respuesta += "+++[COMANDO-FDISK]: Partición \"" + name + "\" eliminada con éxito\n"
				return respuesta
			}
		}
	}
	return respuesta
}

//Función para administrar las funciones del parámetro "add" (quitar y agregar espacio a una partición)
func administrarEspacioParticion(agregarInt int, driveletter string, name string, unit string) string{
	respuesta := ""
	mbr, res := LeerDisco("./MIA/P1/" + driveletter + ".dsk")
	respuesta += res
	particionBuscada, res := BuscarParticion(*mbr, driveletter, name)
	respuesta += res
	if particionBuscada == nil {
		respuesta += "---[ERROR-FDISK]: No existe la partición con el nombre \"" + name + "\"\n"
		return respuesta
	}
	if Comparar(unit, "K") {
		agregarInt = agregarInt * 1024
	} else if Comparar(unit, "M") {
		agregarInt = agregarInt * 1024 * 1024
	}
	if agregarInt < 0 {
		respuesta += quitarEspacioParticion(mbr, particionBuscada, agregarInt, driveletter, name)
	} else {
		respuesta += agregarEspacioParticion(mbr, particionBuscada, agregarInt, driveletter, name)
	}
	return respuesta
}

//Función que se encarga de quitar espacio a una partición determinada
func quitarEspacioParticion(mbr *Estructuras.MBR, particionBuscada *Estructuras.Partition, quitar int, driveletter string, name string) string{
	respuesta := ""
	if particionBuscada.Part_type == 'P' || particionBuscada.Part_type == 'E' {
		if mbr.Mbr_partition_1.Part_name == particionBuscada.Part_name {
			particionBuscada = &mbr.Mbr_partition_1
		} else if mbr.Mbr_partition_2.Part_name == particionBuscada.Part_name {
			particionBuscada = &mbr.Mbr_partition_2
		} else if mbr.Mbr_partition_3.Part_name == particionBuscada.Part_name {
			particionBuscada = &mbr.Mbr_partition_3
		} else if mbr.Mbr_partition_4.Part_name == particionBuscada.Part_name {
			particionBuscada = &mbr.Mbr_partition_4
		}
		var nuevoTamanoParticion int
		if particionBuscada.Part_type == 'P' {
			nuevoTamanoParticion = int(particionBuscada.Part_s) + quitar
		} else {
			arregloEBR, res := ObtenerParticionesLogicas(*particionBuscada, driveletter)
			respuesta += res
			espacioOcupadoExt := int(arregloEBR[len(arregloEBR)-1].Part_start - particionBuscada.Part_start)
			if int(particionBuscada.Part_s)-espacioOcupadoExt+quitar <= 0 {
				respuesta += "---[ERROR-FDISK]: No se pudo quitar espacio de la partición. Espacio a eliminar mayor al tamaño de la partición\n"
				return respuesta
			}
			nuevoTamanoParticion = int(particionBuscada.Part_s) + quitar
		}
		if nuevoTamanoParticion <= 0 {
			respuesta += "---[ERROR-FDISK]: No se pudo quitar espacio de la partición. Espacio a eliminar mayor al tamaño de la partición\n"
			return respuesta
		}
		particionBuscada.Part_s = int64(nuevoTamanoParticion)
		file, err := os.OpenFile(strings.ReplaceAll("./MIA/P1/"+driveletter+".dsk", "\"", ""), os.O_RDWR, 0644)
		defer file.Close()
		if err != nil {
			respuesta += "---[ERROR-FDISK]: Error al abrir el archivo\n"
			return respuesta
		}
		//Se actualiza la información del MBR
		file.Seek(0, 0)
		var binario1 bytes.Buffer
		binary.Write(&binario1, binary.BigEndian, mbr)
		EscribirBytes(file, binario1.Bytes())
		//Escribir los 0 en el espacio que se quitó de la partición
		var vacio int8 = 0
		refVacio := &vacio
		inicio := particionBuscada.Part_start + particionBuscada.Part_s
		file.Seek(inicio, 0)
		var binario2 bytes.Buffer
		binary.Write(&binario2, binary.BigEndian, refVacio)
		for i := 0; i < quitar; i++ {
			EscribirBytes(file, binario2.Bytes())
		}
		respuesta += "+++[COMANDO-FDISK]: Se quitó espacio a la partición \"" + name + "\" con éxito\n"
		return respuesta
	} else {
		particionExtendida := Estructuras.NuevaPartition()
		if mbr.Mbr_partition_1.Part_type == 'E' {
			particionExtendida = mbr.Mbr_partition_1
		} else if mbr.Mbr_partition_2.Part_type == 'E' {
			particionExtendida = mbr.Mbr_partition_2
		} else if mbr.Mbr_partition_3.Part_type == 'E' {
			particionExtendida = mbr.Mbr_partition_3
		} else if mbr.Mbr_partition_4.Part_type == 'E' {
			particionExtendida = mbr.Mbr_partition_4
		}
		arregloEBR, res := ObtenerParticionesLogicas(particionExtendida, driveletter)
		respuesta += res
		fin := 0
		nombreParticion := ""
		for i := 0; i < len(arregloEBR); i++ {
			fin = bytes.IndexByte(arregloEBR[i].Part_name[:], 0)
			nombreParticion = string(arregloEBR[i].Part_name[:fin])
			if nombreParticion == name {
				nuevoTamanoParticion := int(arregloEBR[i].Part_s) + quitar
				if nuevoTamanoParticion <= 0 {
					respuesta += "---[ERROR-FDISK]: No se pudo quitar espacio de la partición. Espacio a eliminar mayor al tamaño de la partición\n"
					return respuesta
				}
				arregloEBR[i].Part_s = int64(nuevoTamanoParticion)
				file, err := os.OpenFile(strings.ReplaceAll("./MIA/P1/"+driveletter+".dsk", "\"", ""), os.O_RDWR, 0644)
				defer file.Close()
				if err != nil {
					respuesta += "---[ERROR-FDISK]: Error al abrir el archivo\n"
					return respuesta
				}
				//Se actualiza el EBR
				file.Seek(arregloEBR[i].Part_start-int64(unsafe.Sizeof(Estructuras.EBR{})), 0)
				var binario bytes.Buffer
				binary.Write(&binario, binary.BigEndian, arregloEBR[i])
				EscribirBytes(file, binario.Bytes())
				//Escribir los 0 en el espacio que se quitó de la partición
				var vacio int8 = 0
				refVacio := &vacio
				inicio := arregloEBR[i].Part_start + arregloEBR[i].Part_s
				file.Seek(inicio, 0)
				var binario2 bytes.Buffer
				binary.Write(&binario2, binary.BigEndian, refVacio)
				for i := 0; i < quitar; i++ {
					EscribirBytes(file, binario2.Bytes())
				}
				respuesta += "+++[COMANDO-FDISK]: Se quitó espacio a la partición \"" + name + "\" con éxito\n"
				return respuesta
			}
		}
	}
	return respuesta
}

//Función que se encarga de agregar espacio a una partición determinada
func agregarEspacioParticion(mbr *Estructuras.MBR, particionBuscada *Estructuras.Partition, agregar int, driveletter string, name string) string{
	respuesta := ""
	if particionBuscada.Part_type == 'P' || particionBuscada.Part_type == 'E' {
		var arregloTransitions []Transition
		arregloTransitions, _, _, _ = AdministrarTransitions(*mbr, ObtenerArregloParticiones(*mbr), arregloTransitions)
		if mbr.Mbr_partition_1.Part_name == particionBuscada.Part_name {
			particionBuscada = &mbr.Mbr_partition_1
		} else if mbr.Mbr_partition_2.Part_name == particionBuscada.Part_name {
			particionBuscada = &mbr.Mbr_partition_2
		} else if mbr.Mbr_partition_3.Part_name == particionBuscada.Part_name {
			particionBuscada = &mbr.Mbr_partition_3
		} else if mbr.Mbr_partition_4.Part_name == particionBuscada.Part_name {
			particionBuscada = &mbr.Mbr_partition_4
		}
		var transitionUsar Transition
		for i := 0; i < len(arregloTransitions); i++ {
			if arregloTransitions[i].Start == int(particionBuscada.Part_start) {
				transitionUsar = arregloTransitions[i]
			}
		}
		var nuevoTamanoParticion int
		if transitionUsar.After < agregar {
			respuesta += "---[ERROR-FDISK]: No se pudo agregar espacio a la partición. Espacio a agregar mayor al espacio disponible después de la partición\n"
			return respuesta
		}
		nuevoTamanoParticion = int(particionBuscada.Part_s) + agregar
		particionBuscada.Part_s = int64(nuevoTamanoParticion)
		file, err := os.OpenFile(strings.ReplaceAll("./MIA/P1/"+driveletter+".dsk", "\"", ""), os.O_RDWR, 0644)
		defer file.Close()
		if err != nil {
			respuesta += "---[ERROR-FDISK]: Error al abrir el archivo\n"
			return respuesta
		}
		//Se actualiza la información del MBR
		file.Seek(0, 0)
		var binario1 bytes.Buffer
		binary.Write(&binario1, binary.BigEndian, mbr)
		EscribirBytes(file, binario1.Bytes())
		respuesta += "+++[COMANDO-FDISK]: Se agregó espacio a la partición \"" + name + "\" con éxito\n"
		return respuesta
	} else {
		particionExtendida := Estructuras.NuevaPartition()
		if mbr.Mbr_partition_1.Part_type == 'E' {
			particionExtendida = mbr.Mbr_partition_1
		} else if mbr.Mbr_partition_2.Part_type == 'E' {
			particionExtendida = mbr.Mbr_partition_2
		} else if mbr.Mbr_partition_3.Part_type == 'E' {
			particionExtendida = mbr.Mbr_partition_3
		} else if mbr.Mbr_partition_4.Part_type == 'E' {
			particionExtendida = mbr.Mbr_partition_4
		}
		arregloEBR, res := ObtenerParticionesLogicas(particionExtendida, driveletter)
		respuesta += res
		fin := 0
		nombreParticion := ""
		for i := 0; i < len(arregloEBR); i++ {
			fin = bytes.IndexByte(arregloEBR[i].Part_name[:], 0)
			nombreParticion = string(arregloEBR[i].Part_name[:fin])
			if nombreParticion == name {
				espacioLibreAfter := int(arregloEBR[i].Part_next - particionBuscada.Part_start + particionBuscada.Part_s)
				if espacioLibreAfter <= 0 || espacioLibreAfter < agregar {
					respuesta += "---[ERROR-FDISK]: No se pudo agregar espacio a la partición. Espacio a agregar mayor al espacio disponible después de la partición\n"
					return respuesta
				}
				nuevoTamanoParticion := int(arregloEBR[i].Part_s) + agregar
				arregloEBR[i].Part_s = int64(nuevoTamanoParticion)
				file, err := os.OpenFile(strings.ReplaceAll("./MIA/P1/"+driveletter+".dsk", "\"", ""), os.O_RDWR, 0644)
				defer file.Close()
				if err != nil {
					respuesta += "---[ERROR-FDISK]: Error al abrir el archivo\n"
					return respuesta
				}
				//Se actualiza el EBR
				file.Seek(arregloEBR[i].Part_start-int64(unsafe.Sizeof(Estructuras.EBR{})), 0)
				var binario bytes.Buffer
				binary.Write(&binario, binary.BigEndian, arregloEBR[i])
				EscribirBytes(file, binario.Bytes())
				respuesta += "+++[COMANDO-FDISK]: Se agregó espacio a la partición \"" + name + "\" con éxito\n"
				return respuesta
			}
		}
	}
	return respuesta
}

//Función que retorna un arreglo con la información de las particiones del disco
func ObtenerArregloParticiones(mbr Estructuras.MBR) []Estructuras.Partition {
	var arregloParticiones []Estructuras.Partition
	arregloParticiones = append(arregloParticiones, mbr.Mbr_partition_1)
	arregloParticiones = append(arregloParticiones, mbr.Mbr_partition_2)
	arregloParticiones = append(arregloParticiones, mbr.Mbr_partition_3)
	arregloParticiones = append(arregloParticiones, mbr.Mbr_partition_4)
	return arregloParticiones
}

/*
Función que se encarga de buscar entre todas las particiones aquella que tenga por nombre un valor
que coincida con el parámetro name. Se retorna dicha partición.
*/
func BuscarParticion(mbr Estructuras.MBR, driveletter string, name string) (*Estructuras.Partition, string) {
	respuesta := ""
	particiones := ObtenerArregloParticiones(mbr)
	existeParticionExtendida := false
	particionExtendida := Estructuras.NuevaPartition()
	for i := 0; i < len(particiones); i++ {
		fin := bytes.IndexByte(particiones[i].Part_name[:], 0)
		nombreParticion := string(particiones[i].Part_name[:fin])
		if Comparar(nombreParticion, name) {
			return &particiones[i], respuesta
		} else if Comparar(string(particiones[i].Part_type), "E") {
			existeParticionExtendida = true
			particionExtendida = particiones[i]
		}
	}
	if existeParticionExtendida {
		arregloEBR, res := ObtenerParticionesLogicas(particionExtendida, driveletter)
		respuesta += res
		for i := 0; i < len(arregloEBR); i++ {
			fin := bytes.IndexByte(arregloEBR[i].Part_name[:], 0)
			nombreParticion := string(arregloEBR[i].Part_name[:fin])
			if Comparar(nombreParticion, name) {
				particionTemporal := Estructuras.NuevaPartition()
				particionTemporal.Part_status = arregloEBR[i].Part_mount
				particionTemporal.Part_type = 'L'
				particionTemporal.Part_fit = arregloEBR[i].Part_fit
				particionTemporal.Part_start = arregloEBR[i].Part_start
				particionTemporal.Part_s = arregloEBR[i].Part_s
				copy(particionTemporal.Part_name[:], arregloEBR[i].Part_name[:])
				return &particionTemporal, respuesta
			}
		}
	}
	return nil, respuesta
}

//Función que retorna un arreglo de los EBRs de la partición extendida
func ObtenerParticionesLogicas(particionExtendida Estructuras.Partition, driveletter string) ([]Estructuras.EBR, string) {
	respuesta := ""
	var arregloEBR []Estructuras.EBR
	file, err1 := os.Open("./MIA/P1/" + driveletter + ".dsk")
	defer file.Close()
	if err1 != nil {
		respuesta += "---[ERROR-FDISK]: No pudo abrirse el archivo\n"
		return nil, respuesta
	}
	//ebrTemportal contendrá por el momento la información del primer EBR en la partición extendida
	ebrTemporal := Estructuras.NuevoEBR()
	file.Seek(particionExtendida.Part_start, 0)
	datos := LeerBytes(file, int(unsafe.Sizeof(Estructuras.EBR{})))
	buffer := bytes.NewBuffer(datos)
	err2 := binary.Read(buffer, binary.BigEndian, &ebrTemporal)
	if err2 != nil {
		respuesta += "---[ERROR-FDISK]: No se pudo leer el archivo\n"
		return nil, respuesta
	}
	for {
		//IF -> Si al EBR le sigue otro EBR
		if int(ebrTemporal.Part_next) != -1 {
			arregloEBR = append(arregloEBR, ebrTemporal)
			file.Seek(ebrTemporal.Part_next, 0)
			//Se lee el siguiente EBR
			datos = LeerBytes(file, int(unsafe.Sizeof(Estructuras.EBR{})))
			buffer = bytes.NewBuffer(datos)
			err2 = binary.Read(buffer, binary.BigEndian, &ebrTemporal)
			if err2 != nil {
				respuesta += "---[ERROR-FDISK]: No se pudo leer el archivo\n"
				return nil, respuesta
			}
		} else {
			arregloEBR = append(arregloEBR, ebrTemporal)
			break
		}
	}
	return arregloEBR, respuesta
}

//Función para gestionar el arreglo de Transitions (estructura que contiene información de las particiones)
func AdministrarTransitions(mbr Estructuras.MBR, arregloParticiones []Estructuras.Partition, arregloTransitions []Transition) ([]Transition, Estructuras.Partition, int, int) {
	particiones := arregloParticiones
	cantParticionesEnUso := 0
	cantParticionesExtendidas := 0
	ocupadoAntes := int(unsafe.Sizeof(Estructuras.MBR{}))
	particionExtendida := Estructuras.NuevaPartition()
	/*
		Se recorren las particiones para determinar las particiones creadas, su inicio y fin, y el espacio disponible antes y después de las mismas.
		Asimismo, para comprobar si existe alguna partición extendida en el disco.
	*/
	for i := 0; i < len(particiones); i++ {
		if int(particiones[i].Part_s) != 0 {
			var transition Transition
			transition.Partition = i
			transition.Start = int(particiones[i].Part_start)
			transition.End = int(particiones[i].Part_start + particiones[i].Part_s)
			transition.Before = transition.Start - ocupadoAntes
			ocupadoAntes = transition.End
			if cantParticionesEnUso != 0 {
				arregloTransitions[cantParticionesEnUso-1].After = transition.Start - (arregloTransitions[cantParticionesEnUso-1].End)
			}
			arregloTransitions = append(arregloTransitions, transition)
			cantParticionesEnUso++
			if Comparar(string(particiones[i].Part_type), "E") {
				cantParticionesExtendidas++
				particionExtendida = particiones[i]
			}
		}
	}
	if cantParticionesEnUso != 0 {
		arregloTransitions[len(arregloTransitions)-1].After = int(mbr.Mbr_tamano) - arregloTransitions[len(arregloTransitions)-1].End
	}
	return arregloTransitions, particionExtendida, cantParticionesExtendidas, cantParticionesEnUso
}

//Función que se encarga de crear una partición lógica dentro de una partición extendida
func crearParticionLogica(particionLogica Estructuras.Partition, particionExtendida Estructuras.Partition, driveletter string) string{
	respuesta := ""
	//Se crea un EBR con la información de la partición lógica a insertar
	ebr := Estructuras.NuevoEBR()
	ebr.Part_fit = particionLogica.Part_fit
	ebr.Part_s = particionLogica.Part_s
	copy(ebr.Part_name[:], particionLogica.Part_name[:])
	file, err1 := os.Open("./MIA/P1/" + driveletter + ".dsk")
	defer file.Close()
	if err1 != nil {
		respuesta += "---[ERROR-FDISK]: No pudo abrirse el archivo\n"
		return respuesta
	}
	//Se lee la información del primer EBR en la partición extendida y se almacena el EBRTemporal
	EBRTemporal := Estructuras.NuevoEBR()
	file.Seek(particionExtendida.Part_start, 0)
	datos := LeerBytes(file, int(unsafe.Sizeof(Estructuras.EBR{})))
	buffer := bytes.NewBuffer(datos)
	err2 := binary.Read(buffer, binary.BigEndian, &EBRTemporal)
	if err2 != nil {
		respuesta += "---[ERROR-FDISK]: No pudo leerse el archivo\n"
		return respuesta
	}
	var size int64 = 0
	for {
		//Espacio ocupado en la partición extendida
		size = size + int64(unsafe.Sizeof(Estructuras.EBR{})) + EBRTemporal.Part_s
		//Si el espacio disponible en la partición externa es menor o igual al tamaño de la partición a insertar
		if (particionExtendida.Part_s - size) <= ebr.Part_s {
			respuesta += "---[ERROR-FDISK]: No hay espacio para crear la partición lógica\n"
			return respuesta
		}
		/*
			Si el EBR leído del archivo no está definido o asociado a una partición lógica (no tiene tamaño),
			entonces se escribe el EBR en el temporal y se escribe un nuevo EBR al final.
		*/
		if EBRTemporal.Part_s == 0 {
			ebr.Part_start = EBRTemporal.Part_start
			if EBRTemporal.Part_next == -1 {
				ebr.Part_next = ebr.Part_start + ebr.Part_s
			} else {
				ebr.Part_next = EBRTemporal.Part_next
			}
			archivo, err3 := os.OpenFile("./MIA/P1/"+driveletter+".dsk", os.O_WRONLY, os.ModeAppend)
			if err3 != nil {
				respuesta += "---[ERROR-FDISK]: No pudo abrirse el archivo\n"
				return respuesta
			}
			//Se escribe en el archivo el EBR con la información de la partición a insertar
			archivo.Seek(ebr.Part_start-int64(unsafe.Sizeof(Estructuras.EBR{})), 0)
			var binario bytes.Buffer
			binary.Write(&binario, binary.BigEndian, ebr)
			EscribirBytes(archivo, binario.Bytes())
			if EBRTemporal.Part_next == -1 {
				//Se escribe en el archivo un nuevo EBR, no asociado a alguna partición
				archivo.Seek(ebr.Part_next, 0)
				nuevoEBR := Estructuras.NuevoEBR()
				nuevoEBR.Part_start = ebr.Part_next + int64(unsafe.Sizeof(Estructuras.EBR{}))
				archivo.Seek(ebr.Part_next, 0)
				var binarioNuevoEBR bytes.Buffer
				binary.Write(&binarioNuevoEBR, binary.BigEndian, nuevoEBR)
				EscribirBytes(archivo, binarioNuevoEBR.Bytes())
			}
			fin := bytes.IndexByte(particionLogica.Part_name[:], 0)
			nombreParticion := string(particionLogica.Part_name[:fin])
			respuesta += "+++[COMANDO-FDISK]: Se ha creado la partición lógica \"" + nombreParticion + "\" con éxito\n"
			return respuesta
		}
		file, err1 = os.Open("./MIA/P1/" + driveletter + ".dsk")
		if err1 != nil {
			respuesta += "---[ERROR-FDISK]: No pudo abrirse el archivo\n"
			return respuesta
		}
		//Se obtiene la información del siguiente EBR
		file.Seek(EBRTemporal.Part_next, 0)
		datos = LeerBytes(file, int(unsafe.Sizeof(Estructuras.EBR{})))
		buffer = bytes.NewBuffer(datos)
		err2 = binary.Read(buffer, binary.BigEndian, &EBRTemporal)
		if err2 != nil {
			respuesta += "---[ERROR-FDISK]: No pudo leerse el archivo\n"
			return respuesta
		}
	}
}

//Función que se encarga de generar la partición en el disco
func crearParticion(size int, driveletter string, name string, tipo string, fit string) string{
	respuesta := ""
	comienzoPartInsertar = 0
	mbr, res := LeerDisco("./MIA/P1/" + driveletter + ".dsk")
	respuesta += res
	particionBuscada, res := BuscarParticion(*mbr, driveletter, name)
	respuesta += res
	if particionBuscada != nil {
		respuesta += "---[ERROR-FDISK]: Ya existe una partición con el nombre \"" + name + "\"\n"
		return respuesta
	}
	particionesMBR := ObtenerArregloParticiones(*mbr)
	var arregloTransitions []Transition
	var particionExtendida Estructuras.Partition
	var cantParticionesExtendidas int
	var cantParticionesEnUso int
	arregloTransitions, particionExtendida, cantParticionesExtendidas, cantParticionesEnUso = AdministrarTransitions(*mbr, particionesMBR, arregloTransitions)
	//Si no hay particiones extendidas y se desea agregar una lógica
	if cantParticionesExtendidas == 0 && Comparar(tipo, "L") {
		respuesta += "---[ERROR-FDISK]: No puede crearse una partición lógica si no existe una partición extendida\n"
		return respuesta
	}
	//Si hay 1 partición extendida y se desea agregar otra extendida
	if cantParticionesExtendidas == 1 && Comparar(tipo, "E") {
		respuesta += "---[ERROR-FDISK]: Solamente puede tenerse una partición extendida en el disco\n"
		return respuesta
	}
	//Si hay 4 particiones creadas en el MBR y se desea agregar una primaria o extendida
	if cantParticionesEnUso == 4 && !Comparar(tipo, "L") {
		respuesta += "---[ERROR-FDISK]: Se ha alcanzado el número límite de particiones\n"
		return respuesta
	}
	//Crear la partición
	particionInsertar := Estructuras.NuevaPartition()
	particionInsertar.Part_type = strings.ToUpper(tipo)[0]
	particionInsertar.Part_fit = strings.ToUpper(fit)[0]
	particionInsertar.Part_s = int64(size)
	copy(particionInsertar.Part_name[:], name)
	//Crear la partición lógica
	if Comparar(tipo, "L") {
		respuesta += crearParticionLogica(particionInsertar, particionExtendida, driveletter)
		return respuesta
	}
	if arregloTransitions == nil && mbr.Mbr_tamano < int64(unsafe.Sizeof(Estructuras.MBR{}))+int64(size) {
		respuesta += "---[ERROR-FDISK]: El tamaño de la partición es mayor al tamaño del disco\n"
		return respuesta
	}
	//Crear partición primaria o extendida
	mbr, res = ajustarParticiones(*mbr, particionInsertar, arregloTransitions, cantParticionesEnUso)
	respuesta += res
	if mbr == nil {
		return respuesta
	}
	file, err := os.OpenFile(strings.ReplaceAll("./MIA/P1/"+driveletter+".dsk", "\"", ""), os.O_WRONLY, os.ModeAppend)
	defer file.Close()
	if err != nil {
		respuesta += "---[ERROR-FDISK]: Error al abrir el archivo\n"
		return respuesta
	}
	//Se actualiza el MBR con la información de la partición creada
	file.Seek(0, 0)
	var binario bytes.Buffer
	binary.Write(&binario, binary.BigEndian, mbr)
	EscribirBytes(file, binario.Bytes())
	if Comparar(tipo, "E") {
		//Si se creó una partición extendida, se le crea el primer EBR
		ebr := Estructuras.NuevoEBR()
		ebr.Part_start = int64(comienzoPartInsertar) + int64(unsafe.Sizeof(Estructuras.EBR{}))
		file.Seek(int64(comienzoPartInsertar), 0)
		var binario2 bytes.Buffer
		binary.Write(&binario2, binary.BigEndian, ebr)
		EscribirBytes(file, binario2.Bytes())
		respuesta += "+++[COMANDO-FDISK]: Se ha creado la partición extendida \"" + name + "\" con éxito\n"
		return respuesta
	}
	respuesta += "+++[COMANDO-FDISK]: Se ha creado la partición primaria \"" + name + "\" con éxito\n"
	return respuesta
}

//Función que se encarga de realizar los ajustes (Primer Ajuste, Mejor Ajuste y Peor Ajuste) de las particiones en el disco (Primarias y Extendidas)
func ajustarParticiones(mbr Estructuras.MBR, particion Estructuras.Partition, arregloTransitions []Transition, cantParticionesEnUso int) (*Estructuras.MBR, string) {
	respuesta := ""
	//Si no hay particiones en uso en el disco, se define la partición creada (primaria o extendida) como la primera del MBR
	if cantParticionesEnUso == 0 {
		particion.Part_start = int64(unsafe.Sizeof(mbr))
		comienzoPartInsertar = int(particion.Part_start)
		particion.Part_correlative = int64(0)
		mbr.Mbr_partition_1 = particion
		return &mbr, respuesta
	} else {
		var usar Transition
		/*
			Se recorre arregloTransitions (que tiene información de las particiones del disco).
			Básicamente, busca la partición guía para insertar la nueva partición antes o después de esta
		*/
		for i := 0; i < len(arregloTransitions); i++ {
			transition := arregloTransitions[i]
			if i == 0 {
				//"usar" tomará el valor del primer elemento en arregloTransitions
				usar = transition
				continue
			}
			if Comparar(string(mbr.Dsk_fit[:]), "F") {
				//Si hay espacio disponible antes o después de la partición del transition anterior para alojar la partición nueva
				if int64(usar.Before) >= particion.Part_s || int64(usar.After) >= particion.Part_s {
					break
				}
				//Si no hay espacio disponible, "usar" pasará a la siguiente partición del disco
				usar = transition
			} else if Comparar(string(mbr.Dsk_fit[:]), "B") {
				if int64(usar.Before) >= particion.Part_s || int64(usar.After) >= particion.Part_s {
					if int64(transition.Before) >= particion.Part_s || int64(transition.After) >= particion.Part_s {
						//espacioAntesUsar, espacioDespuesUsar, espacioAntesTransition, espacioDespuesTransition
						espacioAU := usar.Before - int(particion.Part_s)
						espacioDU := usar.After - int(particion.Part_s)
						espacioAT := transition.Before - int(particion.Part_s)
						espacioDT := transition.After - int(particion.Part_s)
						if espacioAU < 0 {
							//Si el espacioDespuesUsar es menor al espacio antes y después de Transition
							if espacioAT < 0 {
								if espacioDU < espacioDT {
									continue
								}
							} else if espacioDT < 0 {
								if espacioDU < espacioAT {
									continue
								}
							} else {
								if espacioDU < espacioAT && espacioDU < espacioDT {
									continue
								}
							}
						} else if espacioDU < 0 {
							//Si el espacioAntesUsar es menor al espacio antes y después de Transition
							if espacioAT < 0 {
								if espacioAU < espacioDT {
									continue
								}
							} else if espacioDT < 0 {
								if espacioAU < espacioAT {
									continue
								}
							} else {
								if espacioAU < espacioAT && espacioAU < espacioDT {
									continue
								}
							}
						} else {
							//Si el espacioAntesUsar y espacioDespuesUsar es menor al espacio antes y después de Transition
							if espacioAT < 0 {
								if espacioAU < espacioDT && espacioDU < espacioDT {
									continue
								}
							} else if espacioDT < 0 {
								if espacioAU < espacioAT && espacioDU < espacioAT {
									continue
								}
							} else {
								if (espacioAU < espacioAT && espacioAU < espacioDT) || (espacioDU < espacioAT && espacioDU < espacioDT) {
									continue
								}
							}
						}
						//Si en "usar" el espacio disponible es mayor al de Transition pasará a la siguiente partición del disco
						usar = transition
					} else {
						continue
					}
				} else {
					//Si no hay espacio disponible, "usar" pasará a la siguiente partición del disco
					usar = transition
				}
			} else if Comparar(string(mbr.Dsk_fit[:]), "W") {
				if int64(usar.Before) >= particion.Part_s || int64(usar.After) >= particion.Part_s {
					if int64(transition.Before) >= particion.Part_s || int64(transition.After) >= particion.Part_s {
						//espacioAntesUsar, espacioDespuesUsar, espacioAntesTransition, espacioDespuesTransition
						espacioAU := usar.Before - int(particion.Part_s)
						espacioDU := usar.After - int(particion.Part_s)
						espacioAT := transition.Before - int(particion.Part_s)
						espacioDT := transition.After - int(particion.Part_s)
						if espacioAU < 0 {
							//Si el espacioDespuesUsar es mayor al espacio antes y después de Transition
							if espacioAT < 0 {
								if espacioDU > espacioDT {
									continue
								}
							} else if espacioDT < 0 {
								if espacioDU > espacioAT {
									continue
								}
							} else {
								if espacioDU > espacioAT && espacioDU > espacioDT {
									continue
								}
							}
						} else if espacioDU < 0 {
							//Si el espacioAntesUsar es mayor al espacio antes y después de Transition
							if espacioAT < 0 {
								if espacioAU > espacioDT {
									continue
								}
							} else if espacioDT < 0 {
								if espacioAU > espacioAT {
									continue
								}
							} else {
								if espacioAU > espacioAT && espacioAU > espacioDT {
									continue
								}
							}
						} else {
							//Si el espacioAntesUsar y espacioDespuesUsar es mayor al espacio antes y después de Transition
							if espacioAT < 0 {
								if espacioAU > espacioDT && espacioDU > espacioDT {
									continue
								}
							} else if espacioDT < 0 {
								if espacioAU > espacioAT && espacioDU > espacioAT {
									continue
								}
							} else {
								if (espacioAU > espacioAT && espacioAU > espacioDT) || (espacioDU > espacioAT && espacioDU > espacioDT) {
									continue
								}
							}
						}
						//Si en "usar" el espacio disponible es menor al de Transition pasará a la siguiente partición del disco
						usar = transition
					} else {
						continue
					}
				} else {
					//Si no hay espacio disponible, "usar" pasará a la siguiente partición del disco
					usar = transition
				}
			}
		}
		particiones := ObtenerArregloParticiones(mbr)
		//Si hay espacio disponible antes o después de la partición para alojar la partición nueva, se acomoda según los ajustes
		if usar.Before >= int(particion.Part_s) || usar.After >= int(particion.Part_s) {
			if Comparar(string(mbr.Dsk_fit[:]), "F") {
				if usar.Before >= int(particion.Part_s) {
					//La partición nueva empezaría desde el byte donde comienza el espacio disponible antes de llegar a la partición "usar"
					particion.Part_start = int64(usar.Start - usar.Before)
				} else {
					//La partición nueva empezaría desde el byte donde termina la partición "usar"
					particion.Part_start = int64(usar.End)
				}
			} else if Comparar(string(mbr.Dsk_fit[:]), "B") {
				//La partición se almacenará en el espacio donde esta quepa y que tenga menor tamaño (que al insertarla sobre menos espacio)
				espacioAntes := usar.Before - int(particion.Part_s)
				espacioDespues := usar.After - int(particion.Part_s)
				if (usar.Before >= int(particion.Part_s) && espacioAntes < espacioDespues) || usar.After < int(particion.Part_s) {
					particion.Part_start = int64(usar.Start - usar.Before)
				} else {
					particion.Part_start = int64(usar.End)
				}
			} else if Comparar(string(mbr.Dsk_fit[:]), "W") {
				//La partición se almacenará en el espacio donde esta quepa y que tenga mayor tamaño (que al insertarla sobre más espacio)
				espacioAntes := usar.Before - int(particion.Part_s)
				espacioDespues := usar.After - int(particion.Part_s)
				if (usar.Before >= int(particion.Part_s) && espacioAntes > espacioDespues) || usar.After < int(particion.Part_s) {
					particion.Part_start = int64(usar.Start - usar.Before)
				} else {
					particion.Part_start = int64(usar.End)
				}
			}
			comienzoPartInsertar = int(particion.Part_start)
			//Se recorren las particiones del disco, y se almacena la nueva partición en la posición disponible
			for i := 0; i < len(particiones); i++ {
				if int(particiones[i].Part_s) == 0 {
					particion.Part_correlative = int64(i)
					particiones[i] = particion
					break
				}
			}
			mbr.Mbr_partition_1 = particiones[0]
			mbr.Mbr_partition_2 = particiones[1]
			mbr.Mbr_partition_3 = particiones[2]
			mbr.Mbr_partition_4 = particiones[3]
			return &mbr, respuesta
		} else {
			respuesta += "---[ERROR-FDISK]: No hay espacio suficiente\n"
			return nil, respuesta
		}
	}
}