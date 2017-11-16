# Kataras PKG

## Directory

| Name | Description | Depends On | Version  | 
|---|---|---|---|
| [geoloc](geoloc) | Fetch geolocation and language information from a remote machine based on its IP | [kataras/chronos](https://github.com/kataras/chronos), [kataras/iris](https://github.com/kataras/iris) | 0.0.1 |
| [zerocheck](zerocheck) | One function; `IsZero` returns true if exported fields are zero from a struct, or slice/map is empty or user-defined `IsZero` function returns true, otherwise false | [go std library](https://golang.org/pkg/) and **only**  | 0.0.1 |
| [structcopy](structcopy) | Copies struct's fields to another struct, including embedded and anonymous fields | [jinzhu/copier](https://github.com/jinzhu/copier) and **only** | 0.0.1 |