package repository

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/ggwhite/go-masker"
	"github.com/stretchr/testify/assert"
)

func TestMask(t *testing.T) {
	jsonMap := map[string]interface{}{
		"name":     "emre huseyin",
		"tel":      "535 555 55 55",
		"telefon2": "905555555555",
		"telefon3": "0535 555 55 55",
	}

	assert.Equal(t, masker.Telephone(fmt.Sprintf("%v", jsonMap["tel"])), "(53)5555-****")
	assert.Equal(t, masker.Telephone(fmt.Sprintf("%v", jsonMap["telefon"])), "<nil>")
	assert.Equal(t, masker.Telephone(fmt.Sprintf("%v", jsonMap["telefon2"])), "(53)5555-****")
	assert.Equal(t, masker.Telephone(fmt.Sprintf("%v", jsonMap["telefon3"])), "(53)5555-****")
	assert.Equal(t, masker.Name(fmt.Sprintf("%v", jsonMap["name"])), "e**e h**eyin")
}

func TestParseExtraParam(t *testing.T) {
	str := `
{
   "numara ":  "0535 646 87 47 ",
   "aciklama ":  "battaniye çadır baklıyat, un şeker, jeneratör çok acil 5 Aileler çok sıkıntı çekiyorlar ",
   "il ":  "hatay ",
   "ilce ":  "samandağ ",
   "excel_id ": 241,
   "durum ":  "teyitli "
}`

	var jsonMap map[string]interface{}

	assert.NoError(t, json.Unmarshal([]byte(str), &jsonMap))
}

func TestParse(t *testing.T) {
	str := `
{"name_surname": "Tugay Özalpay", "tel": "0532 423 0571", "additional_notes": "2'si bebek 4 kişi 71 saattir enkaz altında", "status": "Enkaz", "manual_confirmation": "nan"}
`

	var jsonMap map[string]interface{}

	assert.NoError(t, json.Unmarshal([]byte(str), &jsonMap))
}
