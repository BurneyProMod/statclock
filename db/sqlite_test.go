package db

import (
	"path/filepath"
	"testing"
)

func TestSaveAndLoadPlayer(t *testing.T) {
	d := OpenDB(filepath.Join(t.TempDir(), "test.db"))
	defer d.Close()

	if got := d.LoadPlayer("missing"); got != nil {
		t.Errorf("LoadPlayer(missing) = %+v, want nil", got)
	}

	d.SavePlayer(Player{PlayerID: "p1", Nickname: "old-name", SteamID: "765"})
	d.SavePlayer(Player{PlayerID: "p1", Nickname: "new-name", SteamID: "765"}) // same ID updates the row

	got := d.LoadPlayer("p1")
	want := Player{PlayerID: "p1", Nickname: "new-name", SteamID: "765"}
	if got == nil || *got != want {
		t.Errorf("LoadPlayer(p1) = %+v, want %+v", got, want)
	}
}
