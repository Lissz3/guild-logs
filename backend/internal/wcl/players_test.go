package wcl

import (
	"encoding/json"
	"testing"
)

func TestParsePlayerDetails(t *testing.T) {
	raw := json.RawMessage(`{"data":{"playerDetails":{
	  "tanks":[{"name":"Tanky","id":1,"type":"Warrior","specs":[{"spec":"Protection","count":1}]}],
	  "healers":[{"name":"Holy","id":3,"type":"Priest","specs":[{"spec":"Holy","count":1}]}],
	  "dps":[{"name":"Mago","id":8,"type":"Mage","specs":[{"spec":"Frost","count":1},{"spec":"Fire","count":5}]}]
	}}}`)
	ps, err := ParsePlayerDetails(raw)
	if err != nil {
		t.Fatal(err)
	}
	if len(ps) != 3 {
		t.Fatalf("esperaba 3 jugadores, hay %d", len(ps))
	}
	for _, p := range ps {
		if p.Name == "Mago" && (p.Spec != "Fire" || p.Role != "dps") {
			t.Errorf("mago mal parseado: %+v", p)
		}
		if p.Name == "Tanky" && p.Role != "tank" {
			t.Errorf("tank mal parseado: %+v", p)
		}
	}
}
