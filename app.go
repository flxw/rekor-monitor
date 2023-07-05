package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"log"
	"sync"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"github.com/go-openapi/runtime"

	"github.com/sigstore/rekor/pkg/client"
	rekorClient "github.com/sigstore/rekor/pkg/generated/client"
	"github.com/sigstore/rekor/pkg/generated/client/entries"
	"github.com/sigstore/rekor/pkg/generated/models"
	"github.com/sigstore/rekor/pkg/types"

	// these imports are to call the packages' init methods
	_ "github.com/sigstore/rekor/pkg/types/alpine/v0.0.1"
	_ "github.com/sigstore/rekor/pkg/types/cose/v0.0.1"
	_ "github.com/sigstore/rekor/pkg/types/dsse/v0.0.1"
	_ "github.com/sigstore/rekor/pkg/types/hashedrekord/v0.0.1"
	_ "github.com/sigstore/rekor/pkg/types/helm/v0.0.1"
	_ "github.com/sigstore/rekor/pkg/types/intoto/v0.0.1"
	_ "github.com/sigstore/rekor/pkg/types/intoto/v0.0.2"
	_ "github.com/sigstore/rekor/pkg/types/jar/v0.0.1"
	_ "github.com/sigstore/rekor/pkg/types/rekord/v0.0.1"
	_ "github.com/sigstore/rekor/pkg/types/rfc3161/v0.0.1"
	_ "github.com/sigstore/rekor/pkg/types/rpm/v0.0.1"
	_ "github.com/sigstore/rekor/pkg/types/tuf/v0.0.1"

	_ "github.com/go-sql-driver/mysql"
)

var TIMEOUT, _ = time.ParseDuration("30s")
var SLEEP_DURATION, _ = time.ParseDuration("20s")

type CrawledEntry struct {
	gorm.Model
	Ts          time.Time
	Idx         int64
	Sub         string
	Pubkey_hash string
}

type EntryInfo struct {
	Timestamp int64
	Index     int64
	EntryImpl types.EntryImpl
}

func main() {
	connectionString := fmt.Sprintf("%s?parseTime=True", Config.Database.String)
	db, err := gorm.Open(mysql.Open(connectionString), &gorm.Config{})
	if err != nil {
		log.Panicln("failed to connect database")
	}

	db.AutoMigrate(&CrawledEntry{})

	rekorClient, _ := client.GetRekorClient(
		Config.Rekor.Url,
		client.WithUserAgent(Config.Rekor.UserAgent),
		client.WithRetryCount(Config.Rekor.RetryCount))

	for {
		ExecuteCrawlRun(rekorClient, db)
		time.Sleep(SLEEP_DURATION)
	}
}

func ExecuteCrawlRun(rekorClient *rekorClient.Rekor, db *gorm.DB) {
	var mostRecentlyCrawledIndex int64 = DetermineStartIndex(db)
	var maximumIndex = mostRecentlyCrawledIndex + 100 //CalculateCurrentMaximumIndex(rekorClient)
	log.Println("Crawling until index:", maximumIndex)

	rekordQueue := make(chan CrawledEntry)
	go SpawnRekorCrawlerRoutines(mostRecentlyCrawledIndex, maximumIndex, rekorClient, rekordQueue)

	for entry := range rekordQueue {
		db.Create(&entry)
	}
}

func CalculateSha256Of(s string) string {
	h := sha256.New()
	h.Write([]byte(s))
	bs := h.Sum(nil)
	return fmt.Sprintf("%x", bs)
}

func SpawnRekorCrawlerRoutines(fromIndex int64, toIndex int64, rekorClient *rekorClient.Rekor, rekordQueue chan CrawledEntry) {
	defer close(rekordQueue)

	var wg sync.WaitGroup
	var number = int(toIndex - fromIndex - 1)
	wg.Add(number)

	for i := fromIndex + 1; i < toIndex; i++ {
		go FetchEntryByUuid(rekorClient, i, &wg, rekordQueue)
	}
	wg.Wait()
}

func CalculateCurrentMaximumIndex(rekorClient *rekorClient.Rekor) int64 {
	logInfo, _ := rekorClient.Tlog.GetLogInfo(nil)
	var inactiveIndexCount int64 = 0

	for _, inactiveShard := range logInfo.Payload.InactiveShards {
		inactiveIndexCount += *inactiveShard.TreeSize
	}

	inactiveIndexCount += *logInfo.Payload.TreeSize

	return inactiveIndexCount
}

func FetchEntryByUuid(rekorClient *rekorClient.Rekor, index int64, wg *sync.WaitGroup, elementQueue chan CrawledEntry) {
	defer wg.Done()

	params := entries.NewGetLogEntryByIndexParams()
	params.SetTimeout(TIMEOUT)
	params.LogIndex = index

	resp, err := rekorClient.Entries.GetLogEntryByIndex(params)

	if err != nil {
		fmt.Println(err)
		return
	}

	for _, entry := range resp.Payload {
		b, _ := base64.StdEncoding.DecodeString(entry.Body.(string))

		pe, err := models.UnmarshalProposedEntry(bytes.NewReader(b), runtime.JSONConsumer())
		if err != nil {
			log.Println(err)
			return
		}

		eimpl, err := types.UnmarshalEntry(pe)
		if err != nil {
			log.Println(err)
			return
		}

		verifier, _ := eimpl.Verifier()
		subject := ""
		if len(verifier.Subjects()) > 0 {
			subject = verifier.Subjects()[0]
		}

		identities, _ := verifier.Identities()
		pubkey_hash := CalculateSha256Of(identities[0])

		elementQueue <- CrawledEntry{
			Ts:          time.Unix(*entry.IntegratedTime, 0),
			Idx:         *entry.LogIndex,
			Sub:         subject,
			Pubkey_hash: pubkey_hash,
		}
	}
}

func DetermineStartIndex(db *gorm.DB) int64 {
	var index int64 = -1
	query := "SELECT MAX(idx) FROM crawled_entries"

	db.Raw(query).Scan(&index)
	if index == -1 {
		index = Config.Rekor.StartIndex
		log.Println("Could not fetch index from db, falling back to config start index:", index)
	}

	return index
}
