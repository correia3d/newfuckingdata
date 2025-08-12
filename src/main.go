package main

import (
	"log"
	"strconv"
	"sync/atomic"

	"github.com/TibiaData/tibiadata-api-go/src/cache"
	"github.com/TibiaData/tibiadata-api-go/src/validation"
)

var (
	// application readyz endpoint value for k8s
	isReady atomic.Bool

	// TibiaDataDefaultVoc - default vocation when not specified in request
	TibiaDataDefaultVoc string = "all"

	// TibiaData app flags for running
	TibiaDataAPIversion      int = 4
	TibiaDataDebug           bool
	TibiaDataRestrictionMode bool

	// TibiaData app settings
	TibiaDataAPIDetails APIDetails // containing information from build
	TibiaDataHost       string     // set through env TIBIADATA_HOST
	TibiaDataProtocol   = "https"  // can be overridden by env TIBIADATA_PROTOCOL

	// TibiaData app details set to release/build on GitHub
	TibiaDataBuildRelease = "unknown"     // will be set by GitHub Actions (to release number)
	TibiaDataBuildBuilder = "manual"      // will be set by GitHub Actions
	TibiaDataBuildCommit  = "-"           // will be set by GitHub Actions (to git commit)
	TibiaDataBuildEdition = "open-source" //
)

// @title           TibiaData API
// @version         edge
// @description     This is the API documentation for the TibiaData API.
// @description     The documentation contains version 3 and above.
// @termsOfService  https://tibiadata.com/terms/

// @contact.name   TibiaData
// @contact.url    https://tibiadata.com/contact/
// @contact.email  tobias@tibiadata.com

// @license.name  MIT
// @license.url   https://github.com/TibiaData/tibiadata-api-go/blob/main/LICENSE

// @schemes   http
// @host      localhost:8080
// @BasePath  /

func init() {
	// logging init of TibiaData
	log.Printf("[info] TibiaData API initializing..")

	// Logging build information
	log.Printf("[info] TibiaData API release: %s", TibiaDataBuildRelease)
	log.Printf("[info] TibiaData API build: %s", TibiaDataBuildBuilder)
	log.Printf("[info] TibiaData API commit: %s", TibiaDataBuildCommit)
	log.Printf("[info] TibiaData API edition: %s", TibiaDataBuildEdition)

	TibiaDataAPIDetails = APIDetails{
		Version: TibiaDataAPIversion,
		Release: TibiaDataBuildRelease,
		Commit:  TibiaDataBuildCommit,
	}

	// Setting tibiadata-application to log much less if DEBUG_MODE is false (default is false)
	if getEnvAsBool("DEBUG_MODE", false) {
		// Setting debug to true for more logging
		TibiaDataDebug = true
	}
	log.Printf("[info] TibiaData API debug-mode: %t", TibiaDataDebug)

	// Running the TibiaDataInitializer function
	TibiaDataInitializer()

	// Generating TibiaDataUserAgent with TibiaDataUserAgentGenerator function
	TibiaDataUserAgent = TibiaDataUserAgentGenerator(TibiaDataAPIversion)

	if TibiaDataDebug {
		// Logging user-agent string
		log.Printf("[debug] TibiaData API User-Agent: %s", TibiaDataUserAgent)
	}

	// Initiate the validator
	err := validation.Initiate(TibiaDataUserAgent)
	if err != nil {
		panic(err)
	}

	// Initialize Redis cache
	initializeCache()
}

func main() {
	// logging start of TibiaData
	log.Printf("[info] TibiaData API starting..")

	// Starting the webserver
	runWebServer()
}

// initializeCache configura o Redis cache baseado nas variáveis de ambiente
func initializeCache() {
	// Tentar configurar via REDIS_URL primeiro
	if redisURL := getEnv("REDIS_URL", ""); redisURL != "" {
		if err := cache.SetupWithURL(redisURL); err != nil {
			log.Printf("[warning] Failed to setup Redis cache with URL: %v", err)
			return
		}
		log.Printf("[info] Redis cache initialized successfully with URL")
		return
	}

	// Fallback para configuração manual
	redisAddr := getEnv("REDIS_ADDR", "")
	if redisAddr != "" {
		redisPassword := getEnv("REDIS_PASSWORD", "")
		redisDB := 0 // Usar DB 0 por padrão
		if dbStr := getEnv("REDIS_DB", ""); dbStr != "" {
			if db, err := strconv.Atoi(dbStr); err == nil {
				redisDB = db
			}
		}

		cache.Setup(redisAddr, redisPassword, redisDB)
		log.Printf("[info] Redis cache initialized successfully at %s", redisAddr)
		return
	}

	log.Printf("[info] Redis cache not configured - running without cache")
}

// TibiaDataInitializer set the background for the webserver
func TibiaDataInitializer() {
	// Setting TibiaDataBuildEdition
	if isEnvExist("TIBIADATA_EDITION") {
		TibiaDataBuildEdition = getEnv("TIBIADATA_EDITION", "open-source")
	}

	// Adding information of host
	if isEnvExist("TIBIADATA_HOST") {
		TibiaDataHost = getEnv("TIBIADATA_HOST", "")
		log.Println("[info] TibiaData API hostname: " + TibiaDataHost)
	}
	if isEnvExist("TIBIADATA_PROTOCOL") {
		TibiaDataProtocol = getEnv("TIBIADATA_PROTOCOL", "https")
		log.Println("[info] TibiaData API protocol: " + TibiaDataProtocol)
	}

	// Setting TibiaDataProxyDomain
	if isEnvExist("TIBIADATA_PROXY") {

		TibiaDataProxyProtocol := getEnv("TIBIADATA_PROXY_PROTOCOL", "https")
		switch TibiaDataProxyProtocol {
		case "http":
			TibiaDataProxyProtocol = "http"
		}

		TibiaDataProxyDomain = TibiaDataProxyProtocol + "://" + getEnv("TIBIADATA_PROXY", "www.tibia.com") + "/"
		log.Printf("[info] TibiaData API proxy: %s", TibiaDataProxyDomain)
	}

	// Run some functions that are empty but required for documentation to be done
	_ = tibiaNewslistArchive()
	_ = tibiaNewslistArchiveDays()
	_ = tibiaNewslistLatest()

	// Run functions for v3 documentation to work
	_ = tibiaBoostableBossesV3()
	_ = tibiaCharactersCharacterV3()
	_ = tibiaCreaturesOverviewV3()
	_ = tibiaCreaturesCreatureV3()
	_ = tibiaFansitesV3()
	_ = tibiaGuildsGuildV3()
	_ = tibiaGuildsOverviewV3()
	_ = tibiaHighscoresV3()
	_ = tibiaHousesHouseV3()
	_ = tibiaHousesOverviewV3()
	_ = tibiaKillstatisticsV3()
	_ = tibiaNewslistArchiveV3()
	_ = tibiaNewslistArchiveDaysV3()
	_ = tibiaNewslistLatestV3()
	_ = tibiaNewslistV3()
	_ = tibiaNewsV3()
	_ = tibiaSpellsOverviewV3()
	_ = tibiaSpellsSpellV3()
	_ = tibiaWorldsOverviewV3()
	_ = tibiaWorldsWorldV3()
}
