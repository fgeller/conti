package main

import (
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCreateAndListTask(t *testing.T) {
	tmp, err := ioutil.TempFile("", "jsondb")
	require.Nil(t, err)
	tmp.Close()

	db, err := newJSONDatabase(tmp.Name())
	require.Nil(t, err)

	id, err := db.createTask("hans", "ls -ltr")
	require.Nil(t, err)

	ts, err := db.readTasks()
	require.Nil(t, err)

	var foundID bool
	for _, t := range ts {
		if t.ID == id {
			foundID = true
		}
	}
	require.True(t, foundID, "readTasks result should contain created task")
}
