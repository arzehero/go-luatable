package luatable

import (
	"fmt"
	"strings"
	"testing"
)

func checkTable(t *testing.T, table interface{}, expected string) {
	lua, err := Encode(table)

	if err != nil {
		t.Error(err)
	}

	if lua != expected {
		t.Error("Expected table encoding doesn't match")
		t.Error("Expected:")
		fmt.Println(expected)
		t.Error("Got:")
		fmt.Println(lua)
	}
}

func TestSimpleTable(t *testing.T) {
	type Table struct {
		StringValue string  `lua:"string_value"`
		NumberValue int     `lua:"int_value"`
		BoolValue   bool    `lua:"bool_value"`
		FloatValue  float32 `lua:"float_value"`
	}
	table := Table{}
	expected := `return {
  string_value = '',
  int_value = 0,
  bool_value = false,
  float_value = 0,
}`

	checkTable(t, table, expected)
}

func TestArray(t *testing.T) {
	ints := []int{1, 2, 3, 4, 5}
	expected := `return { 1, 2, 3, 4, 5, }`

	checkTable(t, ints, expected)

	strings := []string{"a", "b'", "c"}
	expected = `return {
  'a',
  'b\'',
  'c',
}`
	checkTable(t, strings, expected)
}

func TestStructWithArray(t *testing.T) {
	type Table struct {
		Ints []int `lua:"ints"`
	}
	table := Table{
		Ints: []int{1, 2, 3},
	}
	expected := `return {
  ints = { 1, 2, 3, },
}`

	checkTable(t, table, expected)
}

func TestTablesArray(t *testing.T) {
	type IntsTable struct {
		Ints []int `lua:"ints"`
	}
	type Table struct {
		Tables []IntsTable `lua:"tables"`
	}
	table := Table{
		Tables: []IntsTable{
			{Ints: []int{1, 2, 3}},
			{Ints: []int{4, 5, 6}},
		},
	}
	expected := `return {
  tables = {
    {
      ints = { 1, 2, 3, },
    },
    {
      ints = { 4, 5, 6, },
    },
  },
}`

	checkTable(t, table, expected)
}

func TestNestedTables(t *testing.T) {
	type PokemonStats struct {
		Hp    int `lua:"hp"`
		Atk   int `lua:"atk"`
		Def   int `lua:"def"`
		SpAtk int `lua:"sp_atk"`
		SpDef int `lua:"sp_def"`
		Speed int `lua:"speed"`
	}
	type PokemonMoveDescriptions struct {
		English string `lua:"english"`
		Spanish string `lua:"spanish"`
	}
	type PokemonMove struct {
		Name         string                  `lua:"name"`
		Descriptions PokemonMoveDescriptions `lua:"descriptions"`
	}
	type Pokemon struct {
		Dex   int           `lua:"dex"`
		Name  string        `lua:"name"`
		Types []string      `lua:"types"`
		Stats PokemonStats  `lua:"stats"`
		Moves []PokemonMove `lua:"moves"`
	}

	acidEnglish := "Opposing Pokémon are attacked with a spray of harsh\nacid. This may also lower their Sp. Def stats."
	acidSpanish := "Rocía a los enemigos con un ácido corrosivo. \nPuede bajar la Defensa Especial."
	emberEnglish := "The target is attacked with small flames. This may\nalso leave the target with a burn."
	emberSpanish := "Ataca con llamas pequeñas que pueden causar \nquemaduras."
	table := Pokemon{
		Dex:   607,
		Name:  "Litwick",
		Types: []string{"ghost", "fire"},
		Stats: PokemonStats{
			Hp:    50,
			Atk:   30,
			Def:   55,
			SpAtk: 65,
			SpDef: 55,
			Speed: 20,
		},
		Moves: []PokemonMove{
			{
				Name: "Acid",
				Descriptions: PokemonMoveDescriptions{
					English: acidEnglish,
					Spanish: acidSpanish,
				},
			},
			{
				Name: "Ember",
				Descriptions: PokemonMoveDescriptions{
					English: emberEnglish,
					Spanish: emberSpanish,
				},
			},
		},
	}

	expected := `return {
  dex = 607,
  name = 'Litwick',
  types = {
    'ghost',
    'fire',
  },
  stats = {
    hp = 50,
    atk = 30,
    def = 55,
    sp_atk = 65,
    sp_def = 55,
    speed = 20,
  },
  moves = {
    {
      name = 'Acid',
      descriptions = {
        english = 'Opposing Pokémon are attacked with a spray of harsh\nacid. This may also lower their Sp. Def stats.',
        spanish = 'Rocía a los enemigos con un ácido corrosivo. \nPuede bajar la Defensa Especial.',
      },
    },
    {
      name = 'Ember',
      descriptions = {
        english = 'The target is attacked with small flames. This may\nalso leave the target with a burn.',
        spanish = 'Ataca con llamas pequeñas que pueden causar \nquemaduras.',
      },
    },
  },
}`

	checkTable(t, table, expected)
}

func TestMapTable(t *testing.T) {
	stringmap := map[string]string{
		"key":              "value",
		"key with spaces":  "value",
		"key-with-hyphens": "value",
	}
	expected := `return {
  key = 'value',
  ['key with spaces'] = 'value',
  ['key-with-hyphens'] = 'value',
}`

	checkTable(t, stringmap, expected)

	arraymap := map[string][]int{
		"key":              {1, 2, 3},
		"key with spaces":  {4, 5, 6},
		"key-with-hyphens": {7, 8, 9},
	}
	expected = `return {
  key = { 1, 2, 3, },
  ['key with spaces'] = { 4, 5, 6, },
  ['key-with-hyphens'] = { 7, 8, 9, },
}`
	checkTable(t, arraymap, expected)

	type Table struct {
		Id int `lua:"id"`
	}
	structmap := map[string]Table{
		"key":              {Id: 1},
		"key with spaces":  {Id: 2},
		"key-with-hyphens": {Id: 3},
	}
	expected = `return {
  key = {
    id = 1,
  },
  ['key with spaces'] = {
    id = 2,
  },
  ['key-with-hyphens'] = {
    id = 3,
  },
}`
	checkTable(t, structmap, expected)

	arraystructmap := map[string][]Table{
		"key":              {{Id: 1}, {Id: 2}, {Id: 3}},
		"key with spaces":  {{Id: 4}, {Id: 5}, {Id: 6}},
		"key-with-hyphens": {{Id: 7}, {Id: 8}, {Id: 9}},
	}
	expected = `return {
  key = {
    {
      id = 1,
    },
    {
      id = 2,
    },
    {
      id = 3,
    },
  },
  ['key with spaces'] = {
    {
      id = 4,
    },
    {
      id = 5,
    },
    {
      id = 6,
    },
  },
  ['key-with-hyphens'] = {
    {
      id = 7,
    },
    {
      id = 8,
    },
    {
      id = 9,
    },
  },
}`
	checkTable(t, arraystructmap, expected)
}

func TestMapInStruct(t *testing.T) {
	type Table struct {
		Capitals     map[string]string   `lua:"capitals"`
		Temperatures map[string][]int    `lua:"temperatures"`
		Cities       map[string][]string `lua:"cities"`
	}
	table := Table{
		Capitals: map[string]string{
			"United States": "value",
			"Brazil":        "value",
			"México":        "value",
		},
		Temperatures: map[string][]int{
			"United States": {1, 2, 3},
			"Brazil":        {4, 5, 6},
			"México":        {7, 8, 9},
		},
		Cities: map[string][]string{
			"United States": {"A", "B", "C"},
			"Brazil":        {"D", "E", "F"},
			"México":        {"G", "H", "I"},
		},
	}
	expected := ``

	checkTable(t, table, expected)
}

func escape(str string) string {
	return strings.ReplaceAll(str, "\n", "\\n")
}
