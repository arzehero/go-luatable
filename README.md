# go-luatable

go-luatable is a Go library for encoding structs into Lua tables.

## Usage

```go
import "github.com/arzehero/go-luatable/luatable"
```

The luatable package exports a single `Encode()` function. A simple example:

```go
type Unit struct {
  Hp int `lua:"hp"`
  Name string `lua:"name"`
}

knight := Unit {
  Hp = 100,
  Name = "Knight",
}
lua, err := luatable.Encode(knight)
```

This will result in:

```lua
return {
  hp = 100,
  name = 'Knight'
}
```

