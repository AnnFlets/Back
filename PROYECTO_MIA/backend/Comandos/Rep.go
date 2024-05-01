package Comandos

import (
	"PROYECTO_MIA/Estructuras"
	"bytes"
	"encoding/binary"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"unsafe"
)

//Función para comprobar los parámetros del comando REP
func ComprobarParametrosREP(parametros []string) string{
	respuesta := ""
	name := ""
	path := ""
	id := ""
	ruta := ""
	for i := 0; i < len(parametros); i++ {
		datosParametro := strings.Split(parametros[i], "=")
		if Comparar(datosParametro[0], "name") {
			if name == "" {
				name = datosParametro[1]
			} else {
				respuesta += "---[ERROR-REP]: Parámetro \"name\" repetido\n"
				return respuesta
			}
		} else if Comparar(datosParametro[0], "path") {
			if path == "" {
				path = datosParametro[1]
			} else {
				respuesta += "---[ERROR-REP]: Parámetro \"path\" repetido\n"
				return respuesta
			}
		} else if Comparar(datosParametro[0], "id") {
			if id == "" {
				id = datosParametro[1]
			} else {
				respuesta += "---[ERROR-REP]: Parámetro \"id\" repetido\n"
				return respuesta
			}
		} else if Comparar(datosParametro[0], "ruta") {
			if ruta == "" {
				ruta = datosParametro[1]
			} else {
				respuesta += "---[ERROR-REP]: Parámetro \"ruta\" repetido\n"
				return respuesta
			}
		} else {
			respuesta += "---[ERROR-REP]: No se esperaba el parámetro \"" + datosParametro[0] + "\"\n"
			return respuesta
		}
	}
	if name == "" {
		respuesta += "---[ERROR-REP]: Se requiere el parámetro \"name\" para este comando\n"
		return respuesta
	}
	if path == "" {
		respuesta += "---[ERROR-REP]: Se requiere el parámetro \"path\" para este comando\n"
		return respuesta
	}
	if id == "" {
		respuesta += "---[ERROR-REP]: Se requiere el parámetro \"id\" para este comando\n"
		return respuesta
	}
	if !VerificarArchivoExiste("./MIA/P1/" + string(id[0]) + ".dsk") {
		respuesta += "---[ERROR-REP]: No se encontró el disco solicitado\n"
		return respuesta
	}
	if !verificarExisteId(id) {
		respuesta += "---[ERROR-REP]: No se encontró el ID ingresado\n"
		return respuesta
	}
	name = strings.ReplaceAll(name, "\"", "")
	if Comparar(name, "mbr") {
		respuesta += generarReporteMBR(name, path, id)
	} else if Comparar(name, "disk") {
		respuesta += generarReporteDISK(name, path, id)
	} else if Comparar(name, "inode"){
		respuesta += generarReporteInodos(name, path, id)
	} else if Comparar(name, "block"){
		respuesta += generarReporteBloques(name, path, id)
	} else if Comparar(name, "bm_inode"){
		respuesta += generarReporteBMInodos(name, path, id)
	} else if Comparar(name, "bm_block"){
		respuesta += generarReporteBMBloques(name, path, id)	
	} else if Comparar(name, "sb"){
		respuesta += generarReporteSuperbloque(name, path, id)
	} else {
		respuesta += "---[ERROR-REP]: El parámetro \"name\" posee un valor no esperado\n"
	}
	return respuesta
}

//Función para verificar que exista el ID ingresado
func verificarExisteId(id string) bool {
	for i := 0; i < len(ParticionesMontadas); i++ {
		if id == string(ParticionesMontadas[i].Id[:]) {
			return true
		}
	}
	return false
}

//Función encargada de crear un archivo .dot
func crearArchivo(path string) string{
	respuesta := ""
	path = strings.ReplaceAll(path, "\"", "")
	if VerificarArchivoExiste(path) {
		err := os.Remove(path)
		if err != nil {
			respuesta += "---[ERROR-REP]: No pudo eliminarse el archivo\n"
			return respuesta
		}
	}
	//Busca el archivo, si no existe lo crea
	_, err1 := os.Stat(path)
	if os.IsNotExist(err1) {
		file, err2 := os.Create(path)
		defer file.Close()
		if err2 != nil {
			respuesta += "---[ERROR-REP]: No pudo crearse el archivo\n"
			return respuesta
		}
	}
	return respuesta
}

//Función para escribir el contenido recibido en el archivo .dot
func escribirArchivo(contenido string, path string) string{
	respuesta := ""
	//Abre el archivo creado con permisos de lectura y escritura
	file, err1 := os.OpenFile(strings.ReplaceAll(path, "\"", ""), os.O_RDWR, 0644)
	defer file.Close()
	if err1 != nil {
		respuesta += "---[ERROR-REP]: Error al abrir el archivo\n"
		return respuesta
	}
	//Escribe el contenido deseado en el archivo
	_, err1 = file.WriteString(contenido)
	if err1 != nil {
		respuesta += "---[ERROR-REP]: Error al escribir en el archivo\n"
		return respuesta
	}
	return respuesta
}

//Función para generar el .png, .jpg o .pdf de acuerdo al .dot
func ejecutar(path string, archivo string, extension string) string{
	respuesta := ""
	ruta, _ := exec.LookPath("dot")
	var cmd []byte
	if extension == "jpg" {
		cmd, _ = exec.Command(ruta, "-Tjpg", archivo).Output()
	} else if extension == "png" {
		cmd, _ = exec.Command(ruta, "-Tpng", archivo).Output()
	} else if extension == "pdf" {
		cmd, _ = exec.Command(ruta, "-Tpdf", archivo).Output()
	} else {
		respuesta += "---[ERROR-REP]: No se tiene contemplada la extensión \"" + extension + "\"\n"
		return respuesta
	}
	mode := 0777
	_ = os.WriteFile(path, cmd, os.FileMode(mode))
	return respuesta
}

//Función para generar el reporte del MBR y los EBR
func generarReporteMBR(name string, path string, id string) string{
	respuesta := ""
	datosPath := strings.Split(path, ".")
	respuesta += crearArchivo("./Reportes/" + datosPath[0] + ".dot")
	mbr, res := LeerDisco("./MIA/P1/" + string(id[0]) + ".dsk")
	respuesta += res
	contenido := "digraph G{\n" +
		"rankdir=\"LR\"\n" +
		"mbr[shape=none label=<\n" +
		"<TABLE border=\"1\" cellpadding=\"5\">\n" +
		"\t<TR><TD bgcolor=\"#013687\" COLSPAN=\"2\"><FONT COLOR=\"white\">REPORTE DE MBR</FONT></TD></TR>\n" +
		"\t<TR><TD bgcolor=\"white\">mbr_tamano</TD><TD bgcolor=\"white\">" + strconv.FormatInt(mbr.Mbr_tamano, 10) + "</TD></TR>\n" +
		"\t<TR><TD bgcolor=\"#D9E8FF\">mbr_fecha_creacion</TD><TD bgcolor=\"#D9E8FF\">" + string(mbr.Mbr_fecha_creacion[:]) + "</TD></TR>\n" +
		"\t<TR><TD bgcolor=\"white\">mbr_disk_signature</TD><TD bgcolor=\"white\">" + strconv.FormatInt(mbr.Mbr_disk_signature, 10) + "</TD></TR>\n" +
		"\t<TR><TD bgcolor=\"#D9E8FF\">mbr_fit</TD><TD bgcolor=\"#D9E8FF\">" + string(mbr.Dsk_fit[:]) + "</TD></TR>\n"
	particiones := ObtenerArregloParticiones(*mbr)
	var valor byte
	var fin int
	nombreParticion := ""
	for i := 0; i < len(particiones); i++ {
		if particiones[i].Part_s != 0 {
			fin = bytes.IndexByte(particiones[i].Part_name[:], 0)
			nombreParticion = string(particiones[i].Part_name[:fin])
			contenido += "\t<TR><TD bgcolor=\"#9B00BD\" COLSPAN=\"2\"><FONT COLOR=\"white\">PARTICIÓN</FONT></TD></TR>\n" +
				"\t<TR><TD bgcolor=\"white\">part_status</TD><TD bgcolor=\"white\">" + string(particiones[i].Part_status) + "</TD></TR>\n" +
				"\t<TR><TD bgcolor=\"#F7D7FF\">part_type</TD><TD bgcolor=\"#F7D7FF\">" + string(particiones[i].Part_type) + "</TD></TR>\n" +
				"\t<TR><TD bgcolor=\"white\">part_fit</TD><TD bgcolor=\"white\">" + string(particiones[i].Part_fit) + "</TD></TR>\n" +
				"\t<TR><TD bgcolor=\"#F7D7FF\">part_start</TD><TD bgcolor=\"#F7D7FF\">" + strconv.FormatInt(particiones[i].Part_start, 10) + "</TD></TR>\n" +
				"\t<TR><TD bgcolor=\"white\">part_s</TD><TD bgcolor=\"white\">" + strconv.FormatInt(particiones[i].Part_s, 10) + "</TD></TR>\n" +
				"\t<TR><TD bgcolor=\"#F7D7FF\">part_name</TD><TD bgcolor=\"#F7D7FF\">" + nombreParticion + "</TD></TR>\n" +
				"\t<TR><TD bgcolor=\"white\">part_correlative</TD><TD bgcolor=\"white\">" + strconv.FormatInt(particiones[i].Part_correlative, 10) + "</TD></TR>\n"
			if particiones[i].Part_id[0] == valor {
				contenido += "\t<TR><TD bgcolor=\"#F7D7FF\">part_id</TD><TD bgcolor=\"#F7D7FF\">-</TD></TR>\n"
			} else {
				contenido += "\t<TR><TD bgcolor=\"#F7D7FF\">part_id</TD><TD bgcolor=\"#F7D7FF\">" + string(particiones[i].Part_id[:]) + "</TD></TR>\n"
			}
			if Comparar(string(particiones[i].Part_type), "E") {
				arregloEBR, res := ObtenerParticionesLogicas(particiones[i], string(id[0]))
				respuesta += res
				for j := 0; j < len(arregloEBR); j++ {
					if arregloEBR[j].Part_s != 0 {
						fin = bytes.IndexByte(arregloEBR[j].Part_name[:], 0)
						nombreParticion = string(arregloEBR[j].Part_name[:fin])
						contenido += "\t<TR><TD bgcolor=\"#E49126\" COLSPAN=\"2\"><FONT COLOR=\"white\">PARTICIÓN LÓGICA</FONT></TD></TR>\n" +
							"\t<TR><TD bgcolor=\"white\">part_status</TD><TD bgcolor=\"white\">" + string(arregloEBR[j].Part_mount) + "</TD></TR>\n" +
							"\t<TR><TD bgcolor=\"#FFEACF\">part_type</TD><TD bgcolor=\"#FFEACF\">L</TD></TR>\n" +
							"\t<TR><TD bgcolor=\"white\">part_fit</TD><TD bgcolor=\"white\">" + string(arregloEBR[j].Part_fit) + "</TD></TR>\n" +
							"\t<TR><TD bgcolor=\"#FFEACF\">part_start</TD><TD bgcolor=\"#FFEACF\">" + strconv.FormatInt(arregloEBR[j].Part_start, 10) + "</TD></TR>\n" +
							"\t<TR><TD bgcolor=\"white\">part_s</TD><TD bgcolor=\"white\">" + strconv.FormatInt(arregloEBR[j].Part_s, 10) + "</TD></TR>\n" +
							"\t<TR><TD bgcolor=\"#FFEACF\">part_next</TD><TD bgcolor=\"#FFEACF\">" + strconv.FormatInt(arregloEBR[j].Part_next, 10) + "</TD></TR>\n" +
							"\t<TR><TD bgcolor=\"white\">part_name</TD><TD bgcolor=\"white\">" + nombreParticion + "</TD></TR>\n"
					}
				}
			}
		}
	}
	contenido += "</TABLE>>];\n}"
	respuesta += escribirArchivo(contenido, "./Reportes/"+datosPath[0]+".dot")
	respuesta += ejecutar("./Reportes/"+path, "./Reportes/"+datosPath[0]+".dot", datosPath[1])
	respuesta += "+++[COMANDO-REP]: Reporte \"" + name + "\" generado con éxito\n"
	return respuesta
}

//Función para generar el reporte del disco
func generarReporteDISK(name string, path string, id string) string{
	respuesta := ""
	datosPath := strings.Split(path, ".")
	respuesta += crearArchivo("./Reportes/" + datosPath[0] + ".dot")
	mbr, res := LeerDisco("./MIA/P1/" + string(id[0]) + ".dsk")
	respuesta += res
	var tamanio int
	var porcentaje int
	porcentajeTotal := 0
	contenido := "digraph G {\n" +
		"disk[shape=none label=<\n" +
		"<TABLE BORDER=\"1\" CELLBORDER=\"1\" CELLSPACING=\"3\" CELLPADDING=\"10\">\n" +
		"\t<TR>\n" +
		"\t\t<TD bgcolor=\"#FF6767\" ROWSPAN=\"2\"><FONT POINT-SIZE=\"12\">MBR</FONT></TD>\n"
	particiones := ObtenerArregloParticiones(*mbr)
	var arregloTransitions []Transition
	var particionExtendida Estructuras.Partition
	var arregloEBR []Estructuras.EBR
	var cantidadEspaciosEBR int
	porcentajeExtendida := 0
	arregloTransitions, particionExtendida, _, _ = AdministrarTransitions(*mbr, particiones, arregloTransitions)
	if len(arregloTransitions) == 0 {
		contenido += "\t\t<TD bgcolor=\"#A3FF67\" ROWSPAN=\"2\"><FONT POINT-SIZE=\"12\">Libre</FONT><BR/><FONT COLOR=\"red\" POINT-SIZE=\"8\"><B>100% del disco</B></FONT></TD>\n"
	}
	if int(particionExtendida.Part_s) != 0 {
		arregloEBR, res = ObtenerParticionesLogicas(particionExtendida, string(id[0]))
		respuesta += res
		cantidadEspaciosEBR = len(arregloEBR)
	}
	diskTamano := int(mbr.Mbr_tamano) - int(unsafe.Sizeof(Estructuras.MBR{})) - int(unsafe.Sizeof(Estructuras.EBR{}))*len(arregloEBR)
	for i := 0; i < len(arregloTransitions); i++ {
		transicion := arregloTransitions[i]
		if transicion.Before != 0 {
			porcentaje = int(float64(transicion.Before) / float64(diskTamano) * 100)
			porcentajeTotal += porcentaje
			contenido += "\t\t<TD bgcolor=\"#A3FF67\" ROWSPAN=\"2\"><FONT POINT-SIZE=\"12\">Libre</FONT><BR/><FONT COLOR=\"red\" POINT-SIZE=\"8\"><B>" + strconv.Itoa(porcentaje) + "% del disco</B></FONT></TD>\n"
		}
		tamanio = transicion.End - transicion.Start
		porcentaje = int(float64(tamanio) / float64(diskTamano) * 100)
		porcentajeTotal += porcentaje
		if int64(transicion.Start) == particionExtendida.Part_start {
			porcentajeExtendida = porcentaje
			for j := 0; j < len(arregloEBR); j++ {
				if arregloEBR[j].Part_next == -1 {
					cantidadEspaciosEBR += 1
				} else if arregloEBR[j].Part_s == 0 {
					cantidadEspaciosEBR += 1
				} else if (arregloEBR[j].Part_start + arregloEBR[j].Part_s) == arregloEBR[j].Part_next {
					cantidadEspaciosEBR += 1
				} else {
					cantidadEspaciosEBR += 2
				}
			}
			contenido += "\t\t<TD bgcolor=\"#E669FF\" COLSPAN=\"" + strconv.Itoa(cantidadEspaciosEBR) + "\"><FONT POINT-SIZE=\"12\">Extendida</FONT><FONT COLOR=\"red\" POINT-SIZE=\"8\"><B> " + strconv.Itoa(porcentaje) + "% del disco</B></FONT></TD>\n"
		} else {
			contenido += "\t\t<TD bgcolor=\"#27E5FF\" ROWSPAN=\"2\"><FONT POINT-SIZE=\"12\">Primaria</FONT><BR/><FONT COLOR=\"red\" POINT-SIZE=\"8\"><B>" + strconv.Itoa(porcentaje) + "% del disco</B></FONT></TD>\n"
		}
		if i == len(arregloTransitions)-1 {
			if transicion.After != 0 {
				porcentaje = int(float64(transicion.After) / float64(diskTamano) * 100)
				porcentajeTotal += porcentaje
				if porcentajeTotal != 100 {
					porcentaje += (100 - porcentajeTotal)
				}
				contenido += "\t\t<TD bgcolor=\"#A3FF67\" ROWSPAN=\"2\"><FONT POINT-SIZE=\"12\">Libre</FONT><BR/><FONT COLOR=\"red\" POINT-SIZE=\"8\"><B>" + strconv.Itoa(porcentaje) + "% del disco</B></FONT></TD>\n"
			}
		}
	}
	contenido += "\t</TR>\n"
	if len(arregloEBR) != 0 {
		contenido += "\t<TR>\n"
		var porcentajeLogicasExt int
		var porcentajeLogicasDisk int
		porcentajeLogicaExtTotal := 0
		porcentajeLogicasDiskTotal := 0
		for i := 0; i < len(arregloEBR); i++ {
			contenido += "\t\t<TD bgcolor=\"#FFBA67\"><FONT POINT-SIZE=\"12\">EBR</FONT></TD>\n"
			if arregloEBR[i].Part_next == -1 {
				porcentajeLogicasExt = int(float64(particionExtendida.Part_s-arregloEBR[i].Part_start) / float64(particionExtendida.Part_s) * 100)
				porcentajeLogicasDisk = int(float64(particionExtendida.Part_s-arregloEBR[i].Part_start) / float64(diskTamano) * 100)
				porcentajeLogicaExtTotal += porcentajeLogicasExt
				if porcentajeLogicaExtTotal != 100 {
					porcentajeLogicasExt += (100 - porcentajeLogicaExtTotal)
				}
				porcentajeLogicasDiskTotal += porcentajeLogicasDisk
				if porcentajeLogicasDiskTotal != porcentajeExtendida {
					porcentajeLogicasDisk += (porcentajeExtendida - porcentajeLogicasDiskTotal)
				}
				contenido += "\t\t<TD bgcolor=\"#A3FF67\"><FONT POINT-SIZE=\"12\">Libre</FONT><BR/><FONT COLOR=\"purple\" POINT-SIZE=\"8\"><B>" + strconv.Itoa(porcentajeLogicasExt) + "% de la partición</B></FONT><BR/><FONT COLOR=\"blue\" POINT-SIZE=\"8\"><B>" + strconv.Itoa(porcentajeLogicasDisk) + "% del disco</B></FONT></TD>\n"
			} else if arregloEBR[i].Part_s == 0 {
				porcentajeLogicasExt = int(float64(arregloEBR[i].Part_next-arregloEBR[i].Part_start) / float64(particionExtendida.Part_s) * 100)
				porcentajeLogicasDisk = int(float64(arregloEBR[i].Part_next-arregloEBR[i].Part_start) / float64(diskTamano) * 100)
				porcentajeLogicaExtTotal += porcentajeLogicasExt
				porcentajeLogicasDiskTotal += porcentajeLogicasDisk
				contenido += "\t\t<TD bgcolor=\"#A3FF67\"><FONT POINT-SIZE=\"12\">Libre</FONT><BR/><FONT COLOR=\"purple\" POINT-SIZE=\"8\"><B>" + strconv.Itoa(porcentajeLogicasExt) + "% de la partición</B></FONT><BR/><FONT COLOR=\"blue\" POINT-SIZE=\"8\"><B>" + strconv.Itoa(porcentajeLogicasDisk) + "% del disco</B></FONT></TD>\n"
			} else if (arregloEBR[i].Part_start + arregloEBR[i].Part_s) == arregloEBR[i].Part_next {
				porcentajeLogicasExt = int(float64(arregloEBR[i].Part_s) / float64(particionExtendida.Part_s) * 100)
				porcentajeLogicasDisk = int(float64(arregloEBR[i].Part_s) / float64(diskTamano) * 100)
				porcentajeLogicaExtTotal += porcentajeLogicasExt
				porcentajeLogicasDiskTotal += porcentajeLogicasDisk
				contenido += "\t\t<TD bgcolor=\"#FFF467\"><FONT POINT-SIZE=\"12\">Lógica</FONT><BR/><FONT COLOR=\"purple\" POINT-SIZE=\"8\"><B>" + strconv.Itoa(porcentajeLogicasExt) + "% de la partición</B></FONT><BR/><FONT COLOR=\"blue\" POINT-SIZE=\"8\"><B>" + strconv.Itoa(porcentajeLogicasDisk) + "% del disco</B></FONT></TD>\n"
			} else {
				porcentajeLogicasExt = int(float64(arregloEBR[i].Part_s) / float64(particionExtendida.Part_s) * 100)
				porcentajeLogicasDisk = int(float64(arregloEBR[i].Part_s) / float64(diskTamano) * 100)
				porcentajeLogicaExtTotal += porcentajeLogicasExt
				porcentajeLogicasDiskTotal += porcentajeLogicasDisk
				contenido += "\t\t<TD bgcolor=\"#FFF467\"><FONT POINT-SIZE=\"12\">Lógica</FONT><BR/><FONT COLOR=\"purple\" POINT-SIZE=\"8\"><B>" + strconv.Itoa(porcentajeLogicasExt) + "% de la partición</B></FONT><BR/><FONT COLOR=\"blue\" POINT-SIZE=\"8\"><B>" + strconv.Itoa(porcentajeLogicasDisk) + "% del disco</B></FONT></TD>\n"
				porcentajeLogicasExt = int(float64(arregloEBR[i].Part_next-(arregloEBR[i].Part_start+arregloEBR[i].Part_s)) / float64(particionExtendida.Part_s) * 100)
				porcentajeLogicasDisk = int(float64(arregloEBR[i].Part_next-(arregloEBR[i].Part_start+arregloEBR[i].Part_s)) / float64(diskTamano) * 100)
				porcentajeLogicaExtTotal += porcentajeLogicasExt
				porcentajeLogicasDiskTotal += porcentajeLogicasDisk
				contenido += "\t\t<TD bgcolor=\"#A3FF67\"><FONT POINT-SIZE=\"12\">Libre</FONT><BR/><FONT COLOR=\"purple\" POINT-SIZE=\"8\"><B>" + strconv.Itoa(porcentajeLogicasExt) + "% de la partición</B></FONT><BR/><FONT COLOR=\"blue\" POINT-SIZE=\"8\"><B>" + strconv.Itoa(porcentajeLogicasDisk) + "% del disco</B></FONT></TD>\n"
			}
		}
		contenido += "\t</TR>\n"
	}
	contenido += "</TABLE>>];\n}"
	respuesta += escribirArchivo(contenido, "./Reportes/"+datosPath[0]+".dot")
	respuesta += ejecutar("./Reportes/"+path, "./Reportes/"+datosPath[0]+".dot", datosPath[1])
	respuesta += "+++[COMANDO-REP]: Reporte \"" + name + "\" generado con éxito\n"
	return respuesta
}

// Función para generar el reporte del superbloque
func generarReporteSuperbloque(name string, path string, id string) string{
	respuesta := ""
	datosPath := strings.Split(path, ".")
	crearArchivo("./Reportes/" + datosPath[0] + ".dot")
	mbr, res := LeerDisco("./MIA/P1/" + string(id[0]) + ".dsk")
	respuesta += res
	particion := Particiones{}
	for i := 0; i < len(ParticionesMontadas); i++ {
		if id == string(ParticionesMontadas[i].Id[:]) {
			particion = ParticionesMontadas[i]
			break
		}
	}
	fin := bytes.IndexByte(particion.Nombre[:], 0)
	nombreParticion := string(particion.Nombre[:fin])
	particionBuscada, res := BuscarParticion(*mbr, string(id[0]), nombreParticion)
	respuesta += res
	file, _ := os.Open("./MIA/P1/" + string(id[0]) + ".dsk")
	defer file.Close()
	superbloque := Estructuras.NuevoSuperbloque()
	file.Seek(particionBuscada.Part_start, 0)
	informacionSuperbloque := LeerBytes(file, int(unsafe.Sizeof(Estructuras.Superbloque{})))
	buffer := bytes.NewBuffer(informacionSuperbloque)
	binary.Read(buffer, binary.BigEndian, &superbloque)
	if superbloque.S_filesystem_type == 0 {
		respuesta += "---[ERROR-REP]: La partición no cuenta con un sistema de archivos\n"
		return respuesta
	}
	contenido := "digraph G{\n" +
		"rankdir=\"LR\"\n" +
		"mbr[shape=none label=<\n" +
		"<TABLE border=\"1\" cellpadding=\"5\">\n" +
		"\t<TR><TD bgcolor=\"#013687\" COLSPAN=\"2\"><FONT COLOR=\"white\">REPORTE DE SUPERBLOQUE</FONT></TD></TR>\n" +
		"\t<TR><TD bgcolor=\"white\">s_filesystem_type</TD><TD bgcolor=\"white\">" + strconv.FormatInt(superbloque.S_filesystem_type, 10) + "</TD></TR>\n" +
		"\t<TR><TD bgcolor=\"#D9E8FF\">s_inodes_count</TD><TD bgcolor=\"#D9E8FF\">" + strconv.FormatInt(superbloque.S_inodes_count, 10) + "</TD></TR>\n" +
		"\t<TR><TD bgcolor=\"white\">s_blocks_count</TD><TD bgcolor=\"white\">" + strconv.FormatInt(superbloque.S_blocks_count, 10) + "</TD></TR>\n" +
		"\t<TR><TD bgcolor=\"#D9E8FF\">s_free_blocks_count</TD><TD bgcolor=\"#D9E8FF\">" + strconv.FormatInt(superbloque.S_free_blocks_count, 10) + "</TD></TR>\n" +
		"\t<TR><TD bgcolor=\"white\">s_free_inodes_count</TD><TD bgcolor=\"white\">" + strconv.FormatInt(superbloque.S_free_inodes_count, 10) + "</TD></TR>\n" +
		"\t<TR><TD bgcolor=\"#D9E8FF\">s_mtime</TD><TD bgcolor=\"#D9E8FF\">" + string(superbloque.S_mtime[:]) + "</TD></TR>\n" +
		"\t<TR><TD bgcolor=\"white\">s_umtime</TD><TD bgcolor=\"white\">" + string(superbloque.S_umtime[:]) + "</TD></TR>\n" +
		"\t<TR><TD bgcolor=\"#D9E8FF\">s_mnt_count</TD><TD bgcolor=\"#D9E8FF\">" + strconv.FormatInt(superbloque.S_mnt_count, 10) + "</TD></TR>\n" +
		"\t<TR><TD bgcolor=\"white\">s_magic</TD><TD bgcolor=\"white\">" + strconv.FormatInt(superbloque.S_magic, 10) + "</TD></TR>\n" +
		"\t<TR><TD bgcolor=\"#D9E8FF\">s_inode_s</TD><TD bgcolor=\"#D9E8FF\">" + strconv.FormatInt(superbloque.S_inode_s, 10) + "</TD></TR>\n" +
		"\t<TR><TD bgcolor=\"white\">s_block_s</TD><TD bgcolor=\"white\">" + strconv.FormatInt(superbloque.S_block_s, 10) + "</TD></TR>\n" +
		"\t<TR><TD bgcolor=\"#D9E8FF\">s_first_ino</TD><TD bgcolor=\"#D9E8FF\">" + strconv.FormatInt(superbloque.S_first_ino, 10) + "</TD></TR>\n" +
		"\t<TR><TD bgcolor=\"white\">s_first_blo</TD><TD bgcolor=\"white\">" + strconv.FormatInt(superbloque.S_first_blo, 10) + "</TD></TR>\n" +
		"\t<TR><TD bgcolor=\"#D9E8FF\">s_bm_inode_start</TD><TD bgcolor=\"#D9E8FF\">" + strconv.FormatInt(superbloque.S_bm_inode_start, 10) + "</TD></TR>\n" +
		"\t<TR><TD bgcolor=\"white\">s_bm_block_start</TD><TD bgcolor=\"white\">" + strconv.FormatInt(superbloque.S_bm_block_start, 10) + "</TD></TR>\n" +
		"\t<TR><TD bgcolor=\"#D9E8FF\">s_inode_start</TD><TD bgcolor=\"#D9E8FF\">" + strconv.FormatInt(superbloque.S_inode_start, 10) + "</TD></TR>\n" +
		"\t<TR><TD bgcolor=\"white\">s_block_start</TD><TD bgcolor=\"white\">" + strconv.FormatInt(superbloque.S_block_start, 10) + "</TD></TR>\n"
	contenido += "</TABLE>>];\n}"
	escribirArchivo(contenido, "./Reportes/"+datosPath[0]+".dot")
	ejecutar("./Reportes/"+path, "./Reportes/"+datosPath[0]+".dot", datosPath[1])
	respuesta += "+++[COMANDO-REP]: Reporte \"" + name + "\" generado con éxito\n"
	return respuesta
}

// Función para generar el reporte de los inodos
func generarReporteInodos(name string, path string, id string) string{
	respuesta := ""
	datosPath := strings.Split(path, ".")
	crearArchivo("./Reportes/" + datosPath[0] + ".dot")
	mbr, res := LeerDisco("./MIA/P1/" + string(id[0]) + ".dsk")
	respuesta += res
	particion := Particiones{}
	for i := 0; i < len(ParticionesMontadas); i++ {
		if id == string(ParticionesMontadas[i].Id[:]) {
			particion = ParticionesMontadas[i]
			break
		}
	}
	fin := bytes.IndexByte(particion.Nombre[:], 0)
	nombreParticion := string(particion.Nombre[:fin])
	particionBuscada, res := BuscarParticion(*mbr, string(id[0]), nombreParticion)
	respuesta += res
	file, _ := os.Open("./MIA/P1/" + string(id[0]) + ".dsk")
	defer file.Close()
	superbloque := Estructuras.NuevoSuperbloque()
	file.Seek(particionBuscada.Part_start, 0)
	informacionSuperbloque := LeerBytes(file, int(unsafe.Sizeof(Estructuras.Superbloque{})))
	buffer := bytes.NewBuffer(informacionSuperbloque)
	err1 := binary.Read(buffer, binary.BigEndian, &superbloque)
	if err1 != nil {
		respuesta += "---[ERROR-REP]: No pudo leerse el archivo\n"
		return respuesta
	}
	if superbloque.S_filesystem_type == 0 {
		respuesta += "---[ERROR-REP]: La partición no cuenta con un sistema de archivos\n"
		return respuesta
	}
	contenido := "digraph G{\n" +
		"rankdir=\"TB\"\n"
	for i := 0; i < int(superbloque.S_inodes_count-superbloque.S_free_inodes_count); i++ {
		inodo := Estructuras.NuevoInodo()
		file.Seek(superbloque.S_inode_start+int64(int(unsafe.Sizeof(Estructuras.Inodo{}))*i), 0)
		informacionInodo := LeerBytes(file, int(unsafe.Sizeof(Estructuras.Inodo{})))
		buffer = bytes.NewBuffer(informacionInodo)
		err2 := binary.Read(buffer, binary.BigEndian, &inodo)
		if err2 != nil {
			respuesta += "---[ERROR-REP]: No pudo leerse el archivo\n"
			return respuesta
		}
		contenido += "inodo" + strconv.Itoa(i) + "[shape=none label=<\n" +
			"<TABLE border=\"1\" cellpadding=\"5\">\n" +
			"\t<TR><TD bgcolor=\"#013687\" COLSPAN=\"2\"><FONT COLOR=\"white\">INODO " + strconv.Itoa(i) + "</FONT></TD></TR>\n" +
			"\t<TR><TD bgcolor=\"white\">I_uid</TD><TD bgcolor=\"white\">" + strconv.FormatInt(inodo.I_uid, 10) + "</TD></TR>\n" +
			"\t<TR><TD bgcolor=\"#D9E8FF\">I_gid</TD><TD bgcolor=\"#D9E8FF\">" + strconv.FormatInt(inodo.I_gid, 10) + "</TD></TR>\n" +
			"\t<TR><TD bgcolor=\"white\">I_s</TD><TD bgcolor=\"white\">" + strconv.FormatInt(inodo.I_s, 10) + "</TD></TR>\n" +
			"\t<TR><TD bgcolor=\"#D9E8FF\">I_atime</TD><TD bgcolor=\"#D9E8FF\">" + string(inodo.I_atime[:]) + "</TD></TR>\n" +
			"\t<TR><TD bgcolor=\"white\">I_ctime</TD><TD bgcolor=\"white\">" + string(inodo.I_ctime[:]) + "</TD></TR>\n" +
			"\t<TR><TD bgcolor=\"#D9E8FF\">I_mtime</TD><TD bgcolor=\"#D9E8FF\">" + string(inodo.I_mtime[:]) + "</TD></TR>\n"
		for j := 0; j < len(inodo.I_block); j++ {
			contenido += "\t<TR><TD bgcolor=\"white\">I_block_" + strconv.Itoa(j) + "</TD><TD bgcolor=\"white\">" + strconv.FormatInt(inodo.I_block[j], 10) + "</TD></TR>\n"
		}
		contenido += "\t<TR><TD bgcolor=\"#D9E8FF\">I_type</TD><TD bgcolor=\"#D9E8FF\">" + strconv.FormatInt(inodo.I_type, 10) + "</TD></TR>\n" +
			"\t<TR><TD bgcolor=\"white\">I_perm</TD><TD bgcolor=\"white\">" + strconv.FormatInt(inodo.I_perm, 10) + "</TD></TR>\n" +
			"</TABLE>>];\n"
	}
	contenido += "}"
	escribirArchivo(contenido, "./Reportes/"+datosPath[0]+".dot")
	ejecutar("./Reportes/"+path, "./Reportes/"+datosPath[0]+".dot", datosPath[1])
	respuesta += "+++[COMANDO-REP]: Reporte \"" + name + "\" generado con éxito\n"
	return respuesta
}

// Función para generar el reporte de los bloques
func generarReporteBloques(name string, path string, id string) string{
	respuesta := ""
	datosPath := strings.Split(path, ".")
	crearArchivo("./Reportes/" + datosPath[0] + ".dot")
	mbr, res := LeerDisco("./MIA/P1/" + string(id[0]) + ".dsk")
	respuesta += res
	particion := Particiones{}
	for i := 0; i < len(ParticionesMontadas); i++ {
		if id == string(ParticionesMontadas[i].Id[:]) {
			particion = ParticionesMontadas[i]
			break
		}
	}
	fin := bytes.IndexByte(particion.Nombre[:], 0)
	nombreParticion := string(particion.Nombre[:fin])
	particionBuscada, res := BuscarParticion(*mbr, string(id[0]), nombreParticion)
	respuesta += res
	file, _ := os.Open("./MIA/P1/" + string(id[0]) + ".dsk")
	defer file.Close()
	superbloque := Estructuras.NuevoSuperbloque()
	file.Seek(particionBuscada.Part_start, 0)
	informacionSuperbloque := LeerBytes(file, int(unsafe.Sizeof(Estructuras.Superbloque{})))
	buffer := bytes.NewBuffer(informacionSuperbloque)
	err1 := binary.Read(buffer, binary.BigEndian, &superbloque)
	if err1 != nil {
		respuesta += "---[ERROR-REP]: No pudo leerse el archivo\n"
		return respuesta
	}
	if superbloque.S_filesystem_type == 0 {
		respuesta += "---[ERROR-REP]: La partición no cuenta con un sistema de archivos\n"
		return respuesta
	}
	contenido := "digraph G{\n" +
		"rankdir=\"TB\"\n"
	for i := 0; i < int(superbloque.S_inodes_count-superbloque.S_free_inodes_count); i++ {
		inodo := Estructuras.NuevoInodo()
		file.Seek(superbloque.S_inode_start+int64(int(unsafe.Sizeof(Estructuras.Inodo{}))*i), 0)
		informacionInodo := LeerBytes(file, int(unsafe.Sizeof(Estructuras.Inodo{})))
		buffer = bytes.NewBuffer(informacionInodo)
		err2 := binary.Read(buffer, binary.BigEndian, &inodo)
		if err2 != nil {
			respuesta += "---[ERROR-REP]: No pudo leerse el archivo\n"
			return respuesta
		}
		bloqueArchivo := Estructuras.BloqueArchivo{}
		bloqueCarpeta := Estructuras.BloqueCarpeta{}
		var contenidoBloque string
		for j := 0; j < len(inodo.I_block); j++ {
			if inodo.I_block[0] == -1 {
				break
			}
			if inodo.I_type == 1 {
				file.Seek(superbloque.S_block_start+int64(unsafe.Sizeof(Estructuras.BloqueCarpeta{})), 0)
				informacion := LeerBytes(file, int(unsafe.Sizeof(Estructuras.BloqueArchivo{})))
				buffer = bytes.NewBuffer(informacion)
				binary.Read(buffer, binary.BigEndian, &bloqueArchivo)
				fin = bytes.IndexByte(bloqueArchivo.B_content[:], 0)
				contenidoBloque = string(bloqueArchivo.B_content[:fin])
				contenido += "bloqueArchivo" + strconv.Itoa(j) + "[shape=none label=<\n" +
					"<TABLE border=\"1\" cellpadding=\"5\">\n" +
					"\t<TR><TD bgcolor=\"#013687\" COLSPAN=\"2\"><FONT COLOR=\"white\">BLOQUE ARCHIVO " + strconv.Itoa(j) + "</FONT></TD></TR>\n" +
					"\t<TR><TD bgcolor=\"white\">B_content</TD><TD bgcolor=\"white\">" + contenidoBloque + "</TD></TR>\n" +
					"</TABLE>>];\n"
				break
			} else {
				file.Seek(superbloque.S_block_start, 0)
				informacion := LeerBytes(file, int(unsafe.Sizeof(Estructuras.BloqueCarpeta{})))
				buffer = bytes.NewBuffer(informacion)
				binary.Read(buffer, binary.BigEndian, &bloqueCarpeta)
				contenido += "bloqueCarpeta" + strconv.Itoa(j) + "[shape=none label=<\n" +
					"<TABLE border=\"1\" cellpadding=\"5\">\n" +
					"\t<TR><TD bgcolor=\"#013687\" COLSPAN=\"2\"><FONT COLOR=\"white\">BLOQUE CARPETA " + strconv.Itoa(j) + "</FONT></TD></TR>\n"
				for k := 0; k < 4; k++ {
					if string(bloqueCarpeta.B_content[k].B_name[:]) == "" {
						contenido += "\t<TR><TD bgcolor=\"white\"> - </TD><TD bgcolor=\"white\">" + strconv.Itoa(int(bloqueCarpeta.B_content[k].B_inodo)) + "</TD></TR>\n"
					} else {
						fin = bytes.IndexByte(bloqueCarpeta.B_content[k].B_name[:], 0)
						contenidoBloque = string(bloqueCarpeta.B_content[k].B_name[:fin])
						contenido += "\t<TR><TD bgcolor=\"white\">" + contenidoBloque + "</TD><TD bgcolor=\"white\">" + strconv.Itoa(int(bloqueCarpeta.B_content[k].B_inodo)) + "</TD></TR>\n"
					}
				}
				contenido += "</TABLE>>];\n"
				break
			}
		}
	}
	contenido += "}"
	escribirArchivo(contenido, "./Reportes/"+datosPath[0]+".dot")
	ejecutar("./Reportes/"+path, "./Reportes/"+datosPath[0]+".dot", datosPath[1])
	respuesta += "+++[COMANDO-REP]: Reporte \""+name+"\" generado con éxito\n"
	return respuesta
}

// Función que genera el reporte del bitmap de inodos en .txt
func generarReporteBMInodos(name string, path string, id string) string{
	respuesta := ""
	datosPath := strings.Split(path, ".")
	if VerificarArchivoExiste("./Reportes/" + datosPath[0] + ".txt") {
		err := os.Remove("./Reportes/" + datosPath[0] + ".txt")
		if err != nil {
			respuesta += "---[ERROR-REP]: No pudo eliminarse el archivo\n"
			return respuesta
		}
	}
	archivo, err := os.Create("./Reportes/" + datosPath[0] + ".txt")
	defer archivo.Close()
	if err != nil {
		respuesta += "---[ERROR-REP]: No pudo crearse el archivo\n"
		return respuesta
	}
	mbr, res := LeerDisco("./MIA/P1/" + string(id[0]) + ".dsk")
	respuesta += res
	particion := Particiones{}
	for i := 0; i < len(ParticionesMontadas); i++ {
		if id == string(ParticionesMontadas[i].Id[:]) {
			particion = ParticionesMontadas[i]
			break
		}
	}
	fin := bytes.IndexByte(particion.Nombre[:], 0)
	nombreParticion := string(particion.Nombre[:fin])
	particionBuscada, res := BuscarParticion(*mbr, string(id[0]), nombreParticion)
	respuesta += res
	file, _ := os.Open("./MIA/P1/" + string(id[0]) + ".dsk")
	defer file.Close()
	superbloque := Estructuras.NuevoSuperbloque()
	file.Seek(particionBuscada.Part_start, 0)
	informacionSuperbloque := LeerBytes(file, int(unsafe.Sizeof(Estructuras.Superbloque{})))
	buffer := bytes.NewBuffer(informacionSuperbloque)
	err1 := binary.Read(buffer, binary.BigEndian, &superbloque)
	if err1 != nil {
		respuesta += "---[ERROR-REP]: No pudo leerse el archivo\n"
		return respuesta
	}
	if superbloque.S_filesystem_type == 0 {
		respuesta += "---[ERROR-REP]: La partición no cuenta con un sistema de archivos\n"
		return respuesta
	}
	cuenta := 1
	texto := ""
	file, err1 = os.OpenFile(strings.ReplaceAll("./Reportes/"+datosPath[0]+".txt", "\"", ""), os.O_RDWR, 0644)
	defer file.Close()
	if err1 != nil {
		respuesta += "---[ERROR-REP]: Error al abrir el archivo\n"
		return respuesta
	}
	for i := 0; i < int(superbloque.S_inodes_count-superbloque.S_free_inodes_count); i++ {
		if cuenta == 21 {
			texto += "1."
			cuenta = 1
		} else {
			texto += "1"
		}
		cuenta = cuenta + 1
	}
	cuenta = cuenta + 1
	for i := 0; i < int(superbloque.S_free_inodes_count); i++ {
		if cuenta == 21 {
			texto += "0."
			cuenta = 1
		} else {
			texto += "0"
		}
		cuenta = cuenta + 1
	}
	_, err1 = file.WriteString(texto)
	if err1 != nil {
		respuesta += "---[ERROR-REP]: Error al escribir el archivo\n"
		return respuesta
	}
	respuesta += "+++[COMANDO-REP]: Reporte \""+name+"\" generado con éxito\n"
	return respuesta
}

// Función que genera el reporte del bitmap de bloques en .txt
func generarReporteBMBloques(name string, path string, id string) string{
	respuesta := ""
	datosPath := strings.Split(path, ".")
	if VerificarArchivoExiste("./Reportes/" + datosPath[0] + ".txt") {
		err := os.Remove("./Reportes/" + datosPath[0] + ".txt")
		if err != nil {
			respuesta += "---[ERROR-REP]: No pudo eliminarse el archivo\n"
			return respuesta
		}
	}
	archivo, err := os.Create("./Reportes/" + datosPath[0] + ".txt")
	defer archivo.Close()
	if err != nil {
		respuesta += "---[ERROR-REP]: No pudo crearse el archivo\n"
		return respuesta
	}
	mbr, res := LeerDisco("./MIA/P1/" + string(id[0]) + ".dsk")
	respuesta += res
	particion := Particiones{}
	for i := 0; i < len(ParticionesMontadas); i++ {
		if id == string(ParticionesMontadas[i].Id[:]) {
			particion = ParticionesMontadas[i]
			break
		}
	}
	fin := bytes.IndexByte(particion.Nombre[:], 0)
	nombreParticion := string(particion.Nombre[:fin])
	particionBuscada, res := BuscarParticion(*mbr, string(id[0]), nombreParticion)
	respuesta += res
	file, _ := os.Open("./MIA/P1/" + string(id[0]) + ".dsk")
	defer file.Close()
	superbloque := Estructuras.NuevoSuperbloque()
	file.Seek(particionBuscada.Part_start, 0)
	informacionSuperbloque := LeerBytes(file, int(unsafe.Sizeof(Estructuras.Superbloque{})))
	buffer := bytes.NewBuffer(informacionSuperbloque)
	err1 := binary.Read(buffer, binary.BigEndian, &superbloque)
	if err1 != nil {
		respuesta += "---[ERROR-REP]: No pudo leerse el archivo\n"
		return respuesta
	}
	if superbloque.S_filesystem_type == 0 {
		respuesta += "---[ERROR-REP]: La partición no cuenta con un sistema de archivos\n"
		return respuesta
	}
	cuenta := 1
	texto := ""
	file, err1 = os.OpenFile(strings.ReplaceAll("./Reportes/"+datosPath[0]+".txt", "\"", ""), os.O_RDWR, 0644)
	defer file.Close()
	if err1 != nil {
		respuesta += "---[ERROR-REP]: Error al abrir el archivo\n"
		return respuesta
	}
	for i := 0; i < int(superbloque.S_blocks_count-superbloque.S_free_blocks_count); i++ {
		if cuenta == 21 {
			texto += "1."
			cuenta = 1
		} else {
			texto += "1"
		}
		cuenta = cuenta + 1
	}
	cuenta = cuenta + 1
	for i := 0; i < int(superbloque.S_free_blocks_count); i++ {
		if cuenta == 21 {
			texto += "0."
			cuenta = 1
		} else {
			texto += "0"
		}
		cuenta = cuenta + 1
	}
	_, err1 = file.WriteString(texto)
	if err1 != nil {
		respuesta += "---[ERROR-REP]: Error al escribir en el archivo\n"
		return respuesta
	}
	respuesta += "+++[COMANDO-REP]: Reporte \""+name+"\" generado con éxito\n"
	return respuesta
}