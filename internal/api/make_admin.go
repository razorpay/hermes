package api

import (
	"encoding/json"
	"fmt"
	"github.com/hashicorp-forge/hermes/pkg/models"
	"github.com/hashicorp/go-hclog"
	"gorm.io/gorm"
	"net/http"
)

// MakeUserAdminHandler handles the API request to make a user an admin.
func MakeUserAdminHandler(log hclog.Logger, db *gorm.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		switch r.Method {
		case "POST":
			// Decode the request to get the email ID of the user to be made an admin.
			var req struct {
				Email string `json:"email"`
			}
			if err := decodeRequest(r, &req); err != nil {
				log.Error("error decoding make user admin request", "error", err)
				http.Error(w, fmt.Sprintf("Bad request: %q", err), http.StatusBadRequest)
				return
			}

			var user models.User
			if err := db.Where("email_address = ?", req.Email).First(&user).Error; err != nil {
				log.Error("error finding user", "error", err)
				http.Error(w, "User not found", http.StatusNotFound)
				return
			}

			// Update the user's role to "admin" (assuming "Role" is the field in the User model that stores roles).
			user.Role = models.Admin
			if err := db.Save(&user).Error; err != nil {
				log.Error("error updating user role", "error", err)
				http.Error(w, "Error updating user role", http.StatusInternalServerError)
				return
			}

			// Send success response with success message.
			response := struct {
				Message string `json:"message"`
			}{
				Message: "User is now an admin",
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			enc := json.NewEncoder(w)
			err := enc.Encode(response)
			if err != nil {
				log.Error("error encoding response", "error", err)
				http.Error(w, "Error encoding response", http.StatusInternalServerError)
				return
			}
		}
	})
}
