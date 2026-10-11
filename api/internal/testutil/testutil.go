// Package testutil holds what the tests of every package share: a disposable database,
// fake external services and a few fixtures. It must not import the packages under test.
package testutil

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"dev/internal/db"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	AccessSecret  = "test-access-secret"
	RefreshSecret = "test-refresh-secret"
	AdminEmail    = "admin@example.test"
	Origin        = "http://localhost:5173"

	prefix = "dev_test_"
)

var client *mongo.Client

// URI returns TEST_DB_URI, or the DB_URI of api/.env so tests run against the local
// development server. Tests only ever touch databases named dev_test_*.
func URI() string {
	if uri := os.Getenv("TEST_DB_URI"); uri != "" {
		return uri
	}
	dir, _ := os.Getwd()
	for range 6 {
		if values, err := godotenv.Read(filepath.Join(dir, ".env")); err == nil && values["DB_URI"] != "" {
			return values["DB_URI"]
		}
		dir = filepath.Dir(dir)
	}
	return "mongodb://localhost:27017"
}

// Main runs the tests of a package against its own database, dropped afterwards.
func Main(m *testing.M) {
	gin.SetMode(gin.TestMode)
	gin.DefaultWriter = io.Discard
	log.SetOutput(io.Discard)

	os.Setenv("ACCESS_TOKEN_SECRET", AccessSecret)
	os.Setenv("REFRESH_TOKEN_SECRET", RefreshSecret)
	os.Setenv("ADMIN_EMAIL", AdminEmail)
	os.Setenv("BREVO_API_KEY", "test-brevo-key")
	os.Setenv("BREVO_CONTACT_LIST_ID", "7")
	for _, name := range []string{"GITHUB_CLIENT_ID", "GITHUB_CLIENT_SECRET", "GOOGLE_CLIENT_ID", "GOOGLE_CLIENT_SECRET"} {
		os.Setenv(name, "test-"+strings.ToLower(name))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	connected, err := mongo.Connect(ctx, options.Client().ApplyURI(URI()).SetServerSelectionTimeout(5*time.Second))
	if err == nil {
		err = connected.Ping(ctx, nil)
	}
	cancel()
	if err != nil {
		fmt.Fprintln(os.Stderr, "These tests need a MongoDB server: start one (docker compose up -d db) or set TEST_DB_URI.")
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	client = connected

	dropStaleDatabases()
	database := client.Database(fmt.Sprintf("%s%d_%d", prefix, os.Getpid(), time.Now().UnixNano()))
	db.Use(database)

	code := m.Run()

	database.Drop(context.Background())
	client.Disconnect(context.Background())
	os.Exit(code)
}

// dropStaleDatabases removes what interrupted runs left behind
func dropStaleDatabases() {
	ctx := context.Background()
	names, _ := client.ListDatabaseNames(ctx, bson.M{"name": bson.M{"$regex": "^" + prefix}})
	for _, name := range names {
		_, created, _ := strings.Cut(strings.TrimPrefix(name, prefix), "_")
		nanoseconds, err := strconv.ParseInt(created, 10, 64)
		if err == nil && time.Since(time.Unix(0, nanoseconds)) > time.Hour {
			client.Database(name).Drop(ctx)
		}
	}
}

// Clean empties every collection, keeping their indexes.
func Clean(t testing.TB) {
	t.Helper()
	ctx := context.Background()
	names, err := db.GetDB().ListCollectionNames(ctx, bson.M{})
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range names {
		if _, err := db.GetDB().Collection(name).DeleteMany(ctx, bson.M{}); err != nil {
			t.Fatal(err)
		}
	}
}

// BreakDB makes every database call fail until the test ends.
func BreakDB(t testing.TB) {
	t.Helper()
	working := db.GetDB()
	broken, err := mongo.Connect(context.Background(), options.Client().ApplyURI(URI()))
	if err != nil {
		t.Fatal(err)
	}
	broken.Disconnect(context.Background())
	db.Use(broken.Database(working.Name()))
	t.Cleanup(func() { db.Use(working) })
}

type Request struct {
	Method string
	Host   string
	Path   string
	Query  string
	Header http.Header
	Body   string
}

// Internet stands in for every external service: requests to a registered host go to its
// handler, requests to this machine go through, anything else fails.
type Internet struct {
	mutex    sync.Mutex
	handlers map[string]http.Handler
	requests []Request
	fallback http.RoundTripper
}

func FakeInternet(t testing.TB) *Internet {
	t.Helper()
	internet := &Internet{handlers: map[string]http.Handler{}, fallback: http.DefaultTransport}
	http.DefaultTransport = internet
	t.Cleanup(func() { http.DefaultTransport = internet.fallback })
	return internet
}

func (internet *Internet) Handle(host string, handler http.HandlerFunc) {
	internet.mutex.Lock()
	defer internet.mutex.Unlock()
	internet.handlers[host] = handler
}

// JSON answers every request to host with the given status and body.
func (internet *Internet) JSON(host string, status int, body string) {
	internet.Handle(host, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		io.WriteString(w, body)
	})
}

func (internet *Internet) Requests(host string) []Request {
	internet.mutex.Lock()
	defer internet.mutex.Unlock()
	matching := []Request{}
	for _, request := range internet.requests {
		if request.Host == host {
			matching = append(matching, request)
		}
	}
	return matching
}

func (internet *Internet) RoundTrip(req *http.Request) (*http.Response, error) {
	host := req.URL.Hostname()
	internet.mutex.Lock()
	handler := internet.handlers[host]
	internet.mutex.Unlock()

	if handler == nil {
		if ip := net.ParseIP(host); host == "localhost" || (ip != nil && ip.IsLoopback()) {
			return internet.fallback.RoundTrip(req)
		}
		return nil, fmt.Errorf("unexpected request to %s", req.URL)
	}

	body := ""
	if req.Body != nil {
		content, _ := io.ReadAll(req.Body)
		req.Body.Close()
		body = string(content)
	}
	internet.mutex.Lock()
	internet.requests = append(internet.requests, Request{
		Method: req.Method, Host: host, Path: req.URL.Path, Query: req.URL.RawQuery, Header: req.Header.Clone(), Body: body,
	})
	internet.mutex.Unlock()

	recorder := httptest.NewRecorder()
	replay := req.Clone(req.Context())
	replay.Body = io.NopCloser(strings.NewReader(body))
	handler.ServeHTTP(recorder, replay)
	return recorder.Result(), nil
}
