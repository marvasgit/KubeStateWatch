package msteam

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"testing"

	"github.com/marvasgit/kubernetes-statemonitor/config"
	"github.com/marvasgit/kubernetes-statemonitor/pkg/event"
	"github.com/mohae/deepcopy"
	"github.com/stretchr/testify/assert"
)

func TestSendCard_Success(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	httptest.NewServer(handler)
	initCard := &TeamsMessageCard{
		Type:    messageType,
		Context: context,
		Summary: "statemonitor notification received",
		Title:   "statemonitor",

		Sections: []TeamsMessageCardSection{
			{
				Markdown: true,
			},
		},
	}

	server := httptestConfig(t, *initCard, "SendCard")
	defer server.Close()

	expectedCard := &TeamsMessageCard{
		Type:    messageType,
		Context: context,
		Summary: "statemonitor notification received",
		Title:   "statemonitor",

		Sections: []TeamsMessageCardSection{
			{
				Markdown: true,
			},
		},
	}

	ms := &MSTeams{
		TeamsWebhookURL: server.URL,
		Title:           "statemonitor",
	}

	response, err := sendCard(ms, expectedCard)
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, http.StatusOK, response.StatusCode)
}

func TestSendCard_EncodingError(t *testing.T) {
	ms := &MSTeams{
		TeamsWebhookURL: "invalid_url",
	}

	card := &TeamsMessageCard{
		Type: messageType,
		// ... initialize card fields ...
	}

	response, err := sendCard(ms, card)
	assert.Error(t, err)
	assert.Nil(t, response)
}

func TestSendCard_HTTPError(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	server := httptest.NewServer(handler)
	defer server.Close()

	ms := &MSTeams{
		TeamsWebhookURL: server.URL,
	}

	card := &TeamsMessageCard{
		Type: messageType,
		// ... initialize card fields ...
	}

	response, err := sendCard(ms, card)
	assert.Error(t, err)
	assert.Nil(t, response)
}

func TestInit_WithConfig(t *testing.T) {
	ms := &MSTeams{}
	config := &config.Config{
		Handler: config.Handler{
			MSTeams: config.MSTeams{
				WebhookURL: "test_url",
			},
		},
	}

	err := ms.Init(config)
	assert.NoError(t, err)
	assert.Equal(t, "test_url", ms.TeamsWebhookURL)
}

func TestInit_WithEnvVariable(t *testing.T) {
	ms := &MSTeams{}
	config := &config.Config{}
	os.Setenv("KW_MSTEAMS_WEBHOOKURL", "env_test_url")
	defer os.Unsetenv("KW_MSTEAMS_WEBHOOKURL")

	err := ms.Init(config)
	assert.NoError(t, err)
	assert.Equal(t, "env_test_url", ms.TeamsWebhookURL)
}

func TestInit_MissingWebhookURL(t *testing.T) {
	ms := &MSTeams{}
	config := &config.Config{}

	err := ms.Init(config)
	assert.Error(t, err)
}

var msTeamsTestMessage = event.StatemonitorEvent{
	Name:      "foo",
	Namespace: "new",
	Kind:      "pod",
}

// Tests the Init() function
func TestInit(t *testing.T) {
	s := &MSTeams{}
	expectedError := fmt.Errorf(msteamsErrMsg, "Missing MS teams webhook URL")
	var Tests = []struct {
		ms  config.MSTeams
		err error
	}{
		{config.MSTeams{WebhookURL: "somepath"}, nil},
		{config.MSTeams{}, expectedError},
	}

	for _, tt := range Tests {
		c := &config.Config{}
		c.Message.Title = "statemonitor"
		c.Handler.MSTeams = tt.ms
		if err := s.Init(c); !reflect.DeepEqual(err, tt.err) {
			t.Fatalf("Init(): %v", err)
		}
	}
}

// Tests ObjectCreated() by passing v1.Pod
func TestObjectCreated(t *testing.T) {
	e := deepcopy.Copy(msTeamsTestMessage).(event.StatemonitorEvent)
	e.Reason = "Created"
	e.Namespace = "new"
	e.Status = "Normal"

	expectedCard := TeamsMessageCard{
		Type:       messageType,
		Context:    context,
		ThemeColor: msTeamsColors["Normal"],
		Summary:    "statemonitor notification received",
		Title:      "statemonitor",
		Text:       "",

		Sections: []TeamsMessageCardSection{
			{
				Markdown: true,
				Facts:    getFacts(e),
			},
		},
	}

	ts := httptestConfig(t, expectedCard, "ObjectCreated")
	defer ts.Close()

	ms := &MSTeams{TeamsWebhookURL: ts.URL, Title: "statemonitor"}
	p := deepcopy.Copy(msTeamsTestMessage).(event.StatemonitorEvent)
	p.Reason = e.Reason
	p.Status = "Normal"
	ms.Handle(p)
}

// Tests ObjectDeleted() by passing v1.Pod
func TestObjectDeleted(t *testing.T) {
	e := deepcopy.Copy(msTeamsTestMessage).(event.StatemonitorEvent)
	e.Reason = "Deleted"
	e.Status = "Danger"
	expectedCard := TeamsMessageCard{
		Type:       messageType,
		Context:    context,
		ThemeColor: msTeamsColors["Danger"],
		Summary:    "statemonitor notification received",
		Title:      "statemonitor",
		Text:       "",
		Sections: []TeamsMessageCardSection{
			{
				Markdown: true,
				Facts:    getFacts(e),
			},
		},
	}

	ts := httptestConfig(t, expectedCard, "ObjectDeleted")
	defer ts.Close()

	ms := &MSTeams{TeamsWebhookURL: ts.URL, Title: "statemonitor"}

	p := deepcopy.Copy(msTeamsTestMessage).(event.StatemonitorEvent)
	p.Status = "Danger"
	p.Reason = "Deleted"

	ms.Handle(p)
}

// Tests ObjectUpdated() by passing v1.Pod
func TestObjectUpdated(t *testing.T) {
	e := deepcopy.Copy(msTeamsTestMessage).(event.StatemonitorEvent)
	e.Status = "Warning"
	e.Reason = "Updated"

	expectedCard := TeamsMessageCard{
		Type:       messageType,
		Context:    context,
		ThemeColor: msTeamsColors["Warning"],
		Summary:    "statemonitor notification received",
		Title:      "statemonitor",
		Text:       "",
		Sections: []TeamsMessageCardSection{
			{
				Markdown: true,
				Facts:    getFacts(e),
			},
		},
	}

	ts := httptestConfig(t, expectedCard, "ObjectUpdated")
	defer ts.Close()

	ms := &MSTeams{TeamsWebhookURL: ts.URL, Title: "statemonitor"}

	oldP := deepcopy.Copy(msTeamsTestMessage).(event.StatemonitorEvent)
	oldP.Reason = "Updated"
	oldP.Status = "Warning"

	ms.Handle(oldP)
}

func httptestConfig(t *testing.T, expectedCard TeamsMessageCard, action string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if r.Method != "POST" {
			t.Errorf("expected a POST request for %s()", action)
		}
		decoder := json.NewDecoder(r.Body)
		var c TeamsMessageCard
		if err := decoder.Decode(&c); err != nil {
			t.Errorf("%v", err)
		}
		if !reflect.DeepEqual(c, expectedCard) {
			t.Errorf("expected %v, got %v", expectedCard, c)
		}
	}))
}
func getFacts(e event.StatemonitorEvent) []TeamsMessageCardSectionFacts {
	return []TeamsMessageCardSectionFacts{
		{
			Name:  "Type",
			Value: e.Kind,
		},
		{
			Name:  "Name",
			Value: e.Name,
		},
		{
			Name:  "Action",
			Value: e.Reason,
		},
		{
			Name:  "Namespace",
			Value: e.Namespace,
		},
		{
			Name:  "Status",
			Value: e.Status,
		},
	}
}
