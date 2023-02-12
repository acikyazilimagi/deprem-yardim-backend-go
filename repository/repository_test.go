package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"testing"

	"github.com/ggwhite/go-masker"
	"github.com/pashagolub/pgxmock/v2"
	"github.com/stretchr/testify/assert"
)

func TestMask(t *testing.T) {
	jsonMap := map[string]interface{}{
		"name": "emre huseyin",
		"tel":  "535 555 55 55",
	}

	assert.Equal(t, masker.Telephone(fmt.Sprintf("%v", jsonMap["tel"])), "(53)5555-****")
	assert.Equal(t, masker.Telephone(fmt.Sprintf("%v", jsonMap["telefon"])), "<nil>")
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

func Test_UpdateLocationIntent(t *testing.T) {
	//Given
	ctx := context.Background()
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatal(err)
	}
	defer mock.Close()

	id := int64(333)
	reason := "kara"

	mock.ExpectExec(regexp.QuoteMeta(`UPDATE feeds_location SET reason = $1 WHERE entry_id=$2;`)).
		WithArgs(reason, id).
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))

	repo := New(mock)

	//When
	err = repo.UpdateLocationIntent(ctx, id, reason)

	//Then
	if err != nil {
		t.Errorf("error was not expected while updating: %s", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}
