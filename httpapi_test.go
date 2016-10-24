package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCreateAndListTaskViaHTTP(t *testing.T) {
	target := newHTTPAPI(&memDatabase{})
	srv := httptest.NewServer(target)
	defer srv.Close()

	res, err := http.Get(srv.URL + "/tasks")
	require.Nil(t, err)

	var readResp readTasksResponse
	err = json.NewDecoder(res.Body).Decode(&readResp)
	require.Nil(t, err)
	require.Empty(t, readResp.Tasks, "tasks should be empty")

	var payload map[string]interface{}
	buf, _ := json.Marshal(createTaskRequest{Name: "hans", Script: "ls -la"})
	res, err = http.Post(srv.URL+"/tasks", "application/json", bytes.NewReader(buf))
	require.Nil(t, err)

	err = json.NewDecoder(res.Body).Decode(&payload)
	require.Nil(t, err)

	taskID, ok := payload["id"]
	require.True(t, ok, "response should contains id property")

	res, err = http.Get(srv.URL + "/tasks")
	require.Nil(t, err)

	err = json.NewDecoder(res.Body).Decode(&readResp)
	require.Nil(t, err)
	require.NotEmpty(t, readResp.Tasks, "tasks should not be empty")

	var foundTaskID bool
	for _, tsk := range readResp.Tasks {
		if tsk.ID == taskID {
			foundTaskID = true
		}
	}
	require.True(t, foundTaskID, "response should contain created task")
}
