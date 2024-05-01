package Comandos

import (
	"PROYECTO_MIA/Estructuras"
	"bytes"
	"encoding/binary"
	"math"
	"os"
	"strings"
	"time"
	"unsafe"
)

// Función para comprobar los parámetros del comando MKFS
func ComprobarParametrosMKFS(parametros []string) string{
	respuesta := ""
	id := ""
	tipo := ""
	fs := ""
	for i := 0; i < len(parametros); i++ {
		datosParametro := strings.Split(parametros[i], "=")
		if Comparar(datosParametro[0], "id") {
			if id == "" {
				id = datosParametro[1]
			} else {
				respuesta += "---[ERROR-MKFS]: Parámetro \"id\" repetido\n"
				return respuesta
			}
		} else if Comparar(datosParametro[0], "type") {
			if tipo == "" {
				tipo = datosParametro[1]
			} else {
				respuesta += "---[ERROR-MKFS]: Parámetro \"type\" repetido\n"
				return respuesta
			}
		} else if Comparar(datosParametro[0], "fs") {
			if fs == "" {
				fs = datosParametro[1]
			} else {
				respuesta += "---[ERROR-MKFS]: Parámetro \"fs\" repetido\n"
				return respuesta
			}
		} else {
			respuesta += "---[ERROR-MKFS]: No se esperaba el parámetro \"" + datosParametro[0] + "\"\n"
			return respuesta
		}
	}
	if id == "" {
		respuesta += "---[ERROR-MKFS]: Se requiere el parámetro \"id\"\n"
		return respuesta
	}
	if tipo != "" {
		if !Comparar(tipo, "full") {
			respuesta += "---[ERROR-MKFS]: Valor inválido. El parámetro \"type\" espera el valor \"full\"\n"
			return respuesta
		}
	}
	if fs != "" {
		if Comparar(fs, "2fs") {
			if Comparar(fs, "3fs") {
				respuesta += "---[ERROR-MKFS]: Valor inválido. El parámetro \"fs\" espera el valor \"2fs\" o \"3fs\"\n"
				return respuesta
			}
		}
	}
	if !VerificarArchivoExiste("./MIA/P1/" + string(id[0]) + ".dsk") {
		respuesta += "---[ERROR-MKFS]: El disco no existe\n"
		return respuesta
	}
	if !VerificarParticionEstaMontada(id) {
		respuesta += "---[ERROR-MKFS]: No existe una partición montada con el id \"" + id + "\"\n"
		return respuesta
	}
	if tipo == "" {
		tipo = "full"
	}
	if fs == "" {
		fs = "2fs"
	}
	var particion Particiones
	for i := 0; i < len(ParticionesMontadas); i++ {
		if string(ParticionesMontadas[i].Id[:]) == id {
			particion = ParticionesMontadas[i]
			break
		}
	}
	fin := bytes.IndexByte(particion.Nombre[:], 0)
	nombreParticion := string(particion.Nombre[:fin])
	mbr, res := LeerDisco("./MIA/P1/" + string(id[0]) + ".dsk")
	respuesta += res
	particionBuscada, res := BuscarParticion(*mbr, string(id[0]), nombreParticion)
	respuesta += res
	if particionBuscada == nil {
		respuesta += "---[ERROR-MKFS]: No existe una partición con el nombre \"" + nombreParticion + "\"\n"
		return respuesta
	}
	var n int
	if Comparar(fs, "2fs") {
		n = int(math.Floor(float64(particionBuscada.Part_s-int64(unsafe.Sizeof(Estructuras.Superbloque{}))) / float64(4+unsafe.Sizeof(Estructuras.Inodo{})+3*unsafe.Sizeof(Estructuras.BloqueArchivo{}))))
		if n <= 0 {
			respuesta += "---[ERROR-MKFS]: No fue posible crear el sistema de archivos \"EXT2\" en la partición\n"
			return respuesta
		}
		respuesta += FormatearParticionEXT2(id, n, particionBuscada)
	} else {
		n = int(math.Floor(float64(particionBuscada.Part_s-int64(unsafe.Sizeof(Estructuras.Superbloque{}))) / float64(4+unsafe.Sizeof(Estructuras.Journaling{})+unsafe.Sizeof(Estructuras.Inodo{})+3*unsafe.Sizeof(Estructuras.BloqueArchivo{}))))
		if n <= 0 {
			respuesta += "---[ERROR-MKFS]: No fue posible crear el sistema de archivos \"EXT3\" en la partición\n"
			return respuesta
		}
		respuesta += FormatearParticionEXT3(id, n, particionBuscada)
	}
	return respuesta
}

// Función que se encarga de formatear la partición y generar un sistema de archivos EXT2 en la misma
func FormatearParticionEXT2(id string, n int, particion *Estructuras.Partition) string{
	respuesta := ""
	superBloque := Estructuras.NuevoSuperbloque()
	superBloque.S_filesystem_type = 2
	superBloque.S_bm_inode_start = particion.Part_start + int64(unsafe.Sizeof(Estructuras.Superbloque{}))
	respuesta += EscribirEstructurasEXT(id, superBloque, n, particion)
	respuesta += "+++[COMANDO-MKFS]: Se ha formateado correctamente la partición \"" + string(particion.Part_name[:]) + "\"\n"
	return respuesta
}

// Función que se encarga de formatear la partición y generar un sistema de archivos EXT3 en la misma
func FormatearParticionEXT3(id string, n int, particion *Estructuras.Partition) string{
	respuesta := ""
	superBloque := Estructuras.NuevoSuperbloque()
	superBloque.S_filesystem_type = 3
	superBloque.S_bm_inode_start = particion.Part_start + int64(unsafe.Sizeof(Estructuras.Superbloque{})) + int64(unsafe.Sizeof(Estructuras.Journaling{}))
	respuesta += EscribirEstructurasEXT(id, superBloque, n, particion)
	file, err := os.OpenFile("./MIA/P1/"+string(id[0])+".dsk", os.O_WRONLY, os.ModeAppend)
	defer file.Close()
	if err != nil {
		respuesta += "---[ERROR-MKFS]: No se pudo encontrar el disco\n"
		return respuesta
	}
	//Escribir el Journaling en la partición
	journaling := Estructuras.NuevoJournaling()
	file.Seek(particion.Part_start+int64(unsafe.Sizeof(Estructuras.Superbloque{})), 0)
	var binarioJournaling bytes.Buffer
	binary.Write(&binarioJournaling, binary.BigEndian, journaling)
	EscribirBytes(file, binarioJournaling.Bytes())
	respuesta += "+++[COMANDO-MKFS]: Se ha formateado correctamente la partición \"" + string(particion.Part_name[:]) + "\"\n"
	return respuesta
}

func EscribirEstructurasEXT(id string, superBloque Estructuras.Superbloque, n int, particion *Estructuras.Partition) string{
	respuesta := ""
	superBloque.S_inodes_count = int64(n)
	superBloque.S_blocks_count = int64(3 * n)
	superBloque.S_free_inodes_count = int64(n)
	superBloque.S_free_blocks_count = int64(3 * n)
	fecha := time.Now().String()
	copy(superBloque.S_mtime[:], fecha)
	copy(superBloque.S_umtime[:], fecha)
	superBloque.S_mnt_count = 1
	superBloque.S_magic = 0xEF53
	superBloque.S_inode_s = int64(unsafe.Sizeof(Estructuras.Inodo{}))
	superBloque.S_block_s = int64(unsafe.Sizeof(Estructuras.BloqueCarpeta{}))
	superBloque.S_bm_block_start = superBloque.S_bm_inode_start + int64(n)
	superBloque.S_inode_start = superBloque.S_bm_block_start + int64(3*n)
	superBloque.S_block_start = superBloque.S_bm_inode_start + int64(n*int(unsafe.Sizeof(Estructuras.Inodo{})))
	file, err := os.OpenFile("./MIA/P1/"+string(id[0])+".dsk", os.O_WRONLY, os.ModeAppend)
	defer file.Close()
	if err != nil {
		respuesta += "---[ERROR-MKFS]: No se pudo encontrar el disco\n"
		return respuesta
	}
	//Escribir 0's en el espacio del bitmap de inodos y del bitmap de bloques
	var vacio int8 = 0
	refVacio := &vacio
	var binario2 bytes.Buffer
	binary.Write(&binario2, binary.BigEndian, refVacio)
	file.Seek(superBloque.S_bm_inode_start, 0)
	for i := 0; i < n; i++ {
		EscribirBytes(file, binario2.Bytes())
	}
	file.Seek(superBloque.S_bm_block_start, 0)
	for i := 0; i < 3*n; i++ {
		EscribirBytes(file, binario2.Bytes())
	}
	//Escribir la estructura de los inodos en el espacio correspondiente a estos
	inodo := Estructuras.NuevoInodo()
	var binario3 bytes.Buffer
	binary.Write(&binario3, binary.BigEndian, inodo)
	file.Seek(superBloque.S_inode_start, 0)
	for i := 0; i < n; i++ {
		EscribirBytes(file, binario3.Bytes())
	}
	//Escribir la estructura de los bloques de carpetas en el espacio correspondiente a estos
	bloqueCarpeta := Estructuras.NuevoBloqueCarpeta()
	file.Seek(superBloque.S_block_start, 0)
	var binario4 bytes.Buffer
	binary.Write(&binario4, binary.BigEndian, bloqueCarpeta)
	for i := 0; i < 3*n; i++ {
		EscribirBytes(file, binario4.Bytes())
	}
	//Directorio raiz
	inodo.I_uid = 1
	inodo.I_gid = 1
	inodo.I_s = 0
	copy(inodo.I_atime[:], fecha)
	copy(inodo.I_ctime[:], fecha)
	copy(inodo.I_mtime[:], fecha)
	inodo.I_block[0] = 0
	inodo.I_type = 0
	inodo.I_perm = 664
	//Bloque carpeta
	carpeta := Estructuras.NuevoBloqueCarpeta()
	copy(carpeta.B_content[0].B_name[:], ".")
	carpeta.B_content[0].B_inodo = int32(0)
	copy(carpeta.B_content[1].B_name[:], "..")
	carpeta.B_content[1].B_inodo = int32(0)
	copy(carpeta.B_content[2].B_name[:], "users.txt")
	carpeta.B_content[2].B_inodo = int32(1)
	//Inodo del archivo users.txt y el bloque con la información del mismo
	informacionArchivo := "1,G,root\n1,U,root,root,123\n"
	inodoUsers := Estructuras.NuevoInodo()
	inodoUsers.I_uid = 1
	inodoUsers.I_gid = 1
	inodoUsers.I_s = int64(unsafe.Sizeof(informacionArchivo) + unsafe.Sizeof(Estructuras.BloqueCarpeta{}))
	copy(inodoUsers.I_atime[:], fecha)
	copy(inodoUsers.I_ctime[:], fecha)
	copy(inodoUsers.I_mtime[:], fecha)
	inodoUsers.I_block[0] = 1
	inodoUsers.I_type = 1
	inodoUsers.I_perm = 664
	inodoUsers.I_s = inodoUsers.I_s + int64(unsafe.Sizeof(Estructuras.BloqueCarpeta{})) + int64(unsafe.Sizeof(Estructuras.Inodo{}))
	var archivoUsers Estructuras.BloqueArchivo
	copy(archivoUsers.B_content[:], informacionArchivo)
	//Escribir los inodos activos en el bitmap de inodos
	file.Seek(superBloque.S_bm_inode_start, 0)
	valor := '1'
	var binario5 bytes.Buffer
	binary.Write(&binario5, binary.BigEndian, valor)
	//Inodo 0
	EscribirBytes(file, binario5.Bytes())
	//Inodo 1
	EscribirBytes(file, binario5.Bytes())
	//Escribir los bloques activos en el bitmap de bloques
	file.Seek(superBloque.S_bm_block_start, 0)
	//Bloque 0
	EscribirBytes(file, binario5.Bytes())
	//Bloque 1
	EscribirBytes(file, binario5.Bytes())
	//Escribir los inodos
	file.Seek(superBloque.S_inode_start, 0)
	var binario6 bytes.Buffer
	binary.Write(&binario6, binary.BigEndian, inodo)
	EscribirBytes(file, binario6.Bytes())
	file.Seek(superBloque.S_inode_start+int64(unsafe.Sizeof(Estructuras.Inodo{})), 0)
	var binario7 bytes.Buffer
	binary.Write(&binario7, binary.BigEndian, inodoUsers)
	EscribirBytes(file, binario7.Bytes())
	//Escribir los bloques
	file.Seek(superBloque.S_block_start, 0)
	var binario8 bytes.Buffer
	binary.Write(&binario8, binary.BigEndian, carpeta)
	EscribirBytes(file, binario8.Bytes())
	file.Seek(superBloque.S_block_start+int64(unsafe.Sizeof(Estructuras.BloqueCarpeta{})), 0)
	var binario9 bytes.Buffer
	binary.Write(&binario9, binary.BigEndian, archivoUsers)
	EscribirBytes(file, binario9.Bytes())
	//Escribir el superbloque al inicio de la partición
	//Actualizar la cantidad de inodos y bloques libres
	superBloque.S_free_inodes_count = superBloque.S_free_inodes_count - 2
	superBloque.S_free_blocks_count = superBloque.S_free_blocks_count - 2
	//Actualizar el primer inodo y bloque libres
	superBloque.S_first_ino = superBloque.S_first_ino + 2
	superBloque.S_first_blo = superBloque.S_first_blo + 2
	file.Seek(particion.Part_start, 0)
	var binario1 bytes.Buffer
	binary.Write(&binario1, binary.BigEndian, superBloque)
	EscribirBytes(file, binario1.Bytes())
	return respuesta
}
