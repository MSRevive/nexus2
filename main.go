package main

import (
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"time"

	entd "entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/schema"
	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
	"github.com/msrevive/nexus2/controller"
	"github.com/msrevive/nexus2/ent"
	"github.com/msrevive/nexus2/ent/player"
	"github.com/msrevive/nexus2/log"
	"github.com/msrevive/nexus2/middleware"
	"github.com/msrevive/nexus2/system"
	"golang.org/x/crypto/acme"
	"golang.org/x/crypto/acme/autocert"
)

func initPrint() {
	fmt.Printf(`
    _   __                    ___
   / | / /__  _  ____  Nexus2|__ \
  /  |/ / _ \| |/_/ / / / ___/_/ /
 / /|  /  __/>  </ /_/ (__  ) __/
/_/ |_/\___/_/|_|\__,_/____/____/

Copyright © %d, Team MSRebirth

Version: %s
Website: https://msrebirth.net/
License: GPL-3.0 https://github.com/MSRevive/nexus2/blob/main/LICENSE %s`, time.Now().Year(), system.Version, "\n\n")
}

func main() {
	var cfile string
	var migrateConfig bool //migrates ini/toml config to yaml
	flag.StringVar(&cfile, "cfile", "./runtime/config.yaml", "Location of via config file")
	flag.BoolVar(&system.Dbg, "dbg", false, "Run with debug mode.")
	flag.BoolVar(&migrateConfig, "m", false, "Migrate the ini/toml config to YAML")
	flag.Parse()

	if err := system.LoadConfig(cfile); err != nil {
		panic(err)
	}

	if migrateConfig {
		fmt.Println("Running migration...")
		if err := system.MigrateConfig(); err != nil {
			fmt.Printf("Migration error: %s", err)
		}
		fmt.Println("Finished migration, starting server...")
	}

	//initial print
	initPrint()

	//Initiate logging
	log.InitLogging("server.log", system.Config.Log.Dir, system.Config.Log.Level, system.Config.Log.ExpireTime)

	if system.Dbg {
		log.Log.Warnln("Running in Debug mode, do not use in production!")
	}

	//Max threads allowed.
	if system.Config.Core.MaxThreads != 0 {
		runtime.GOMAXPROCS(system.Config.Core.MaxThreads)
	}

	//Load json files.
	if system.Config.ApiAuth.EnforceIP {
		log.Log.Printf("Loading IP list from %s", system.Config.ApiAuth.IPListFile)
		if err := system.LoadIPList(system.Config.ApiAuth.IPListFile); err != nil {
			log.Log.Warnln("Failed to load IP list.")
		}
	}

	if system.Config.Verify.EnforceMap {
		log.Log.Printf("Loading Map list from %s", system.Config.Verify.MapListFile)
		if err := system.LoadMapList(system.Config.Verify.MapListFile); err != nil {
			log.Log.Warnln("Failed to load Map list.")
		}
	}

	if system.Config.Verify.EnforceBan {
		log.Log.Printf("Loading Ban list from %s", system.Config.Verify.BanListFile)
		if err := system.LoadBanList(system.Config.Verify.BanListFile); err != nil {
			log.Log.Warnln("Failed to load Ban list.")
		}
	}

	log.Log.Printf("Loading Admin list from %s", system.Config.Verify.AdminListFile)
	if err := system.LoadAdminList(system.Config.Verify.AdminListFile); err != nil {
		log.Log.Warnln("Failed to load Admin list.")
	}

	//Connect database.
	log.Log.Println("Connecting to database")
	// client, err = ent.Open("sqlite3", system.Config.Core.DBString)
	// if err != nil {
	// 	log.Log.Fatalf("failed to open connection to sqlite3: %v", err)
	// }
	// if err := client.Schema.Create(ctx, schema.WithAtlas(true)); err != nil {
	// 	log.Log.Fatalf("failed to create schema resources: %v", err)
	// }
	// system.Client = client
	// defer system.Client.Close()
	tmpMigration()
	defer system.Client.Close()

	//variables for web server
	var srv *http.Server
	router := mux.NewRouter()
	srv = &http.Server{
		Handler:      router,
		Addr:         system.Config.Core.Address + ":" + strconv.Itoa(system.Config.Core.Port),
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
	router.Use(middleware.PanicRecovery)
	router.Use(middleware.Log)
	if system.Config.RateLimit.Enable {
		router.Use(middleware.RateLimit)
	}

	//api routes
	apic := controller.New(router.PathPrefix(system.Config.Core.RootPath).Subrouter())
	apic.R.HandleFunc("/ping", middleware.Lv2Auth(apic.GetPing)).Methods(http.MethodGet)
	apic.R.HandleFunc("/map/{name}/{hash}", middleware.Lv1Auth(apic.GetMapVerify)).Methods(http.MethodGet)
	apic.R.HandleFunc("/ban/{steamid:[0-9]+}", middleware.Lv1Auth(apic.GetBanVerify)).Methods(http.MethodGet)
	apic.R.HandleFunc("/sc/{hash}", middleware.Lv1Auth(apic.GetSCVerify)).Methods(http.MethodGet)

	//character routes
	charc := controller.New(router.PathPrefix(system.Config.Core.RootPath + "/character").Subrouter())
	charc.R.HandleFunc("/", middleware.Lv1Auth(charc.GetAllCharacters)).Methods(http.MethodGet)
	charc.R.HandleFunc("/id/{uid}", middleware.Lv1Auth(charc.GetCharacterByID)).Methods(http.MethodGet)
	charc.R.HandleFunc("/{steamid:[0-9]+}", middleware.Lv1Auth(charc.GetCharacters)).Methods(http.MethodGet)
	charc.R.HandleFunc("/{steamid:[0-9]+}/{slot:[0-9]}", middleware.Lv1Auth(charc.GetCharacter)).Methods(http.MethodGet)
	charc.R.HandleFunc("/export/{steamid:[0-9]+}/{slot:[0-9]}", middleware.Lv1Auth(charc.ExportCharacter)).Methods(http.MethodGet)
	charc.R.HandleFunc("/", middleware.Lv2Auth(charc.PostCharacter)).Methods(http.MethodPost)
	charc.R.HandleFunc("/{uid}", middleware.Lv2Auth(charc.PutCharacter)).Methods(http.MethodPut)
	charc.R.HandleFunc("/{uid}", middleware.Lv2Auth(charc.DeleteCharacter)).Methods(http.MethodDelete)
	charc.R.HandleFunc("/{uid}/restore", middleware.Lv2Auth(charc.RestoreCharacter)).Methods(http.MethodPatch)
	charc.R.HandleFunc("/{steamid:[0-9]+}/{slot:[0-9]}/versions", middleware.Lv2Auth(charc.CharacterVersions)).Methods(http.MethodGet)
	charc.R.HandleFunc("/{steamid:[0-9]+}/{slot:[0-9]}/rollback/{version:[0-9]+}", middleware.Lv2Auth(charc.RollbackCharacter)).Methods(http.MethodPatch)

	if system.Config.Cert.Enable {
		cm := autocert.Manager{
			Prompt:     autocert.AcceptTOS,
			HostPolicy: autocert.HostWhitelist(system.Config.Cert.Domain),
			Cache:      autocert.DirCache("./runtime/certs"),
		}

		srv.TLSConfig = &tls.Config{
			GetCertificate: cm.GetCertificate,
			NextProtos:     append(srv.TLSConfig.NextProtos, acme.ALPNProto), // enable tls-alpn ACME challenges
		}

		go func() {
			if err := http.ListenAndServe(":http", cm.HTTPHandler(nil)); err != nil {
				log.Log.Fatalf("failed to serve autocert server: %v", err)
			}
		}()

		log.Log.Printf("Listening on: %s TLS", srv.Addr)
		if err := srv.ListenAndServeTLS("", ""); err != nil {
			log.Log.Fatalf("failed to serve over HTTPS: %v", err)
		}
	} else {
		log.Log.Printf("Listening on: %s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil {
			log.Log.Fatalf("failed to serve over HTTP: %v", err)
		}
	}
}

// tmpMigration will convert all current characters to the new schema
// If any failure is detected, the current database will not be affected
func tmpMigration() {
	ctx := context.Background()
	dbFileName := "./runtime/chars.db"
	oldDbFileName := "./runtime/old_chars.db"
	dbBakFileName := dbFileName + ".bak"
	dbBakConnStr := "file:" + oldDbFileName + "?cache=shared&mode=rwc&_fk=1"

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
			client, err := ent.Open("sqlite3", system.Config.Core.DBString)
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
		client, err = ent.Open("sqlite3", system.Config.Core.DBString)
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

	// happy path cleanup
	os.Remove(dbBakFileName)
	os.Remove(oldDbFileName)
}
