package controllers

import (
	"fmt"
	//"time"
	//"strconv"
	//"strings"
	//"encoding/json"
	"github.com/astaxie/beego"
	//. "github.com/mndrix/golog"
	"github.com/udistrital/administrativa_mid_api/models"

)

//GestionDocumentoResolucionController operations for Preliquidacion
type GestionDocumentoResolucionController struct {
	beego.Controller
}

// URLMapping ...
func (c *GestionDocumentoResolucionController) URLMapping() {


}

// GestionPrevinculacionesController ...
// @Title ListarDocentesCargaHoraria
// @Description create GetContenidoResolucion
// @Param id_resolucion query string false "año a consultar"
// @Param id_facultad query string false "periodo a listar"
// @Success 201 {object} models.ResolucionCompleta
// @Failure 403 body is empty
// @router get_contenido_resolucion [get]
func (c *GestionDocumentoResolucionController) GetContenidoResolucion() {
	id_resolucion := c.GetString("id_resolucion")
	id_facultad := c.GetString("id_facultad")
	var contenidoResolucion models.ResolucionCompleta
	var ordenador_gasto []models.OrdenadorGasto
	var jefe_dependencia []models.JefeDependencia
	var query string

	if err2 := getJson("http://"+beego.AppConfig.String("UrlcrudAdmin")+"/"+beego.AppConfig.String("NscrudAdmin")+"/contenido_resolucion/"+id_resolucion, &contenidoResolucion); err2 == nil {
		query = "?limit=-1&query=DependenciaId:"+id_facultad

		if err := getJson("http://"+beego.AppConfig.String("UrlcrudCore")+"/"+beego.AppConfig.String("NscrudCore")+"/ordenador_gasto"+query, &ordenador_gasto); err == nil {
			if(ordenador_gasto == nil){
				if err := getJson("http://"+beego.AppConfig.String("UrlcrudCore")+"/"+beego.AppConfig.String("NscrudCore")+"/ordenador_gasto/1", &ordenador_gasto); err == nil {
					contenidoResolucion.OrdenadorGasto = ordenador_gasto[0]
				}else{
							fmt.Println("Error al consultar ordenador 1", err2)
					}
			}else{
				contenidoResolucion.OrdenadorGasto = ordenador_gasto[0]
			}


		} else {

			fmt.Println("Error al consultar ordenador del gasto", err2)
		}

	} else {
		fmt.Println("Error al consultar contenido", err2)
	}

	query="?query=DependenciaId:"+id_facultad
	if err := getJson("http://"+beego.AppConfig.String("UrlcrudCore")+"/"+beego.AppConfig.String("NscrudCore")+"/jefe_dependencia"+query, &jefe_dependencia); err == nil {
		contenidoResolucion.OrdenadorGasto.NombreOrdenador = BuscarNombreProveedor(jefe_dependencia[0].TerceroId)
	}else{

	}

	c.Ctx.Output.SetStatus(201)
	c.Data["json"] = contenidoResolucion
	c.ServeJSON()
	
}