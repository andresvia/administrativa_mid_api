package controllers

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/astaxie/beego"
	"github.com/udistrital/administrativa_mid_api/models"
)

// PreliquidacionController operations for Preliquidacion
type GestionDesvinculacionesController struct {
	beego.Controller
}

// URLMapping ...
func (c *GestionDesvinculacionesController) URLMapping() {

	c.Mapping("ActualizarVinculaciones", c.ActualizarVinculaciones)
	c.Mapping("AdicionarHoras", c.AdicionarHoras)

}

// GestionDesvinculacionesController ...
// @Title ListarDocentesDesvinculados
// @Description create ListarDocentesDesvinculados
// @Param id_resolucion query string false "resolucion a consultar"
// @Success 201 {int} models.VinculacionDocente
// @Failure 403 body is empty
// @router /docentes_desvinculados [get]
func (c *GestionDesvinculacionesController) ListarDocentesDesvinculados() {
	fmt.Println("docentes desvinculados")
	id_resolucion := c.GetString("id_resolucion")
	query := "?limit=-1&query=IdResolucion.Id:" + id_resolucion + ",Estado:false"
	var v []models.VinculacionDocente

	if err2 := getJson(beego.AppConfig.String("ProtocolAdmin")+"://"+beego.AppConfig.String("UrlcrudAdmin")+"/"+beego.AppConfig.String("NscrudAdmin")+"/vinculacion_docente"+query, &v); err2 == nil {
		for x, pos := range v {
			documento_identidad, _ := strconv.Atoi(pos.IdPersona)
			v[x].NombreCompleto = BuscarNombreProveedor(documento_identidad)
			v[x].NumeroDisponibilidad = BuscarNumeroDisponibilidad(pos.Disponibilidad)
			v[x].Dedicacion = BuscarNombreDedicacion(pos.IdDedicacion.Id)
			v[x].LugarExpedicionCedula = BuscarLugarExpedicion(pos.IdPersona)
		}

	} else {
		fmt.Println("Error de consulta en vinculacion", err2)
	}

	c.Ctx.Output.SetStatus(201)
	c.Data["json"] = v
	c.ServeJSON()

}

// ActualizarVinculaciones ...
// @Title ActualizarVinculaciones
// @Description create ActualizarVinculaciones
// @Success 201 {string}
// @Failure 403 body is empty
// @router /actualizar_vinculaciones [post]
func (c *GestionDesvinculacionesController) ActualizarVinculaciones() {

	var v models.Objeto_Desvinculacion
	var respuesta interface{}

	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &v); err == nil {
		fmt.Println("para poner en false", v)

		for _, pos := range v.DocentesDesvincular {
			if err2 := sendJson(beego.AppConfig.String("ProtocolAdmin")+"://"+beego.AppConfig.String("UrlcrudAdmin")+"/"+beego.AppConfig.String("NscrudAdmin")+"/vinculacion_docente/"+strconv.Itoa(pos.Id), "PUT", &respuesta, pos); err2 == nil {
				fmt.Println("respuesta", respuesta)
			} else {
				fmt.Println("error en json", err2)
			}
		}

		fmt.Println("Id para modificacion,res", v.IdModificacionResolucion)

		for _, pos := range v.DocentesDesvincular {
			temp := models.ModificacionVinculacion{ModificacionResolucion: &models.ModificacionResolucion{Id: v.IdModificacionResolucion}, VinculacionDocenteCancelada: &models.VinculacionDocente{Id: pos.Id}}
			if err2 := sendJson(beego.AppConfig.String("ProtocolAdmin")+"://"+beego.AppConfig.String("UrlcrudAdmin")+"/"+beego.AppConfig.String("NscrudAdmin")+"/modificacion_vinculacion/", "POST", respuesta, temp); err2 == nil {
				fmt.Println("respuesta", respuesta)
			} else {
				fmt.Println("error en json de modificacion vinculacion", err2)
			}
		}

		c.Data["json"] = respuesta
	} else {
		fmt.Println("ERROR")
		fmt.Println(err)
		c.Data["json"] = "Error al leer json para desvincular"
	}

	c.ServeJSON()

}

// AdicionarHoras ...
// @Title AdicionarHoras
// @Description create AdicionarHoras
// @Success 201 {string}
// @Failure 403 body is empty
// @router /adicionar_horas [post]
func (c *GestionDesvinculacionesController) AdicionarHoras() {

	var v models.Objeto_Desvinculacion
	var respuesta_mod_vin models.ModificacionVinculacion
	var respuesta string
	var vinculacion_nueva int
	var temp_vinculacion [1]models.VinculacionDocente

	err := json.Unmarshal(c.Ctx.Input.RequestBody, &v)
	if err != nil {
		beego.Error(err)
		c.Data["json"] = "Error al leer json para desvincular"
	}

	//CAMBIAR ESTADO DE VINCULACIÓN DOCNETE
	for _, pos := range v.DocentesDesvincular {
		err := sendJson(beego.AppConfig.String("ProtocolAdmin")+"://"+beego.AppConfig.String("UrlcrudAdmin")+"/"+beego.AppConfig.String("NscrudAdmin")+"/vinculacion_docente/"+strconv.Itoa(pos.Id), "PUT", &respuesta, pos)
		//TODO: unificar errores
		if err != nil {
			err = fmt.Errorf("error al cambiar estado en vinculación docente al adicionar horas %s", err)
			beego.Error(err)
			c.Abort("403")
		}
		beego.Info("respuesta", respuesta)

		temp_vinculacion[0] = models.VinculacionDocente{
			IdPersona:            v.DocentesDesvincular[0].IdPersona,
			NumeroHorasSemanales: v.DocentesDesvincular[0].NumeroHorasNuevas,
			NumeroSemanas:        v.DocentesDesvincular[0].NumeroSemanasNuevas,
			IdResolucion:         &models.ResolucionVinculacionDocente{Id: v.IdNuevaResolucion},
			IdDedicacion:         v.DocentesDesvincular[0].IdDedicacion,
			IdProyectoCurricular: v.DocentesDesvincular[0].IdProyectoCurricular,
			Categoria:            v.DocentesDesvincular[0].Categoria,
			Dedicacion:           v.DocentesDesvincular[0].Dedicacion,
			NivelAcademico:       v.DocentesDesvincular[0].NivelAcademico,
			Disponibilidad:       v.DisponibilidadNueva,
			Vigencia:             v.DocentesDesvincular[0].Vigencia,
		}

		//CREAR NUEVA Vinculacion
		vinculacion_nueva, err = InsertarDesvinculaciones(temp_vinculacion)
		if err != nil {
			beego.Error("error al realizar vinculacion nueva", err)
			c.Abort("400")
		}

		//INSERCION  TABLA  DE TRAZA MODIFICACION VINCULACION
		for _, pos := range v.DocentesDesvincular {
			temp := models.ModificacionVinculacion{
				ModificacionResolucion:       &models.ModificacionResolucion{Id: v.IdModificacionResolucion},
				VinculacionDocenteCancelada:  &models.VinculacionDocente{Id: pos.Id},
				VinculacionDocenteRegistrada: &models.VinculacionDocente{Id: vinculacion_nueva},
				Horas: pos.NumeroHorasNuevas,
			}
			err := sendJson(beego.AppConfig.String("ProtocolAdmin")+"://"+beego.AppConfig.String("UrlcrudAdmin")+"/"+beego.AppConfig.String("NscrudAdmin")+"/modificacion_vinculacion/", "POST", &respuesta_mod_vin, temp)

			if err != nil {
				beego.Error("error en actualizacion de modificacion vinculacion de modificacion vinculacion", err)
				respuesta = "error"
			} else {
				beego.Info("respuesta modificacion vin", respuesta_mod_vin)
				respuesta = "OK"
			}
		}
	}

	c.Data["json"] = respuesta

	c.ServeJSON()

}

// GestionDesvinculacionesController ...
// @Title AnularDesvinculacionDocente
// @Description create AnularDesvinculacionDocente
// @Success 201 {string}
// @Failure 403 body is empty
// @router /anular_desvinculacion [post]
func (c *GestionDesvinculacionesController) AnularDesvinculacionDocente() {
	fmt.Println("anular desvinculacion")
	var v models.Objeto_Desvinculacion
	var respuesta_vinculacion string
	var respuesta_delete string
	var respuesta_total string
	var respuesta_modificacion_vinculacion []models.ModificacionVinculacion

	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &v); err == nil {
		respuesta_total = "OK"
		if err2 := sendJson(beego.AppConfig.String("ProtocolAdmin")+"://"+beego.AppConfig.String("UrlcrudAdmin")+"/"+beego.AppConfig.String("NscrudAdmin")+"/vinculacion_docente/"+strconv.Itoa(v.DocentesDesvincular[0].Id), "PUT", &respuesta_vinculacion, v.DocentesDesvincular[0]); err2 == nil {
			respuesta_total = "OK"
		} else {
			respuesta_total = "error"
		}

		query := "?limit=-1&query=ModificacionResolucion.Id:" + strconv.Itoa(v.IdModificacionResolucion) + ",VinculacionDocenteCancelada.Id:" + strconv.Itoa(v.DocentesDesvincular[0].Id)
		if err2 := getJson(beego.AppConfig.String("ProtocolAdmin")+"://"+beego.AppConfig.String("UrlcrudAdmin")+"/"+beego.AppConfig.String("NscrudAdmin")+"/modificacion_vinculacion"+query, &respuesta_modificacion_vinculacion); err2 == nil {
			respuesta_total = "OK"
		} else {
			respuesta_total = "error"
		}

		if err2 := sendJson(beego.AppConfig.String("ProtocolAdmin")+"://"+beego.AppConfig.String("UrlcrudAdmin")+"/"+beego.AppConfig.String("NscrudAdmin")+"/modificacion_vinculacion/"+strconv.Itoa(respuesta_modificacion_vinculacion[0].Id), "DELETE", &respuesta_delete, respuesta_modificacion_vinculacion[0]); err2 == nil {
			respuesta_total = "OK"
		} else {
			respuesta_total = "error"
		}

	} else {
		respuesta_total = "error"
	}

	c.Data["json"] = respuesta_total
	c.ServeJSON()
}

// GestionDesvinculacionesController ...
// @Title AnularAdicionDocente
// @Description create AnularAdicionDocente
// @Success 201 {string}
// @Failure 403 body is empty
// @router /anular_adicion [post]
func (c *GestionDesvinculacionesController) AnularAdicionDocente() {
	fmt.Println("anular adicion")
	var v models.Objeto_Desvinculacion
	var respuesta_vinculacion string
	var vinculacion_cancelada []models.VinculacionDocente
	var respuesta_delete_vin string
	var respuesta_delete string
	var respuesta_total string
	var respuesta_modificacion_vinculacion []models.ModificacionVinculacion

	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &v); err == nil {
		respuesta_total = "OK"

		//Se trae información de tabla de traza modificacion_vinculacion, para saber cuál vinculación hay que poner en true y cuál eliminar
		query := "?limit=-1&query=ModificacionResolucion.Id:" + strconv.Itoa(v.IdModificacionResolucion) + ",VinculacionDocenteRegistrada.Id:" + strconv.Itoa(v.DocentesDesvincular[0].Id)
		if err2 := getJson(beego.AppConfig.String("ProtocolAdmin")+"://"+beego.AppConfig.String("UrlcrudAdmin")+"/"+beego.AppConfig.String("NscrudAdmin")+"/modificacion_vinculacion"+query, &respuesta_modificacion_vinculacion); err2 == nil {
			fmt.Println("modificacion_vinculacion", respuesta_modificacion_vinculacion)
			respuesta_total = "OK"
		} else {
			respuesta_total = "error"
		}

		//se trae informacion de vinculacion que fue cancelada
		query2 := "?limit=-1&query=Id:" + strconv.Itoa(respuesta_modificacion_vinculacion[0].VinculacionDocenteCancelada.Id)
		if err2 := getJson(beego.AppConfig.String("ProtocolAdmin")+"://"+beego.AppConfig.String("UrlcrudAdmin")+"/"+beego.AppConfig.String("NscrudAdmin")+"/vinculacion_docente"+query2, &vinculacion_cancelada); err2 == nil {
			fmt.Println("vinculacion_cancelada", vinculacion_cancelada)
			respuesta_total = "OK"
		} else {
			respuesta_total = "error"
		}
		//se cambia a true vinculacion que fue cancelada
		vinculacion_cancelada[0].Estado = true
		fmt.Println("nuevo estado de vinculacion cancelada", vinculacion_cancelada)

		//Se le cambia estado en bd a vinculacion cancelada

		if err2 := sendJson(beego.AppConfig.String("ProtocolAdmin")+"://"+beego.AppConfig.String("UrlcrudAdmin")+"/"+beego.AppConfig.String("NscrudAdmin")+"/vinculacion_docente/"+strconv.Itoa(vinculacion_cancelada[0].Id), "PUT", &respuesta_vinculacion, vinculacion_cancelada[0]); err2 == nil {
			fmt.Println("respuesta_vinculacion", respuesta_vinculacion)
			respuesta_total = "OK"
		} else {
			respuesta_total = "error"
		}

		//se elimina registro en modificacion_vinculacion

		if err2 := sendJson(beego.AppConfig.String("ProtocolAdmin")+"://"+beego.AppConfig.String("UrlcrudAdmin")+"/"+beego.AppConfig.String("NscrudAdmin")+"/modificacion_vinculacion/"+strconv.Itoa(respuesta_modificacion_vinculacion[0].Id), "DELETE", &respuesta_delete, respuesta_modificacion_vinculacion[0]); err2 == nil {
			respuesta_total = "OK"
		} else {
			respuesta_total = "error"
		}

		//Se elimina vinculacion nueva
		if err2 := sendJson(beego.AppConfig.String("ProtocolAdmin")+"://"+beego.AppConfig.String("UrlcrudAdmin")+"/"+beego.AppConfig.String("NscrudAdmin")+"/vinculacion_docente/"+strconv.Itoa(v.DocentesDesvincular[0].Id), "DELETE", &respuesta_delete_vin, v.DocentesDesvincular[0]); err2 == nil {
			fmt.Println("respuesta_eliminar_vin_nueva", respuesta_delete_vin)
			respuesta_total = "OK"
		} else {
			respuesta_total = "error"
		}

	} else {
		respuesta_total = "error"
	}

	c.Data["json"] = respuesta_total
	c.ServeJSON()
}

func InsertarDesvinculaciones(v [1]models.VinculacionDocente) (id int, err error) {
	var d []models.VinculacionDocente
	json_ejemplo, err := json.Marshal(v)
	if err != nil {
		beego.Error(err)
		return id, err
	}
	err = json.Unmarshal(json_ejemplo, &d)

	if err != nil {
		beego.Error(err)
		return id, err
	}

	beego.Debug("docentes a contratar", d)
	//TODO: unificar cont con error
	d, err = CalcularSalarioPrecontratacion(d)
	if err != nil {
		return id, err
	}

	err = sendJson(beego.AppConfig.String("ProtocolAdmin")+"://"+beego.AppConfig.String("UrlcrudAdmin")+"/"+beego.AppConfig.String("NscrudAdmin")+"/vinculacion_docente/InsertarVinculaciones/", "POST", &id, &d)
	if err != nil {
		beego.Error(err)
		return 0, err
	}
	return id, err
}

// GestionCanceladosController ...
// @Title ListarDocentesCancelados
// @Description create ListarDocentesCancelados
// @Param id_resolucion query string false "resolucion a consultar"
// @Success 201 {int} models.VinculacionDocente
// @Failure 403 body is empty
// @router /docentes_cancelados [get]
func (c *GestionDesvinculacionesController) ListarDocentesCancelados() {
	id_resolucion := c.GetString("id_resolucion")
	var v []models.VinculacionDocente
	var modRes []models.ModificacionResolucion
	var modVin []models.ModificacionVinculacion
	var cv models.VinculacionDocente
	// if 3 - modificacion_resolucion
	if err := getJson(beego.AppConfig.String("ProtocolAdmin")+"://"+beego.AppConfig.String("UrlcrudAdmin")+"/"+beego.AppConfig.String("NscrudAdmin")+"/modificacion_resolucion/?query=resolucionNueva:"+id_resolucion, &modRes); err == nil {
		// if 2 - modificacion_vinculacion
		t := beego.AppConfig.String("ProtocolAdmin") + "://" + beego.AppConfig.String("UrlcrudAdmin") + "/" + beego.AppConfig.String("NscrudAdmin") + "/modificacion_vinculacion/?limit=-1&query=modificacion_resolucion:" + strconv.Itoa(modRes[0].Id)
		beego.Info(t)
		if err := getJson(t, &modVin); err == nil {
			//for vinculaciones
			for _, vinculacion := range modVin {
				beego.Info(fmt.Sprintf("%+v", vinculacion.VinculacionDocenteCancelada))
				// if 1 - vinculacion_docente
				if err := getJson(beego.AppConfig.String("ProtocolAdmin")+"://"+beego.AppConfig.String("UrlcrudAdmin")+"/"+beego.AppConfig.String("NscrudAdmin")+"/vinculacion_docente/"+strconv.Itoa(vinculacion.VinculacionDocenteCancelada.Id), &cv); err == nil {
					documento_identidad, _ := strconv.Atoi(vinculacion.VinculacionDocenteCancelada.IdPersona)
					cv.NombreCompleto = BuscarNombreProveedor(documento_identidad)
					cv.NumeroDisponibilidad = BuscarNumeroDisponibilidad(vinculacion.VinculacionDocenteCancelada.Disponibilidad)
					cv.Dedicacion = BuscarNombreDedicacion(vinculacion.VinculacionDocenteCancelada.IdDedicacion.Id)
					cv.LugarExpedicionCedula = BuscarLugarExpedicion(vinculacion.VinculacionDocenteCancelada.IdPersona)
				} else { // if 1 - vinculacion_docente
					fmt.Println("Error de consulta en vinculacion, solucioname!!!, if 1 - vinculacion_docente: ", err)
				}
				v = append(v, cv)
			} //fin for vinculaciones
		} else { // if 2 - modificacion_vinculacion
			fmt.Println("Error de consulta en modificacion_vinculacion, solucioname!!!, if 2 - modificacion_vinculacion: ", err)
		}
	} else { // if 3 - modificacion_resolucion
		fmt.Println("Error de consulta en modificacion_resolucion, solucioname!!!, if 3 - modificacion_resolucion: ", err)
	}
	c.Ctx.Output.SetStatus(201)
	c.Data["json"] = v
	c.ServeJSON()
}
