package lib

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

const profileSittingJSON = `{"id":"51182256","type":"meal_sitting","attributes":{"summary":"ZZTEST meal","description":"","note":"","rrule":null,"draft":false,"recurring":false,"instances":["2026-11-17"]},"relationships":{"meal_category":{"data":{"id":"2941306","type":"meal_category"}},"meal_recipe":{"data":null},"profiles":{"data":[{"id":"6751001","type":"category"},{"id":"6751003","type":"category"}]}}}`

const sittingInclude = "meal_category,meal_recipe,profiles"

func profileTestClient(t *testing.T, h http.HandlerFunc) *Client {
	t.Helper()
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)
	old := SkylightURL
	SkylightURL = srv.URL + "/api"
	t.Cleanup(func() { SkylightURL = old })
	c, err := NewClientWithToken("u", "t")
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func TestCreateMealSitting_ProfileIDs(t *testing.T) {
	var body map[string]any
	c := profileTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/frames/f1/meals/sittings" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		b, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(b, &body)
		_, _ = w.Write([]byte(`{"data":[` + profileSittingJSON + `],"included":[]}`))
	})

	got, err := c.CreateMealSitting(context.Background(), "f1", MealSittingData{
		Summary: "ZZTEST meal", MealCategoryID: "2941306", Date: "2026-11-17",
		ProfileIDs: []string{"6751001", "6751003"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(body["profile_ids"], []any{"6751001", "6751003"}) {
		t.Errorf("profile_ids body = %v", body["profile_ids"])
	}
	if !reflect.DeepEqual(got.ProfileIDs, []string{"6751001", "6751003"}) {
		t.Errorf("ProfileIDs = %v", got.ProfileIDs)
	}
	if got.RecipeID != "" || got.MealCategoryID != "2941306" || got.Date != "2026-11-17" {
		t.Errorf("unexpected sitting %+v", got)
	}
}

func TestListMealSittings_Profiles(t *testing.T) {
	c := profileTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if q.Get("include") != sittingInclude || q.Get("date_min") != "2026-11-01" || q.Get("date_max") != "2026-12-01" {
			t.Errorf("unexpected query %v", q)
		}
		_, _ = w.Write([]byte(`{"data":[` + profileSittingJSON + `]}`))
	})
	got, err := c.ListMealSittings(context.Background(), "f1", MealSittingListOptions{DateMin: "2026-11-01", DateMax: "2026-12-01"})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || !reflect.DeepEqual(got[0].ProfileIDs, []string{"6751001", "6751003"}) {
		t.Errorf("got %+v", got)
	}
}

func TestUpdateMealSitting_Instance(t *testing.T) {
	var body map[string]any
	c := profileTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch || r.URL.Path != "/api/frames/f1/meals/sittings/51182256/instances/2026-11-17" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		q := r.URL.Query()
		if q.Get("include") != sittingInclude || q.Get("date_min") != "2026-11-17" || q.Get("date_max") != "2026-11-18" {
			t.Errorf("unexpected query %v", q)
		}
		b, _ := io.ReadAll(r.Body)
		body = nil
		_ = json.Unmarshal(b, &body)
		_, _ = w.Write([]byte(`{"data":` + profileSittingJSON + `}`))
	})

	ids := []string{"6751001", "6751003", "6750983"}
	if _, err := c.UpdateMealSitting(context.Background(), "f1", "51182256", "2026-11-17", MealSittingUpdateData{ProfileIDs: &ids}); err != nil {
		t.Fatal(err)
	}
	if len(body) != 1 || !reflect.DeepEqual(body["profile_ids"], []any{"6751001", "6751003", "6750983"}) {
		t.Errorf("body = %v", body)
	}

	// An explicit empty list clears profiles and is still sent.
	empty := []string{}
	if _, err := c.UpdateMealSitting(context.Background(), "f1", "51182256", "2026-11-17", MealSittingUpdateData{ProfileIDs: &empty}); err != nil {
		t.Fatal(err)
	}
	if v, ok := body["profile_ids"]; !ok || !reflect.DeepEqual(v, []any{}) {
		t.Errorf("empty body = %v", body)
	}

	// An omitted profile list is not sent.
	if _, err := c.UpdateMealSitting(context.Background(), "f1", "51182256", "2026-11-17", MealSittingUpdateData{Summary: "x"}); err != nil {
		t.Fatal(err)
	}
	if _, ok := body["profile_ids"]; ok {
		t.Errorf("profile_ids should be omitted: %v", body)
	}

	if _, err := c.UpdateMealSitting(context.Background(), "f1", "51182256", "bad", MealSittingUpdateData{}); err == nil {
		t.Error("expected invalid date error")
	}
}

func TestGetMealSitting_Instances(t *testing.T) {
	for name, resp := range map[string]string{
		"array":  `{"data":[` + profileSittingJSON + `]}`,
		"object": `{"data":` + profileSittingJSON + `}`,
	} {
		t.Run(name, func(t *testing.T) {
			c := profileTestClient(t, func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodGet || r.URL.Path != "/api/frames/f1/meals/sittings/51182256/instances" {
					t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
				}
				q := r.URL.Query()
				if q.Get("date_min") != "2026-11-17" || q.Get("date_max") != "2026-11-18" || q.Get("include") != sittingInclude {
					t.Errorf("unexpected query %v", q)
				}
				_, _ = w.Write([]byte(resp))
			})
			got, err := c.GetMealSitting(context.Background(), "f1", "51182256", "2026-11-17")
			if err != nil {
				t.Fatal(err)
			}
			if got.ID != "51182256" || !reflect.DeepEqual(got.ProfileIDs, []string{"6751001", "6751003"}) {
				t.Errorf("got %+v", got)
			}
		})
	}

	t.Run("empty array", func(t *testing.T) {
		c := profileTestClient(t, func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write([]byte(`{"data":[]}`)) })
		if _, err := c.GetMealSitting(context.Background(), "f1", "1", "2026-11-17"); err == nil {
			t.Error("expected error")
		}
	})
}
