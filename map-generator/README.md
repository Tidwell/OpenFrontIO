# MapGenerator

This is a tool to generate map files for OpenFront.

## Installation

1. Install go https://go.dev/doc/install
2. Install dependencies: `go mod download`
3. Run the generator: `go run .`

## Creating a new map

1. Create a new folder in assets/maps/<map_name>
2. Create image.png
3. Create info.json with name and countries
4. Add the map name in `main.go`
5. Run the generator: `go run .`
6. Find the output folder at `../resources/maps/<map_name>` when running default (non-test) maps.
   Test maps are written to `../tests/testdata/maps/<map_name>`.

## Create image.png

1. Download world map (warning very large file) https://drive.google.com/file/d/1W2oMPj1L5zWRyPhh8LfmnY3_kve-FBR2/view?usp=sharing
2. Crop the file (recommend Gimp), we recommend roughly 2 million pixels for performance reasons. Do not go over 4 million pixels.

## Create info.json

- Look at existing info.json for structure
- Use country codes found here: https://en.wikipedia.org/wiki/List_of_ISO_3166_country_codes

## Notes

- Islands smaller than 30 tiles (pixels) are automatically removed by the script.
- Bodies of water smaller than 200 tiles (pixels) are also removed.

## Using the `-maps` flag

An optional CLI flag `-maps` can be used to override the builtin list of maps and process only a subset of maps. The value is a comma-separated list of map names. Prefix a name with `test:` to mark that map as a test map (it will be written to the test output directory and small-island removal is disabled for tests).

Examples:

- Process two default maps:

```
go run . -maps="asia,africa"
```

- Process one test map and one normal map:

```
go run . -maps="test:plains,world"
```

Notes:

- Map names must match the folder name under `map-generator/assets/maps` (or `map-generator/assets/test_maps` for test maps).
- If `-maps` is not provided the built-in list in `main.go` is used.

## Verbose logging

The generator supports a `-verbose` CLI flag that enables additional logging about removed small islands and lakes. When enabled the generator will print a short summary for each removed entity showing its size and a representative starting pixel coordinate.

Examples:

- Run the generator with verbose logging for the default built-in maps:

```
go run . -verbose
```

- Run a single map with verbose logging:

```
go run . -maps="asia" -verbose
```

Notes:

- Verbose output is purposely compact: it prints the size and a single starting pixel for each removed island or lake (to avoid extremely long coordinate lists).
- Use `-verbose` when you need to debug or inspect which small features are being removed during generation.
