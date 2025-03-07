package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/hashicorp-forge/hermes/internal/config"
	"github.com/hashicorp-forge/hermes/pkg/algolia"
	"github.com/hashicorp-forge/hermes/pkg/models"
	"github.com/hashicorp/go-hclog"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type TeamRequest struct {
	TeamName string `json:"teamName,omitempty"`
	TeamBU   string `json:"teamBU,omitempty"`
}

type TeamData struct {
	BU             string                 `json:"BU"`
	PerDocTypeData interface{}            `json:"perDocDataType"`
	Projects       map[string]interface{} `json:"projects"`
}

// TeamsHandler returns the product mappings to the Hermes frontend.
func TeamsHandler(cfg *config.Config, ar *algolia.Client,
	aw *algolia.Client, db *gorm.DB, log hclog.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		switch r.Method {
		case "POST":
			// Decode request.
			var req TeamRequest
			if err := decodeRequest(r, &req); err != nil {
				log.Error("error decoding teams request", "error", err)
				http.Error(w, fmt.Sprintf("Bad request: %q", err),
					http.StatusBadRequest)
				return
			}

			// Add the data to both algolia and the Postgres Database
			err := AddNewTeams(ar, aw, db, req)
			if err != nil {
				log.Error("error inserting new product/Business Unit", "error", err)
				http.Error(w, "Error inserting products",
					http.StatusInternalServerError)
				return
			}

			// Send success response
			// Send success response with success message
			response := struct {
				Message string `json:"message"`
			}{
				Message: "Team/Pod Inserted successfully",
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			enc := json.NewEncoder(w)
			err = enc.Encode(response)
			if err != nil {
				log.Error("error encoding teams response", "error", err)
				http.Error(w, "Error creating new teams",
					http.StatusInternalServerError)
				return
			}

		case "GET":
			// Get teams and associated data from Algolia
			products, err := getTeamsData(db)
			if err != nil {
				log.Error("error getting teams from database", "error", err)
				http.Error(w, "Error getting product mappings",
					http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)

			enc := json.NewEncoder(w)
			err = enc.Encode(products)
			if err != nil {
				log.Error("error encoding teams response", "error", err)
				http.Error(w, "Error getting products",
					http.StatusInternalServerError)
				return
			}
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
			return

		}

	})
}

// getTeamsData gets the teams and their associated
// data from Database
func getTeamsData(db *gorm.DB) (map[string]TeamData, error) {
	var teams []models.Team

	if err := db.Preload(clause.Associations).Find(&teams).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch teams: %w", err)
	}

	teamsData := make(map[string]TeamData)

	for _, team := range teams {
		projectDataList := make(map[string]interface{})
		for _, project := range team.Projects {
			projectData := map[string]interface{}{
				"teamid": project.TeamID,
				// Add any other project-related data you want to include here
			}
			projectDataList[project.Name] = projectData
		}

		teamsData[team.Name] = TeamData{
			BU:             team.BU.Name,
			PerDocTypeData: nil,
			Projects:       projectDataList,
		}
	}

	return teamsData, nil
}

// AddNewTeams This helper function adds the newly added product and upserts it
// in the postgres Database
func AddNewTeams(ar *algolia.Client,
	aw *algolia.Client, db *gorm.DB, req TeamRequest) error {
	pm := models.Team{
		Name: req.TeamName,
	}
	if err := pm.Upsert(db, req.TeamBU); err != nil {
		return fmt.Errorf("error upserting product: %w", err)
	}

	return nil
}
