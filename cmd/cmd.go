package cmd

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
	"errors"
	"syscall"
	"os/signal"

	"github.com/msrevive/nexus2/cmd/app"
	"github.com/msrevive/nexus2/internal/controller"
	"github.com/msrevive/nexus2/internal/middleware"
	"github.com/msrevive/nexus2/ent"
	"github.com/msrevive/nexus2/ent/player"

	"github.com/saintwish/auralog"
	entd "entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/schema"
	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/acme"
	"golang.org/x/crypto/acme/autocert"
)

var (
	logCore *auralog.Logger // Logs for core/server
	logAPI *auralog.Logger // Logs for endpoints/middleware
)

type flags struct {
	address string
	port int
	configFile string
	migrateConfig bool
	debug bool
}

func doFlags(args []string) *flags {
	flgs := &flags{}

	flagSet := flag.NewFlagSet(args[0], flag.ContinueOnError)
	flagSet.StringVar(&flgs.address, "addr", "127.0.0.1", "The address of the server.")
	flagSet.IntVar(&flgs.port, "port", 1337, "The port this should run on.")
	flagSet.StringVar(&flgs.configFile, "cfile", "./runtime/config.yaml", "Location of via config file")
	flagSet.BoolVar(&flgs.debug, "d", false, "Run with debug mode.")
	flagSet.BoolVar(&flgs.migrateConfig, "m", false, "Migrate the ini/toml config to YAML")
	flagSet.Parse(args[1:])

	return flgs
}

func initLoggers(filename string, dir string, level string, expire string) {
	ex, _ := time.ParseDuration(expire)
	flags := auralog.Ldate | auralog.Ltime | auralog.Lmicroseconds
	flagsWarn := auralog.Ldate | auralog.Ltime | auralog.Lmicroseconds
	flagsError := auralog.Ldate | auralog.Ltime | auralog.Lmicroseconds | auralog.Lshortfile
	flagsDebug := auralog.Ltime | auralog.Lmicroseconds | auralog.Lshortfile

	file := &auralog.RotateWriter{
		Dir: dir,
		Filename: filename,
		ExTime: ex,
		MaxSize: 5 * auralog.Megabyte,
	}

	logCore = auralog.New(auralog.Config{
		Output: io.MultiWriter(os.Stdout, file),
		Prefix: "[CORE] ",
		Level: auralog.ToLogLevel(level),
		Flag: flags,
		WarnFlag: flagsWarn,
		ErrorFlag: flagsError,
		DebugFlag: flagsDebug,
	})

	logAPI = auralog.New(auralog.Config{
		Output: io.MultiWriter(os.Stdout, file),
		Prefix: "[API] ",
		Level: auralog.ToLogLevel(level),
		Flag: flags,
		WarnFlag: flagsWarn,
		ErrorFlag: flagsError,
		DebugFlag: flagsDebug,
	})
}

func tmpMigration(apps *app.App) {
	ctx := context.Background()
	dbstring := apps.Config.Core.DBString
	dbFileName := "./runtime/chars.db"
	oldDbFileName := "./runtime/old_chars.db"
	dbBakFileName := dbFileName + ".bak"
	dbBakConnStr := "file:" + oldDbFileName + "?cache=shared&mode=rwc&_fk=1"
	
	//if file doesn't exists then no migration is needed.
	if _, ferr := os.Stat(dbFileName); errors.Is(ferr, os.ErrNotExist) {
		client, err := ent.Open("sqlite3", dbstring)
		if err != nil {
			logCore.Fatalf("failed to open connection to sqlite3: %v", err)
		}
		if err := client.Schema.Create(ctx, schema.WithAtlas(true)); err != nil {
			logCore.Fatalf("failed to create schema resources: %v", err)
		}
		apps.Client = client
	} else {
		err := func() error {
			/////////////////////////////////////////////////
			/////////// Check if migration is required
			// 1. Check if the 'players' table exists
			sqlClient, err := entd.Open("sqlite3", dbFileName)
			if err != nil {
				logCore.Errorf("failed to open connection to sqlite3: %v", err)
				return err
			}
	
			rows, err := sqlClient.DB().Query("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='players';")
			if err != nil {
				logCore.Errorf("failed to get table: %v", err)
				return err
			}
	
			var count int64
			for rows.Next() {
				rows.Scan(&count)
			}
			if count > 0 {
				logCore.Println("DB already migrated")
	
				// Set the connection as normal
				client, err := ent.Open("sqlite3", dbstring)
				if err != nil {
					logCore.Errorf("failed to open connection to sqlite3: %v", err)
					return err
				}
				if err := client.Schema.Create(ctx, schema.WithAtlas(true)); err != nil {
					logCore.Errorf("failed to create schema resources: %v", err)
					return err
				}
				apps.Client = client
	
				return nil
			}
			sqlClient.Close()
	
			logCore.Println("DB not migrated")
	
			/////////////////////////////////////////////////
			/////////// Prepare DB files
			// 1. Copy current DB file (prevents irreversible changes)
			// 2. Rename current DB file (prevents conflicts with new DB)
			currentDbFile, err := os.Open(dbFileName)
			if err != nil {
				logCore.Errorf("failed to open current DB file: %v", err)
				return err
			}
	
			backupDbFile, err := os.Create(dbBakFileName)
			if err != nil {
				logCore.Errorf("failed to create backup DB file: %v", err)
				return err
			}
	
			_, err = io.Copy(backupDbFile, currentDbFile)
			if err != nil {
				logCore.Errorf("failed to backup DB file: %v", err)
				return err
			}
			currentDbFile.Close()
			backupDbFile.Close()
	
			if err := os.Rename(dbFileName, oldDbFileName); err != nil {
				logCore.Errorf("failed to rename original DB file: %v", err)
				return err
			}
			logCore.Println("Backed up DB file")
	
			/////////////////////////////////////////////////
			/////////// Prepare DB schema
			// 1. Rename 'characters' table (allows the use of ORM)
			sqlClient, err = entd.Open("sqlite3", dbBakConnStr)
			if err != nil {
				logCore.Errorf("failed to open connection to old sqlite3: %v", err)
				return err
			}
			logCore.Println("Connected to old DB")
	
			_, err = sqlClient.DB().Exec("ALTER TABLE characters RENAME TO old_characters;")
			if err != nil {
				logCore.Errorf("failed to rename old table: %v", err)
				return err
			}
			sqlClient.Close()
			logCore.Println("Renamed table")
	
			/////////////////////////////////////////////////
			/////////// Collect old data
			// 1. Load all characters into memory using ORM
			client, err := ent.Open("sqlite3", dbBakConnStr)
			if err != nil {
				logCore.Errorf("failed to open connection to old sqlite3: %v", err)
				return err
			}
			logCore.Println("Connected to old db")
	
			oldCharacters, err := client.DeprecatedCharacter.Query().All(ctx)
			if err != nil {
				logCore.Errorf("failed to get all old characters: %v", err)
				return err
			}
			client.Close()
			logCore.Println("Loaded all characters")
	
			/////////////////////////////////////////////////
			/////////// Migrate characters to new DB
			// 1. Connect to new DB (creating file in the process)
			// 2. Loop through old characters
			// 3. Convert old character to new Player + Versioned Character
			// 4. Log any conversion errors
			client, err = ent.Open("sqlite3", dbstring)
			if err != nil {
				logCore.Errorf("failed to open connection to sqlite3: %v", err)
				return err
			}
			if err := client.Schema.Create(ctx, schema.WithAtlas(true)); err != nil {
				logCore.Errorf("failed to create schema resources: %v", err)
				return err
			}
			apps.Client = client
			logCore.Println("connected to new db")
	
			errLog, err := os.Create("./runtime/logs/migration_errors.log")
			if err != nil {
				logCore.Errorf("failed to create migration error log: %v", err)
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
						logCore.Errorf("failed to save Player: %v", err)
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
				logCore.Printf("completed migration with %d errors!\n", failed)
			}
			logCore.Println("Migrated all characters")
	
			return nil
		}()
		
		if err != nil {
			// Error detected, revert db changes if possible
			if _, err := os.Stat(dbBakFileName); err == nil {
				os.Remove(dbFileName)
				os.Remove(oldDbFileName)
				os.Rename(dbBakFileName, dbFileName)
			}
			logCore.Fatalln("failed to migrate DB")
		}
		
		// happy path cleanup, we don't want to delete old one incase we need to revert suddenly.
		os.Remove(dbBakFileName)
	}
}

func Run(args []string) error {
	flgs := doFlags(args)

	config, err := app.LoadConfig(flgs.configFile)
	if err != nil {
		return err
	}

	apps := app.New(config);

	if flgs.migrateConfig {
		fmt.Println("Running migration...")
		if err := apps.MigrateConfig(); err != nil {
			fmt.Printf("Migration error: %s", err)
		}
		fmt.Println("Finished migration, starting server...")
	}

	//Initiate logging
	initLoggers("server.log", apps.Config.Log.Dir, apps.Config.Log.Level, apps.Config.Log.ExpireTime)
	apps.SetupLoggers(logCore, logAPI)

	//Max threads allowed.
	if apps.Config.Core.MaxThreads != 0 {
		runtime.GOMAXPROCS(apps.Config.Core.MaxThreads)
	}

	//Load json files.
	if apps.Config.ApiAuth.EnforceIP {
		logCore.Printf("Loading IP list from %s", apps.Config.ApiAuth.IPListFile)
		if err := apps.LoadIPList(apps.Config.ApiAuth.IPListFile); err != nil {
			logCore.Warnln("Failed to load IP list.")
		}
	}

	if apps.Config.Verify.EnforceMap {
		logCore.Printf("Loading Map list from %s", apps.Config.Verify.MapListFile)
		if err := apps.LoadMapList(apps.Config.Verify.MapListFile); err != nil {
			logCore.Warnln("Failed to load Map list.")
		}
	}

	if apps.Config.Verify.EnforceBan {
		logCore.Printf("Loading Ban list from %s", apps.Config.Verify.BanListFile)
		if err := apps.LoadBanList(apps.Config.Verify.BanListFile); err != nil {
			logCore.Warnln("Failed to load Ban list.")
		}
	}

	logCore.Printf("Loading Admin list from %s", apps.Config.Verify.AdminListFile)
	if err := apps.LoadAdminList(apps.Config.Verify.AdminListFile); err != nil {
		logCore.Warnln("Failed to load Admin list.")
	}

	//Connect database.
	logCore.Println("Connecting to database")
	tmpMigration(apps)
	defer apps.Client.Close()

	//variables for web server
	var srv *http.Server
	router := mux.NewRouter()
	srv = &http.Server{
		Handler:      router,
		Addr:         apps.Config.Core.Address + ":" + strconv.Itoa(apps.Config.Core.Port),
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
	mw := middleware.New(apps)

	router.Use(mw.PanicRecovery)
	router.Use(mw.Log)
	if apps.Config.RateLimit.Enable {
		router.Use(mw.RateLimit)
	}

	//api routes
	apic := controller.New(router.PathPrefix(apps.Config.Core.RootPath).Subrouter(), apps)
	apic.R.HandleFunc("/ping", mw.Lv2Auth(apic.GetPing)).Methods(http.MethodGet)
	apic.R.HandleFunc("/map/{name}/{hash}", mw.Lv1Auth(apic.GetMapVerify)).Methods(http.MethodGet)
	apic.R.HandleFunc("/ban/{steamid:[0-9]+}", mw.Lv1Auth(apic.GetBanVerify)).Methods(http.MethodGet)
	apic.R.HandleFunc("/sc/{hash}", mw.Lv1Auth(apic.GetSCVerify)).Methods(http.MethodGet)

	//character routes
	charc := controller.New(router.PathPrefix(apps.Config.Core.RootPath + "/character").Subrouter(), apps)
	charc.R.HandleFunc("/", mw.Lv1Auth(charc.GetAllCharacters)).Methods(http.MethodGet)
	charc.R.HandleFunc("/id/{uid}", mw.Lv1Auth(charc.GetCharacterByID)).Methods(http.MethodGet)
	charc.R.HandleFunc("/{steamid:[0-9]+}", mw.Lv1Auth(charc.GetCharacters)).Methods(http.MethodGet)
	charc.R.HandleFunc("/{steamid:[0-9]+}/{slot:[0-9]}", mw.Lv1Auth(charc.GetCharacter)).Methods(http.MethodGet)
	charc.R.HandleFunc("/export/{steamid:[0-9]+}/{slot:[0-9]}", mw.Lv1Auth(charc.ExportCharacter)).Methods(http.MethodGet)
	charc.R.HandleFunc("/", mw.Lv2Auth(charc.PostCharacter)).Methods(http.MethodPost)
	charc.R.HandleFunc("/{uid}", mw.Lv2Auth(charc.PutCharacter)).Methods(http.MethodPut)
	charc.R.HandleFunc("/{uid}", mw.Lv2Auth(charc.DeleteCharacter)).Methods(http.MethodDelete)
	charc.R.HandleFunc("/{uid}/restore", mw.Lv1Auth(charc.RestoreCharacter)).Methods(http.MethodPatch)
	charc.R.HandleFunc("/{steamid:[0-9]+}/{slot:[0-9]}/versions", mw.Lv1Auth(charc.CharacterVersions)).Methods(http.MethodGet)
	charc.R.HandleFunc("/{steamid:[0-9]+}/{slot:[0-9]}/rollback/{version:[0-9]+}", mw.Lv1Auth(charc.RollbackCharacter)).Methods(http.MethodPatch)
	charc.R.HandleFunc("/{steamid:[0-9]+}/{slot:[0-9]}/rollback/latest", mw.Lv1Auth(charc.RollbackLatestCharacter)).Methods(http.MethodPatch)
	charc.R.HandleFunc("/{steamid:[0-9]+}/{slot:[0-9]}/rollback", mw.Lv1Auth(charc.DeleteRollbacksCharacter)).Methods(http.MethodDelete)

	if apps.Config.Cert.Enable {
		cm := autocert.Manager{
			Prompt:     autocert.AcceptTOS,
			HostPolicy: autocert.HostWhitelist(apps.Config.Cert.Domain),
			Cache:      autocert.DirCache("./runtime/certs"),
		}

		srv.TLSConfig = &tls.Config{
			GetCertificate: cm.GetCertificate,
			NextProtos:     append(srv.TLSConfig.NextProtos, acme.ALPNProto), // enable tls-alpn ACME challenges
		}

		go func() {
			if err := http.ListenAndServe(":http", cm.HTTPHandler(nil)); err != nil {
				logCore.Errorf("failed to serve autocert server: %v", err)
			}
		}()
		
		go func() {
			logCore.Printf("Listening on: %s TLS", srv.Addr)
			if err := srv.ListenAndServeTLS("", ""); err != nil && err != http.ErrServerClosed {
				errMsg := errors.New(fmt.Sprintf("failed to serve over HTTPS: %v", err))
				panic(errMsg)
			}
		}()
	} else {
		go func() {
			logCore.Printf("Listening on: %s", srv.Addr)
			if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				errMsg := errors.New(fmt.Sprintf("failed to serve over HTTP: %v", err))
				panic(errMsg)
			}
		}()
	}

	s := make(chan os.Signal, 1)
	signal.Notify(s, os.Interrupt)
	<-s

	//wait 5 seconds before timing out
	ctx, cancel := context.WithTimeout(context.Background(), time.Second * 5)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		return err // failure/timeout shutting down the server gracefully
	}

	return nil
}
