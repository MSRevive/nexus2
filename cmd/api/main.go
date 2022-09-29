package main

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"time"

	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
	"github.com/msrevive/nexus2/internal/config"
	"github.com/msrevive/nexus2/internal/controller"
	"github.com/msrevive/nexus2/internal/middleware"
	"github.com/msrevive/nexus2/internal/service"
	"github.com/saintwish/auralog"
	"golang.org/x/crypto/acme"
	"golang.org/x/crypto/acme/autocert"
)

var spMsg string = `
    _   __                    ___
   / | / /__  _  ____  Nexus2|__ \
  /  |/ / _ \| |/_/ / / / ___/_/ /
 / /|  /  __/>  </ /_/ (__  ) __/
/_/ |_/\___/_/|_|\__,_/____/____/

Copyright © %d, Team MSRebirth

Version: %s
Website: https://msrebirth.net/
License: GPL-3.0 https://github.com/MSRevive/nexus2/blob/main/LICENSE %s
`

var Version = "canary"

func main() {
	fmt.Printf(spMsg, time.Now().Year(), Version, "\n\n")

	if err := run(os.Args); err != nil {
		fmt.Printf("critical error detected: %s", err)
		os.Exit(1)
	}
}

var (
	logCore *auralog.Logger // Logs for core/server
	logAPI  *auralog.Logger // Logs for endpoints/middleware
)

func initLoggers(filename string, dir string, level string, expire string) {
	ex, _ := time.ParseDuration(expire)
	flags := auralog.Ldate | auralog.Ltime | auralog.Lmicroseconds
	flagsWarn := auralog.Ldate | auralog.Ltime | auralog.Lmicroseconds
	flagsError := auralog.Ldate | auralog.Ltime | auralog.Lmicroseconds | auralog.Lshortfile
	flagsDebug := auralog.Ltime | auralog.Lmicroseconds | auralog.Lshortfile

	file := &auralog.RotateWriter{
		Dir:      dir,
		Filename: filename,
		ExTime:   ex,
		MaxSize:  5 * auralog.Megabyte,
	}

	logCore = auralog.New(auralog.Config{
		Output:    io.MultiWriter(os.Stdout, file),
		Prefix:    "[CORE] ",
		Level:     auralog.ToLogLevel(level),
		Flag:      flags,
		WarnFlag:  flagsWarn,
		ErrorFlag: flagsError,
		DebugFlag: flagsDebug,
	})

	logAPI = auralog.New(auralog.Config{
		Output:    io.MultiWriter(os.Stdout, file),
		Prefix:    "[API] ",
		Level:     auralog.ToLogLevel(level),
		Flag:      flags,
		WarnFlag:  flagsWarn,
		ErrorFlag: flagsError,
		DebugFlag: flagsDebug,
	})
}

func run(args []string) error {
	flgs := config.InitializeFlags(args)

	cfg, err := config.LoadConfig(flgs.ConfigFile, flgs.Debug)
	if err != nil {
		return err
	}

	if flgs.MigrateConfig {
		fmt.Println("Running config migration...")
		if err := cfg.Migrate(); err != nil {
			fmt.Printf("Config migration error: %s", err)
		}
		fmt.Println("Finished config migration, starting server...")
	}

	if flgs.Debug {
		fmt.Println("Running in Debug mode, do not use in production!")
	}

	fmt.Println("Initiating Loggers...")
	initLoggers("server.log", cfg.Log.Dir, cfg.Log.Level, cfg.Log.ExpireTime)

	//Max threads allowed.
	if cfg.Core.MaxThreads != 0 {
		runtime.GOMAXPROCS(cfg.Core.MaxThreads)
	}

	//Load JSON files.
	if cfg.ApiAuth.EnforceIP {
		logCore.Printf("Loading IP list from %s", cfg.ApiAuth.IPListFile)
		if err := cfg.LoadIPList(); err != nil {
			logCore.Warnln("Failed to load IP list.")
		}
	}

	if cfg.Verify.EnforceMap {
		logCore.Printf("Loading Map list from %s", cfg.Verify.MapListFile)
		if err := cfg.LoadMapList(); err != nil {
			logCore.Warnln("Failed to load Map list.")
		}
	}

	if cfg.Verify.EnforceBan {
		logCore.Printf("Loading Ban list from %s", cfg.Verify.BanListFile)
		if err := cfg.LoadBanList(); err != nil {
			logCore.Warnln("Failed to load Ban list.")
		}
	}

	logCore.Printf("Loading Admin list from %s", cfg.Verify.AdminListFile)
	if err := cfg.LoadAdminList(); err != nil {
		logCore.Warnln("Failed to load Admin list.")
	}

	logCore.Println("Connecting to database")
	db, err := config.InitializeDb(cfg.Core.DBString)
	if err != nil {
		return errors.New(fmt.Sprintf("failed to connect to database, %s", err))
	}
	defer db.Close()

	//create tables for new database file.
	service.New(context.Background(), db).CreateTables()

	//TODO: old database migration.

	//variables for web server
	//var srv *http.Server
	router := mux.NewRouter()
	srv := &http.Server{
		Handler:      router,
		Addr:         fmt.Sprintf("%s:%d", flgs.Address, flgs.Port),
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
		// DefaultTLSConfig sets sane defaults to use when configuring the internal
		// webserver to listen for public connections.
		//
		// @see https://blog.cloudflare.com/exposing-go-on-the-internet
		// credit to https://github.com/pterodactyl/wings/blob/develop/config/config.go
		TLSConfig: &tls.Config{
			NextProtos: []string{"h2", "http/1.1"},
			CipherSuites: []uint16{
				tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
				tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
				tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256,
				tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256,
			},
			PreferServerCipherSuites: true,
			MinVersion:               tls.VersionTLS12,
			MaxVersion:               tls.VersionTLS13,
			CurvePreferences:         []tls.CurveID{tls.X25519, tls.CurveP256},
		},
	}

	//middleware
	mw := middleware.New(logAPI)
	router.Use(mw.PanicRecovery)
	router.Use(mw.Log)
	if cfg.RateLimit.Enable {
		router.Use(mw.RateLimit)
	}

	//API Routes
	api := controller.New(router.PathPrefix(cfg.Core.RootPath).Subrouter(), db, logAPI)
	api.R.HandleFunc("/ping", mw.Lv2Auth(api.GetPing)).Methods(http.MethodGet)
	api.R.HandleFunc("/map/{name}/{hash}", mw.Lv1Auth(api.GetMapVerify)).Methods(http.MethodGet)
	api.R.HandleFunc("/ban/{steamid:[0-9]+}", mw.Lv1Auth(api.GetBanVerify)).Methods(http.MethodGet)
	api.R.HandleFunc("/sc/{hash}", mw.Lv1Auth(api.GetSCVerify)).Methods(http.MethodGet)

	//Character Routes
	capi := controller.New(router.PathPrefix(cfg.Core.RootPath+"/character").Subrouter(), db, logAPI)
	capi.R.HandleFunc("/", mw.Lv1Auth(capi.GetAllCharacters)).Methods(http.MethodGet)
	capi.R.HandleFunc("/id/{uid}", mw.Lv1Auth(capi.GetCharacterByID)).Methods(http.MethodGet)
	capi.R.HandleFunc("/{steamid:[0-9]+}", mw.Lv1Auth(capi.GetCharacters)).Methods(http.MethodGet)
	capi.R.HandleFunc("/{steamid:[0-9]+}/{slot:[0-9]}", mw.Lv1Auth(capi.GetCharacter)).Methods(http.MethodGet)
	capi.R.HandleFunc("/export/{steamid:[0-9]+}/{slot:[0-9]}", mw.Lv1Auth(capi.ExportCharacter)).Methods(http.MethodGet)
	capi.R.HandleFunc("/", mw.Lv2Auth(capi.PostCharacter)).Methods(http.MethodPost)
	capi.R.HandleFunc("/{uid}", mw.Lv2Auth(capi.PutCharacter)).Methods(http.MethodPut)
	capi.R.HandleFunc("/{uid}", mw.Lv2Auth(capi.DeleteCharacter)).Methods(http.MethodDelete)
	capi.R.HandleFunc("/{uid}/restore", mw.Lv1Auth(capi.RestoreCharacter)).Methods(http.MethodPatch)
	capi.R.HandleFunc("/{steamid:[0-9]+}/{slot:[0-9]}/versions", mw.Lv1Auth(capi.CharacterVersions)).Methods(http.MethodGet)
	capi.R.HandleFunc("/{steamid:[0-9]+}/{slot:[0-9]}/rollback/{version:[0-9]+}", mw.Lv1Auth(capi.RollbackCharacter)).Methods(http.MethodPatch)

	if cfg.Cert.Enable {
		cm := autocert.Manager{
			Prompt:     autocert.AcceptTOS,
			HostPolicy: autocert.HostWhitelist(cfg.Cert.Domain),
			Cache:      autocert.DirCache("./runtime/certs"),
		}

		srv.TLSConfig = &tls.Config{
			GetCertificate: cm.GetCertificate,
			NextProtos:     append(srv.TLSConfig.NextProtos, acme.ALPNProto), // enable tls-alpn ACME challenges
		}

		go func() {
			if err := http.ListenAndServe(":http", cm.HTTPHandler(nil)); err != nil {
				fmt.Printf("failed to serve autocert server: %v\n", err)
			}
		}()

		logCore.Printf("Listening on: %s TLS", srv.Addr)
		if err := srv.ListenAndServeTLS("", ""); err != nil {
			return errors.New(fmt.Sprintf("failed to serve over HTTPS: %v", err))
		}
	} else {
		logCore.Printf("Listening on: %s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil {
			return errors.New(fmt.Sprintf("failed to serve over HTTP: %v", err))
		}
	}

	return nil
}

// tmpMigration will convert all current characters to the new schema
// If any failure is detected, the current database will not be affected
/*
func tmpMigration(dbstring string) {
	ctx := context.Background()
	dbFileName := "./runtime/chars.db"
	oldDbFileName := "./runtime/old_chars.db"
	dbBakFileName := dbFileName + ".bak"
	dbBakConnStr := "file:" + oldDbFileName + "?cache=shared&mode=rwc&_fk=1"

	//if file doesn't exists then no migration is needed.
	if _, ferr := os.Stat(dbFileName); errors.Is(ferr, os.ErrNotExist) {
		client, err := ent.Open("sqlite3", dbstring)
		if err != nil {
			log.Log.Fatalf("failed to open connection to sqlite3: %v", err)
		}
		if err := client.Schema.Create(ctx, schema.WithAtlas(true)); err != nil {
			log.Log.Fatalf("failed to create schema resources: %v", err)
		}
		system.Client = client
	} else {
		err := func() error {
			/////////////////////////////////////////////////
			/////////// Check if migration is required
			// 1. Check if the 'players' table exists
			sqlClient, err := entd.Open("sqlite3", dbFileName)
			if err != nil {
				log.Log.Errorf("failed to open connection to sqlite3: %v", err)
				return err
			}

			rows, err := sqlClient.DB().Query("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='players';")
			if err != nil {
				log.Log.Errorf("failed to get table: %v", err)
				return err
			}

			var count int64
			for rows.Next() {
				rows.Scan(&count)
			}
			if count > 0 {
				log.Log.Println("DB already migrated")

				// Set the connection as normal
				client, err := ent.Open("sqlite3", dbstring)
				if err != nil {
					log.Log.Errorf("failed to open connection to sqlite3: %v", err)
					return err
				}
				if err := client.Schema.Create(ctx, schema.WithAtlas(true)); err != nil {
					log.Log.Errorf("failed to create schema resources: %v", err)
					return err
				}
				system.Client = client

				return nil
			}
			sqlClient.Close()

			log.Log.Println("DB not migrated")

			/////////////////////////////////////////////////
			/////////// Prepare DB files
			// 1. Copy current DB file (prevents irreversible changes)
			// 2. Rename current DB file (prevents conflicts with new DB)
			currentDbFile, err := os.Open(dbFileName)
			if err != nil {
				log.Log.Errorf("failed to open current DB file: %v", err)
				return err
			}

			backupDbFile, err := os.Create(dbBakFileName)
			if err != nil {
				log.Log.Errorf("failed to create backup DB file: %v", err)
				return err
			}

			_, err = io.Copy(backupDbFile, currentDbFile)
			if err != nil {
				log.Log.Errorf("failed to backup DB file: %v", err)
				return err
			}
			currentDbFile.Close()
			backupDbFile.Close()

			if err := os.Rename(dbFileName, oldDbFileName); err != nil {
				log.Log.Errorf("failed to rename original DB file: %v", err)
				return err
			}
			log.Log.Println("Backed up DB file")

			/////////////////////////////////////////////////
			/////////// Prepare DB schema
			// 1. Rename 'characters' table (allows the use of ORM)
			sqlClient, err = entd.Open("sqlite3", dbBakConnStr)
			if err != nil {
				log.Log.Errorf("failed to open connection to old sqlite3: %v", err)
				return err
			}
			log.Log.Println("Connected to old DB")

			_, err = sqlClient.DB().Exec("ALTER TABLE characters RENAME TO old_characters;")
			if err != nil {
				log.Log.Errorf("failed to rename old table: %v", err)
				return err
			}
			sqlClient.Close()
			log.Log.Println("Renamed table")

			/////////////////////////////////////////////////
			/////////// Collect old data
			// 1. Load all characters into memory using ORM
			client, err := ent.Open("sqlite3", dbBakConnStr)
			if err != nil {
				log.Log.Errorf("failed to open connection to old sqlite3: %v", err)
				return err
			}
			log.Log.Println("Connected to old db")

			oldCharacters, err := client.DeprecatedCharacter.Query().All(ctx)
			if err != nil {
				log.Log.Errorf("failed to get all old characters: %v", err)
				return err
			}
			client.Close()
			log.Log.Println("Loaded all characters")

			/////////////////////////////////////////////////
			/////////// Migrate characters to new DB
			// 1. Connect to new DB (creating file in the process)
			// 2. Loop through old characters
			// 3. Convert old character to new Player + Versioned Character
			// 4. Log any conversion errors
			client, err = ent.Open("sqlite3", dbstring)
			if err != nil {
				log.Log.Errorf("failed to open connection to sqlite3: %v", err)
				return err
			}
			if err := client.Schema.Create(ctx, schema.WithAtlas(true)); err != nil {
				log.Log.Errorf("failed to create schema resources: %v", err)
				return err
			}
			system.Client = client
			log.Log.Println("connected to new db")

			errLog, err := os.Create("./runtime/logs/migration_errors.log")
			if err != nil {
				log.Log.Errorf("failed to create migration error log: %v", err)
				return err
			}

			cLen := len(oldCharacters)
			var failed int
			for i, c := range oldCharacters {
				fmt.Printf("\033[1A\033[Kmigrating %d of %d; steamId='%s'\n", i+1, cLen, c.Steamid)
				player, err := client.Player.Query().Where(player.Steamid(c.Steamid)).Only(ctx)
				if ent.IsNotFound(err) {
					player, err = client.Player.Create().
						SetSteamid(c.Steamid).
						Save(ctx)
					if err != nil {
						log.Log.Errorf("failed to save Player: %v", err)
						return err
					}
				}
				_, err = client.Character.Create().
					SetID(c.ID).
					SetPlayer(player).
					SetVersion(1).
					SetSlot(c.Slot).
					SetSize(c.Size).
					SetData(c.Data).
					Save(ctx)
				if err != nil {
					failed++
					errLog.WriteString(fmt.Sprintf("failed to save character! %v\n\t%+v\n", err, c))
				}
			}
			errLog.Close()
			if failed > 0 {
				log.Log.Printf("completed migration with %d errors!\n", failed)
			}
			log.Log.Println("Migrated all characters")

			return nil
		}()

		if err != nil {
			// Error detected, revert db changes if possible
			if _, err := os.Stat(dbBakFileName); err == nil {
				os.Remove(dbFileName)
				os.Remove(oldDbFileName)
				os.Rename(dbBakFileName, dbFileName)
			}
			log.Log.Fatalln("failed to migrate DB")
		}

		// happy path cleanup, we don't want to delete old one incase we need to revert suddenly.
		os.Remove(dbBakFileName)
	}
}*/
