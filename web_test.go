package main

import (
	"bytes"
	"encoding/json"
	ctph "github.com/joekir/ssdeepviz/src/ctph"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewHash_withValidLengthField_ReturnsSerializedFHStruct(t *testing.T) {
	var jsonStr = []byte(`{"data_length": 15}`)
	req, err := http.NewRequest("POST", "/NewHash", bytes.NewBuffer(jsonStr))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(NewHash)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusCreated {
		t.Errorf("handler returned wrong status code: got %v want %v\n",
			status, http.StatusCreated)
	}

	if ctype := rr.Header().Get("Content-Type"); ctype != "application/json" {
		t.Errorf("handler returned wrong content type: got %v want %v\n", ctype, "application/json")
	}

	t.Logf("%#v\n", rr.Body.String())
}

func TestStepHash_noSession_Returns500(t *testing.T) {
	var jsonStr = []byte(`{"byte": 103}`)
	req, err := http.NewRequest("POST", "/StepHash", bytes.NewBuffer(jsonStr))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(StepHash)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusPreconditionRequired {
		t.Errorf("handler returned wrong status code: got %v want %v\n",
			status, http.StatusPreconditionRequired)
	}
}

func TestStepHash_withSession_StepsItByOne(t *testing.T) {
	var jsonStr = []byte(`{"data_length": 10}`)
	req, err := http.NewRequest("POST", "/NewHash", bytes.NewBuffer(jsonStr))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(NewHash)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusCreated {
		t.Errorf("handler returned wrong status code: got %v want %v\n",
			status, http.StatusCreated)
	}

	cookies := rr.Result().Cookies()
	var ctphCookie *http.Cookie
	for i := range cookies {
		if cookies[i].Name == CTPH_COOKIE {
			ctphCookie = cookies[i]
		}
	}

	if nil == ctphCookie {
		t.Error("First call did not return a cookie")
	}

	jsonStr = []byte(`{"byte": 103}`)
	req, err = http.NewRequest("POST", "/StepHash", bytes.NewBuffer(jsonStr))
	if err != nil {
		t.Fatal(err)
	}

	rr = httptest.NewRecorder()
	req.AddCookie(ctphCookie)
	handler = http.HandlerFunc(StepHash)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v\n",
			status, http.StatusOK)
	}

	decoder := json.NewDecoder(rr.Body)

	var fh ctph.FuzzyHash
	err = decoder.Decode(&fh)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("%#v\n", fh)

	if fh.Rh.Window[0] != 103 {
		t.Fatalf("expected 103, got %d", fh.Rh.Window[0])
	}

}
