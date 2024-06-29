package handlers

import (
	"net/http"
	"strconv"

	"goapi-template/db"
	"goapi-template/models"

	"github.com/jackc/pgx/v5/pgtype"
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

	result, err := h.Queries.GetPersonById(r.Context(), int32(id))

	if err != nil {
		status, body := h.ErrorToHttpResult(err)
		h.JSON(w, status, body)
		return
	}
	body := models.Person{
		ID:         int(result.ID),
		Name:       result.Name,
		Email:      result.Email,
		CreatedAt:  result.CreatedAt.Time,
		UpdatedAt:  result.UpdatedAt.Time,
		UpdateUser: result.UpdateUser,
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

	result, err := h.Queries.InsertPerson(r.Context(), db.InsertPersonParams{
		Name:       body.Name,
		Email:      body.Email,
		CreatedAt:  pgtype.Timestamp{Time: body.CreatedAt, Valid: true},
		UpdatedAt:  pgtype.Timestamp{Time: body.UpdatedAt, Valid: true},
		UpdateUser: body.UpdateUser,
	})

	if err != nil {
		status, body := h.ErrorToHttpResult(err)
		h.JSON(w, status, body)
		return
	}

	h.JSON(w, http.StatusAccepted, &models.IdResult{ID: int(result.ID)})
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

	_, err = h.Queries.UpdatePerson(r.Context(), db.UpdatePersonParams{
		ID:         int32(id),
		Name:       body.Name,
		Email:      body.Email,
		CreatedAt:  pgtype.Timestamp{Time: body.CreatedAt, Valid: true},
		UpdatedAt:  pgtype.Timestamp{Time: body.UpdatedAt, Valid: true},
		UpdateUser: body.UpdateUser,
	})

	if err != nil {
		status, err := h.ErrorToHttpResult(err)
		h.JSON(w, status, err)
		return
	}

	// if result.RowsAffected == 0 {
	// 	h.Status(w, http.StatusNotFound)
	// 	return
	// }

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

	result, err := h.Queries.DeletePerson(r.Context(), int32(id))

	if err != nil {
		status, err := h.ErrorToHttpResult(err)
		h.JSON(w, status, err)
		return
	}

	if result == 0 {
		h.Status(w, http.StatusNotFound)
		return
	}

	h.Status(w, http.StatusAccepted)
}
