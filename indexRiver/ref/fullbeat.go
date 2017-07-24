package ref

import (
	"crypto/aes"
	"crypto/cipher"
	"database/sql"
	"encoding/hex"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/elastic/beats/libbeat/beat"
	"github.com/elastic/beats/libbeat/cfgfile"
	"github.com/elastic/beats/libbeat/common"
	"github.com/elastic/beats/libbeat/logp"

	"github.com/adibendahan/mysqlbeat/config"

	// mysql go driver
	_ "github.com/go-sql-driver/mysql"
)

// Mysqlbeat is a struct to hold the beat config & info
type Fullbeat struct {
	beatConfig       *config.Config
	done             chan struct{}
	period           time.Duration
	hostname         string
	port             string
	username         string
	password         string
	passwordAES      string
	queries          []string
	queryTypes       []string
	deltaWildcard    string
	deltaKeyWildcard string
	needquerys       []string

	ickey_query     string
	aii_query       string
	digikey_query   string
	element14_query string

	ickey_query_test     string
	aii_query_test       string
	digikey_query_test   string
	element14_query_test string

	supliers     common.MapStr
	oldValues    common.MapStr
	oldValuesAge common.MapStr
}

var (
	commonIV = []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f}
)

const (
	// secret length must be 16, 24 or 32, corresponding to the AES-128, AES-192 or AES-256 algorithms
	// you should compile your mysqlbeat with a unique secret and hide it (don't leave it in the code after compiled)
	// you can encrypt your password with github.com/adibendahan/mysqlbeat-password-encrypter just update your secret
	// (and commonIV if you choose to change it) and compile.
	secret = "github.com/stanxii/fullbeat"

	// default values
	defaultPeriod           = "10s"
	defaultHostname         = "10.8.11.225"
	defaultPort             = "3306"
	defaultUsername         = "root"
	defaultPassword         = "mysqlbeat_pass"
	defaultDeltaWildcard    = "__DELTA"
	defaultDeltaKeyWildcard = "__DELTAKEY"

	// query types values
	queryTypeSingleRow    = "single-row"
	queryTypeMultipleRows = "multiple-rows"
	queryTypeTwoColumns   = "two-columns"
	queryTypeSlaveDelay   = "show-slave-delay"

	// special column names values
	columnNameSlaveDelay = "Seconds_Behind_Master"

	// column types values
	columnTypeString = iota
	columnTypeInt
	columnTypeFloat
)

// New Creates beater
// func New() *Fullbeat {
// 	return &Fullbeat{
// 		done: make(chan struct{}),
// 	}
// }

// Creates beater
func New(b *beat.Beat, cfg *common.Config) (beat.Beater, error) {

	full := &Fullbeat{
		done: make(chan struct{}),
	}

	Config((full)

	return *full, nil
}

///*** Beater interface methods ***///

// Config is a function to read config file
func (bt *Fullbeat) Config(b *beat.Beat) error {

	// Load beater beatConfig
	err := cfgfile.Read(&bt.beatConfig, "")
	if err != nil {
		return fmt.Errorf("Error reading config file: %v", err)
	}

	return nil
}

// Setup is a function to setup all beat config & info into the beat struct
func (bt *Fullbeat) Setup(b *beat.Beat) error {

	if len(bt.beatConfig.Mysqlbeat.Queries) < 1 {
		err := fmt.Errorf("there are no queries to execute")
		return err
	}

	if len(bt.beatConfig.Mysqlbeat.Queries) != len(bt.beatConfig.Mysqlbeat.QueryTypes) {
		err := fmt.Errorf("error on config file, queries array length != queryTypes array length (each query should have a corresponding type on the same index)")
		return err
	}

	// Setting defaults for missing config
	if bt.beatConfig.Mysqlbeat.Period == "" {
		logp.Info("Period not selected, proceeding with '%v' as default", defaultPeriod)
		bt.beatConfig.Mysqlbeat.Period = defaultPeriod
	}

	if bt.beatConfig.Mysqlbeat.Hostname == "" {
		logp.Info("Hostname not selected, proceeding with '%v' as default", defaultHostname)
		bt.beatConfig.Mysqlbeat.Hostname = defaultHostname
	}

	if bt.beatConfig.Mysqlbeat.Port == "" {
		logp.Info("Port not selected, proceeding with '%v' as default", defaultPort)
		bt.beatConfig.Mysqlbeat.Port = defaultPort
	}

	if bt.beatConfig.Mysqlbeat.Username == "" {
		logp.Info("Username not selected, proceeding with '%v' as default", defaultUsername)
		bt.beatConfig.Mysqlbeat.Username = defaultUsername
	}

	if bt.beatConfig.Mysqlbeat.Password == "" && bt.beatConfig.Mysqlbeat.EncryptedPassword == "" {
		logp.Info("Password not selected, proceeding with default password")
		bt.beatConfig.Mysqlbeat.Password = defaultPassword
	}

	if bt.beatConfig.Mysqlbeat.DeltaWildcard == "" {
		logp.Info("DeltaWildcard not selected, proceeding with '%v' as default", defaultDeltaWildcard)
		bt.beatConfig.Mysqlbeat.DeltaWildcard = defaultDeltaWildcard
	}

	if bt.beatConfig.Mysqlbeat.DeltaKeyWildcard == "" {
		logp.Info("DeltaKeyWildcard not selected, proceeding with '%v' as default", defaultDeltaKeyWildcard)
		bt.beatConfig.Mysqlbeat.DeltaKeyWildcard = defaultDeltaKeyWildcard
	}

	// Parse the Period string
	var durationParseError error
	bt.period, durationParseError = time.ParseDuration(bt.beatConfig.Mysqlbeat.Period)
	if durationParseError != nil {
		return durationParseError
	}

	// Handle password decryption and save in the bt
	if bt.beatConfig.Mysqlbeat.Password != "" {
		bt.password = bt.beatConfig.Mysqlbeat.Password
	} else if bt.beatConfig.Mysqlbeat.EncryptedPassword != "" {
		aesCipher, err := aes.NewCipher([]byte(secret))
		if err != nil {
			return err
		}
		cfbDecrypter := cipher.NewCFBDecrypter(aesCipher, commonIV)
		chiperText, err := hex.DecodeString(bt.beatConfig.Mysqlbeat.EncryptedPassword)
		if err != nil {
			return err
		}
		plaintextCopy := make([]byte, len(chiperText))
		cfbDecrypter.XORKeyStream(plaintextCopy, chiperText)
		bt.password = string(plaintextCopy)
	}

	// init the oldValues and oldValuesAge array
	bt.oldValues = common.MapStr{"mysqlbeat": "init"}
	bt.oldValuesAge = common.MapStr{"mysqlbeat": "init"}

	// Save config values to the bt
	bt.hostname = bt.beatConfig.Mysqlbeat.Hostname
	bt.port = bt.beatConfig.Mysqlbeat.Port
	bt.username = bt.beatConfig.Mysqlbeat.Username
	bt.queries = bt.beatConfig.Mysqlbeat.Queries
	bt.queryTypes = bt.beatConfig.Mysqlbeat.QueryTypes
	bt.deltaWildcard = bt.beatConfig.Mysqlbeat.DeltaWildcard
	bt.deltaKeyWildcard = bt.beatConfig.Mysqlbeat.DeltaKeyWildcard

	safeQueries := true

	logp.Info("Total # of queries to execute: %d", len(bt.queries))
	for index, queryStr := range bt.queries {

		strCleanQuery := strings.TrimSpace(strings.ToUpper(queryStr))

		if !strings.HasPrefix(strCleanQuery, "SELECT") && !strings.HasPrefix(strCleanQuery, "SHOW") || strings.ContainsAny(strCleanQuery, ";") {
			safeQueries = false
		}

		logp.Info("Query #%d (type: %s): %s", index+1, bt.queryTypes[index], queryStr)
	}

	if !safeQueries {
		err := fmt.Errorf("Only SELECT/SHOW queries are allowed (the char ; is forbidden)")
		return err
	}

	return nil
}

// Run is a functions that runs the beat
func (bt *Fullbeat) Run(b *beat.Beat) error {
	logp.Info("mysqlbeat is running! Hit CTRL-C to stop it.")

	ticker := time.NewTicker(bt.period)
	for {
		select {
		case <-bt.done:
			return nil
		case <-ticker.C:
		}

		err := bt.beat(b)
		if err != nil {
			return err
		}

		//##############################################
		//Full index only need run once!!!!!!
		return nil
		//for Full index not need loop trick loop period runing
	}
}

// Cleanup is a function that does nothing on this beat :)
func (bt *Fullbeat) Cleanup(b *beat.Beat) error {
	return nil
}

// Stop is a function that runs once the beat is stopped
func (bt *Fullbeat) Stop() {
	close(bt.done)
}

// beat is a function that iterate over the query array, generate and publish events
func (bt *Fullbeat) beat(b *beat.Beat) error {

	// Build the MySQL connection string
	connString := fmt.Sprintf("%v:%v@tcp(%v:%v)/", bt.username, bt.password, bt.hostname, bt.port)

	fmt.Println("beat it , beat it. connectstr %s", connString)

	// Great success!
	return nil
}

// getKeyFromRow is a function that returns a unique key from row
func getKeyFromRow(bt *Fullbeat, values []sql.RawBytes, columns []string) (strKey string, err error) {

	keyFound := false

	// Loop on all columns
	for i, col := range values {
		// Get column name and string value
		if strings.HasSuffix(string(columns[i]), bt.deltaKeyWildcard) {
			strKey += string(col)
			keyFound = true
		}
	}

	if !keyFound {
		err = fmt.Errorf("query type multiple-rows requires at least one delta key column")
	}

	return strKey, err
}

// roundF2I is a function that returns a rounded int64 from a float64
func roundF2I(val float64, roundOn float64) (newVal int64) {
	var round float64

	digit := val
	_, div := math.Modf(digit)
	if div >= roundOn {
		round = math.Ceil(digit)
	} else {
		round = math.Floor(digit)
	}

	return int64(round)
}
