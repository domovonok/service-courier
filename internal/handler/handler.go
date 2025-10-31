package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	sq "github.com/Masterminds/squirrel"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func New(p *pgxpool.Pool) http.Handler {
	h := &courierHandler{pool: p}
	r := chi.NewRouter()
	r.Get("/ping", h.ping)
	r.Head("/healthcheck", h.healthcheck)
	r.Get("/courier/{id}", h.get)
	r.Get("/couriers", h.list)
	r.Post("/courier", h.create)
	r.Put("/courier", h.update)
	return r
}

type courierHandler struct {
	pool *pgxpool.Pool
}

var psql = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

func (h *courierHandler) writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func (h *courierHandler) ping(w http.ResponseWriter, _ *http.Request) {
	h.writeJSON(w, http.StatusOK, map[string]string{"message": "pong"})
}

func (h *courierHandler) healthcheck(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}

func (h *courierHandler) get(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil || id < 1 {
		h.writeError(w, ErrInvalidID)
		return
	}

	var c courier

	query, args, _ := psql.
		Select("id", "name", "phone", "status").
		From("couriers").
		Where(sq.Eq{"id": id}).
		ToSql()

	if err := h.pool.QueryRow(r.Context(), query, args...).
		Scan(&c.ID, &c.Name, &c.Phone, &c.Status); err != nil {

		if errors.Is(err, pgx.ErrNoRows) {
			err = ErrNotFound
		}
		h.writeError(w, err)
		return
	}

	h.writeJSON(w, http.StatusOK, c)
}

func (h *courierHandler) list(w http.ResponseWriter, r *http.Request) {
	query, args, _ := psql.
		Select("id", "name", "phone", "status").
		From("couriers").
		ToSql()

	rows, err := h.pool.Query(r.Context(), query, args...)
	if err != nil {
		h.writeError(w, err)
		return
	}
	defer rows.Close()

	couriers := make([]courier, 0)
	for rows.Next() {
		var c courier
		if err := rows.Scan(&c.ID, &c.Name, &c.Phone, &c.Status); err != nil {
			h.writeError(w, err)
			return
		}
		couriers = append(couriers, c)
	}
	if err := rows.Err(); err != nil {
		h.writeError(w, err)
		return
	}

	h.writeJSON(w, http.StatusOK, couriers)
}

func (h *courierHandler) decodeAndValidate(w http.ResponseWriter, r *http.Request, c *courier) bool {
	if err := json.NewDecoder(r.Body).Decode(c); err != nil || !c.validate() {
		h.writeError(w, ErrInvalidInput)
		return false
	}
	return true
}

func (h *courierHandler) create(w http.ResponseWriter, r *http.Request) {
	c := courier{ID: 1}
	if !h.decodeAndValidate(w, r, &c) {
		return
	}

	query, args, _ := psql.
		Insert("couriers").
		Columns("name", "phone", "status").
		Values(c.Name, c.Phone, c.Status).
		Suffix("RETURNING id").
		ToSql()

	if err := h.pool.QueryRow(r.Context(), query, args...).Scan(&c.ID); err != nil {
		h.handleDBError(w, err)
		return
	}

	h.writeJSON(w, http.StatusCreated, c)
}

func (h *courierHandler) update(w http.ResponseWriter, r *http.Request) {
	var c courier
	if !h.decodeAndValidate(w, r, &c) {
		return
	}

	query, args, _ := psql.
		Update("couriers").
		Set("name", c.Name).
		Set("phone", c.Phone).
		Set("status", c.Status).
		Where(sq.Eq{"id": c.ID}).
		ToSql()

	cmdTag, err := h.pool.Exec(r.Context(), query, args...)
	if err != nil {
		h.handleDBError(w, err)
		return
	}
	if cmdTag.RowsAffected() == 0 {
		h.writeError(w, ErrNotFound)
		return
	}

	h.writeJSON(w, http.StatusOK, c)
}
