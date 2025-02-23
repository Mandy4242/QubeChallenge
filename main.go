package main

import (
	"encoding/csv"
	"fmt"
	"os"
)

type Region struct {
	Name     string
	Code     string
	Parent   *Region
	Children map[string]*Region
}

type Distributor struct {
	Name     string
	Includes map[string]struct{}
	Excludes map[string]struct{}
	Parent   *Distributor
}

var regions = make(map[string]*Region)

var distributors = make(map[string]*Distributor)

func LoadRegionsFromCSV(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	_, err = reader.Read() // Skip header
	if err != nil {
		return err
	}

	records, err := reader.ReadAll()
	if err != nil {
		return err
	}

	for _, record := range records {
		cityCode, stateCode, countryCode := record[0], record[1], record[2]
		cityName, stateName, countryName := record[3], record[4], record[5]

		fullStateCode := stateCode + "-" + countryCode
		fullCityCode := cityCode + "-" + fullStateCode

		// Ensure country exists
		if _, exists := regions[countryCode]; !exists {
			regions[countryCode] = &Region{Name: countryName, Code: countryCode, Children: make(map[string]*Region)}
		}

		// Ensure state exists
		if _, exists := regions[fullStateCode]; !exists {
			regions[fullStateCode] = &Region{Name: stateName, Code: fullStateCode, Parent: regions[countryCode], Children: make(map[string]*Region)}
			regions[countryCode].Children[fullStateCode] = regions[fullStateCode]
		}

		// Ensure city exists
		if _, exists := regions[fullCityCode]; !exists {
			regions[fullCityCode] = &Region{Name: cityName, Code: fullCityCode, Parent: regions[fullStateCode], Children: make(map[string]*Region)}
			regions[fullStateCode].Children[fullCityCode] = regions[fullCityCode]
		}
	}
	return nil
}

func AddDistributor(name string, parent *Distributor) *Distributor {
	d := &Distributor{Name: name, Includes: make(map[string]struct{}), Excludes: make(map[string]struct{}), Parent: parent}
	distributors[name] = d
	return d
}

func (d *Distributor) SetPermissions(include []string, exclude []string) {
	for _, region := range include {
		if _, exists := regions[region]; exists {
			d.Includes[region] = struct{}{}
		} else {
			fmt.Println("Warning: Included region does not exist:", region)
		}
	}
	for _, region := range exclude {
		if _, exists := regions[region]; exists {
			delete(d.Includes, region)
			d.Excludes[region] = struct{}{}
		} else {
			fmt.Println("Warning: Excluded region does not exist:", region)
		}
	}
}

func (d *Distributor) HasPermission(location string) bool {
	fmt.Println("Checking permission for:", location)

	region, exists := regions[location]
	if !exists {
		fmt.Println("Unknown region:", location)
		return false
	}

	// Print the region hierarchy for debugging
	for r := region; r != nil; r = r.Parent {
		fmt.Println("-> Checking region:", r.Code)
		if _, excluded := d.Excludes[r.Code]; excluded {
			fmt.Println("Excluded region:", r.Code)
			return false
		}
		if _, included := d.Includes[r.Code]; included {
			fmt.Println("Included region:", r.Code)
			return true
		}
	}

	// Check parent's permissions
	if d.Parent != nil {
		return d.Parent.HasPermission(location)
	}
	return false
}

func main() {
	if err := LoadRegionsFromCSV("cities.csv"); err != nil {
		fmt.Println("Error loading regions:", err)
		return
	}

	parentDist := AddDistributor("DISTRIBUTOR1", nil)
	parentDist.SetPermissions([]string{"US", "IN"}, []string{"KA-IN", "TN-IN"})

}
