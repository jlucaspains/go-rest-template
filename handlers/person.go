package handlers

import (
	"net/http"
	"strconv"

	"goapi-template/models"

	"github.com/gin-gonic/gin"
)

// GetPerson godoc
//
//	@Summary		Retrieves a single person by id
//	@Description	get person by ID
//
//	@Security		OAuth2Implicit
//
//	@Tags			person
//	@Produce		json
//	@Param			id				path		int	true	"Person ID"
//	@Success		200				{object}	models.Person
//	@Failure		400				{object}	models.ErrorResult
//	@Router			/person/{id}	[get]
func (h Handlers) GetPerson(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 32)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, &models.ErrorResult{Errors: []string{"ID is invalid"}})
		return
	}

	body := models.Person{}
	if result := h.DB.First(&body, id); result.Error != nil {
		c.AbortWithStatusJSON(h.ErrorToHttpResult(result.Error))
		return
	}

	c.JSON(http.StatusOK, body)
}

// AddAccount godoc
//
//	@Summary		Add person
//	@Description	add by json person
//
//	@Security		OAuth2Implicit
//
//	@Tags			person
//	@Accept			json
//	@Produce		json
//	@Param			person	body		models.Person	true	"Add person"
//	@Success		202		{object}	models.Person
//	@Failure		400		{object}	[]string
//	@Router			/person [post]
func (h Handlers) PostPerson(c *gin.Context) {
	body := models.Person{}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.AbortWithStatusJSON(h.ErrorToHttpResult(err))
		return
	}

	body.ID = 0 // ensure we leverage auto increment

	if result := h.DB.Create(&body); result.Error != nil {
		c.AbortWithStatusJSON(h.ErrorToHttpResult(result.Error))
		return
	}

	c.JSON(http.StatusAccepted, &models.IdResult{ID: body.ID})
}

// PutPerson godoc
//
//	@Summary		Update person
//	@Description	update by json person
//
//	@Security		OAuth2Implicit
//
//	@Tags			person
//	@Accept			json
//	@Produce		json
//	@Param			id		path		int				true	"Person ID"
//	@Param			person	body		models.Person	true	"Update person"
//	@Success		202		{object}	models.Person
//	@Failure		400		{object}	models.ErrorResult
//	@Router			/person/{id} [put]
func (h Handlers) PutPerson(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 32)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, &models.ErrorResult{Errors: []string{"ID is invalid"}})
		return
	}
	body := models.Person{}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.AbortWithStatusJSON(h.ErrorToHttpResult(err))
		return
	}

	result := h.DB.Model(&models.Person{ID: int(id)}).Updates(&body)
	if result.Error != nil {
		c.AbortWithStatusJSON(h.ErrorToHttpResult(err))
		return
	}

	if result.RowsAffected == 0 {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	c.Status(http.StatusAccepted)
}

// DeletePerson godoc
//
//	@Summary		Delete person
//	@Description	Delete by id person
//	@Security		OAuth2Implicit
//	@Tags			person
//	@Accept			json
//	@Produce		json
//	@Param			id	path	int	true	"Person ID"
//	@Success		202
//	@Failure		400	{object}	models.ErrorResult
//	@Router			/person/{id} [delete]
func (h Handlers) DeletePerson(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 32)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, &models.ErrorResult{Errors: []string{"ID is invalid"}})
		return
	}

	result := h.DB.Delete(&models.Person{}, id)

	if result.Error != nil {
		c.AbortWithStatusJSON(h.ErrorToHttpResult(err))
		return
	}

	if result.RowsAffected == 0 {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	c.Status(http.StatusAccepted)
}
