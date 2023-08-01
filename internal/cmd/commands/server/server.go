package server

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"time"

	"github.com/hashicorp/go-hclog"

	"github.com/algolia/algoliasearch-client-go/algoliasearch"
	"github.com/algolia/algoliasearch-client-go/v3/algolia/search"
	"github.com/hashicorp-forge/hermes/internal/api"
	"github.com/hashicorp-forge/hermes/internal/auth"
	"github.com/hashicorp-forge/hermes/internal/cmd/base"
	"github.com/hashicorp-forge/hermes/internal/config"
	"github.com/hashicorp-forge/hermes/internal/db"
	"github.com/hashicorp-forge/hermes/internal/email"
	"github.com/hashicorp-forge/hermes/internal/pkg/doctypes"
	"github.com/hashicorp-forge/hermes/internal/pub"
	slackbot "github.com/hashicorp-forge/hermes/internal/slack-bot"
	"github.com/hashicorp-forge/hermes/internal/structs"
	"github.com/hashicorp-forge/hermes/pkg/algolia"
	gw "github.com/hashicorp-forge/hermes/pkg/googleworkspace"
	hcd "github.com/hashicorp-forge/hermes/pkg/hashicorpdocs"
	"github.com/hashicorp-forge/hermes/pkg/links"
	"github.com/hashicorp-forge/hermes/pkg/models"
	"github.com/hashicorp-forge/hermes/web"
	"github.com/joho/godotenv"
	"gorm.io/gorm"
)

type Command struct {
	*base.Command

	flagAddr              string
	flagBaseURL           string
	flagConfig            string
	flagOktaAuthServerURL string
	flagOktaClientID      string
	flagOktaDisabled      bool
}

type endpoint struct {
	pattern string
	handler http.Handler
}

func (c *Command) Synopsis() string {
	return "Run the server"
}

func (c *Command) Help() string {
	return `Usage: hermes server

  This command runs the Hermes web server.` + c.Flags().Help()
}

func (c *Command) Flags() *base.FlagSet {
	f := base.NewFlagSet(flag.NewFlagSet("server", flag.ExitOnError))

	f.StringVar(
		&c.flagAddr, "addr", "127.0.0.1:8000",
		"[HERMES_SERVER_ADDR] Address to bind to for listening.",
	)
	f.StringVar(
		&c.flagBaseURL, "base-url", "http://localhost:8000",
		"[HERMES_BASE_URL] Base URL used for building links.",
	)
	f.StringVar(
		&c.flagConfig, "config", "", "Path to Hermes config file",
	)
	f.StringVar(
		&c.flagOktaAuthServerURL, "okta-auth-server-url", "",
		"[HERMES_SERVER_OKTA_AUTH_SERVER_URL] URL to the Okta authorization server.",
	)
	f.StringVar(
		&c.flagOktaClientID, "okta-client-id", "",
		"[HERMES_SERVER_OKTA_CLIENT_ID] Okta client ID.",
	)
	f.BoolVar(
		&c.flagOktaDisabled, "okta-disabled", false,
		"[HERMES_SERVER_OKTA_DISABLED] Disable Okta authorization.",
	)

	return f
}

func (c *Command) Run(args []string) int {
	f := c.Flags()
	if err := f.Parse(args); err != nil {
		c.UI.Error(fmt.Sprintf("error parsing flags: %v", err))
		return 1
	}

	var (
		cfg *config.Config
		err error
	)
	if c.flagConfig != "" {
		cfg, err = config.NewConfig(c.flagConfig)
		if err != nil {
			c.UI.Error(fmt.Sprintf("error parsing config file: %v: config=%q",
				err, c.flagConfig))
			return 1
		}
	}

	/* Remove this just for explicitly setting up the env variables*/
	err1 := godotenv.Load()
	if err1 != nil {
		log.Fatal("Error loading .env file")
	}

	// Access and print the environment variables
	//fmt.Println(os.LookupEnv("ALGOLIA_APPLICATION_ID"))
	//panic("")
	//fmt.Println(os.Getenv("ALGOLIA_SEARCH_API_KEY"))
	//fmt.Println(os.Getenv("ALGOLIA_WRITE_API_KEY"))
	//fmt.Println(os.Getenv("GOOGLE_WORKSPACE_OAUTH2_CLIENT_ID"))
	//fmt.Println(os.Getenv("GOOGLE_WORKSPACE_OAUTH2_HD"))
	//fmt.Println(os.Getenv("GOOGLE_WORKSPACE_OAUTH2_REDIRECT_URI"))
	//fmt.Println(os.Getenv("POSTGRES_PASSWORD"))
	//fmt.Println(os.Getenv("POSTGRES_USER"))

	// Get the sensitive details if present in the environment
	if val, ok := os.LookupEnv("ALGOLIA_APPLICATION_ID"); ok {
		cfg.Algolia.ApplicationID = val
	} else {
		c.UI.Error("ALGOLIA_APPLICATION_ID must be provided as an env variable!")
		return 1
	}
	if val, ok := os.LookupEnv("ALGOLIA_SEARCH_API_KEY"); ok {
		cfg.Algolia.SearchAPIKey = val
	} else {
		c.UI.Error("ALGOLIA_SEARCH_API_KEY must be provided as an env variable!")
		return 1
	}

	if val, ok := os.LookupEnv("ALGOLIA_WRITE_API_KEY"); ok {
		cfg.Algolia.WriteAPIKey = val
	} else {
		c.UI.Error("ALGOLIA_SEARCH_API_KEY must be provided as an env variable!")
		return 1
	}
	if val, ok := os.LookupEnv("GOOGLE_WORKSPACE_OAUTH2_CLIENT_ID"); ok {
		cfg.GoogleWorkspace.OAuth2.ClientID = val
	} else {
		c.UI.Error("GOOGLE_WORKSPACE_OAUTH2_CLIENT_ID must be provided as an env variable!")
		return 1
	}

	if val, ok := os.LookupEnv("GOOGLE_WORKSPACE_OAUTH2_HD"); ok {
		cfg.GoogleWorkspace.OAuth2.HD = val
	} else {
		c.UI.Error("GOOGLE_WORKSPACE_OAUTH2_HD must be provided as an env variable!")
		return 1
	}
	if val, ok := os.LookupEnv("GOOGLE_WORKSPACE_OAUTH2_REDIRECT_URI"); ok {
		cfg.GoogleWorkspace.OAuth2.RedirectURI = val
	} else {
		c.UI.Error("GOOGLE_WORKSPACE_OAUTH2_REDIRECT_URI must be provided as an env variable!")
		return 1
	}

	if val, ok := os.LookupEnv("GOOGLE_WORKSPACE_AUTH_CLIENT_EMAIL"); ok {
		cfg.GoogleWorkspace.Auth.ClientEmail = val
	} else {
		c.UI.Error("GOOGLE_WORKSPACE_AUTH_CLIENT_EMAIL must be provided as an env variable!")
		return 1
	}
	if val, ok := os.LookupEnv("GOOGLE_WORKSPACE_AUTH_PRIVATE_KEY"); ok {
		cfg.GoogleWorkspace.Auth.PrivateKey = val
	} else {
		c.UI.Error("GOOGLE_WORKSPACE_AUTH_PRIVATE_KEY must be provided as an env variable!")
		return 1
	}

	if val, ok := os.LookupEnv("GOOGLE_WORKSPACE_AUTH_SUBJECT"); ok {
		cfg.GoogleWorkspace.Auth.Subject = val
	} else {
		c.UI.Error("GOOGLE_WORKSPACE_AUTH_SUBJECT must be provided as an env variable!")
		return 1
	}

	if val, ok := os.LookupEnv("POSTGRES_PASSWORD"); ok {
		cfg.Postgres.Password = val
	} else {
		c.UI.Error("POSTGRES_PASSWORD must be provided as an env variable!")
		return 1
	}

	if val, ok := os.LookupEnv("POSTGRES_USER"); ok {
		cfg.Postgres.User = val
	} else {
		c.UI.Error("POSTGRES_USER_Name must be provided as an env variable!")
		return 1
	}

	if val, ok := os.LookupEnv("POSTGRES_DBNAME"); ok {
		cfg.Postgres.DBName = val
	} else {
		c.UI.Error("POSTGRES_dbname must be provided as an env variable!")
		return 1
	}
	if val, ok := os.LookupEnv("POSTGRES_HOST"); ok {
		cfg.Postgres.Host = val
	} else {
		c.UI.Error("POSTGRES_host must be provided as an env variable!")
		return 1
	}

	// scanning doc folder drive ids
	if val, ok := os.LookupEnv("DOCS_DRIVE_FOLDER_ID"); ok {
		cfg.GoogleWorkspace.DocsFolder = val
	} else {
		c.UI.Error("DOCS_DRIVE_FOLDER_ID must be provided as an env variable!")
		return 1
	}
	if val, ok := os.LookupEnv("DRAFTS_DRIVE_FOLDER_ID"); ok {
		cfg.GoogleWorkspace.DraftsFolder = val
	} else {
		c.UI.Error("DRAFTS_DRIVE_FOLDER_ID must be provided as an env variable!")
		return 1
	}
	if val, ok := os.LookupEnv("SHORTCUTS_DRIVE_FOLDER_ID"); ok {
		cfg.GoogleWorkspace.ShortcutsFolder = val
	} else {
		c.UI.Error("SHORTCUTS_DRIVE_FOLDER_ID must be provided as an env variable!")
		return 1
	}
	if val, ok := os.LookupEnv("EMAIL_FROM_ADDRESS"); ok {
		cfg.Email.FromAddress = val
	} else {
		c.UI.Error("EMAIL_FROM_ADDRESS must be provided as an env variable!")
		return 1
	}

	/* Scanned all env variables succesfully */

	// Get configuration from environment variables if not set on the command
	// line.
	// TODO: make this section more DRY and add tests.
	if val, ok := os.LookupEnv("HERMES_SERVER_ADDR"); ok {
		cfg.Server.Addr = val
	}
	if c.flagAddr != f.Lookup("addr").DefValue {
		cfg.Server.Addr = c.flagAddr
	}
	if val, ok := os.LookupEnv("HERMES_BASE_URL"); ok {
		cfg.BaseURL = val
	}
	if c.flagBaseURL != f.Lookup("base-url").DefValue {
		cfg.BaseURL = c.flagBaseURL
	}
	if val, ok := os.LookupEnv("HERMES_SERVER_OKTA_AUTH_SERVER_URL"); ok {
		cfg.Okta.AuthServerURL = val
	}
	if c.flagOktaAuthServerURL != f.Lookup("okta-auth-server-url").DefValue {
		cfg.Okta.AuthServerURL = c.flagOktaAuthServerURL
	}
	if val, ok := os.LookupEnv("HERMES_SERVER_OKTA_CLIENT_ID"); ok {
		cfg.Okta.ClientID = val
	}
	if c.flagOktaClientID != f.Lookup("okta-client-id").DefValue {
		cfg.Okta.ClientID = c.flagOktaClientID
	}
	if val, ok := os.LookupEnv("HERMES_SERVER_OKTA_DISABLED"); ok {
		if val == "" || val == "false" {
			// Keep Okta enabled if the env var value is an empty string or "false".
		} else {
			cfg.Okta.Disabled = true
		}
	}
	if c.flagOktaDisabled {
		cfg.Okta.Disabled = true
	}

	// Validate feature flags defined in configuration
	if cfg.FeatureFlags != nil {
		err := config.ValidateFeatureFlags(cfg.FeatureFlags.FeatureFlag)
		if err != nil {
			c.UI.Error(fmt.Sprintf("error initializing server: %v", err))
			return 1
		}
	}

	// Validate other configuration.
	if cfg.Email != nil && cfg.Email.Enabled {
		if cfg.Email.FromAddress == "" {
			c.UI.Error("email from_address must be set if email is enabled")
			return 1
		}
	}

	// Build configuration for Okta authentication.
	if !cfg.Okta.Disabled {
		// Check for required Okta configuration.
		if cfg.Okta.AuthServerURL == "" {
			c.UI.Error("error initializing server: Okta authorization server URL is required")
			return 1
		}
		if cfg.Okta.AWSRegion == "" {
			c.UI.Error("error initializing server: Okta AWS region is required")
			return 1
		}
		if cfg.Okta.ClientID == "" {
			c.UI.Error("error initializing server: Okta client ID is required")
			return 1
		}
	}

	// Initialize Google Workspace service.
	var goog *gw.Service
	if cfg.GoogleWorkspace.Auth != nil {
		// Use Google Workspace auth if it is defined in the config.
		goog = gw.NewFromConfig(cfg.GoogleWorkspace.Auth)
	} else {
		// Use OAuth if Google Workspace auth is not defined in the config.
		goog = gw.New()
	}

	reqOpts := map[interface{}]string{
		cfg.Algolia.ApplicationID:           "Algolia Application ID is required",
		cfg.Algolia.SearchAPIKey:            "Algolia Search API Key is required",
		cfg.BaseURL:                         "Base URL is required",
		cfg.GoogleWorkspace.DocsFolder:      "Google Workspace Docs Folder is required",
		cfg.GoogleWorkspace.DraftsFolder:    "Google Workspace Drafts Folder is required",
		cfg.GoogleWorkspace.ShortcutsFolder: "Google Workspace Shortcuts Folder is required",
	}
	for r, msg := range reqOpts {
		if r == "" {
			c.UI.Error(fmt.Sprintf("error initializing server: %s", msg))
			return 1
		}
	}

	// Initialize Algolia search client.
	algoSearch, err := algolia.NewSearchClient(cfg.Algolia)
	if err != nil {
		c.UI.Error(fmt.Sprintf("error initializing Algolia search client: %v", err))
		return 1
	}

	// Initialize Algolia write client.
	algoWrite, err := algolia.New(cfg.Algolia)
	if err != nil {
		c.UI.Error(fmt.Sprintf("error initializing Algolia write client: %v", err))
		return 1
	}

	// Initialize database.
	db, err := db.NewDB(*cfg.Postgres)
	if err != nil {
		c.UI.Error(fmt.Sprintf("error initializing database: %v", err))
		return 1
	}

	// Register document types.
	// for _, d := range cfg.DocumentTypes.DocumentType {
	// 	if err := models.RegisterDocumentType(*d, db); err != nil {
	// 		c.UI.Error(fmt.Sprintf("error registering document type: %v", err))
	// 		return 1
	// 	}
	// }
	if err := registerDocumentTypes(*cfg, db); err != nil {
		c.UI.Error(fmt.Sprintf("error registering document types: %v", err))
		return 1
	}

	//// Register products.
	//if err := registerProducts(cfg, algoWrite, db); err != nil {
	//	c.UI.Error(fmt.Sprintf("error registering products: %v", err))
	//	return 1
	//}

	// Register document types.
	// TODO: remove this and use the database for all document type lookups.
	// array of object object to store doctype objects fetched from algolia
	var objectArray []template = GetDocTypeArray(*cfg)
	docTypes := map[string]hcd.Doc{
		// "prd": &hcd.PRD{},
	}
	for i := 0; i < len(objectArray); i++ {
		doctype := objectArray[i].TemplateName
		doctype = strings.ToLower(doctype)
		docTypes[doctype] = &hcd.COMMONTEMPLATE{}
	}

	for name, dt := range docTypes {
		if err = doctypes.Register(name, dt); err != nil {
			c.UI.Error(fmt.Sprintf("error registering %q doc type: %v", name, err))
			return 1
		}
	}

	mux := http.NewServeMux()

	// Define handlers for authenticated endpoints.
	// TODO: stop passing around all these arguments to handlers and use a struct
	// with (functional) options.
	authenticatedEndpoints := []endpoint{
		{"/1/indexes/",
			algolia.AlgoliaProxyHandler(algoSearch, cfg.Algolia, c.Log)},
		{"/api/v1/approvals/",
			api.ApprovalHandler(cfg, c.Log, algoSearch, algoWrite, goog, db)},
		{"/api/v1/document-types", api.DocumentTypesHandler(*cfg, c.Log)},
		{"/api/v1/documents/",
			api.DocumentHandler(cfg, c.Log, algoSearch, algoWrite, goog, db)},
		{"/api/v1/drafts",
			api.DraftsHandler(cfg, c.Log, algoSearch, algoWrite, goog, db)},
		{"/api/v1/drafts/",
			api.DraftsDocumentHandler(cfg, c.Log, algoSearch, algoWrite, goog, db)},
		{"/api/v1/custom-template",
			api.TemplateHandler(cfg, c.Log, algoSearch, algoWrite, goog, db)},
		{"/api/v1/custom-template/",
			api.TemplateUpdateDeleteHandler(cfg, c.Log, algoSearch, algoWrite, goog, db)},
		{"/api/v1/make-admin", api.MakeUserAdminHandler(c.Log, db)},
		{"/api/v1/me", api.MeHandler(c.Log, goog, db)},
		{"/api/v1/me/recently-viewed-docs",
			api.MeRecentlyViewedDocsHandler(cfg, c.Log, db)},
		{"/api/v1/me/subscriptions",
			api.MeSubscriptionsHandler(cfg, c.Log, goog, db)},
		{"/api/v1/people", api.PeopleDataHandler(cfg, c.Log, goog)},
		{"/api/v1/products", api.ProductsHandler(cfg, algoSearch, algoWrite, db, c.Log)},
		{"/api/v1/teams", api.TeamsHandler(cfg, algoSearch, algoWrite, db, c.Log)},
		{"/api/v1/projects", api.ProjectsHandler(cfg, algoSearch, algoWrite, db, c.Log)},
		{"/api/v1/reviews/",
			api.ReviewHandler(cfg, c.Log, algoSearch, algoWrite, goog, db)},
		{"/api/v1/web/analytics", api.AnalyticsHandler(c.Log)},
	}

	// Define a slice of patterns that require admin access.
	adminRequiredPatterns := []string{
		"/api/v1/products",
		"/api/v1/teams",
		"/api/v1/custom-template",
		"/api/v1/custom-template/",
		"/api/v1/make-admin",
		// Add more patterns here if needed.
	}

	// Define handlers for unauthenticated endpoints.
	unauthenticatedEndpoints := []endpoint{
		{"/health", healthHandler()},
		{"/pub/", http.StripPrefix("/pub/", pub.Handler())},
	}

	// Web endpoints are conditionally authenticated based on if Okta is enabled.
	webEndpoints := []endpoint{
		{"/", web.Handler()},
		{"/api/v1/web/config", web.ConfigHandler(cfg, algoSearch, c.Log)},
		{"/l/", links.RedirectHandler(algoSearch, cfg.Algolia, c.Log)},
	}

	// If Okta is enabled, add the web endpoints for the single page app as
	// authenticated endpoints.
	if cfg.Okta != nil && !cfg.Okta.Disabled {
		authenticatedEndpoints = append(authenticatedEndpoints, webEndpoints...)
	} else {
		// If Okta is disabled, we need to add the web endpoints for the SPA as
		// unauthenticated endpoints so the application will load.
		unauthenticatedEndpoints = append(unauthenticatedEndpoints, webEndpoints...)
	}

	// Register handlers.
	for _, e := range authenticatedEndpoints {
		// Check if the current pattern requires admin access.
		if containsPattern(adminRequiredPatterns, e.pattern) {
			// If it does, apply the isAdminForProduct middleware after AuthenticateRequest.
			mux.Handle(
				e.pattern,
				auth.AuthenticateRequest(*cfg, goog, c.Log, isAdminForProduct(e.handler, db, c.Log)),
			)
		} else {
			// For other endpoints, use the existing authentication middleware.
			mux.Handle(
				e.pattern,
				auth.AuthenticateRequest(*cfg, goog, c.Log, e.handler),
			)
		}
	}

	for _, e := range unauthenticatedEndpoints {
		mux.Handle(e.pattern, e.handler)
	}

	server := &http.Server{
		Addr:    cfg.Server.Addr,
		Handler: mux,
	}
	go func() {
		c.Log.Info(fmt.Sprintf("listening on %s...", cfg.Server.Addr))

		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			c.Log.Error(fmt.Sprintf("error starting listener: %v", err))
			os.Exit(1)
		}
	}()

	go func() {
		sendDailyReviewReminderNotificationToAllRunsDaily(cfg, c.Log, algoSearch, algoWrite, goog, db)
		// reviewReminderNotifiactionToAllDocs(cfg, c.Log, algoSearch, algoWrite, goog, db)
	}()

	return c.WaitForInterrupt(c.ShutdownServer(server))
}

func sendDailyReviewReminderNotificationToAllRunsDaily(cfg *config.Config, l hclog.Logger, ar *algolia.Client, aw *algolia.Client, s *gw.Service, db *gorm.DB) {
	// Desired time to run the function (change this to your desired time)
	desiredTime := "10:00"

	// Calculate the duration until the desired time for the first execution
	calculateDurationUntilDesiredTime := func(desiredTime string) time.Duration {
		// Parse the desired time string to obtain the hour and minute values
		parseDesiredTime := func(desiredTime string) (hour, minute int, err error) {
			parsedTime, err := time.Parse("15:04", desiredTime)
			if err != nil {
				return 0, 0, err
			}
			return parsedTime.Hour(), parsedTime.Minute(), nil
		}

		hour, minute, err := parseDesiredTime(desiredTime)
		if err != nil {
			fmt.Println("Error parsing desired time:", err)
			return 0
		}

		// Get the current time
		now := time.Now()

		// Set the desired time for today
		desired := time.Date(now.Year(), now.Month(), now.Day(), hour, minute, 0, 0, now.Location())

		// If the desired time has already passed for today, set it for tomorrow
		if desired.Before(now) {
			desired = desired.Add(24 * time.Hour)
		}

		// Calculate the duration until the desired time
		durationUntilDesiredTime := desired.Sub(now)

		return durationUntilDesiredTime
	}

	// Wait until the first execution time
	time.Sleep(calculateDurationUntilDesiredTime(desiredTime))

	// Create a ticker with a 24-hour interval
	ticker := time.NewTicker(24 * time.Hour)

	// Start a goroutine to execute the function periodically
	for {

		// Call your function here (replace the print statement with your actual function call)
		fmt.Println("Function executed at", time.Now())
		reviewReminderNotifiactionToAllDocs(cfg, l, ar, aw, s, db)
		// Wait for the next tick
		<-ticker.C
	}
}

func reviewReminderNotifiactionToAllDocs(cfg *config.Config, l hclog.Logger, ar *algolia.Client, aw *algolia.Client, s *gw.Service, db *gorm.DB) {
	fmt.Println("sending reminder to all the reviewers who have not reviewed yet")
	params := algoliasearch.Map{
		"filters": "status:In-Review",
	}

	// Perform the search
	res, err := ar.Docs.Search("", params)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	// Access the search results
	fmt.Println("Total hits:", res.NbHits)
	fmt.Println("Hits:")
	// Access the search results and print reviewers
	for _, hit := range res.Hits {
		sendReviewReminderPerDoc(cfg, l, ar, aw, s, db, hit)
	}
}

func sendReviewReminderPerDoc(
	cfg *config.Config,
	l hclog.Logger,
	ar *algolia.Client,
	aw *algolia.Client,
	s *gw.Service,
	db *gorm.DB,
	hit map[string]interface{},
) {
	title, ok := hit["title"].(string)
	if !ok {
		fmt.Println("Title not found in hit")
	}
	fmt.Println("title :")
	fmt.Println(title)
	// Access the "reviewers" field from the hit
	reviewers, _ := hit["reviewers"].([]interface{})

	reviewedBy, ok := hit["reviewedBy"].([]interface{})
	if !ok {
		reviewedBy = []interface{}{}
	}

	// Create a new array to store the reviewersWhoHaveNotReviewed
	var reviewersWhoHaveNotReviewed []interface{}

	// Loop through array a and check each element against array b
	for _, reviewer := range reviewers {
		found := false

		for _, reviewedReviewer := range reviewedBy {
			if reviewer == reviewedReviewer {
				found = true
				break
			}
		}

		// If the element from a is not found in b, add it to the reviewersWhoHaveNotReviewed array
		if !found {
			reviewersWhoHaveNotReviewed = append(reviewersWhoHaveNotReviewed, reviewer)
		}
	}

	// Convert the []interface{} to an array of strings
	// and these are the reviewers who have not reviewed yet
	reviewersToEmail := make([]string, len(reviewersWhoHaveNotReviewed))
	for i, val := range reviewersWhoHaveNotReviewed {
		if str, ok := val.(string); ok {
			reviewersToEmail[i] = str
		} else {
			// Handle the case if the element is not a string
			fmt.Println("Element at index", i, "is not a string:", val)
		}
	}

	// Convert the "hit" map to JSON data (a []byte slice)
	jsonData, err := json.Marshal(hit)
	if err != nil {
		fmt.Println("Error marshaling JSON:", err)
		return
	}
	// Create an instance of the Document struct
	var docObj hcd.BaseDoc

	// Unmarshal the JSON data into the Document struct
	err = json.Unmarshal(jsonData, &docObj)
	if err != nil {
		fmt.Println("Error unmarshaling JSON:", err)
		return
	}
	docURL, err := getDocumentURL(cfg.BaseURL, docObj.GetObjectID())
	if err != nil {
		l.Error("error getting the base doc url!", "error", err)
	}
	// Send emails to reviewers.
	if cfg.Email != nil && cfg.Email.Enabled && len(reviewersToEmail) > 0 {
		var err error
		for _, reviewerEmail := range reviewersToEmail {
			err := email.SendReviewReminderEmail(
				email.ReviewRequestedEmailData{
					BaseURL:            cfg.BaseURL,
					DocumentOwner:      docObj.Owners[0],
					DocumentType:       docObj.GetDocType(),
					DocumentShortName:  docObj.GetDocNumber(),
					DocumentTitle:      docObj.GetTitle(),
					DocumentURL:        docURL,
					DocumentProd:       docObj.GetProduct(),
					DocumentTeam:       docObj.GetTeam(),
					DocumentOwnerEmail: docObj.GetOwners()[0],
				},
				[]string{reviewerEmail},
				cfg.Email.FromAddress,
				s,
			)
			if err != nil {
				l.Error("error sending reviewer email",
					"error", err,
				)
				return
			}
			l.Info("doc reviewer email sent")
		}
		if err != nil {
			fmt.Printf("Some error occured while sendind the message: %s", err)
		} else {
			fmt.Println("Succesfully! Delivered the message to all reviewers who have not reviewed")
		}

		// Also send the slack message tagginhg all the reviewers in the
		// dedicated channel
		// tagging all reviewers emails
		emails := make([]string, len(reviewersToEmail))
		for i, c := range reviewersToEmail {
			emails[i] = c
		}
		err = slackbot.SendSlackMessage_ReminderReviewer(slackbot.ReviewerRequestedSlackData{
			BaseURL:            cfg.BaseURL,
			DocumentOwner:      docObj.Owners[0],
			DocumentType:       docObj.GetDocType(),
			DocumentShortName:  docObj.GetDocNumber(),
			DocumentTitle:      docObj.GetTitle(),
			DocumentURL:        docURL,
			DocumentProd:       docObj.GetProduct(),
			DocumentTeam:       docObj.GetTeam(),
			DocumentOwnerEmail: docObj.GetOwners()[0],
		}, emails,
		)
		//handle error gracefully
		if err != nil {
			fmt.Printf("Some error occured while sendind the message: %s", err)
		} else {
			fmt.Println("Succesfully! Delivered the message to all reviewers who have not reviewed")
		}
	}
}

// getDocumentURL returns a Hermes document URL.
func getDocumentURL(baseURL, docID string) (string, error) {
	docURL, err := url.Parse(baseURL)
	if err != nil {
		return "", fmt.Errorf("error parsing base URL: %w", err)
	}

	docURL.Path = path.Join(docURL.Path, "document", docID)
	docURLString := docURL.String()
	docURLString = strings.TrimRight(docURLString, "/")

	return docURLString, nil
}

// healthHandler responds with the health of the service.
func healthHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
}

// ShutdownServer gracefully shuts down the HTTP server.
func (c *Command) ShutdownServer(s *http.Server) func() {
	return func() {
		c.Log.Debug("shutting down HTTP server...")
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if err := s.Shutdown(ctx); err != nil {
			c.Log.Error(fmt.Sprintf("error shutting down HTTP server: %v", err))
		}
	}
}

// registerDocumentTypes registers all products configured in the application
// config in the database.
// Define a variable to hold the retrieved object
type template struct {
	ObjectId     string `json:"objectId"`
	Description  string `json:"description,omitempty"`
	TemplateName string `json:"templateName"`
	DocId        string `json:"docId"`
}

func GetDocTypeArray(cfg config.Config) []template {
	// Initialize the Algolia client
	appID := cfg.Algolia.ApplicationID
	apiKey := cfg.Algolia.WriteAPIKey
	// Initialize the search client
	client := search.NewClient(appID, apiKey)

	// Specify the index name
	indexName := cfg.Algolia.TemplateIndexName
	// Create a search index instance
	index := client.InitIndex(indexName)
	// fetch all objectIds and append in
	var record template
	// use BrowseObjects to Get all records as an iterator
	it, err := index.BrowseObjects()
	if err != nil {
		fmt.Println("error browsing document types")
	}
	//objectArray contains array of template objects
	var objectArray []template
	// loop to traverse through all the template objects
	for {
		rec, err := it.Next(&record)
		if err != nil {
			break
		}
		jsonData, err := json.Marshal(rec)
		if err != nil {
			fmt.Println("Error converting to JSON:", err)
			break
		}
		// temporary object variable to store and append to object Array
		var object template
		error := json.Unmarshal([]byte(jsonData), &object)
		if error != nil {
			fmt.Println("Error converting JSON to object:", err)
			break
		}
		objectArray = append(objectArray, object)
	}
	// fmt.Printf("objectArray: %v\n", objectArray)
	return objectArray
}

func registerDocumentTypes(cfg config.Config, db *gorm.DB) error {
	var objectArray []template = GetDocTypeArray(cfg)
	for _, d := range objectArray {
		dt := models.DocumentType{
			Name:         d.TemplateName,
			Description:  d.Description,
			Checks:       nil,
			CustomFields: nil,
		}
		// Upsert document type.
		if err := dt.Upsert(db); err != nil {
			return fmt.Errorf("error upserting document type: %w", err)
		}
	}
	return nil
}

// registerProducts registers all products configured in the application config
// in the database and Algolia.
// TODO: products are currently needed in Algolia for legacy reasons - remove
// this when possible.
func registerProducts(
	cfg *config.Config, algo *algolia.Client, db *gorm.DB) error {

	productsObj := structs.Products{
		ObjectID: "products",
		Data:     make(map[string]structs.ProductData, 0),
	}

	for _, p := range cfg.Products.Product {
		// Upsert product in database.
		pm := models.Product{
			Name: p.Name,
		}
		if err := pm.Upsert(db); err != nil {
			return fmt.Errorf("error upserting product: %w", err)
		}

		// Add product to Algolia products object.
		productsObj.Data[p.Name] = structs.ProductData{}
	}

	// Save Algolia products object.
	res, err := algo.Internal.SaveObject(&productsObj)
	if err != nil {
		return fmt.Errorf("error saving Algolia products object: %w", err)
	}
	err = res.Wait()
	if err != nil {
		return fmt.Errorf("error saving Algolia products object: %w", err)
	}

	return nil
}

// isAdminForProduct is a middleware function that checks if the user is an admin
// for POST requests within the only admin access endpoint.
func isAdminForProduct(next http.Handler, db *gorm.DB, log hclog.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if the request method is POST.
		if r.Method != http.MethodGet {
			if r.Context().Value("userEmail") == nil {
				log.Error("userEmail is not set in the request context",
					"method", r.Method,
					"path", r.URL.Path)
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}
			userEmail := r.Context().Value("userEmail").(string)
			var user models.User = models.User{EmailAddress: userEmail}
			// Check if the user is an admin (you need to implement this logic).
			isAdmin, err := user.IsUserAdmin(db)
			if err != nil {
				http.Error(w, "Error Fetching User details from Database!", http.StatusInternalServerError)
			}
			if !isAdmin {
				http.Error(w, "Access denied: You must be an admin to perform this action.", http.StatusForbidden)
				return
			}
		}

		// If the request is not a Get request or the user is an admin for a POST request or any other crud,
		// call the next handler in the chain.
		next.ServeHTTP(w, r)
	})
}

// Helper function to check if a pattern is present in the slice.
func containsPattern(patterns []string, target string) bool {
	for _, p := range patterns {
		if p == target {
			return true
		}
	}
	return false
}
