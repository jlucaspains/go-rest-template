package handlers

import (
	"net/http"
	"strconv"

	"goapi-template/models"
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
func (h Handlers) GetPerson(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(h.Param(r, "id"), 10, 32)
	if err != nil {
		h.JSON(w, http.StatusBadRequest, &models.ErrorResult{Errors: []string{"ID is invalid"}})
		return
	}

	body := models.Person{}
	if result := h.DB.First(&body, id); result.Error != nil {
		status, body := h.ErrorToHttpResult(result.Error)
		h.JSON(w, status, body)
		return
	}

	h.JSON(w, http.StatusOK, body)
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
func (h Handlers) PostPerson(w http.ResponseWriter, r *http.Request) {
	body := &models.Person{}
	if err := h.BindJSON(r, body); err != nil {
		status, body := h.ErrorToHttpResult(err)
		h.JSON(w, status, body)
		return
	}

	body.ID = 0 // ensure we leverage auto increment
	body.UpdateUser = h.GetUserEmail(r)

	if result := h.DB.Create(&body); result.Error != nil {
		status, body := h.ErrorToHttpResult(result.Error)
		h.JSON(w, status, body)
		return
	}

	h.JSON(w, http.StatusAccepted, &models.IdResult{ID: body.ID})
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
func (h Handlers) PutPerson(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(h.Param(r, "id"), 10, 32)
	if err != nil {
		h.JSON(w, http.StatusBadRequest, &models.ErrorResult{Errors: []string{"ID is invalid"}})
		return
	}
	body := &models.Person{}
	if err := h.BindJSON(r, body); err != nil {
		status, err := h.ErrorToHttpResult(err)
		h.JSON(w, status, err)
		return
	}

	body.UpdateUser = h.GetUserEmail(r)

	result := h.DB.Model(&models.Person{ID: int(id)}).Updates(&body)
	if result.Error != nil {
		status, err := h.ErrorToHttpResult(err)
		h.JSON(w, status, err)
		return
	}

	if result.RowsAffected == 0 {
		h.Status(w, http.StatusNotFound)
		return
	}

	h.Status(w, http.StatusAccepted)
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
func (h Handlers) DeletePerson(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(h.Param(r, "id"), 10, 32)
	if err != nil {
		h.JSON(w, http.StatusBadRequest, &models.ErrorResult{Errors: []string{"ID is invalid"}})
		return
	}

	result := h.DB.Delete(&models.Person{}, id)

	if result.Error != nil {
		status, err := h.ErrorToHttpResult(err)
		h.JSON(w, status, err)
		return
	}

	if result.RowsAffected == 0 {
		h.Status(w, http.StatusNotFound)
		return
	}

	h.Status(w, http.StatusAccepted)
}
