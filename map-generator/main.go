package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type MapItem struct {
	Name   string
	IsTest bool
}

var maps = []MapItem{
	{Name: "africa"},
	{Name: "asia"},
	{Name: "australia"},
	{Name: "achiran"},
	{Name: "baikal"},
	{Name: "baikalnukewars"},
	{Name: "betweentwoseas"},
	{Name: "blacksea"},
	{Name: "britannia"},
	{Name: "deglaciatedantarctica"},
	{Name: "eastasia"},
	{Name: "europe"},
	{Name: "europeclassic"},
	{Name: "falklandislands"},
	{Name: "faroeislands"},
	{Name: "fourislands"},
	{Name: "gatewaytotheatlantic"},
	{Name: "giantworldmap"},
	{Name: "halkidiki"},
	{Name: "iceland"},
	{Name: "italia"},
	{Name: "japan"},
	{Name: "mars"},
	{Name: "mena"},
	{Name: "montreal"},
	{Name: "northamerica"},
	{Name: "oceania"},
	{Name: "pangaea"},
	{Name: "pluto"},
	{Name: "southamerica"},
	{Name: "straitofgibraltar"},
	{Name: "world"},
	{Name: "big_plains", IsTest: true},
	{Name: "half_land_half_ocean", IsTest: true},
	{Name: "ocean_and_land", IsTest: true},
	{Name: "plains", IsTest: true},
	{Name: "giantworldmap", IsTest: true},
}

func outputMapDir(isTest bool) (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get working directory: %w", err)
	}
	if isTest {
		return filepath.Join(cwd, "..", "tests", "testdata", "maps"), nil
	}
	return filepath.Join(cwd, "..", "resources", "maps"), nil
}

func inputMapDir(isTest bool) (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get working directory: %w", err)
	}
	if isTest {
		return filepath.Join(cwd, "assets", "test_maps"), nil
	} else {
		return filepath.Join(cwd, "assets", "maps"), nil
	}
}

func processMap(name string, isTest bool) error {
	outputMapBaseDir, err := outputMapDir(isTest)
	if err != nil {
		return fmt.Errorf("failed to get map directory: %w", err)
	}

	inputMapDir, err := inputMapDir(isTest)
	if err != nil {
		return fmt.Errorf("failed to get input map directory: %w", err)
	}

	inputPath := filepath.Join(inputMapDir, name, "image.png")
	imageBuffer, err := os.ReadFile(inputPath)
	if err != nil {
		return fmt.Errorf("failed to read map file %s: %w", inputPath, err)
	}

	// Read the info.json file
	manifestPath := filepath.Join(inputMapDir, name, "info.json")
	manifestBuffer, err := os.ReadFile(manifestPath)
	if err != nil {
		return fmt.Errorf("failed to read info file %s: %w", manifestPath, err)
	}

	// Parse the info buffer as dynamic JSON
	var manifest map[string]interface{}
	if err := json.Unmarshal(manifestBuffer, &manifest); err != nil {
		return fmt.Errorf("failed to parse info.json for %s: %w", name, err)
	}

	// Generate maps
	result, err := GenerateMap(GeneratorArgs{
		ImageBuffer: imageBuffer,
		RemoveSmall: !isTest, // Don't remove small islands for test maps
		Name:        name,
	})
	if err != nil {
		return fmt.Errorf("failed to generate map for %s: %w", name, err)
	}

	manifest["map"] = map[string]interface{}{
		"width":          result.Map.Width,
		"height":         result.Map.Height,
		"num_land_tiles": result.Map.NumLandTiles,
	}
	manifest["map4x"] = map[string]interface{}{
		"width":          result.Map4x.Width,
		"height":         result.Map4x.Height,
		"num_land_tiles": result.Map4x.NumLandTiles,
	}
	manifest["map16x"] = map[string]interface{}{
		"width":          result.Map16x.Width,
		"height":         result.Map16x.Height,
		"num_land_tiles": result.Map16x.NumLandTiles,
	}

	mapDir := filepath.Join(outputMapBaseDir, name)
	if err := os.MkdirAll(mapDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory for %s: %w", name, err)
	}
	if err := os.WriteFile(filepath.Join(mapDir, "map.bin"), result.Map.Data, 0644); err != nil {
		return fmt.Errorf("failed to write combined binary for %s: %w", name, err)
	}
	if err := os.WriteFile(filepath.Join(mapDir, "map4x.bin"), result.Map4x.Data, 0644); err != nil {
		return fmt.Errorf("failed to write combined binary for %s: %w", name, err)
	}
	if err := os.WriteFile(filepath.Join(mapDir, "map16x.bin"), result.Map16x.Data, 0644); err != nil {
		return fmt.Errorf("failed to write combined binary for %s: %w", name, err)
	}
	if err := os.WriteFile(filepath.Join(mapDir, "thumbnail.webp"), result.Thumbnail, 0644); err != nil {
		return fmt.Errorf("failed to write thumbnail for %s: %w", name, err)
	}

	// Serialize the updated manifest to JSON
	updatedManifest, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to serialize manifest for %s: %w", name, err)
	}

	if err := os.WriteFile(filepath.Join(mapDir, "manifest.json"), updatedManifest, 0644); err != nil {
		return fmt.Errorf("failed to write manifest for %s: %w", name, err)
	}
	return nil
}

func loadTerrainMaps() error {
	var wg sync.WaitGroup
	errChan := make(chan error, len(maps))

	// Process maps concurrently
	for _, mapItem := range maps {
		wg.Add(1)
		// pass mapItem into closure to avoid loop variable capture
		go func(mi MapItem) {
			defer wg.Done()
			if err := processMap(mi.Name, mi.IsTest); err != nil {
				errChan <- err
			}
		}(mapItem)
	}

	// Wait for all goroutines to complete
	wg.Wait()
	close(errChan)

	// Check for errors
	for err := range errChan {
		if err != nil {
			return err
		}
	}

	return nil
}

func main() {
	// Add a CLI flag to optionally override the built-in `maps` list.
	mapsFlag := flag.String("maps", "", "Comma-separated list of map names to process. Prefix a name with 'test:' to mark it as a test map.")
	flag.Parse()

	if *mapsFlag != "" {
		parts := strings.Split(*mapsFlag, ",")
		newMaps := make([]MapItem, 0, len(parts))
		for _, p := range parts {
			p = strings.TrimSpace(p)
			if p == "" {
				continue
			}
			isTest := false
			if strings.HasPrefix(p, "test:") {
				isTest = true
				p = strings.TrimPrefix(p, "test:")
			}
			newMaps = append(newMaps, MapItem{Name: p, IsTest: isTest})
		}
		if len(newMaps) > 0 {
			maps = newMaps
		}
	}

	if err := loadTerrainMaps(); err != nil {
		log.Fatalf("Error generating terrain maps: %v", err)
	}

	fmt.Println("Terrain maps generated successfully")
}
