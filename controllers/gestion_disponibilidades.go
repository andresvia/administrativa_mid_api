package controllers

import (
	"encoding/json"
	"fmt"

	"github.com/astaxie/beego"
	"github.com/udistrital/administrativa_mid_api/models"
)

// PreliquidacionController operations for Preliquidacion
type GestionDisponibilidadController struct {
	beego.Controller
}

// URLMapping ...
func (c *GestionDisponibilidadController) URLMapping() {
	c.Mapping("ListarApropiaciones", c.ListarApropiaciones)

}

// ListarApropiaciones ...
// @Title ListarApropiaciones
// @Description create ListarApropiaciones
// @Success 201 {int} models.DisponibilidadApropiacion
// @Failure 403 body is empty
// @router /listar_apropiaciones [post]
func (c *GestionDisponibilidadController) ListarApropiaciones() {

	var v []models.DisponibilidadApropiacion
	var respuesta models.DatosApropiacion
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &v); err == nil {

		for x, pos := range v {

			if err2 := sendJson(beego.AppConfig.String("ProtocolAdmin")+"://"+beego.AppConfig.String("UrlcrudKronos")+"/"+beego.AppConfig.String("NscrudKronos")+"/disponibilidad/SaldoCdp", "POST", &respuesta, &pos); err2 == nil {
				v[x].Apropiacion.Saldo = int(respuesta.Saldo)
				fmt.Println("respuesta", respuesta)

			} else {
				fmt.Println("error en json", err2)
			}
		}
		c.Data["json"] = v
	} else {
		fmt.Println("ERROR")
		fmt.Println(err)
		c.Data["json"] = "Error al listar disponibilidades"
	}

	c.ServeJSON()
}
