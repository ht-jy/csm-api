package handler

import (
	"csm-api/auth"
	"net/http"
)

type JwtValidHandler struct {
	Jwt *auth.JWTUtils
}

func (handler *JwtValidHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	claims, err := handler.Jwt.ValidateJWT(r)
	if err != nil {
		RespondJSON(
			r.Context(),
			w,
			ErrResponse{
				Result:         Failure,
				Message:        err.Error(),
				Details:        InvalidToken,
				HttpStatusCode: http.StatusUnauthorized,
			},
			http.StatusOK,
		)
		return
	}

	rsp := Response{
		Result: Success,
		Values: struct {
			Message string         `json:"message"`
			Claims  auth.JWTClaims `json:"claims"`
		}{Message: "jwt Validate ok", Claims: *claims},
	}

	RespondJSON(ctx, w, &rsp, http.StatusOK)
}
