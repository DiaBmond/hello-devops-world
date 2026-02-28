package http

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"go-prod-app/internal/domain"
	"go-prod-app/internal/repository"
	"go-prod-app/internal/service"

	"github.com/google/uuid"
)

type Handler struct {
	userService *service.UserService
}

func NewHandler(userService *service.UserService) *Handler {
	return &Handler{userService: userService}
}

func (h *Handler) users(w http.ResponseWriter, r *http.Request) {
	switch r.Method {

	case http.MethodGet:

		q := r.URL.Query()

		// limit
		limit := 10
		if l := q.Get("limit"); l != "" {
			if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
				limit = parsed
			}
		}

		// cursor (decode from base64 JSON)
		var cursor *repository.Cursor
		if c := q.Get("cursor"); c != "" {

			raw, err := base64.StdEncoding.DecodeString(c)
			if err != nil {
				writeError(w, http.StatusBadRequest, "invalid cursor")
				return
			}

			var decoded repository.Cursor
			if err := json.Unmarshal(raw, &decoded); err != nil {
				writeError(w, http.StatusBadRequest, "invalid cursor")
				return
			}

			cursor = &decoded
		}

		// email filter (*string)
		var emailPtr *string
		if e := q.Get("email"); e != "" {
			emailPtr = &e
		}

		filter := repository.UserFilter{
			Email: emailPtr,
		}

		users, nextCursor, err := h.userService.ListUsers(
			r.Context(),
			filter,
			cursor,
			limit,
		)
		if err != nil {
			handleServiceError(w, err)
			return
		}

		var resp []UserResponse
		for _, u := range users {
			resp = append(resp, toUserResponse(u))
		}

		response := map[string]any{
			"data": resp,
		}

		// encode next cursor back to base64 JSON
		if nextCursor != nil {
			b, _ := json.Marshal(nextCursor)
			response["next_cursor"] = base64.StdEncoding.EncodeToString(b)
		}

		writeJSON(w, http.StatusOK, response)

	case http.MethodPost:

		var req CreateUserRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		user, err := h.userService.CreateUser(r.Context(), req.Name, req.Email)
		if err != nil {
			handleServiceError(w, err)
			return
		}

		writeJSON(w, http.StatusCreated, toUserResponse(user))

	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (h *Handler) userByID(w http.ResponseWriter, r *http.Request) {

	id := r.URL.Path[len("/users/"):]

	if _, err := uuid.Parse(id); err != nil {
		writeError(w, http.StatusBadRequest, "invalid id format")
		return
	}

	switch r.Method {

	case http.MethodGet:

		user, err := h.userService.GetUser(r.Context(), domain.UserID(id))
		if err != nil {
			handleServiceError(w, err)
			return
		}

		writeJSON(w, http.StatusOK, toUserResponse(user))

	case http.MethodPut:

		var req UpdateUserRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		user, err := h.userService.UpdateUser(
			r.Context(),
			domain.UserID(id),
			req.Name,
			req.Email,
		)
		if err != nil {
			handleServiceError(w, err)
			return
		}

		writeJSON(w, http.StatusOK, toUserResponse(user))

	case http.MethodDelete:

		if err := h.userService.DeleteUser(
			r.Context(),
			domain.UserID(id),
		); err != nil {
			handleServiceError(w, err)
			return
		}

		w.WriteHeader(http.StatusNoContent)

	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (h *Handler) health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{
		"status": "ok",
	})
}

func (h *Handler) ready(w http.ResponseWriter, r *http.Request) {
	if err := h.userService.Ping(r.Context()); err != nil {
		writeError(w, http.StatusServiceUnavailable, "not ready")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{
		"status": "ready",
	})
}

func handleServiceError(w http.ResponseWriter, err error) {

	switch {
	case errors.Is(err, service.ErrInvalidInput):
		writeError(w, http.StatusBadRequest, err.Error())

	case errors.Is(err, service.ErrDuplicateEmail):
		writeError(w, http.StatusConflict, err.Error())

	case errors.Is(err, service.ErrUserNotFound):
		writeError(w, http.StatusNotFound, err.Error())

	case errors.Is(err, service.ErrConflict):
		writeError(w, http.StatusConflict, err.Error())

	default:
		writeError(w, http.StatusInternalServerError, "internal server error")
	}
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{
		"error": message,
	})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
