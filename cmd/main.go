package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"github.com/wazupwiddat/postrack/server/config"
	"github.com/wazupwiddat/postrack/server/controllers"
	"github.com/wazupwiddat/postrack/server/schwab"
	"github.com/wazupwiddat/postrack/server/stock"
	"github.com/wazupwiddat/postrack/server/transaction"
	"github.com/wazupwiddat/postrack/server/user"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var db *gorm.DB

func main() {
	// Config
	cfg, err := config.NewConfig("./config.yml")
	if err != nil {
		log.Fatal(err)
		return
	}
	// connect to the database
	dsn := cfg.MySQLDNS()
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}

	db.AutoMigrate(&user.User{}, &transaction.Transaction{}, &stock.Stock{}, &schwab.SchwabAccess{})

	router := mux.NewRouter()
	controller := controllers.InitController(db, cfg)
	c := cors.AllowAll()

	router.HandleFunc("/signup", controller.HandleSignup).Methods("POST")
	router.HandleFunc("/login", controller.HandleLogin).Methods("POST")

	protected := router.PathPrefix("/protected").Subrouter()
	protected.HandleFunc("/stock", controller.HandleStockSimpleView).Methods("GET")
	protected.HandleFunc("/stock/{symbol}", controller.HandleStockAdd).Methods("POST")
	protected.HandleFunc("/stock/{symbol}", controller.HandleStockRemove).Methods("DELETE")
	protected.HandleFunc("/summary", controller.HandleSummary).Methods("GET")
	protected.HandleFunc("/import", controller.HandleImport).Methods("POST")
	protected.HandleFunc("/inspect", controller.HandleInspect).Methods("GET")
	protected.HandleFunc("/inspect/{symbol}", controller.HandleInspectSymbol).Methods("GET")
	protected.HandleFunc("/schwabaccess", controller.HandleSchwabAccess).Methods("POST")
	protected.HandleFunc("/schwabimporttrans", controller.HandleSchwabImportTrans).Methods("POST")
	protected.Use(controller.VerifyJWT)

	http.ListenAndServe(fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port), c.Handler(router))
}
