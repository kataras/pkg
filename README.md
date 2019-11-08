# Kataras PKG

## Directory

| Name | Description | Depends On | Version  | 
|---|---|---|---|
| [geoloc](geoloc) | Fetch geolocation and language information from a remote machine based on its IP | [kataras/chronos](https://github.com/kataras/chronos), [kataras/iris](https://github.com/kataras/iris) |0.0.2 |
| [zerocheck](zerocheck) | One function; `IsZero` returns true if exported fields are zero from a struct, or slice/map is empty or user-defined `IsZero` function returns true, otherwise false | [go std library](https://golang.org/pkg/) and **only**  |0.0.2 |
| [structcopy](structcopy) | Copies struct's fields to another struct, including embedded and anonymous fields | [jinzhu/copier](https://github.com/jinzhu/copier) and **only** |0.0.2 |
| [config](config) | Config and protected settings made easy; load from any file (yaml by default), missing field or password? It fills the missing fields from `os.Stdin` if necessary (beauty!) | [kataras/pkg/zerocheck](zerocheck), [AlecAivazis/survey](https://github.com/AlecAivazis/survey), [gopkg.in/yaml.v2](https://gopkg.in/yaml.v2) |0.0.3 |
| [sched](sched) | Sched is a simple task/job scheduler that should be executed once on the future, i.e send an e-mail to a client after a year. | [go std library](https://golang.org/pkg/) and **only** | 0.0. |