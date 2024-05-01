package Comandos

import (
	"PROYECTO_MIA/Estructuras"
	"bytes"
	"encoding/binary"
	"os"
	"strings"
	"unsafe"
)

var UsuarioActivo Usuario

// Estructura para almacenar la información del usuario logueado y la partición
type Usuario struct {
	Gid        int64
	Uid        int64
	Grupo      string
	Nombre     string
	Contrasena string
}

// Función para comprobar los parámetros del comando LOGIN
func ComprobarParametrosLogin(parametros []string) string {
	respuesta := ""
	user := ""
	pass := ""
	id := ""
	for i := 0; i < len(parametros); i++ {
		datosParametro := strings.Split(parametros[i], "=")
		if Comparar(datosParametro[0], "user") {
			if user == "" {
				user = datosParametro[1]
			} else {
				respuesta += "---[ERROR-LOGIN]: Parámetro \"user\" repetido\n"
				return respuesta
			}
		} else if Comparar(datosParametro[0], "pass") {
			if pass == "" {
				pass = datosParametro[1]
			} else {
				respuesta += "---[ERROR-LOGIN]: Parámetro \"pass\" repetido\n"
				return respuesta
			}
		} else if Comparar(datosParametro[0], "id") {
			if id == "" {
				id = datosParametro[1]
			} else {
				respuesta += "---[ERROR-LOGIN]: Parámetro \"id\" repetido\n"
				return respuesta
			}
		} else {
			respuesta += "---[ERROR-LOGIN]: No se esperaba el parámetro \"" + datosParametro[0] + "\"\n"
			return respuesta
		}
	}
	if user == "" {
		respuesta += "---[ERROR-LOGIN]: Se requiere el parámetro \"user\" para este comando\n"
		return respuesta
	}
	if pass == "" {
		respuesta += "---[ERROR-LOGIN]: Se requiere el parámetro \"pass\" para este comando\n"
		return respuesta
	}
	if id == "" {
		respuesta += "---[ERROR-LOGIN]: Se requiere el parámetro \"id\" para este comando\n"
		return respuesta
	}
	if !VerificarArchivoExiste("./MIA/P1/" + string(id[0]) + ".dsk") {
		respuesta += "---[ERROR-LOGIN]: El disco \"" + string(id[0]) + "\" no existe\n"
		return respuesta
	}
	if !VerificarParticionEstaMontada(id) {
		respuesta += "---[ERROR-LOGIN]: No existe una partición montada con el id \"" + id + "\"\n"
		return respuesta
	}
	if UsuarioActivo.Nombre != "" {
		respuesta += "---[ERROR-LOGIN]: El usuario ya se encuentra logueado. Debe cerrar sesión\n"
		return respuesta
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
		respuesta += "---[ERROR-LOGIN]: No existe una partición con el nombre \"" + nombreParticion + "\"\n"
		return respuesta
	}
	respuesta += IniciarSesion(user, pass, id, particionBuscada)
	return respuesta
}

func IniciarSesion(user string, pass string, id string, particion *Estructuras.Partition) string{
	respuesta := ""
	file, err1 := os.Open("./MIA/P1/" + string(id[0]) + ".dsk")
	defer file.Close()
	if err1 != nil {
		respuesta += "---[ERROR-LOGIN]: No se pudo encontrar el disco \"" + string(id[0]) + "\"\n"
		return respuesta
	}
	//Se lee el superbloque de la partición
	superbloque := Estructuras.NuevoSuperbloque()
	file.Seek(particion.Part_start, 0)
	informacion := LeerBytes(file, int(unsafe.Sizeof(Estructuras.Superbloque{})))
	buffer := bytes.NewBuffer(informacion)
	err2 := binary.Read(buffer, binary.BigEndian, &superbloque)
	if err2 != nil {
		respuesta += "---[ERROR-LOGIN]: No pudo leerse el disco\n"
		return respuesta
	}
	if superbloque.S_filesystem_type == 0 {
		respuesta += "---[ERROR-LOGIN]: La partición no cuenta con un sistema de archivos\n"
		return respuesta
	}
	//Se lee el inodo de users.txt
	inodoUsers := Estructuras.NuevoInodo()
	//Se ubicará el puntero en el byte donde empieza el inodo 1, que es el que contiene users.txt
	file.Seek(superbloque.S_inode_start+int64(unsafe.Sizeof(Estructuras.Inodo{})), 0)
	informacion = LeerBytes(file, int(unsafe.Sizeof(Estructuras.Inodo{})))
	buffer = bytes.NewBuffer(informacion)
	err2 = binary.Read(buffer, binary.BigEndian, &inodoUsers)
	if err2 != nil {
		respuesta += "---[ERROR-LOGIN]: No pudo leerse el archivo\n"
		return respuesta
	}
	//Se lee la información del bloque de archivo con los usuarios y grupos
	bloqueUsers := Estructuras.BloqueArchivo{}
	archivoLeido := false
	informacionUsers := ""
	for i := 0; i < len(inodoUsers.I_block); i++ {
		if inodoUsers.I_block[0] == -1 {
			break
		}
		file.Seek(superbloque.S_block_start+int64(unsafe.Sizeof(Estructuras.BloqueCarpeta{}))+int64(unsafe.Sizeof(Estructuras.BloqueArchivo{}))*int64(i), 0)
		informacion = LeerBytes(file, int(unsafe.Sizeof(Estructuras.BloqueArchivo{})))
		buffer = bytes.NewBuffer(informacion)
		err3 := binary.Read(buffer, binary.BigEndian, &bloqueUsers)
		if err3 != nil {
			respuesta += "---[ERROR-LOGIN]: Error al leer el archivo\n"
			return respuesta
		}
		for j := 0; j < len(bloqueUsers.B_content); j++ {
			if bloqueUsers.B_content[j] != 0 {
				informacionUsers += string(bloqueUsers.B_content[j])
			} else {
				archivoLeido = true
			}
		}
		if archivoLeido {
			break
		}
	}
	lineas := strings.Split(informacionUsers, "\n")
	var usuario Usuario
	grupoExiste := false
	for i := 0; i < len(lineas); i++ {
		if lineas[i] != "" {
			datosLineaUser := strings.Split(lineas[i], ",")
			if Comparar(datosLineaUser[1], "u") {
				if !Comparar(datosLineaUser[0], "0") && (datosLineaUser[3] == user) && (datosLineaUser[4] == pass) {
					usuario.Nombre = user
					usuario.Contrasena = pass
					usuario.Grupo = datosLineaUser[2]
					usuario.Uid = int64(1)
					for j := 0; j < len(lineas); j++ {
						if lineas[j] != "" {
							datosLineaGrupo := strings.Split(lineas[j], ",")
							if Comparar(datosLineaGrupo[1], "g") {
								if !Comparar(datosLineaGrupo[0], "0") && Comparar(datosLineaGrupo[2], usuario.Grupo) {
									usuario.Gid = int64(1)
									grupoExiste = true
									break
								}
							}
						}
					}
					if !grupoExiste {
						respuesta += "---[ERROR-LOGIN]: No se encontró el grupo \"" + usuario.Grupo + "\"\n"
						return respuesta
					}
					UsuarioActivo = usuario
					respuesta += "+++[COMANDO-LOGIN]: Usuario \"" + user + "\" logueado con éxito\n"
					return respuesta
				}
			}
		}
	}
	respuesta += "---[ERROR-LOGIN]: No se encontró el usuario \"" + user + "\"\n"
	return respuesta
}

//Función para cerra una sesión
func CerrarSesion() string{
	respuesta := ""
	if UsuarioActivo.Nombre != "" {
		respuesta += "+++[COMANDO-LOGOUT]: Se ha cerrado la sesión del usuario \"" + UsuarioActivo.Nombre + "\"\n"
		UsuarioActivo = Usuario{}
	} else {
		respuesta += "---[ERROR-LOGOUT]: No hay sesiones activas\n"
	}
	return respuesta
}

//Función que devuelve el contenido del archivo users.txt
func LeerArchivoUsers(id string, particion *Estructuras.Partition) string{
	file, _ := os.Open("./MIA/P1/" + id + ".dsk")
	defer file.Close()
	superbloque := Estructuras.NuevoSuperbloque()
	file.Seek(particion.Part_start, 0)
	informacion := LeerBytes(file, int(unsafe.Sizeof(Estructuras.Superbloque{})))
	buffer := bytes.NewBuffer(informacion)
	binary.Read(buffer, binary.BigEndian, &superbloque)
	inodoUsers := Estructuras.NuevoInodo()
	file.Seek(superbloque.S_inode_start+int64(unsafe.Sizeof(Estructuras.Inodo{})), 0)
	informacion = LeerBytes(file, int(unsafe.Sizeof(Estructuras.Inodo{})))
	buffer = bytes.NewBuffer(informacion)
	binary.Read(buffer, binary.BigEndian, &inodoUsers)
	bloqueUsers := Estructuras.BloqueArchivo{}
	archivoLeido := false
	informacionUsers := ""
	for i := 0; i < len(inodoUsers.I_block); i++ {
		if inodoUsers.I_block[0] == -1 {
			break
		}
		file.Seek(superbloque.S_block_start+int64(unsafe.Sizeof(Estructuras.BloqueCarpeta{}))+int64(unsafe.Sizeof(Estructuras.BloqueArchivo{}))*int64(i), 0)
		informacion = LeerBytes(file, int(unsafe.Sizeof(Estructuras.BloqueArchivo{})))
		buffer = bytes.NewBuffer(informacion)
		binary.Read(buffer, binary.BigEndian, &bloqueUsers)
		for j := 0; j < len(bloqueUsers.B_content); j++ {
			if bloqueUsers.B_content[j] != 0 {
				informacionUsers += string(bloqueUsers.B_content[j])
			} else {
				archivoLeido = true
			}
		}
		if archivoLeido {
			break
		}
	}
	return informacionUsers
}