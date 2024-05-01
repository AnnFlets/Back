package main

import (
	"PROYECTO_MIA/Comandos"
	"PROYECTO_MIA/Estructuras"
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"unsafe"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

var respuesta = ""

func main() {
	app := fiber.New()
	app.Use(cors.New())
	app.Post("/", RecibirComandos)
	app.Get("/discos", EnviarDiscos)
	app.Post("/particiones", EnviarParticiones)
	app.Post("/particion", EntrarParticion)
	app.Post("/login", VerificarLogin)
	app.Get("/logout", Logout)
	app.Get("/enviarreportes", MandarReportes)
	app.Post("/eliminarreporte", EliminarReporte)
	app.Listen(":4000")
}

/*
Función que recibe los comandos enviados desde la consola del frontend y retorna las respuestas, en un
arreglo, de las ejecuciones de los comandos
*/
func RecibirComandos(c *fiber.Ctx) error {
	var comandos Estructuras.Comando
	c.BodyParser(&comandos)
	VerificarComandos(comandos.Comand)
	respuestaArreglo := strings.Split(respuesta, "\n")
	respuestaArreglo = respuestaArreglo[:(len(respuestaArreglo)-1)]
	return c.JSON(&fiber.Map{
		"comandos": respuestaArreglo,
	})
}

//Función que se encarga de enviar un arreglo con los discos existentes
func EnviarDiscos(c *fiber.Ctx) error {
	path := ""
	arregloDiscos := []string{}
	letras := []string{"A", "B", "C", "D", "E", "F", "G", "H", "I", "J", "K", "L", "M", "N", "O", "P", "Q", "R", "S", "T", "U", "V", "W", "X", "Y", "Z"}
	for i := 0; i < len(letras); i++ {
		path = "./MIA/P1/" + letras[i] + ".dsk"
		if Comandos.VerificarArchivoExiste(path) {
			arregloDiscos = append(arregloDiscos, letras[i] + ".dsk")
		}
	}
	return c.JSON(&fiber.Map{
		"discos": arregloDiscos,
	})
}

//Función que se encarga de enviar un arreglo con las particiones de determinado disco
func EnviarParticiones(c *fiber.Ctx) error{
	arregloParticiones := []string{}
	fin := 0
	nombreParticion := ""
	var disco Estructuras.DiscoPeticion
	c.BodyParser(&disco)
	mbr, _ := Comandos.LeerDisco("./MIA/P1/" + disco.Driveletter)
	particiones := Comandos.ObtenerArregloParticiones(*mbr)
	for i := 0; i < len(particiones); i++{
		if particiones[i].Part_status == '1'{
			fin = bytes.IndexByte(particiones[i].Part_name[:], 0)
			nombreParticion = string(particiones[i].Part_name[:fin])
			arregloParticiones = append(arregloParticiones, nombreParticion)
		}
	}
	return c.JSON(&fiber.Map{
		"particiones": arregloParticiones,
	})
}

//Función para verificar si la partición cuenta con un sistema de archivos para poder hacer login
func EntrarParticion(c *fiber.Ctx) error{
	var particion Estructuras.ParticionPeticion
	c.BodyParser(&particion)
	mbr, _ := Comandos.LeerDisco("./MIA/P1/" + particion.Disco)
	particionBuscada, _ := Comandos.BuscarParticion(*mbr, particion.Disco, particion.Nombre)
	file, _ := os.Open("./MIA/P1/" + particion.Disco)
	defer file.Close()
	superbloque := Estructuras.NuevoSuperbloque()
	file.Seek(particionBuscada.Part_start, 0)
	informacionSuperbloque := Comandos.LeerBytes(file, int(unsafe.Sizeof(Estructuras.Superbloque{})))
	buffer := bytes.NewBuffer(informacionSuperbloque)
	binary.Read(buffer, binary.BigEndian, &superbloque)
	if superbloque.S_filesystem_type == 0 {
		return c.JSON(&fiber.Map{
			"status": 0,
			"id": string(particionBuscada.Part_id[:]),
		})
	}
	return c.JSON(&fiber.Map{
		"status": 1,
		"id": string(particionBuscada.Part_id[:]),
	})
}

//Función que recibe los parámetros de user, password e id para poder hacer login en la partición
func VerificarLogin(c *fiber.Ctx) error{
	var usuario Estructuras.InicioSesion
	c.BodyParser(&usuario)
	if Comandos.UsuarioActivo.Nombre != ""{
		return c.JSON(&fiber.Map{
			"status": 2,
		})
	}
	mbr, _ := Comandos.LeerDisco("./MIA/P1/" + string(usuario.Id[0]) + ".dsk")
	particionBuscada, _ := Comandos.BuscarParticion(*mbr, string(usuario.Id[0]), usuario.Nombre)
	Comandos.IniciarSesion(usuario.User, usuario.Password, usuario.Id, particionBuscada)
	archivoUsers := Comandos.LeerArchivoUsers(string(usuario.Id[0]), particionBuscada)
	arregloArchivo := strings.Split(archivoUsers, "\n")
	if Comandos.UsuarioActivo.Nombre != ""{
		return c.JSON(&fiber.Map{
			"status": 1,
			"nombre": "users.txt",
			"contenido": arregloArchivo,
		})
	}
	return c.JSON(&fiber.Map{
		"status": 0,
	})
}

//Función que se encarga de cerrar sesión
func Logout(c *fiber.Ctx) error {
	Comandos.CerrarSesion()
	if Comandos.UsuarioActivo.Nombre != ""{
		return c.JSON(&fiber.Map{
			"status": 0,
		})
	}
	return c.JSON(&fiber.Map{
		"status": 1,
	})
}

/*
Función que se encarga de mandar un arreglo con la información de los reportes existentes, enviando
información del nombre y el reporte en base64 o txt
*/
func MandarReportes(c *fiber.Ctx) error{
	arregloReportes := []Estructuras.ReporteEnviar{}
	archivos, err := ioutil.ReadDir("./Reportes")
    if err != nil {
        log.Fatal(err)
    }
    for _, archivo := range archivos {
		datosArchivo := strings.Split(archivo.Name(), ".")
		if datosArchivo[1] == "jpg" || datosArchivo[1] == "png" || datosArchivo[1] == "txt"{
			var reporte Estructuras.ReporteEnviar
			reporte.Nombre = archivo.Name()
			reporte.Extension = datosArchivo[1]
			if datosArchivo[1] == "jpg" || datosArchivo[1] == "png"{
				var imgBase string
				imgBytes, err := os.ReadFile("./Reportes/" + archivo.Name())
				if err != nil{
					fmt.Println("No pudo leerse la imagen")
				}
				if datosArchivo[1] == "jpg"{
					imgBase = "data:image/jpg;base64," + base64.StdEncoding.EncodeToString(imgBytes)
				}else {
					imgBase = "data:image/png;base64," + base64.StdEncoding.EncodeToString(imgBytes)
				}
				reporte.Contenido = imgBase
			}else{
				datosComoBytes, err := ioutil.ReadFile("./Reportes/" + archivo.Name())
				if err != nil{
					log.Fatal(err)
				}
				content := string(datosComoBytes)
				reporte.Contenido = content
			}
			arregloReportes = append(arregloReportes, reporte)
		}
    }
	return c.JSON(&fiber.Map{
		"reportes": arregloReportes,
	})
}

//Función que se encarga de eliminar un reporte en específico
func EliminarReporte(c *fiber.Ctx) error{
	var reporte Estructuras.Reporte
	c.BodyParser(&reporte)
	datosReporte := strings.Split(reporte.Nombre, ".")
	if datosReporte[1] == "txt"{
		err1 := os.Remove("./Reportes/" + reporte.Nombre)
		if err1 != nil {
			return c.JSON(&fiber.Map{
			"status": 0,
			})
		}
		return c.JSON(&fiber.Map{
			"status": 1,
		})
	}
	err1 := os.Remove("./Reportes/" + reporte.Nombre)
	err2 := os.Remove("./Reportes/" + datosReporte[0] + ".dot")
	if err1 != nil || err2 != nil {
		return c.JSON(&fiber.Map{
			"status": 0,
		})
	}
	return c.JSON(&fiber.Map{
		"status": 1,
	})
}

//Función que se encarga de dividir los comandos que vengan en la variable 'entrada' y comprobarlos
func VerificarComandos(entrada string) {
	respuesta = ""
	comandos := strings.Split(entrada, "\n")
	for i := 0; i < len(comandos); i++ {
		comandos[i] = strings.TrimSpace(comandos[i])
		if comandos[i] == ""{
			continue
		}
		if Comandos.Comparar(comandos[i], "exit") {
			respuesta += "\n******************** COMANDO EXIT ********************\n"
			respuesta += "+++[COMANDO-EXIT]: exit\n"
			break
		} else if Comandos.Comparar(comandos[i], "pause") {
			pausar()
			continue
		} else if string((comandos[i])[0]) == "#" {
			comentario(comandos[i])
		} else {
			//'comando' corresponde a la primera palabra de la línea del comando
			comando := strings.Split(comandos[i], " ")[0]
			if Comandos.Comparar(comando, "pause") {
				pausar()
			} else {
				entradaSinComando := strings.TrimLeft(comandos[i], comando)
				entradaSinComando = strings.TrimLeft(entradaSinComando, " ")
				parametros := separarParametros(entradaSinComando)
				comprobarComando(comandos[i], comando, parametros)
			}
		}
	}
}

//Función que se encarga de separar los parámetros de una entrada y devuelve un arreglo de strings con los parámetros (sin el signo "-")
func separarParametros(lineaEntrada string) []string {
	var parametros []string
	if lineaEntrada == "" {
		return parametros
	}
	lineaEntrada += " "
	var parametro string
	estado := 0
	//Se recorre la entrada caracter por caracter
	for i := 0; i < len(lineaEntrada); i++ {
		caracterEntrada := string(lineaEntrada[i])
		//El nombre del parámetro tiene antes un signo '-'. Si es así, se pasa al siguiente estado
		if estado == 0 && caracterEntrada == "-" {
			estado = 1
		} else if estado != 0 {
			if caracterEntrada == " " && (estado == 1 || estado == 2) {
				continue
			}
			if estado == 1 {
				if caracterEntrada == "=" {
					estado = 2
				}
			} else if estado == 2 {
				if caracterEntrada == "\"" {
					estado = 3
					continue
				} else {
					estado = 4
				}
			} else if estado == 3 {
				if caracterEntrada == "\"" {
					estado = 4
					continue
				}
			} else if estado == 4 {
				if caracterEntrada == "\"" {
					parametros = []string{}
					continue
				} else if caracterEntrada == " " {
					parametros = append(parametros, parametro)
					estado = 0
					parametro = ""
					continue
				}
			}
			parametro += caracterEntrada
		}
	}
	return parametros
}

//Función para verificar que el comando cuente con la cantidad de parámetros mínimos
func cumpleCantidadParametrosMin(cantParametros int, comando string) bool {
	if cantParametros < 0 {
		respuesta += "---[ERROR-" + strings.ToUpper(comando) + "]: No se ingresaron los parámetros solicitados\n"
		return false
	}
	return true
}

//Función para comprobar el comando ingresado y realizar la función solicitada
func comprobarComando(lineaEntrada string, comando string, parametros []string) {
	if comando != "" {
		if Comandos.Comparar(comando, "mkdisk") {
			respuesta += "\n******************** COMANDO MKDISK ********************\n"
			respuesta += "+++[COMANDO-" + strings.ToUpper(comando) + "]: " + lineaEntrada + "\n"
			if !cumpleCantidadParametrosMin(len(parametros), comando) {
				return
			}
			respuesta += Comandos.ComprobarParametrosMKDISK(parametros)
		} else if Comandos.Comparar(comando, "rmdisk") {
			respuesta += "\n******************** COMANDO RMDISK ********************\n"
			respuesta += "+++[COMANDO-" + strings.ToUpper(comando) + "]: " + lineaEntrada + "\n"
			if !cumpleCantidadParametrosMin(len(parametros), comando) {
				return
			}
			respuesta += Comandos.ComprobarParametrosRMDISK(parametros)
		} else if Comandos.Comparar(comando, "fdisk") {
			respuesta += "\n******************** COMANDO FDISK ********************\n"
			respuesta += "+++[COMANDO-" + strings.ToUpper(comando) + "]: " + lineaEntrada + "\n"
			if !cumpleCantidadParametrosMin(len(parametros), comando) {
				return
			}
			respuesta += Comandos.ComprobarParametrosFDisk(parametros)
		} else if Comandos.Comparar(comando, "mount") {
			respuesta += "\n******************** COMANDO MOUNT ********************\n"
			respuesta += "+++[COMANDO-" + strings.ToUpper(comando) + "]: " + lineaEntrada + "\n"
			if !cumpleCantidadParametrosMin(len(parametros), comando) {
				return
			}
			respuesta += Comandos.ComprobarParametrosMount(parametros)
		} else if Comandos.Comparar(comando, "unmount") {
			respuesta += "\n******************** COMANDO UNMOUNT ********************\n"
			respuesta += "+++[COMANDO-" + strings.ToUpper(comando) + "]: " + lineaEntrada + "\n"
			if !cumpleCantidadParametrosMin(len(parametros), comando) {
				return
			}
			respuesta += Comandos.ComprobarParametrosUnmount(parametros)
		} else if Comandos.Comparar(comando, "listar") {
			respuesta += "\n******************** COMANDO LISTAR ********************\n"
			respuesta += "+++[COMANDO-" + strings.ToUpper(comando) + "]: " + lineaEntrada + "\n"
			if parametros == nil {
				respuesta += Comandos.MostrarParticionesMontadas()
			} else {
				respuesta += "---[ERROR-LISTAR]: Se ingresaron más parámetros de los solicitados\n"
				return
			}
		} else if Comandos.Comparar(comando, "mkfs"){
			respuesta += "\n******************** COMANDO MKFS ********************\n"
			respuesta += "+++[COMANDO-" + strings.ToUpper(comando) + "]: " + lineaEntrada + "\n"
			if !cumpleCantidadParametrosMin(len(parametros), comando) {
				return
			}
			respuesta += Comandos.ComprobarParametrosMKFS(parametros)
		} else if Comandos.Comparar(comando, "login"){
			respuesta += "\n******************** COMANDO LOGIN ********************\n"
			respuesta += "+++[COMANDO-" + strings.ToUpper(comando) + "]: " + lineaEntrada + "\n"
			if !cumpleCantidadParametrosMin(len(parametros), comando) {
				return
			}
			respuesta += Comandos.ComprobarParametrosLogin(parametros)
		} else if Comandos.Comparar(comando, "logout"){
			respuesta += "\n******************** COMANDO LOGOUT ********************\n"
			respuesta += "+++[COMANDO-" + strings.ToUpper(comando) + "]: " + lineaEntrada + "\n"
			respuesta += Comandos.CerrarSesion()
		} else if Comandos.Comparar(comando, "rep") {
			respuesta += "\n******************** COMANDO REP ********************\n"
			respuesta += "+++[COMANDO-" + strings.ToUpper(comando) + "]: " + lineaEntrada + "\n"
			if !cumpleCantidadParametrosMin(len(parametros), comando) {
				return
			}
			respuesta += Comandos.ComprobarParametrosREP(parametros)
		} else {
			respuesta += "\n******************** COMANDO NO IDENTIFICADO ********************\n"
			respuesta += "---[ERROR-COMANDO]: No se reconoce el comando \"" + lineaEntrada + "\"\n"
		}
	}
}

//Función para el comando 'PAUSE'
func pausar() {
	respuesta += "\n******************** PAUSE ********************\n"
	respuesta += "+++[COMANDO-PAUSE]: pause\n"
}

//Función para el comando 'COMENTARIO' #
func comentario(entrada string) {
	respuesta += "\n******************** COMENTARIO ********************\n"
	respuesta += "+++[COMANDO-COMENTARIO]: " + entrada + "\n"
}