package cmd

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/sebrandon1/go-skylight/lib"
)

func parseAlarmFlags(t *testing.T, args ...string) (lib.AlarmData, error) {
	t.Helper()
	cmd := alarmFieldCmd(t, args...)
	if err := cmd.ValidateFlagGroups(); err != nil {
		return lib.AlarmData{}, err
	}
	return alarmDataFromFlags(cmd)
}

func TestAlarmDataFromFlags(t *testing.T) {
	t.Run("no flags yields empty data", func(t *testing.T) {
		data, err := parseAlarmFlags(t)
		if err != nil {
			t.Fatal(err)
		}
		if data != (lib.AlarmData{}) {
			t.Errorf("want empty AlarmData, got %+v", data)
		}
	})

	t.Run("all fields", func(t *testing.T) {
		data, err := parseAlarmFlags(t, "--label", "School", "--time", "6:05", "--days", "mon,fri",
			"--sound", "chimes", "--volume", "40", "--snoozable", "--enabled=false")
		if err != nil {
			t.Fatal(err)
		}
		b, _ := json.Marshal(data)
		want := `{"label":"School","enabled":false,"time":"06:05","rrule":"FREQ=WEEKLY;INTERVAL=1;WKST=SU;BYDAY=MO,FR","sound":"chimes","volume":40,"snoozable":true}`
		if string(b) != want {
			t.Errorf("got  %s\nwant %s", b, want)
		}
	})

	t.Run("raw rrule strips RRULE prefix", func(t *testing.T) {
		data, err := parseAlarmFlags(t, "--rrule", "RRULE:FREQ=DAILY")
		if err != nil {
			t.Fatal(err)
		}
		if data.RRule == nil || *data.RRule != "FREQ=DAILY" {
			t.Errorf("got %v", data.RRule)
		}
	})

	errCases := []struct {
		name string
		args []string
		want string
	}{
		{"bad time", []string{"--time", "25:00"}, "invalid --time"},
		{"bad day", []string{"--days", "funday"}, "invalid --days"},
		{"volume too high", []string{"--volume", "101"}, "invalid --volume"},
		{"days and rrule together", []string{"--days", "mon", "--rrule", "FREQ=DAILY"}, "none of the others"},
	}
	for _, tc := range errCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := parseAlarmFlags(t, tc.args...)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Errorf("want error containing %q, got %v", tc.want, err)
			}
		})
	}
}

func TestPickBuddyDevice(t *testing.T) {
	frame := lib.Device{ID: "1", Role: "frame"}
	buddyA := lib.Device{ID: "2", Role: lib.DeviceRoleBuddy}
	buddyB := lib.Device{ID: "3", Role: lib.DeviceRoleBuddy}

	if id, err := pickBuddyDevice([]lib.Device{frame, buddyA}); err != nil || id != "2" {
		t.Errorf("single buddy: got %q, %v", id, err)
	}
	if _, err := pickBuddyDevice([]lib.Device{frame}); err == nil || !strings.Contains(err.Error(), "no Buddy") {
		t.Errorf("no buddy: got %v", err)
	}
	if _, err := pickBuddyDevice([]lib.Device{buddyA, buddyB}); err == nil || !strings.Contains(err.Error(), "2, 3") {
		t.Errorf("multiple buddies: got %v", err)
	}
}

func TestAlarmTableHelpers(t *testing.T) {
	if got := alarmDaysDisplay("FREQ=WEEKLY;INTERVAL=1;WKST=SU;BYDAY=MO,TU"); got != "MO,TU" {
		t.Errorf("alarmDaysDisplay BYDAY: got %q", got)
	}
	if got := alarmDaysDisplay("FREQ=DAILY"); got != "FREQ=DAILY" {
		t.Errorf("alarmDaysDisplay no BYDAY: got %q", got)
	}
	if got := alarmDaysDisplay(""); got != "-" {
		t.Errorf("alarmDaysDisplay empty: got %q", got)
	}
	if got := rawJSONDisplay(json.RawMessage(`"chimes"`)); got != "chimes" {
		t.Errorf("rawJSONDisplay string: got %q", got)
	}
	if got := rawJSONDisplay(json.RawMessage(`{"id":1}`)); got != `{"id":1}` {
		t.Errorf("rawJSONDisplay object: got %q", got)
	}
	if got := rawJSONDisplay(nil); got != "-" {
		t.Errorf("rawJSONDisplay nil: got %q", got)
	}
}
