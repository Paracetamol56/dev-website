package controllers

import (
	"net/http"
	"testing"

	"dev/internal/db"
	"dev/internal/models"
	"dev/internal/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestGetHealth(t *testing.T) {
	response := do(t, http.MethodGet, "/health", nil)
	assert.Equal(t, http.StatusOK, response.Code)
	assert.Equal(t, `{"status":"ok"}`, response.Body.String())
}

func TestHipparcos(t *testing.T) {
	testutil.Clean(t)
	_, err := db.GetDB().Collection("hipparcos").InsertMany(ctx, []any{
		bson.M{"HIP": 1, "Vmag": 5.0, "Plx": 100.0, "B-V": 0.5},
		bson.M{"HIP": 2, "Vmag": 6.0, "Plx": 100.0, "B-V": 1.5},
		bson.M{"HIP": 3, "Vmag": 6.0, "Plx": 100.0, "B-V": 9.0},
	})
	require.NoError(t, err)

	all := do(t, http.MethodGet, "/hipparcos", nil)
	assert.Equal(t, http.StatusOK, all.Code)
	assert.Len(t, decode[[]models.HipparcosHR](t, all), 2)

	one := do(t, http.MethodGet, "/hipparcos/2", nil)
	assert.Equal(t, http.StatusOK, one.Code)
	assert.JSONEq(t, `{"hip": 2, "Amag": 6, "BV": 1.5}`, one.Body.String())

	assert.Equal(t, http.StatusNotFound, do(t, http.MethodGet, "/hipparcos/3", nil).Code, "outside the diagram")
	assert.Equal(t, http.StatusNotFound, do(t, http.MethodGet, "/hipparcos/99", nil).Code)
	assert.Equal(t, http.StatusBadRequest, do(t, http.MethodGet, "/hipparcos/sirius", nil).Code)

	testutil.BreakDB(t)
	assert.Equal(t, http.StatusInternalServerError, do(t, http.MethodGet, "/hipparcos", nil).Code)
	assert.Equal(t, http.StatusInternalServerError, do(t, http.MethodGet, "/hipparcos/1", nil).Code)
}

func TestIcons(t *testing.T) {
	testutil.Clean(t)
	require.NoError(t, models.ReplaceIcons(ctx, "seed", []models.Icon{
		{Id: "lucide:zap", Name: "zap", Title: "Zap", Source: "lucide", Tags: []string{"flash"}},
		{Id: "lucide:zap-off", Name: "zap-off", Title: "Zap Off", Source: "lucide", Tags: []string{}},
		{Id: "simpleicons:zapier", Name: "zapier", Title: "Zapier", Source: "simpleicons", Tags: []string{}, Hex: "FF4F00"},
	}))

	all := do(t, http.MethodGet, "/icons", nil)
	assert.Equal(t, http.StatusOK, all.Code)
	assert.Equal(t, "public, max-age=3600", all.Header().Get("Cache-Control"))
	assert.Len(t, decode[[]models.Icon](t, all), 3)
	assert.Len(t, decode[[]models.Icon](t, do(t, http.MethodGet, "/icons?source=simpleicons", nil)), 1)
	assert.Equal(t, http.StatusBadRequest, do(t, http.MethodGet, "/icons?source=fontawesome", nil).Code)

	found := do(t, http.MethodGet, "/icons/search?q=zap&source=lucide&source=simpleicons&limit=2", nil)
	assert.Equal(t, http.StatusOK, found.Code)
	assert.Equal(t, "public, max-age=3600", found.Header().Get("Cache-Control"))
	result := decode[models.IconSearchResult](t, found)
	assert.Equal(t, 3, result.Total)
	require.Len(t, result.Items, 2)
	assert.Equal(t, "lucide:zap", result.Items[0].Id)

	assert.Equal(t, 3, decode[models.IconSearchResult](t, do(t, http.MethodGet, "/icons/search", nil)).Total, "defaults")
	assert.Equal(t, 1, decode[models.IconSearchResult](t, do(t, http.MethodGet, "/icons/search?tag=FLASH", nil)).Total)
	for _, query := range []string{"limit=0", "limit=501", "offset=-1", "source=fontawesome", "q=" + repeat("a", 101), "limit=many"} {
		assert.Equal(t, http.StatusBadRequest, do(t, http.MethodGet, "/icons/search?"+query, nil).Code, query)
	}

	testutil.BreakDB(t)
	assert.Equal(t, http.StatusInternalServerError, do(t, http.MethodGet, "/icons", nil).Code)
	assert.Equal(t, http.StatusInternalServerError, do(t, http.MethodGet, "/icons/search", nil).Code)
}

func TestChips(t *testing.T) {
	testutil.Clean(t)
	require.NoError(t, models.ReplaceMicroprocessors(ctx, "wikipedia", []models.Microprocessor{
		{Id: "wikipedia:cpu-intel-4004-1971", Name: "Intel 4004", Type: "CPU", Vendor: "Intel", Transistors: 0.00225,
			Details: map[string]any{"process": map[string]any{"value": 10000.0, "unit": "nm"}}},
		{Id: "wikipedia:gpu-geforce-256-1999", Name: "GeForce 256", Type: "GPU", Vendor: "NVIDIA"},
	}))

	assert.Len(t, decode[[]map[string]any](t, do(t, http.MethodGet, "/chips", nil)), 2)
	gpus := decode[[]map[string]any](t, do(t, http.MethodGet, "/chips?type=GPU", nil))
	require.Len(t, gpus, 1)
	assert.Equal(t, "GeForce 256", gpus[0]["name"])
	assert.Len(t, decode[[]map[string]any](t, do(t, http.MethodGet, "/chips?vendor=Intel", nil)), 1)
	assert.JSONEq(t, `[]`, do(t, http.MethodGet, "/chips?vendor=Intel&type=GPU", nil).Body.String())

	one := do(t, http.MethodGet, "/chips/wikipedia:cpu-intel-4004-1971", nil)
	assert.Equal(t, http.StatusOK, one.Code)
	assert.JSONEq(t, `{"id": "wikipedia:cpu-intel-4004-1971", "name": "Intel 4004", "type": "CPU", "vendor": "Intel",
		"transistors": 0.00225, "process": {"value": 10000, "unit": "nm"}}`, one.Body.String())

	missing := do(t, http.MethodGet, "/chips/"+primitive.NewObjectID().Hex(), nil)
	assert.Equal(t, http.StatusNotFound, missing.Code)
	assert.Equal(t, "Chip not found", errorOf(t, missing))

	testutil.BreakDB(t)
	assert.Equal(t, http.StatusInternalServerError, do(t, http.MethodGet, "/chips", nil).Code)
	assert.Equal(t, http.StatusInternalServerError, do(t, http.MethodGet, "/chips/x", nil).Code)
}

func TestPostContact(t *testing.T) {
	testutil.Clean(t)
	internet := testutil.FakeInternet(t)
	fakeBrevo(internet)
	userId := primitive.NewObjectID()
	body := map[string]any{"userId": userId, "name": "Ada Lovelace", "email": "ada@example.test", "message": "I found a bug in the engine."}

	response := do(t, http.MethodPost, "/contact", body)
	assert.Equal(t, http.StatusCreated, response.Code)
	assert.JSONEq(t, `{"message": "Message sent"}`, response.Body.String())

	emails := sentEmails(t, internet)
	require.Len(t, emails, 1)
	assert.Equal(t, testutil.AdminEmail, emails[0].To[0].Email, "the message goes to the admin")
	assert.Equal(t, "ada@example.test", emails[0].ReplyTo.Email)
	assert.Equal(t, "Ada Lovelace", emails[0].ReplyTo.Name)
	assert.Contains(t, emails[0].TextContent, "I found a bug in the engine.")
	assert.Contains(t, emails[0].TextContent, userId.Hex())
	assert.Empty(t, emails[0].HTMLContent, "user input is never sent as HTML")

	var stored models.Contact
	require.NoError(t, db.GetDB().Collection("contacts").FindOne(ctx, bson.M{}).Decode(&stored))
	assert.Equal(t, userId, stored.UserId)
	assert.Equal(t, "Ada Lovelace", stored.Name)
	assert.Equal(t, "I found a bug in the engine.", stored.Message)
	assert.Equal(t, "<message@brevo>", stored.MailId)

	anonymous := map[string]any{"name": "Ada", "email": "ada@example.test", "message": "No account needed."}
	assert.Equal(t, http.StatusCreated, do(t, http.MethodPost, "/contact", anonymous).Code)
}

func TestPostContactRejections(t *testing.T) {
	testutil.Clean(t)
	internet := testutil.FakeInternet(t)
	fakeBrevo(internet)
	valid := func() map[string]any {
		return map[string]any{"name": "Ada", "email": "ada@example.test", "message": "Long enough message"}
	}

	for name, change := range map[string]func(map[string]any){
		"no name":           func(body map[string]any) { delete(body, "name") },
		"short name":        func(body map[string]any) { body["name"] = "A" },
		"long name":         func(body map[string]any) { body["name"] = repeat("a", 101) },
		"invalid email":     func(body map[string]any) { body["email"] = "ada@" },
		"no message":        func(body map[string]any) { delete(body, "message") },
		"short message":     func(body map[string]any) { body["message"] = "Too short" },
		"long message":      func(body map[string]any) { body["message"] = repeat("a", 1001) },
		"user id not an id": func(body map[string]any) { body["userId"] = "me" },
	} {
		body := valid()
		change(body)
		assert.Equal(t, http.StatusBadRequest, do(t, http.MethodPost, "/contact", body).Code, name)
	}
	assert.Equal(t, http.StatusBadRequest, do(t, http.MethodPost, "/contact", "not json").Code)
	assert.Empty(t, sentEmails(t, internet))
	assert.Zero(t, count(t, "contacts"))

	internet.JSON(brevoHost, http.StatusInternalServerError, `{}`)
	failed := do(t, http.MethodPost, "/contact", valid())
	assert.Equal(t, http.StatusInternalServerError, failed.Code)
	assert.Equal(t, "Error while sending email", errorOf(t, failed))
	assert.Zero(t, count(t, "contacts"), "a message that was not sent is not stored")

	fakeBrevo(internet)
	testutil.BreakDB(t)
	assert.Equal(t, http.StatusInternalServerError, do(t, http.MethodPost, "/contact", valid()).Code)
}
